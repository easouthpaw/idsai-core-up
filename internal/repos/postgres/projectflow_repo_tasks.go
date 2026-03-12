package postgres

import (
	"context"
	"strings"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
)

func (r *ProjectFlowRepo) CreateTask(
	ctx context.Context,
	projectID uuid.UUID,
	title, description string,
	positionID uuid.UUID,
	assigneeUserID *uuid.UUID,
	status string,
	createdBy uuid.UUID,
	dueAt *time.Time,
) (uuid.UUID, error) {
	const q = `
INSERT INTO tasks(tenant_id, project_id, title, description, position_id, assignee_user_id, status, created_by, due_at)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, $3, $4, $5, $6, $7, $8)
RETURNING id;
`
	var taskID uuid.UUID
	if err := r.db.QueryRow(ctx, q, projectID, title, description, positionID, assigneeUserID, status, createdBy, dueAt).Scan(&taskID); err != nil {
		return uuid.Nil, err
	}
	return taskID, nil
}

func (r *ProjectFlowRepo) GetTaskByID(ctx context.Context, projectID, taskID uuid.UUID) (projectflow.Task, error) {
	const q = `
SELECT t.id, t.project_id, t.title, t.description, t.position_id,
       t.assignee_user_id, t.status, t.created_by, t.due_at, t.created_at, t.updated_at,
       p.code, p.name
FROM tasks t
JOIN project_positions p ON p.id = t.position_id
WHERE t.project_id = $1
  AND t.id = $2;
`
	row := r.db.QueryRow(ctx, q, projectID, taskID)
	t, err := scanTaskRow(row)
	return t, mapProjectFlowErr(err)
}

