package postgres

import (
	"context"
	"time"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TasksRepo struct {
	db *pgxpool.Pool
}

func NewTasksRepo(db *pgxpool.Pool) *TasksRepo {
	return &TasksRepo{db: db}
}

func (r *TasksRepo) Create(ctx context.Context, projectID uuid.UUID, title, description string, positionID uuid.UUID, createdBy uuid.UUID, dueAt *time.Time) (uuid.UUID, error) {
	const q = `
INSERT INTO tasks(project_id, title, description, position_id, created_by, due_at)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id;
`
	var id uuid.UUID
	err := r.db.QueryRow(ctx, q, projectID, title, description, positionID, createdBy, dueAt).Scan(&id)
	return id, err
}

func (r *TasksRepo) ListOpenByProject(ctx context.Context, projectID uuid.UUID) ([]domain.Task, error) {
	const q = `
SELECT id, project_id, title, description, position_id, assignee_user_id, status, created_by, due_at, created_at, updated_at
FROM tasks
WHERE project_id=$1 AND status='OPEN'
ORDER BY created_at ASC;
`
	rows, err := r.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domain.Task
	for rows.Next() {
		var t domain.Task
		var assignee *uuid.UUID
		var dueAt *time.Time
		if err := rows.Scan(
			&t.ID, &t.ProjectID, &t.Title, &t.Description,
			&t.PositionID, &assignee, &t.Status, &t.CreatedBy,
			&dueAt, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		t.AssigneeUserID = assignee
		t.DueAt = dueAt
		out = append(out, t)
	}
	return out, rows.Err()
}

// Claim assigns task to user ONLY if:
// - task is OPEN and unassigned
// - user is ACTIVE member of the same project
// - user's position_id == task.position_id
func (r *TasksRepo) Claim(ctx context.Context, projectID, taskID, userID uuid.UUID) error {
	const q = `
UPDATE tasks t
SET assignee_user_id = $3, status='IN_PROGRESS', updated_at=now()
WHERE t.id=$2
  AND t.project_id=$1
  AND t.status='OPEN'
  AND t.assignee_user_id IS NULL
  AND EXISTS (
    SELECT 1 FROM project_members m
    WHERE m.project_id=$1
      AND m.user_id=$3
      AND m.status='ACTIVE'
      AND m.position_id = t.position_id
  );
`
	ct, err := r.db.Exec(ctx, q, projectID, taskID, userID)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return domain.ErrForbidden
	}
	return nil
}
