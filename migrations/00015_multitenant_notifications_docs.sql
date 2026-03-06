-- +goose Up

CREATE TABLE IF NOT EXISTS tenants (
  id UUID PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'ACTIVE',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT tenants_status_check CHECK (status IN ('ACTIVE', 'DISABLED'))
);

INSERT INTO tenants(id, code, name, status)
VALUES (
  '00000000-0000-0000-0000-000000000001',
  'CORE',
  'Core Default Tenant',
  'ACTIVE'
)
ON CONFLICT (code) DO NOTHING;

ALTER TABLE faculties
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE student_groups
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE departments
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE users
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE user_profiles
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE refresh_tokens
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE role_assignments
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001',
  ADD COLUMN IF NOT EXISTS repo_url TEXT NULL;
ALTER TABLE project_positions
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE project_members
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE tasks
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE project_stacks
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';
ALTER TABLE project_criteria
  ADD COLUMN IF NOT EXISTS tenant_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000001';

UPDATE faculties
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE student_groups
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE departments
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE users
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE user_profiles
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE refresh_tokens
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE role_assignments
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE projects
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE project_positions
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE project_members
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE tasks
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE project_stacks
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;
UPDATE project_criteria
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE tenant_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_faculties_tenant ON faculties(tenant_id);
CREATE INDEX IF NOT EXISTS idx_student_groups_tenant ON student_groups(tenant_id);
CREATE INDEX IF NOT EXISTS idx_departments_tenant ON departments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_tenant ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_user_profiles_tenant ON user_profiles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_refresh_tokens_tenant ON refresh_tokens(tenant_id);
CREATE INDEX IF NOT EXISTS idx_role_assignments_tenant ON role_assignments(tenant_id);
CREATE INDEX IF NOT EXISTS idx_projects_tenant ON projects(tenant_id);
CREATE INDEX IF NOT EXISTS idx_project_positions_tenant ON project_positions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_project_members_tenant ON project_members(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tasks_tenant ON tasks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_project_stacks_tenant ON project_stacks(tenant_id);
CREATE INDEX IF NOT EXISTS idx_project_criteria_tenant ON project_criteria(tenant_id);

ALTER TABLE role_assignments DROP CONSTRAINT IF EXISTS role_assignments_scope_type_check;
ALTER TABLE role_assignments
  ADD CONSTRAINT role_assignments_scope_type_check
  CHECK (scope_type IN ('SYSTEM', 'TENANT', 'FACULTY', 'DEPARTMENT', 'PROJECT'));

ALTER TABLE faculties
  ADD CONSTRAINT faculties_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE student_groups
  ADD CONSTRAINT student_groups_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE departments
  ADD CONSTRAINT departments_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE users
  ADD CONSTRAINT users_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE user_profiles
  ADD CONSTRAINT user_profiles_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE refresh_tokens
  ADD CONSTRAINT refresh_tokens_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE role_assignments
  ADD CONSTRAINT role_assignments_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE projects
  ADD CONSTRAINT projects_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE project_positions
  ADD CONSTRAINT project_positions_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE project_members
  ADD CONSTRAINT project_members_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE tasks
  ADD CONSTRAINT tasks_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE project_stacks
  ADD CONSTRAINT project_stacks_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;
ALTER TABLE project_criteria
  ADD CONSTRAINT project_criteria_tenant_fk
  FOREIGN KEY (tenant_id) REFERENCES tenants(id) ON DELETE RESTRICT;

CREATE TABLE IF NOT EXISTS project_files (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  uploaded_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  storage_key TEXT NOT NULL,
  file_name TEXT NOT NULL,
  content_type TEXT NOT NULL DEFAULT 'application/octet-stream',
  size_bytes BIGINT NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_id, storage_key)
);

CREATE INDEX IF NOT EXISTS idx_project_files_project ON project_files(project_id);
CREATE INDEX IF NOT EXISTS idx_project_files_tenant ON project_files(tenant_id);

CREATE TABLE IF NOT EXISTS notifications (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  title TEXT NOT NULL,
  body TEXT NOT NULL,
  payload JSONB NOT NULL DEFAULT '{}'::jsonb,
  is_read BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  read_at TIMESTAMPTZ NULL
);

CREATE INDEX IF NOT EXISTS idx_notifications_user_created
  ON notifications(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_notifications_user_unread
  ON notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_tenant
  ON notifications(tenant_id);

CREATE TABLE IF NOT EXISTS notification_outbox (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  notification_id UUID NOT NULL REFERENCES notifications(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  email_to TEXT NOT NULL,
  subject TEXT NOT NULL,
  body TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'PENDING',
  attempts INT NOT NULL DEFAULT 0,
  next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  last_error TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  sent_at TIMESTAMPTZ NULL,
  CONSTRAINT notification_outbox_status_check CHECK (status IN ('PENDING', 'RETRY', 'SENT', 'DEAD'))
);

CREATE INDEX IF NOT EXISTS idx_outbox_status_next_attempt
  ON notification_outbox(status, next_attempt_at);
CREATE INDEX IF NOT EXISTS idx_outbox_tenant
  ON notification_outbox(tenant_id);
CREATE INDEX IF NOT EXISTS idx_outbox_user
  ON notification_outbox(user_id);

-- +goose Down

DROP TABLE IF EXISTS notification_outbox;
DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS project_files;

ALTER TABLE project_criteria DROP CONSTRAINT IF EXISTS project_criteria_tenant_fk;
ALTER TABLE project_stacks DROP CONSTRAINT IF EXISTS project_stacks_tenant_fk;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS tasks_tenant_fk;
ALTER TABLE project_members DROP CONSTRAINT IF EXISTS project_members_tenant_fk;
ALTER TABLE project_positions DROP CONSTRAINT IF EXISTS project_positions_tenant_fk;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_tenant_fk;
ALTER TABLE role_assignments DROP CONSTRAINT IF EXISTS role_assignments_tenant_fk;
ALTER TABLE refresh_tokens DROP CONSTRAINT IF EXISTS refresh_tokens_tenant_fk;
ALTER TABLE user_profiles DROP CONSTRAINT IF EXISTS user_profiles_tenant_fk;
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_tenant_fk;
ALTER TABLE departments DROP CONSTRAINT IF EXISTS departments_tenant_fk;
ALTER TABLE student_groups DROP CONSTRAINT IF EXISTS student_groups_tenant_fk;
ALTER TABLE faculties DROP CONSTRAINT IF EXISTS faculties_tenant_fk;

ALTER TABLE role_assignments DROP CONSTRAINT IF EXISTS role_assignments_scope_type_check;
ALTER TABLE role_assignments
  ADD CONSTRAINT role_assignments_scope_type_check
  CHECK (scope_type IN ('SYSTEM', 'FACULTY', 'DEPARTMENT', 'PROJECT'));

ALTER TABLE project_criteria DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE project_stacks DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE tasks DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE project_members DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE project_positions DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE projects DROP COLUMN IF EXISTS repo_url;
ALTER TABLE projects DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE role_assignments DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE refresh_tokens DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE users DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE departments DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE student_groups DROP COLUMN IF EXISTS tenant_id;
ALTER TABLE faculties DROP COLUMN IF EXISTS tenant_id;

DROP TABLE IF EXISTS tenants;
