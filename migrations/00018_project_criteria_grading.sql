-- +goose Up

CREATE TABLE IF NOT EXISTS project_criterion_reviews (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  criterion_id UUID NOT NULL REFERENCES project_criteria(id) ON DELETE CASCADE,
  professor_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  is_met BOOLEAN NULL,
  comment TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_id, criterion_id, professor_id)
);

CREATE INDEX IF NOT EXISTS idx_project_criterion_reviews_project_prof
  ON project_criterion_reviews(project_id, professor_id);
CREATE INDEX IF NOT EXISTS idx_project_criterion_reviews_tenant
  ON project_criterion_reviews(tenant_id);

-- +goose Down

DROP INDEX IF EXISTS idx_project_criterion_reviews_tenant;
DROP INDEX IF EXISTS idx_project_criterion_reviews_project_prof;
DROP TABLE IF EXISTS project_criterion_reviews;
