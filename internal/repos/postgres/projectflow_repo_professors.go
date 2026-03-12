package postgres

import (
	"context"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
)

func (r *ProjectFlowRepo) ListProfessorCandidates(
	ctx context.Context,
	facultyID uuid.UUID,
	term string,
	limit int,
	requesterUserID, projectOwnerID uuid.UUID,
) ([]projectflow.ProfessorCandidate, error) {
	const q = `
SELECT u.id,
       COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(u.email, '@', 1)) AS full_name,
       u.email,
       COALESCE(d.code, '') AS department_code
FROM users u
JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN departments d ON d.id = up.department_id
WHERE u.status = 'ACTIVE'
  AND up.faculty_id = $1
  AND u.id <> $4
  AND u.id <> $5
  AND EXISTS (
    SELECT 1
    FROM role_assignments ra
    JOIN roles r ON r.id = ra.role_id
    WHERE ra.user_id = u.id
      AND ra.tenant_id = u.tenant_id
      AND r.code = 'PROFESSOR'
  )
  AND ($2 = '' OR lower(up.full_name) LIKE '%' || $2 || '%' OR lower(u.email) LIKE '%' || $2 || '%')
ORDER BY up.full_name ASC, u.email ASC
LIMIT $3;
`
	rows, err := r.db.Query(ctx, q, facultyID, term, limit, requesterUserID, projectOwnerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.ProfessorCandidate, 0, limit)
	for rows.Next() {
		var item projectflow.ProfessorCandidate
		if err := rows.Scan(&item.UserID, &item.FullName, &item.Email, &item.DepartmentCode); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) IsActiveProfessorInFaculty(ctx context.Context, professorID, facultyID uuid.UUID) (bool, error) {
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
        AND r.code = 'PROFESSOR'
    )
) AS ok;
`
	var ok bool
	if err := r.db.QueryRow(ctx, q, professorID, facultyID).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (r *ProjectFlowRepo) AssignProjectProfessor(ctx context.Context, projectID, professorID uuid.UUID) error {
	const q = `
UPDATE projects
SET professor_id = $2,
    professor_review_status = 'PENDING',
    professor_invited_at = now(),
    professor_responded_at = NULL,
    updated_at = now()
WHERE id = $1;
`
	ct, err := r.db.Exec(ctx, q, projectID, professorID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return projectflow.ErrNotFound
	}
	return nil
}

func (r *ProjectFlowRepo) GetProfessorCandidateByID(
	ctx context.Context,
	professorID, facultyID uuid.UUID,
) (projectflow.ProfessorCandidate, error) {
	const q = `
SELECT u.id,
       COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(u.email, '@', 1)) AS full_name,
       u.email,
       COALESCE(d.code, '') AS department_code
FROM users u
JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN departments d ON d.id = up.department_id
WHERE u.id = $1
  AND up.faculty_id = $2
LIMIT 1;
`
	var item projectflow.ProfessorCandidate
	if err := r.db.QueryRow(ctx, q, professorID, facultyID).Scan(
		&item.UserID,
		&item.FullName,
		&item.Email,
		&item.DepartmentCode,
	); err != nil {
		return projectflow.ProfessorCandidate{}, mapProjectFlowErr(err)
	}
	return item, nil
}

func (r *ProjectFlowRepo) RespondProfessorInvite(
	ctx context.Context,
	projectID, professorID uuid.UUID,
	accept bool,
) (domain.Project, error) {
	nextStatus := "REJECTED"
	if accept {
		nextStatus = "ACCEPTED"
	}
	const q = `
UPDATE projects
SET professor_review_status = $3,
    professor_responded_at = now(),
    updated_at = now()
WHERE id = $1
  AND professor_id = $2
  AND professor_review_status = 'PENDING'
RETURNING id, title, description, status, is_public, created_by, professor_id,
          professor_review_status, faculty_id, visibility, group_id, created_at, updated_at;
`
	p, err := scanProjectRow(r.db.QueryRow(ctx, q, projectID, professorID, nextStatus))
	return p, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) ListProfessorReviewInvites(
	ctx context.Context,
	professorID uuid.UUID,
	term string,
	limit int,
) ([]domain.Project, error) {
	const q = `
SELECT id, title, description, status, is_public, created_by, professor_id,
       professor_review_status, faculty_id, visibility, group_id, created_at, updated_at
FROM projects
WHERE professor_id = $1
  AND professor_review_status = 'PENDING'
  AND ($2 = '' OR lower(title) LIKE '%' || $2 || '%' OR lower(description) LIKE '%' || $2 || '%')
ORDER BY updated_at DESC
LIMIT $3;
`
	rows, err := r.db.Query(ctx, q, professorID, term, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.Project, 0, limit)
	for rows.Next() {
		p, err := scanProjectRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
