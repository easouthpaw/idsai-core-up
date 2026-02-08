-- +goose Up

CREATE TABLE IF NOT EXISTS projects (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'DRAFT', -- DRAFT | REVIEW | RECRUITMENT | ACTIVE | GRADING | ARCHIVE
  is_public BOOLEAN NOT NULL DEFAULT FALSE,

  created_by UUID NOT NULL, -- user_id (пока без FK на users)
  professor_id UUID NULL,   -- assigned professor for the project (later)

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT projects_status_check
    CHECK (status IN ('DRAFT','REVIEW','RECRUITMENT','ACTIVE','GRADING','ARCHIVE'))
);

CREATE INDEX IF NOT EXISTS idx_projects_status ON projects(status);
CREATE INDEX IF NOT EXISTS idx_projects_created_by ON projects(created_by);

-- +goose Down
DROP TABLE IF EXISTS projects;