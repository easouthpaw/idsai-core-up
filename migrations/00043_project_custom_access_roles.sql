-- +goose Up

CREATE TABLE IF NOT EXISTS project_access_roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  created_by UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (tenant_id, project_id, code),
  UNIQUE (role_id)
);

CREATE INDEX IF NOT EXISTS idx_project_access_roles_project
  ON project_access_roles(tenant_id, project_id);

-- +goose Down

DROP INDEX IF EXISTS idx_project_access_roles_project;
DROP TABLE IF EXISTS project_access_roles;
