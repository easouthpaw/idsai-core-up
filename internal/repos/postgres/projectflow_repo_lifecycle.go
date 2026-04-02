package postgres

import (
	"context"
	"errors"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (r *ProjectFlowRepo) ActivateProject(ctx context.Context, projectID uuid.UUID) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	const q = `
UPDATE projects
SET status='ACTIVE', updated_at=now()
WHERE tenant_id = $1
  AND id = $2
  AND status IN ('REVIEW', 'RECRUITMENT');
`
	ct, err := r.db.Exec(ctx, q, tenantID, projectID)
	if err != nil {
		if isUndefinedColumnErr(err, "retake_count") {
			return projectflow.ErrSchemaMissing
		}
		return err
	}
	if ct.RowsAffected() == 0 {
		return projectflow.ErrNotFound
	}
	return nil
}

func (r *ProjectFlowRepo) CountProjectTasksSummary(ctx context.Context, projectID uuid.UUID) (int, int, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return 0, 0, err
	}

	const q = `
SELECT
  COUNT(*) AS total,
  COUNT(*) FILTER (WHERE status = 'DONE') AS done
FROM tasks
WHERE tenant_id = $1
  AND project_id = $2;
`
	var total int
	var done int
	if err := r.db.QueryRow(ctx, q, tenantID, projectID).Scan(&total, &done); err != nil {
		return 0, 0, err
	}
	return total, done, nil
}

func (r *ProjectFlowRepo) MoveProjectToGrading(ctx context.Context, projectID uuid.UUID) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	const q = `
UPDATE projects
SET status = 'GRADING', updated_at = now()
WHERE tenant_id = $1
  AND id = $2
  AND status = 'ACTIVE';
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

func (r *ProjectFlowRepo) ReturnProjectToActive(ctx context.Context, projectID uuid.UUID) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	const q = `
UPDATE projects
SET status = 'ACTIVE',
    retake_count = retake_count + 1,
    updated_at = now()
WHERE tenant_id = $1
  AND id = $2
  AND status = 'GRADING';
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

func (r *ProjectFlowRepo) MoveProjectToCompleted(ctx context.Context, projectID uuid.UUID) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	const q = `
	UPDATE projects
	SET status = 'COMPLETED', updated_at = now()
	WHERE tenant_id = $1
	  AND id = $2
	  AND status = 'GRADING';
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

func (r *ProjectFlowRepo) DeleteOwnedProject(ctx context.Context, projectID, ownerID uuid.UUID) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var currentOwnerID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT created_by FROM projects WHERE tenant_id = $1 AND id = $2`, tenantID, projectID).Scan(&currentOwnerID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return projectflow.ErrNotFound
		}
		return err
	}
	if currentOwnerID != ownerID {
		return domain.ErrForbidden
	}

	if _, err := tx.Exec(ctx, `
DELETE FROM role_assignments
WHERE tenant_id = $1
  AND scope_type = 'PROJECT'
  AND scope_id = $2;
`, tenantID, projectID); err != nil {
		return err
	}

	ct, err := tx.Exec(ctx, `DELETE FROM projects WHERE tenant_id = $1 AND id = $2`, tenantID, projectID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return projectflow.ErrNotFound
	}

	return tx.Commit(ctx)
}
