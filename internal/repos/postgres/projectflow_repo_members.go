package postgres

import (
	"context"

	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
)

func (r *ProjectFlowRepo) IsActiveStudentInFaculty(ctx context.Context, studentID, facultyID uuid.UUID) (bool, error) {
	const q = `
SELECT EXISTS (
  SELECT 1
  FROM users u
  JOIN user_profiles up ON up.user_id = u.id
  WHERE u.id = $1
    AND u.status = 'ACTIVE'
    AND up.tenant_id = u.tenant_id
    AND up.faculty_id = $2
    AND EXISTS (
      SELECT 1
      FROM role_assignments ra
      JOIN roles r ON r.id = ra.role_id
      WHERE ra.user_id = u.id
        AND ra.tenant_id = u.tenant_id
        AND r.code = 'STUDENT'
    )
) AS ok;
`
	var ok bool
	if err := r.db.QueryRow(ctx, q, studentID, facultyID).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (r *ProjectFlowRepo) UpsertInvitedMember(
	ctx context.Context,
	projectID, studentID, invitedBy uuid.UUID,
	comment string,
) (projectflow.Member, error) {
	const q = `
INSERT INTO project_members(tenant_id, project_id, user_id, status, position_id, joined_at, invite_comment, invited_by, responded_at)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, 'INVITED', NULL, NULL, $3, $4, NULL)
ON CONFLICT (project_id, user_id)
DO UPDATE SET status='INVITED', position_id=NULL, joined_at=NULL, invite_comment=EXCLUDED.invite_comment, invited_by=EXCLUDED.invited_by, responded_at=NULL
WHERE project_members.status IN ('INVITED', 'APPLIED', 'REJECTED', 'REMOVED')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, projectID, studentID, comment, invitedBy))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) UpsertAppliedMember(ctx context.Context, projectID, userID uuid.UUID, comment string) (projectflow.Member, error) {
	const q = `
INSERT INTO project_members(tenant_id, project_id, user_id, status, position_id, joined_at, invite_comment, invited_by, responded_at)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, 'APPLIED', NULL, NULL, $3, NULL, now())
ON CONFLICT (project_id, user_id)
DO UPDATE SET status='APPLIED', position_id=NULL, joined_at=NULL, invite_comment=EXCLUDED.invite_comment, invited_by=NULL, responded_at=now()
WHERE project_members.status IN ('APPLIED', 'INVITED', 'REJECTED', 'REMOVED')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, projectID, userID, comment))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]projectflow.Member, error) {
	const q = `
SELECT m.id, m.project_id, m.user_id, m.position_id, m.status, m.invite_comment, m.invited_by, m.responded_at, m.joined_at, m.created_at,
       p.code, p.name,
       COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), '') AS full_name,
       COALESCE(u.email, '') AS email
FROM project_members m
LEFT JOIN project_positions p ON p.id = m.position_id
LEFT JOIN users u ON u.id = m.user_id
LEFT JOIN user_profiles up ON up.user_id = m.user_id
WHERE m.project_id = $1
ORDER BY m.created_at ASC;
`
	rows, err := r.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.Member, 0, 8)
	for rows.Next() {
		var (
			m         projectflow.Member
			memberID  uuid.UUID
			projID    uuid.UUID
			userID    uuid.UUID
			position  *uuid.UUID
			invitedBy *uuid.UUID
			posCode   *string
			posName   *string
		)

		if err := rows.Scan(
			&memberID,
			&projID,
			&userID,
			&position,
			&m.Status,
			&m.InviteComment,
			&invitedBy,
			&m.RespondedAt,
			&m.JoinedAt,
			&m.CreatedAt,
			&posCode,
			&posName,
			&m.FullName,
			&m.Email,
		); err != nil {
			return nil, err
		}

		m.ID = memberID.String()
		m.ProjectID = projID.String()
		m.UserID = userID.String()
		m.PositionID = uuidPtrToStringPtr(position)
		m.InvitedBy = uuidPtrToStringPtr(invitedBy)
		m.PositionCode = posCode
		m.PositionName = posName
		out = append(out, m)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) CountActiveMembersByPosition(
	ctx context.Context,
	projectID, positionID uuid.UUID,
	excludeUserID *uuid.UUID,
) (int, error) {
	const qWithExclude = `
SELECT COUNT(*)
FROM project_members
WHERE project_id = $1
  AND position_id = $2
  AND status = 'ACTIVE'
  AND user_id <> $3;
`
	const q = `
SELECT COUNT(*)
FROM project_members
WHERE project_id = $1
  AND position_id = $2
  AND status = 'ACTIVE';
`
	var total int
	if excludeUserID != nil {
		if err := r.db.QueryRow(ctx, qWithExclude, projectID, positionID, *excludeUserID).Scan(&total); err != nil {
			return 0, err
		}
		return total, nil
	}
	if err := r.db.QueryRow(ctx, q, projectID, positionID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *ProjectFlowRepo) GetProjectMemberStatusAndPosition(ctx context.Context, projectID, userID uuid.UUID) (string, *uuid.UUID, error) {
	const q = `
SELECT status, position_id
FROM project_members
WHERE project_id = $1 AND user_id = $2;
`
	var status string
	var positionID *uuid.UUID
	if err := r.db.QueryRow(ctx, q, projectID, userID).Scan(&status, &positionID); err != nil {
		return "", nil, mapProjectFlowErr(err)
	}
	return status, positionID, nil
}

func (r *ProjectFlowRepo) CountActiveMembersWithPosition(ctx context.Context, projectID uuid.UUID) (int, error) {
	const q = `
SELECT COUNT(*)
FROM project_members
WHERE project_id=$1 AND status='ACTIVE' AND position_id IS NOT NULL;
`
	var total int
	if err := r.db.QueryRow(ctx, q, projectID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *ProjectFlowRepo) ApproveProjectMember(
	ctx context.Context,
	projectID, memberUserID uuid.UUID,
	positionID *uuid.UUID,
) (projectflow.Member, error) {
	const q = `
UPDATE project_members
SET status='ACTIVE', position_id=COALESCE($3, position_id), joined_at=now(), responded_at=now()
WHERE project_id=$1 AND user_id=$2 AND status IN ('APPLIED', 'ACTIVE')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, projectID, memberUserID, positionID))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) RejectProjectMemberApplication(
	ctx context.Context,
	projectID, memberUserID uuid.UUID,
) (projectflow.Member, error) {
	const q = `
UPDATE project_members
SET status='REJECTED', position_id=NULL, joined_at=NULL, responded_at=now()
WHERE project_id=$1 AND user_id=$2 AND status='APPLIED'
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, projectID, memberUserID))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) SetActiveMemberPosition(
	ctx context.Context,
	projectID, memberUserID, positionID uuid.UUID,
) (projectflow.Member, error) {
	const q = `
UPDATE project_members
SET position_id=$3
WHERE project_id=$1 AND user_id=$2 AND status='ACTIVE'
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, projectID, memberUserID, positionID))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) GetInvitedMemberPosition(ctx context.Context, projectID, userID uuid.UUID) (*uuid.UUID, error) {
	const q = `
SELECT position_id
FROM project_members
WHERE project_id = $1
  AND user_id = $2
  AND status = 'INVITED';
`
	var positionID *uuid.UUID
	if err := r.db.QueryRow(ctx, q, projectID, userID).Scan(&positionID); err != nil {
		return nil, mapProjectFlowErr(err)
	}
	return positionID, nil
}

func (r *ProjectFlowRepo) RespondMemberInvite(
	ctx context.Context,
	projectID, userID uuid.UUID,
	accept bool,
) (projectflow.Member, error) {
	const q = `
UPDATE project_members
SET status = CASE WHEN $3::boolean THEN 'ACTIVE' ELSE 'REJECTED' END,
    joined_at = CASE WHEN $3::boolean THEN now() ELSE NULL END,
    responded_at = now()
WHERE project_id = $1
  AND user_id = $2
  AND status = 'INVITED'
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, projectID, userID, accept))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) ListIncomingInvites(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]projectflow.IncomingInvite, error) {
	const q = `
SELECT
  p.id,
  p.title,
  p.status,
  pm.status,
  pm.user_id,
  pm.invite_comment,
  pm.invited_by,
  COALESCE(
    NULLIF(TRIM(inv.full_name), ''),
    split_part(COALESCE(inv_u.email, ''), '@', 1),
    NULLIF(TRIM(app.full_name), ''),
    split_part(COALESCE(app_u.email, ''), '@', 1),
    ''
  ) AS inviter_name,
  COALESCE(inv_u.email, app_u.email, '') AS inviter_email,
  pm.created_at,
  pm.responded_at
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
LEFT JOIN users inv_u ON inv_u.id = pm.invited_by
LEFT JOIN user_profiles inv ON inv.user_id = pm.invited_by
LEFT JOIN users app_u ON app_u.id = pm.user_id
LEFT JOIN user_profiles app ON app.user_id = pm.user_id
WHERE (
    pm.user_id = $1
    AND pm.status = 'INVITED'
  )
  OR (
    p.created_by = $1
    AND pm.status = 'APPLIED'
    AND pm.user_id <> $1
  )
ORDER BY pm.created_at DESC
LIMIT $2;
`
	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.IncomingInvite, 0, limit)
	for rows.Next() {
		var (
			item         projectflow.IncomingInvite
			projectID    uuid.UUID
			memberUserID uuid.UUID
			invitedBy    *uuid.UUID
		)
		if err := rows.Scan(
			&projectID,
			&item.ProjectTitle,
			&item.ProjectStatus,
			&item.Status,
			&memberUserID,
			&item.InviteComment,
			&invitedBy,
			&item.InviterName,
			&item.InviterEmail,
			&item.CreatedAt,
			&item.RespondedAt,
		); err != nil {
			return nil, err
		}
		item.ProjectID = projectID.String()
		item.UserID = memberUserID.String()
		item.InvitedBy = uuidPtrToStringPtr(invitedBy)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) ListOutgoingApplications(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]projectflow.OutgoingApplication, error) {
	const q = `
SELECT
  p.id,
  p.title,
  p.status,
  pm.status,
  pm.created_at,
  pm.responded_at
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
WHERE pm.user_id = $1
  AND pm.invited_by IS NULL
  AND pm.status IN ('APPLIED', 'REJECTED', 'ACTIVE', 'REMOVED')
ORDER BY COALESCE(pm.responded_at, pm.created_at) DESC
LIMIT $2;
`
	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.OutgoingApplication, 0, limit)
	for rows.Next() {
		var (
			item      projectflow.OutgoingApplication
			projectID uuid.UUID
		)
		if err := rows.Scan(
			&projectID,
			&item.ProjectTitle,
			&item.ProjectStatus,
			&item.Status,
			&item.CreatedAt,
			&item.RespondedAt,
		); err != nil {
			return nil, err
		}
		item.ProjectID = projectID.String()
		out = append(out, item)
	}
	return out, rows.Err()
}
