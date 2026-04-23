-- +goose Up

DELETE FROM role_assignments
WHERE role_id = (SELECT id FROM roles WHERE code = 'MODERATOR');

DELETE FROM roles
WHERE code = 'MODERATOR';

DELETE FROM permissions
WHERE code IN ('moderation.approve_free_project', 'moderation.reject_free_project');

DROP TABLE IF EXISTS project_files;

-- +goose Down

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

INSERT INTO permissions(code, description) VALUES
  ('moderation.approve_free_project', 'Approve free project'),
  ('moderation.reject_free_project', 'Reject free project')
ON CONFLICT (code) DO NOTHING;

INSERT INTO roles(code, name)
VALUES ('MODERATOR', 'Moderator')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN ('moderation.approve_free_project', 'moderation.reject_free_project')
WHERE r.code = 'MODERATOR'
ON CONFLICT DO NOTHING;
