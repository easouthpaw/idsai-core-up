-- +goose Up

ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_status_check;

ALTER TABLE projects
  ADD CONSTRAINT projects_status_check
  CHECK (status IN ('DRAFT','REVIEW','RECRUITMENT','ACTIVE','GRADING','COMPLETED','ARCHIVE'));

-- +goose Down

UPDATE projects
SET status = 'ARCHIVE'
WHERE status = 'COMPLETED';

ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_status_check;

ALTER TABLE projects
  ADD CONSTRAINT projects_status_check
  CHECK (status IN ('DRAFT','REVIEW','RECRUITMENT','ACTIVE','GRADING','ARCHIVE'));
