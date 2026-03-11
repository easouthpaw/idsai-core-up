-- +goose Up

CREATE TABLE IF NOT EXISTS task_submissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  comment TEXT NOT NULL DEFAULT '',
  attachments JSONB NOT NULL DEFAULT '[]'::jsonb,
  submitted_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (task_id)
);

CREATE INDEX IF NOT EXISTS idx_task_submissions_project ON task_submissions(project_id);
CREATE INDEX IF NOT EXISTS idx_task_submissions_tenant ON task_submissions(tenant_id);

CREATE TABLE IF NOT EXISTS task_activity_logs (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  task_id UUID NOT NULL REFERENCES tasks(id) ON DELETE CASCADE,
  actor_user_id UUID NULL REFERENCES users(id) ON DELETE SET NULL,
  event_type TEXT NOT NULL,
  from_status TEXT NULL,
  to_status TEXT NULL,
  title TEXT NOT NULL DEFAULT '',
  comment TEXT NOT NULL DEFAULT '',
  attachments JSONB NOT NULL DEFAULT '[]'::jsonb,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT task_activity_event_type_check CHECK (event_type IN ('CREATED', 'ASSIGNED', 'CLAIMED', 'STATUS_CHANGED', 'COMPLETED'))
);

CREATE INDEX IF NOT EXISTS idx_task_activity_logs_project_task_time
  ON task_activity_logs(project_id, task_id, created_at ASC);
CREATE INDEX IF NOT EXISTS idx_task_activity_logs_tenant ON task_activity_logs(tenant_id);

-- +goose Down

DROP INDEX IF EXISTS idx_task_activity_logs_tenant;
DROP INDEX IF EXISTS idx_task_activity_logs_project_task_time;
DROP TABLE IF EXISTS task_activity_logs;

DROP INDEX IF EXISTS idx_task_submissions_tenant;
DROP INDEX IF EXISTS idx_task_submissions_project;
DROP TABLE IF EXISTS task_submissions;
