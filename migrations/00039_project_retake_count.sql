-- +goose Up

ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS retake_count INTEGER NOT NULL DEFAULT 0;

ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_retake_count_check;

ALTER TABLE projects
  ADD CONSTRAINT projects_retake_count_check
  CHECK (retake_count >= 0);

-- +goose Down

ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_retake_count_check;

ALTER TABLE projects
  DROP COLUMN IF EXISTS retake_count;