func (r *ProjectFlowRepo) ListProjectTasks(ctx context.Context, projectID uuid.UUID) ([]projectflow.Task, error) {
	const q = `
SELECT t.id, t.project_id, t.title, t.description, t.position_id,
       t.assignee_user_id, t.status, t.created_by, t.due_at, t.created_at, t.updated_at,
       p.code, p.name
FROM tasks t
JOIN project_positions p ON p.id = t.position_id
WHERE t.project_id = $1
ORDER BY t.created_at ASC;
`
	rows, err := r.db.Query(ctx, q, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.Task, 0, 16)
	for rows.Next() {
		t, err := scanTaskRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) GetTaskStatusAndTitle(ctx context.Context, projectID, taskID uuid.UUID) (string, string, error) {
	const q = `
SELECT status, title
FROM tasks
WHERE project_id=$1 AND id=$2;
`
	var status string
	var title string
	if err := r.db.QueryRow(ctx, q, projectID, taskID).Scan(&status, &title); err != nil {
		return "", "", mapProjectFlowErr(err)
	}
	return status, title, nil
}

func (r *ProjectFlowRepo) UpdateTaskStatus(ctx context.Context, projectID, taskID uuid.UUID, status string) (uuid.UUID, error) {
	const q = `
UPDATE tasks
SET status=$3, updated_at=now()
WHERE project_id=$1 AND id=$2
RETURNING id;
`
	var outTaskID uuid.UUID
	if err := r.db.QueryRow(ctx, q, projectID, taskID, status).Scan(&outTaskID); err != nil {
		return uuid.Nil, mapProjectFlowErr(err)
	}
	return outTaskID, nil
}

func (r *ProjectFlowRepo) GetTaskAssignContext(ctx context.Context, projectID, taskID uuid.UUID) (uuid.UUID, string, string, *uuid.UUID, error) {
	const q = `
SELECT position_id, status, title, assignee_user_id
FROM tasks
WHERE project_id = $1 AND id = $2;
`
	var positionID uuid.UUID
	var prevStatus string
	var taskTitle string
	var prevAssignee *uuid.UUID
	if err := r.db.QueryRow(ctx, q, projectID, taskID).Scan(&positionID, &prevStatus, &taskTitle, &prevAssignee); err != nil {
		return uuid.Nil, "", "", nil, mapProjectFlowErr(err)
	}
	return positionID, prevStatus, taskTitle, prevAssignee, nil
}

func (r *ProjectFlowRepo) AssignTaskToUser(ctx context.Context, projectID, taskID, assigneeUserID uuid.UUID) (uuid.UUID, error) {
	const q = `
UPDATE tasks
SET assignee_user_id = $3,
    updated_at = now()
WHERE project_id=$1 AND id=$2
RETURNING id;
`
	var outTaskID uuid.UUID
	if err := r.db.QueryRow(ctx, q, projectID, taskID, assigneeUserID).Scan(&outTaskID); err != nil {
		return uuid.Nil, mapProjectFlowErr(err)
	}
	return outTaskID, nil
}

func (r *ProjectFlowRepo) ListProjectTaskActivities(ctx context.Context, projectID uuid.UUID, taskID *uuid.UUID) ([]projectflow.TaskActivity, error) {
	baseQ := `
SELECT
  a.id, a.project_id, a.task_id, a.actor_user_id,
  COALESCE(NULLIF(TRIM(up.full_name), ''), split_part(COALESCE(u.email, ''), '@', 1), '') AS actor_name,
  COALESCE(u.email, '') AS actor_email,
  a.event_type,
  COALESCE(a.from_status, '') AS from_status,
  COALESCE(a.to_status, '') AS to_status,
  COALESCE(a.title, '') AS title,
  COALESCE(a.comment, '') AS comment,
  a.attachments,
  a.created_at
FROM task_activity_logs a
LEFT JOIN users u ON u.id = a.actor_user_id
LEFT JOIN user_profiles up ON up.user_id = a.actor_user_id
WHERE a.project_id = $1
`
	args := []any{projectID}
	if taskID != nil {
		baseQ += ` AND a.task_id = $2`
		args = append(args, *taskID)
	}
	baseQ += ` ORDER BY a.created_at ASC;`

	rows, err := r.db.Query(ctx, baseQ, args...)
	if err != nil {
		if isUndefinedRelationErr(err, "task_activity_logs") {
			return []projectflow.TaskActivity{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.TaskActivity, 0, 32)
	for rows.Next() {
		var (
			item           projectflow.TaskActivity
			id             uuid.UUID
			pid            uuid.UUID
			tid            uuid.UUID
			actorID        *uuid.UUID
			rawAttachments []byte
		)
		if err := rows.Scan(
			&id,
			&pid,
			&tid,
			&actorID,
			&item.ActorName,
			&item.ActorEmail,
			&item.EventType,
			&item.FromStatus,
			&item.ToStatus,
			&item.Title,
			&item.Comment,
			&rawAttachments,
			&item.CreatedAt,
		); err != nil {
			return nil, err
		}
		item.ID = id.String()
		item.ProjectID = pid.String()
		item.TaskID = tid.String()
		item.EventType = strings.ToUpper(strings.TrimSpace(item.EventType))
		item.FromStatus = strings.ToUpper(strings.TrimSpace(item.FromStatus))
		item.ToStatus = strings.ToUpper(strings.TrimSpace(item.ToStatus))
		item.Attachments = decodeStringSliceJSON(rawAttachments)
		if actorID != nil {
			v := actorID.String()
			item.ActorUserID = &v
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) GetTaskCompleteContext(ctx context.Context, projectID, taskID uuid.UUID) (*uuid.UUID, string, string, error) {
	const q = `
SELECT assignee_user_id, status, title
FROM tasks
WHERE project_id=$1 AND id=$2;
`
	var assigneeID *uuid.UUID
	var currentStatus string
	var taskTitle string
	if err := r.db.QueryRow(ctx, q, projectID, taskID).Scan(&assigneeID, &currentStatus, &taskTitle); err != nil {
		return nil, "", "", mapProjectFlowErr(err)
	}
	return assigneeID, currentStatus, taskTitle, nil
}

func (r *ProjectFlowRepo) UpsertTaskSubmission(ctx context.Context, projectID, taskID, userID uuid.UUID, comment string, attachments []string) error {
	const q = `
INSERT INTO task_submissions(tenant_id, project_id, task_id, user_id, comment, attachments)
VALUES ((SELECT tenant_id FROM projects WHERE id = $1), $1, $2, $3, $4, $5::jsonb)
ON CONFLICT (task_id)
DO UPDATE SET comment = EXCLUDED.comment, attachments = EXCLUDED.attachments, updated_at = now(), submitted_at = now();
`
	if _, err := r.db.Exec(ctx, q, projectID, taskID, userID, comment, encodeStringSliceJSON(attachments)); err != nil {
		if isUndefinedRelationErr(err, "task_submissions") {
			return nil
		}
		return err
	}
	return nil
}

func (r *ProjectFlowRepo) MarkTaskDone(ctx context.Context, projectID, taskID uuid.UUID) (uuid.UUID, error) {
	const q = `
UPDATE tasks
SET status='DONE', updated_at=now()
WHERE project_id=$1 AND id=$2
RETURNING id;
`
	var outTaskID uuid.UUID
	if err := r.db.QueryRow(ctx, q, projectID, taskID).Scan(&outTaskID); err != nil {
		return uuid.Nil, mapProjectFlowErr(err)
	}
	return outTaskID, nil
}

func (r *ProjectFlowRepo) ClaimTask(ctx context.Context, projectID, taskID, userID uuid.UUID) error {
	const q = `
UPDATE tasks t
SET assignee_user_id = COALESCE(t.assignee_user_id, $3), status='IN_PROGRESS', updated_at=now()
WHERE t.id=$2
  AND t.project_id=$1
  AND t.status='OPEN'
  AND (
    (
      t.assignee_user_id IS NULL
      AND EXISTS (
        SELECT 1 FROM project_members m
        WHERE m.project_id=$1
          AND m.user_id=$3
          AND m.status='ACTIVE'
          AND m.position_id = t.position_id
      )
    )
    OR t.assignee_user_id = $3
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

func (r *ProjectFlowRepo) InsertTaskActivity(
	ctx context.Context,
	projectID, taskID uuid.UUID,
	actorUserID *uuid.UUID,
	eventType, fromStatus, toStatus, title, comment string,
	attachments []string,
) error {
	const q = `
INSERT INTO task_activity_logs(
  tenant_id, project_id, task_id, actor_user_id, event_type,
  from_status, to_status, title, comment, attachments
)
VALUES (
  (SELECT tenant_id FROM projects WHERE id = $1),
  $1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7, $8, $9::jsonb
);
`
	_, err := r.db.Exec(ctx, q, projectID, taskID, actorUserID, eventType, fromStatus, toStatus, title, comment, encodeStringSliceJSON(attachments))
	if err != nil && isUndefinedRelationErr(err, "task_activity_logs") {
		return nil
	}
	return err
}
