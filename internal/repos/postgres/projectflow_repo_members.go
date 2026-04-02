package postgres

import (
	"context"

	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
)

func (r *ProjectFlowRepo) IsActiveStudentInFaculty(ctx context.Context, studentID, facultyID uuid.UUID) (bool, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return false, err
	}

	const q = `
SELECT EXISTS (
  SELECT 1
  FROM users u
  JOIN user_profiles up ON up.user_id = u.id
  WHERE u.tenant_id = $1
    AND u.id = $2
    AND u.status = 'ACTIVE'
    AND up.tenant_id = u.tenant_id
    AND up.faculty_id = $3
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
	if err := r.db.QueryRow(ctx, q, tenantID, studentID, facultyID).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (r *ProjectFlowRepo) UpsertInvitedMember(
	ctx context.Context,
	projectID, studentID, invitedBy uuid.UUID,
	comment string,
) (projectflow.Member, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return projectflow.Member{}, err
	}

	const q = `
INSERT INTO project_members(tenant_id, project_id, user_id, status, position_id, joined_at, invite_comment, invited_by, responded_at)
SELECT $1, $2, $3, 'INVITED', NULL, NULL, $4, $5, NULL
FROM projects p
WHERE p.tenant_id = $1
  AND p.id = $2
ON CONFLICT (project_id, user_id)
DO UPDATE SET status='INVITED', position_id=NULL, joined_at=NULL, invite_comment=EXCLUDED.invite_comment, invited_by=EXCLUDED.invited_by, responded_at=NULL
WHERE project_members.status IN ('INVITED', 'APPLIED', 'REJECTED', 'REMOVED')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, tenantID, projectID, studentID, comment, invitedBy))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) UpsertAppliedMember(ctx context.Context, projectID, userID uuid.UUID, comment string) (projectflow.Member, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return projectflow.Member{}, err
	}

	const q = `
INSERT INTO project_members(tenant_id, project_id, user_id, status, position_id, joined_at, invite_comment, invited_by, responded_at)
SELECT $1, $2, $3, 'APPLIED', NULL, NULL, $4, NULL, now()
FROM projects p
WHERE p.tenant_id = $1
  AND p.id = $2
ON CONFLICT (project_id, user_id)
DO UPDATE SET status='APPLIED', position_id=NULL, joined_at=NULL, invite_comment=EXCLUDED.invite_comment, invited_by=NULL, responded_at=now()
WHERE project_members.status IN ('APPLIED', 'INVITED', 'REJECTED', 'REMOVED')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, tenantID, projectID, userID, comment))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) ListProjectMembers(ctx context.Context, projectID uuid.UUID) ([]projectflow.Member, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT m.id, m.project_id, m.user_id, m.position_id, m.status, m.invite_comment, m.invited_by, m.responded_at, m.joined_at, m.created_at,
       p.code, p.name,
       COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), '') AS full_name,
       COALESCE(u.email, '') AS email
FROM project_members m
LEFT JOIN project_positions p ON p.id = m.position_id
LEFT JOIN users u ON u.id = m.user_id
LEFT JOIN user_profiles up ON up.user_id = m.user_id
WHERE m.tenant_id = $1
  AND (p.id IS NULL OR p.tenant_id = $1)
  AND (u.id IS NULL OR u.tenant_id = $1)
  AND (up.user_id IS NULL OR up.tenant_id = $1)
  AND m.project_id = $2
ORDER BY m.created_at ASC;
`
	rows, err := r.db.Query(ctx, q, tenantID, projectID)
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
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	const qWithExclude = `
SELECT COUNT(*)
FROM project_members
WHERE tenant_id = $1
  AND project_id = $2
  AND position_id = $3
  AND status = 'ACTIVE'
  AND user_id <> $4;
`
	const q = `
SELECT COUNT(*)
FROM project_members
WHERE tenant_id = $1
  AND project_id = $2
  AND position_id = $3
  AND status = 'ACTIVE';
`
	var total int
	if excludeUserID != nil {
		if err := r.db.QueryRow(ctx, qWithExclude, tenantID, projectID, positionID, *excludeUserID).Scan(&total); err != nil {
			return 0, err
		}
		return total, nil
	}
	if err := r.db.QueryRow(ctx, q, tenantID, projectID, positionID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *ProjectFlowRepo) GetProjectMemberStatusAndPosition(ctx context.Context, projectID, userID uuid.UUID) (string, *uuid.UUID, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return "", nil, err
	}

	const q = `
SELECT status, position_id
FROM project_members
WHERE tenant_id = $1
  AND project_id = $2
  AND user_id = $3;
`
	var status string
	var positionID *uuid.UUID
	if err := r.db.QueryRow(ctx, q, tenantID, projectID, userID).Scan(&status, &positionID); err != nil {
		return "", nil, mapProjectFlowErr(err)
	}
	return status, positionID, nil
}

