-- +goose Up
INSERT INTO task_activity_logs (
  tenant_id,
  project_id,
  task_id,
  actor_user_id,
  event_type,
  from_status,
  to_status,
  title,
  comment,
  attachments,
  created_at
)
SELECT
  t.tenant_id,
  t.project_id,
  t.id,
  t.created_by,
  'CREATED',
  NULL,
  t.status,
  t.title,
  'Backfill from 00021',
  '[]'::jsonb,
  t.created_at
FROM tasks t
WHERE NOT EXISTS (
  SELECT 1
  FROM task_activity_logs l
  WHERE l.task_id = t.id
);

-- +goose Down
DELETE FROM task_activity_logs
WHERE event_type = 'CREATED'
  AND comment = 'Backfill from 00021';
