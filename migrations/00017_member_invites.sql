-- +goose Up

ALTER TABLE project_members
  ADD COLUMN IF NOT EXISTS invite_comment TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS invited_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
  ADD COLUMN IF NOT EXISTS responded_at TIMESTAMPTZ NULL;

ALTER TABLE project_members DROP CONSTRAINT IF EXISTS project_members_status_check;
ALTER TABLE project_members
  ADD CONSTRAINT project_members_status_check
  CHECK (status IN ('APPLIED','ACTIVE','REMOVED','INVITED','REJECTED'));

CREATE INDEX IF NOT EXISTS idx_project_members_status ON project_members(project_id, status);
CREATE INDEX IF NOT EXISTS idx_project_members_invited_by ON project_members(invited_by);

-- +goose Down

DROP INDEX IF EXISTS idx_project_members_invited_by;
DROP INDEX IF EXISTS idx_project_members_status;

ALTER TABLE project_members DROP CONSTRAINT IF EXISTS project_members_status_check;
ALTER TABLE project_members
  ADD CONSTRAINT project_members_status_check
  CHECK (status IN ('APPLIED','ACTIVE','REMOVED'));

ALTER TABLE project_members
  DROP COLUMN IF EXISTS responded_at,
  DROP COLUMN IF EXISTS invited_by,
  DROP COLUMN IF EXISTS invite_comment;