func (r *ProjectFlowRepo) CountActiveMembersWithPosition(ctx context.Context, projectID uuid.UUID) (int, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	const q = `
SELECT COUNT(*)
FROM project_members
WHERE tenant_id = $1
  AND project_id = $2
  AND status = 'ACTIVE'
  AND position_id IS NOT NULL;
`
	var total int
	if err := r.db.QueryRow(ctx, q, tenantID, projectID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *ProjectFlowRepo) ApproveProjectMember(
	ctx context.Context,
	projectID, memberUserID uuid.UUID,
	positionID *uuid.UUID,
) (projectflow.Member, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return projectflow.Member{}, err
	}

	const q = `
UPDATE project_members
SET status='ACTIVE', position_id=COALESCE($4, position_id), joined_at=now(), responded_at=now()
WHERE tenant_id = $1
  AND project_id = $2
  AND user_id = $3
  AND status IN ('APPLIED', 'ACTIVE')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, tenantID, projectID, memberUserID, positionID))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) RejectProjectMemberApplication(
	ctx context.Context,
	projectID, memberUserID uuid.UUID,
) (projectflow.Member, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return projectflow.Member{}, err
	}

	const q = `
UPDATE project_members
SET status='REJECTED', position_id=NULL, joined_at=NULL, responded_at=now()
WHERE tenant_id = $1
  AND project_id = $2
  AND user_id = $3
  AND status='APPLIED'
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, tenantID, projectID, memberUserID))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) RemoveProjectMember(
	ctx context.Context,
	projectID, memberUserID uuid.UUID,
) (projectflow.Member, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return projectflow.Member{}, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return projectflow.Member{}, err
	}
	defer tx.Rollback(ctx)

	const q = `
UPDATE project_members
SET status='REMOVED', position_id=NULL, joined_at=NULL, responded_at=now()
WHERE tenant_id = $1
  AND project_id = $2
  AND user_id = $3
  AND status IN ('ACTIVE', 'INVITED')
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(tx.QueryRow(ctx, q, tenantID, projectID, memberUserID))
	if err != nil {
		return projectflow.Member{}, err
	}

	if _, err := tx.Exec(ctx, `
UPDATE tasks
SET assignee_user_id = NULL,
    updated_at = now()
WHERE project_id = $1
  AND assignee_user_id = $2;
`, projectID, memberUserID); err != nil {
		return projectflow.Member{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return projectflow.Member{}, err
	}
	return m, nil
}

func (r *ProjectFlowRepo) SetActiveMemberPosition(
	ctx context.Context,
	projectID, memberUserID, positionID uuid.UUID,
) (projectflow.Member, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return projectflow.Member{}, err
	}

	const q = `
UPDATE project_members
SET position_id = $4
WHERE tenant_id = $1
  AND project_id = $2
  AND user_id = $3
  AND status='ACTIVE'
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, tenantID, projectID, memberUserID, positionID))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) GetInvitedMemberPosition(ctx context.Context, projectID, userID uuid.UUID) (*uuid.UUID, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT position_id
FROM project_members
WHERE tenant_id = $1
  AND project_id = $2
  AND user_id = $3
  AND status = 'INVITED';
`
	var positionID *uuid.UUID
	if err := r.db.QueryRow(ctx, q, tenantID, projectID, userID).Scan(&positionID); err != nil {
		return nil, mapProjectFlowErr(err)
	}
	return positionID, nil
}

func (r *ProjectFlowRepo) RespondMemberInvite(
	ctx context.Context,
	projectID, userID uuid.UUID,
	accept bool,
) (projectflow.Member, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return projectflow.Member{}, err
	}

	const q = `
UPDATE project_members
SET status = CASE WHEN $4::boolean THEN 'ACTIVE' ELSE 'REJECTED' END,
    joined_at = CASE WHEN $4::boolean THEN now() ELSE NULL END,
    responded_at = now()
WHERE tenant_id = $1
  AND project_id = $2
  AND user_id = $3
  AND status = 'INVITED'
RETURNING id, project_id, user_id, position_id, status, invite_comment, invited_by, responded_at, joined_at, created_at;
`
	m, err := scanMemberRow(r.db.QueryRow(ctx, q, tenantID, projectID, userID, accept))
	return m, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) ListIncomingInvites(
	ctx context.Context,
	userID uuid.UUID,
	limit int,
) ([]projectflow.IncomingInvite, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

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
JOIN projects p ON p.id = pm.project_id AND p.tenant_id = pm.tenant_id
LEFT JOIN users inv_u ON inv_u.id = pm.invited_by
LEFT JOIN user_profiles inv ON inv.user_id = pm.invited_by
LEFT JOIN users app_u ON app_u.id = pm.user_id
LEFT JOIN user_profiles app ON app.user_id = pm.user_id
WHERE p.tenant_id = $1
  AND pm.tenant_id = $1
  AND (
    (
      pm.user_id = $2
      AND pm.status = 'INVITED'
    )
    OR (
      p.created_by = $2
      AND pm.status = 'APPLIED'
      AND pm.user_id <> $2
    )
  )
ORDER BY pm.created_at DESC
LIMIT $3;
`
	rows, err := r.db.Query(ctx, q, tenantID, userID, limit)
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
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT
  p.id,
  p.title,
  p.status,
  pm.status,
  pm.created_at,
  pm.responded_at
FROM project_members pm
JOIN projects p ON p.id = pm.project_id AND p.tenant_id = pm.tenant_id
WHERE p.tenant_id = $1
  AND pm.tenant_id = $1
  AND pm.user_id = $2
  AND pm.invited_by IS NULL
  AND pm.status IN ('APPLIED', 'REJECTED', 'ACTIVE', 'REMOVED')
ORDER BY COALESCE(pm.responded_at, pm.created_at) DESC
LIMIT $3;
`
	rows, err := r.db.Query(ctx, q, tenantID, userID, limit)
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
