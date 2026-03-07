-- +goose Up

ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS professor_review_status TEXT NOT NULL DEFAULT 'NONE',
  ADD COLUMN IF NOT EXISTS professor_invited_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS professor_responded_at TIMESTAMPTZ NULL;

ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_professor_review_status_check;
ALTER TABLE projects
  ADD CONSTRAINT projects_professor_review_status_check
  CHECK (professor_review_status IN ('NONE', 'PENDING', 'ACCEPTED', 'REJECTED'));

UPDATE projects
SET professor_review_status = 'ACCEPTED',
    professor_invited_at = COALESCE(professor_invited_at, updated_at),
    professor_responded_at = COALESCE(professor_responded_at, updated_at)
WHERE professor_id IS NOT NULL
  AND professor_review_status = 'NONE';

CREATE INDEX IF NOT EXISTS idx_projects_professor_review_status
  ON projects(professor_review_status);

-- +goose Down

DROP INDEX IF EXISTS idx_projects_professor_review_status;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_professor_review_status_check;
ALTER TABLE projects DROP COLUMN IF EXISTS professor_responded_at;
ALTER TABLE projects DROP COLUMN IF EXISTS professor_invited_at;
ALTER TABLE projects DROP COLUMN IF EXISTS professor_review_status;
