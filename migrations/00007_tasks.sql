-- +goose Up

CREATE TABLE IF NOT EXISTS tasks (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',

  -- задача всегда “для позиции” (Backend/Frontend/QA)
  position_id UUID NOT NULL,

  -- кто взял задачу (может быть null, пока не claimed)
  assignee_user_id UUID NULL,

  status TEXT NOT NULL DEFAULT 'OPEN', -- OPEN | IN_PROGRESS | DONE
  created_by UUID NOT NULL,

  due_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT tasks_status_check
    CHECK (status IN ('OPEN','IN_PROGRESS','DONE')),

  -- гарантирует: position_id принадлежит тому же project_id
  CONSTRAINT tasks_position_fk
    FOREIGN KEY (position_id, project_id)
    REFERENCES project_positions(id, project_id)
    ON DELETE RESTRICT,

  -- гарантирует: если назначен assignee, то он участник этого же проекта
  CONSTRAINT tasks_assignee_fk
    FOREIGN KEY (project_id, assignee_user_id)
    REFERENCES project_members(project_id, user_id)
    ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_project ON tasks(project_id);
CREATE INDEX IF NOT EXISTS idx_tasks_project_status ON tasks(project_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_position ON tasks(project_id, position_id);
CREATE INDEX IF NOT EXISTS idx_tasks_assignee ON tasks(project_id, assignee_user_id);

-- +goose Down

DROP INDEX IF EXISTS idx_tasks_assignee;
DROP INDEX IF EXISTS idx_tasks_position;
DROP INDEX IF EXISTS idx_tasks_project_status;
DROP INDEX IF EXISTS idx_tasks_project;

DROP TABLE IF EXISTS tasks;