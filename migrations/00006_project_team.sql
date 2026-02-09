-- +goose Up

-- Позиции внутри проекта (BACKEND, FRONTEND, QA, DESIGN...)
CREATE TABLE IF NOT EXISTS project_positions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  code TEXT NOT NULL, -- e.g. BACKEND
  name TEXT NOT NULL, -- e.g. Backend Developer
  capacity INT NOT NULL DEFAULT 1,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (project_id, code),
  UNIQUE (id, project_id) -- нужно для составных FK (чтобы position принадлежала проекту)
);

-- Участники проекта
CREATE TABLE IF NOT EXISTS project_members (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id UUID NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id UUID NOT NULL,

  position_id UUID NULL,
  status TEXT NOT NULL DEFAULT 'APPLIED', -- APPLIED | ACTIVE | REMOVED
  joined_at TIMESTAMPTZ NULL,

  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  UNIQUE (project_id, user_id),

  CONSTRAINT project_members_status_check
    CHECK (status IN ('APPLIED','ACTIVE','REMOVED')),

  -- гарантирует: если position_id задан, то эта позиция принадлежит ТОМУ ЖЕ проекту
  CONSTRAINT project_members_position_fk
    FOREIGN KEY (position_id, project_id)
    REFERENCES project_positions(id, project_id)
    ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_positions_project ON project_positions(project_id);
CREATE INDEX IF NOT EXISTS idx_members_project ON project_members(project_id);
CREATE INDEX IF NOT EXISTS idx_members_user ON project_members(user_id);

-- +goose Down

DROP INDEX IF EXISTS idx_members_user;
DROP INDEX IF EXISTS idx_members_project;
DROP INDEX IF EXISTS idx_positions_project;

DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS project_positions;