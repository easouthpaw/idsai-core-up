package postgres

import (
	"context"

	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
)

func (r *ProjectFlowRepo) GetProjectCriteriaWeightSum(ctx context.Context, projectID uuid.UUID) (int, error) {
	const q = `
SELECT COALESCE(SUM(weight), 0)
FROM project_criteria
WHERE project_id = $1;
`
	var total int
	if err := r.db.QueryRow(ctx, q, projectID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *ProjectFlowRepo) CreateProjectCriterion(
	ctx context.Context,
	projectID, userID uuid.UUID,
	title, description string,
	weight int,
) (projectflow.Criterion, error) {
	const q = `
INSERT INTO project_criteria(tenant_id, project_id, title, description, weight, created_by)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, $3, $4, $5)
RETURNING id, project_id, title, description, weight, created_by, created_at;
`
	var (
		c         projectflow.Criterion
		id        uuid.UUID
		pid       uuid.UUID
		createdBy uuid.UUID
	)
	err := r.db.QueryRow(ctx, q, projectID, title, description, weight, userID).Scan(
		&id,
		&pid,
		&c.Title,
		&c.Description,
		&c.Weight,
		&createdBy,
		&c.CreatedAt,
	)
	if err != nil {
		return projectflow.Criterion{}, err
	}
	c.ID = id.String()
	c.ProjectID = pid.String()
	c.CreatedBy = createdBy.String()
	return c, nil
}

func (r *ProjectFlowRepo) ListProjectCriteria(ctx context.Context, projectID uuid.UUID) ([]projectflow.Criterion, error) {
	const q = `
SELECT id, project_id, title, description, weight, created_by, created_at
FROM project_criteria
WHERE project_id=$1
ORDER BY created_at ASC;
`
	rows, err := r.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.Criterion, 0, 8)
	for rows.Next() {
		var (
			c         projectflow.Criterion
			id        uuid.UUID
			pid       uuid.UUID
			createdBy uuid.UUID
		)
		if err := rows.Scan(&id, &pid, &c.Title, &c.Description, &c.Weight, &createdBy, &c.CreatedAt); err != nil {
			return nil, err
		}
		c.ID = id.String()
		c.ProjectID = pid.String()
		c.CreatedBy = createdBy.String()
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) ListProjectCriterionGrades(
	ctx context.Context,
	projectID, professorID uuid.UUID,
) ([]projectflow.CriterionGrade, error) {
	const q = `
SELECT criterion_id, is_met, comment, updated_at
FROM project_criterion_reviews
WHERE project_id = $1
  AND professor_id = $2
ORDER BY updated_at DESC;
`
	rows, err := r.db.Query(ctx, q, projectID, professorID)
	if err != nil {
		if isUndefinedRelationErr(err, "project_criterion_reviews") {
			return []projectflow.CriterionGrade{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.CriterionGrade, 0, 16)
	for rows.Next() {
		var (
			item projectflow.CriterionGrade
			cid  uuid.UUID
		)
		if err := rows.Scan(&cid, &item.IsMet, &item.Comment, &item.UpdatedAt); err != nil {
			return nil, err
		}
		item.CriterionID = cid.String()
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) UpsertProjectCriterionGrades(
	ctx context.Context,
	projectID, professorID uuid.UUID,
	items []projectflow.CriterionGradeUpsert,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const qExists = `
SELECT EXISTS (
  SELECT 1
  FROM project_criteria
  WHERE id = $1
    AND project_id = $2
) AS ok;
`
	const qUpsert = `
INSERT INTO project_criterion_reviews(tenant_id, project_id, criterion_id, professor_id, is_met, comment)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, $3, $4, $5)
ON CONFLICT (project_id, criterion_id, professor_id)
DO UPDATE SET is_met = EXCLUDED.is_met, comment = EXCLUDED.comment, updated_at = now();
`

	for _, item := range items {
		var ok bool
		if err := tx.QueryRow(ctx, qExists, item.CriterionID, projectID).Scan(&ok); err != nil {
			return err
		}
		if !ok {
			return projectflow.ErrInvalidInput
		}
		if _, err := tx.Exec(ctx, qUpsert, projectID, item.CriterionID, professorID, item.IsMet, item.Comment); err != nil {
			if isUndefinedRelationErr(err, "project_criterion_reviews") {
				return projectflow.ErrSchemaMissing
			}
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ProjectFlowRepo) CountProjectCriteria(ctx context.Context, projectID uuid.UUID) (int, error) {
	const q = `
SELECT COUNT(*)
FROM project_criteria
WHERE project_id = $1;
`
	var total int
	if err := r.db.QueryRow(ctx, q, projectID).Scan(&total); err != nil {
		return 0, err
	}
	return total, nil
}

func (r *ProjectFlowRepo) CountProjectGradedCriteria(ctx context.Context, projectID, professorID uuid.UUID) (int, error) {
	const q = `
SELECT COUNT(*)
FROM project_criterion_reviews
WHERE project_id = $1
  AND professor_id = $2
  AND is_met IS NOT NULL;
`
	var total int
	if err := r.db.QueryRow(ctx, q, projectID, professorID).Scan(&total); err != nil {
		if isUndefinedRelationErr(err, "project_criterion_reviews") {
			return 0, projectflow.ErrSchemaMissing
		}
		return 0, err
	}
	return total, nil
}
