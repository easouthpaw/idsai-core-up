-- +goose Up

-- 1) Faculties
CREATE TABLE IF NOT EXISTS faculties (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- 2) Student groups (belongs to a faculty)
CREATE TABLE IF NOT EXISTS student_groups (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  faculty_id UUID NOT NULL REFERENCES faculties(id) ON DELETE CASCADE,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  year INT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (faculty_id, code)
);

-- 3) Project visibility by faculty/group
-- visibility: PUBLIC | FACULTY | GROUP | PRIVATE
ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS faculty_id UUID NULL REFERENCES faculties(id),
  ADD COLUMN IF NOT EXISTS visibility TEXT NOT NULL DEFAULT 'FACULTY',
  ADD COLUMN IF NOT EXISTS group_id UUID NULL REFERENCES student_groups(id);

ALTER TABLE projects
  ADD CONSTRAINT projects_visibility_check
  CHECK (visibility IN ('PUBLIC','FACULTY','GROUP','PRIVATE'));

ALTER TABLE projects
  ADD CONSTRAINT projects_group_visibility_check
  CHECK (
    (visibility = 'GROUP' AND group_id IS NOT NULL)
    OR (visibility <> 'GROUP' AND group_id IS NULL)
  );

-- 4) Seed one default faculty + one default group (для MVP)
-- можно потом добавить другие faculties/groups
INSERT INTO faculties(code, name)
VALUES ('IDSAI_ENU', 'IDSAI (ENU)')
ON CONFLICT (code) DO NOTHING;

INSERT INTO student_groups(faculty_id, code, name, year)
SELECT f.id, 'CS-101', 'CS-101', 1
FROM faculties f
WHERE f.code='IDSAI_ENU'
ON CONFLICT (faculty_id, code) DO NOTHING;

-- 5) Backfill existing projects -> set faculty_id to default faculty
UPDATE projects
SET faculty_id = (SELECT id FROM faculties WHERE code='IDSAI_ENU')
WHERE faculty_id IS NULL;

-- 6) Now faculty_id can be NOT NULL
ALTER TABLE projects
  ALTER COLUMN faculty_id SET NOT NULL;

CREATE INDEX IF NOT EXISTS idx_projects_faculty ON projects(faculty_id);
CREATE INDEX IF NOT EXISTS idx_projects_group ON projects(group_id);
CREATE INDEX IF NOT EXISTS idx_projects_visibility ON projects(visibility);

-- +goose Down

DROP INDEX IF EXISTS idx_projects_visibility;
DROP INDEX IF EXISTS idx_projects_group;
DROP INDEX IF EXISTS idx_projects_faculty;

ALTER TABLE projects
  DROP CONSTRAINT IF EXISTS projects_group_visibility_check;

ALTER TABLE projects
  DROP CONSTRAINT IF EXISTS projects_visibility_check;

ALTER TABLE projects
  DROP COLUMN IF EXISTS group_id,
  DROP COLUMN IF EXISTS visibility,
  DROP COLUMN IF EXISTS faculty_id;

DROP TABLE IF EXISTS student_groups;
DROP TABLE IF EXISTS faculties;