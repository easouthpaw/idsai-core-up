package postgres

import (
	"context"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
)

func (r *ProjectFlowRepo) GetProjectByID(ctx context.Context, projectID uuid.UUID) (domain.Project, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return domain.Project{}, err
	}

	const q = `
SELECT id, title, description, status, is_public, created_by, professor_id,
       professor_review_status,
       faculty_id, visibility, group_id,
       created_at, updated_at
FROM projects
WHERE tenant_id = $1
  AND id = $2;
`
	p, err := scanProjectRow(r.db.QueryRow(ctx, q, tenantID, projectID))
	return p, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) IsActiveProjectMember(ctx context.Context, userID, projectID uuid.UUID) (bool, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return false, err
	}

	const q = `
SELECT EXISTS (
  SELECT 1
  FROM projects p
  WHERE p.tenant_id = $1
    AND p.id = $2
    AND (
      p.created_by = $3
      OR EXISTS (
        SELECT 1
        FROM project_members pm
        WHERE pm.project_id = p.id
          AND pm.tenant_id = p.tenant_id
          AND pm.user_id = $3
          AND pm.status = 'ACTIVE'
      )
    )
) AS ok;
`
	var ok bool
	if err := r.db.QueryRow(ctx, q, tenantID, projectID, userID).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (r *ProjectFlowRepo) HasProjectRole(ctx context.Context, userID, projectID uuid.UUID, roleCode string) (bool, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return false, err
	}

	const q = `
SELECT EXISTS (
	SELECT 1
	FROM role_assignments ra
	JOIN roles r ON r.id = ra.role_id
	WHERE ra.user_id = $1
	  AND ra.tenant_id = $2
	  AND ra.scope_type = 'PROJECT'
	  AND ra.scope_id = $3
	  AND r.code = $4
) AS ok;
`
	var ok bool
	if err := r.db.QueryRow(ctx, q, userID, tenantID, projectID, roleCode).Scan(&ok); err != nil {
		return false, err
	}
	return ok, nil
}

func (r *ProjectFlowRepo) RevokeProjectRole(ctx context.Context, userID, projectID uuid.UUID, roleCode string) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	const q = `
DELETE FROM role_assignments ra
USING roles r
WHERE ra.role_id = r.id
  AND ra.user_id = $1
  AND ra.tenant_id = $2
  AND ra.scope_type = 'PROJECT'
  AND ra.scope_id = $3
  AND r.code = $4;
`
	_, err = r.db.Exec(ctx, q, userID, tenantID, projectID, roleCode)
	return err
}

func (r *ProjectFlowRepo) UpdateProject(
	ctx context.Context,
	projectID uuid.UUID,
	titleSet bool,
	title string,
	descriptionSet bool,
	description string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	const q = `
UPDATE projects
SET title = CASE WHEN $2::boolean THEN $3 ELSE title END,
    description = CASE WHEN $4::boolean THEN $5 ELSE description END,
    updated_at = now()
WHERE tenant_id = $1
  AND id = $6;
`
	ct, err := r.db.Exec(ctx, q, tenantID, titleSet, title, descriptionSet, description, projectID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return projectflow.ErrNotFound
	}
	return nil
}

func (r *ProjectFlowRepo) OpenProjectRecruitment(ctx context.Context, projectID uuid.UUID) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	const q = `
UPDATE projects
SET status = 'RECRUITMENT', updated_at = now()
WHERE tenant_id = $1
  AND id = $2
  AND status IN ('DRAFT','REVIEW','RECRUITMENT');
`
	ct, err := r.db.Exec(ctx, q, tenantID, projectID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return projectflow.ErrNotFound
	}
	return nil
}

func (r *ProjectFlowRepo) ListStudentCandidates(
	ctx context.Context,
	facultyID, projectID, requesterUserID, projectOwnerID uuid.UUID,
	term string,
	limit int,
) ([]projectflow.StudentCandidate, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT u.id,
       COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(u.email, '@', 1)) AS full_name,
       u.email,
       COALESCE(d.code, '') AS department_code
FROM users u
JOIN user_profiles up ON up.user_id = u.id
LEFT JOIN departments d ON d.id = up.department_id
WHERE u.tenant_id = $1
  AND up.tenant_id = $1
  AND u.status = 'ACTIVE'
  AND up.faculty_id = $2
  AND u.id <> $4
  AND u.id <> $5
  AND EXISTS (
    SELECT 1
    FROM role_assignments ra
    JOIN roles r ON r.id = ra.role_id
    WHERE ra.user_id = u.id
      AND ra.tenant_id = u.tenant_id
      AND r.code = 'STUDENT'
  )
  AND NOT EXISTS (
    SELECT 1
    FROM project_members pm
    WHERE pm.tenant_id = $1
      AND pm.project_id = $3
      AND pm.user_id = u.id
      AND pm.status IN ('ACTIVE', 'INVITED', 'APPLIED')
  )
  AND ($6 = '' OR lower(up.full_name) LIKE '%' || $6 || '%' OR lower(u.email) LIKE '%' || $6 || '%')
