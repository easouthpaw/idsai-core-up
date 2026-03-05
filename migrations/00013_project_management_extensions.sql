-- +goose Up

CREATE TABLE IF NOT EXISTS project_stacks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  stack_code TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_id, stack_code)
);

CREATE INDEX IF NOT EXISTS idx_project_stacks_project ON project_stacks(project_id);

CREATE TABLE IF NOT EXISTS project_criteria (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  weight INT NOT NULL DEFAULT 1,
  created_by UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT project_criteria_weight_check CHECK (weight > 0 AND weight <= 100)
);

CREATE INDEX IF NOT EXISTS idx_project_criteria_project ON project_criteria(project_id);

-- +goose Down

DROP INDEX IF EXISTS idx_project_criteria_project;
DROP TABLE IF EXISTS project_criteria;

DROP INDEX IF EXISTS idx_project_stacks_project;
DROP TABLE IF EXISTS project_stacks;