ORDER BY up.full_name ASC, u.email ASC
LIMIT $7;
`
	rows, err := r.db.Query(ctx, q, tenantID, facultyID, projectID, requesterUserID, projectOwnerID, term, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.StudentCandidate, 0, limit)
	for rows.Next() {
		var item projectflow.StudentCandidate
		if err := rows.Scan(&item.UserID, &item.FullName, &item.Email, &item.DepartmentCode); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) ReplaceProjectStacks(ctx context.Context, projectID uuid.UUID, stackCodes []string) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, `DELETE FROM project_stacks WHERE tenant_id = $1 AND project_id = $2`, tenantID, projectID); err != nil {
		if isUndefinedRelationErr(err, "project_stacks") {
			return projectflow.ErrSchemaMissing
		}
		return err
	}
	for _, code := range stackCodes {
		if _, err := tx.Exec(ctx, `
INSERT INTO project_stacks(tenant_id, project_id, stack_code)
SELECT $1, $2, $3
FROM projects p
WHERE p.tenant_id = $1
  AND p.id = $2
`, tenantID, projectID, code); err != nil {
			if isUndefinedRelationErr(err, "project_stacks") {
				return projectflow.ErrSchemaMissing
			}
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *ProjectFlowRepo) ListProjectStackCodes(ctx context.Context, projectID uuid.UUID) ([]string, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx, `SELECT stack_code FROM project_stacks WHERE tenant_id = $1 AND project_id = $2 ORDER BY stack_code ASC`, tenantID, projectID)
	if err != nil {
		if isUndefinedRelationErr(err, "project_stacks") {
			return []string{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, 8)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) CreateProjectPosition(ctx context.Context, projectID uuid.UUID, code, name string, capacity int) (projectflow.Position, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return projectflow.Position{}, err
	}

	const q = `
INSERT INTO project_positions(tenant_id, project_id, code, name, capacity)
SELECT $1, $2, $3, $4, $5
FROM projects p
WHERE p.tenant_id = $1
  AND p.id = $2
RETURNING id, project_id, code, name, capacity, created_at;
`
	var (
		id         uuid.UUID
		positionID uuid.UUID
		p          projectflow.Position
	)
	err = r.db.QueryRow(ctx, q, tenantID, projectID, code, name, capacity).Scan(
		&id,
		&positionID,
		&p.Code,
		&p.Name,
		&p.Capacity,
		&p.CreatedAt,
	)
	if err != nil {
		return projectflow.Position{}, err
	}
	p.ID = id.String()
	p.ProjectID = positionID.String()
	return p, nil
}

func (r *ProjectFlowRepo) ListProjectPositions(ctx context.Context, projectID uuid.UUID) ([]projectflow.Position, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT id, project_id, code, name, capacity, created_at
FROM project_positions
WHERE tenant_id = $1
  AND project_id = $2
ORDER BY created_at ASC;
`
	rows, err := r.db.Query(ctx, q, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.Position, 0, 8)
	for rows.Next() {
		var (
			id         uuid.UUID
			positionID uuid.UUID
			p          projectflow.Position
		)
		if err := rows.Scan(&id, &positionID, &p.Code, &p.Name, &p.Capacity, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.ID = id.String()
		p.ProjectID = positionID.String()
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) GetProjectPositionCapacity(ctx context.Context, projectID, positionID uuid.UUID) (int, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	const q = `
SELECT capacity
FROM project_positions
WHERE tenant_id = $1
  AND project_id = $2
  AND id = $3;
`
	var capacity int
	if err := r.db.QueryRow(ctx, q, tenantID, projectID, positionID).Scan(&capacity); err != nil {
		return 0, mapProjectFlowErr(err)
	}
	return capacity, nil
}

func (r *ProjectFlowRepo) SumProjectPositionCapacities(ctx context.Context, projectID uuid.UUID) (int, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return 0, err
	}

	const q = `SELECT COALESCE(SUM(capacity), 0) FROM project_positions WHERE tenant_id = $1 AND project_id = $2;`
	var total int
	if err := r.db.QueryRow(ctx, q, tenantID, projectID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}
