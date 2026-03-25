-- +goose Up

ALTER TABLE student_groups
  ADD COLUMN IF NOT EXISTS department_id UUID NULL,
  ADD COLUMN IF NOT EXISTS group_code TEXT NULL,
  ADD COLUMN IF NOT EXISTS group_number INT NULL,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'student_groups_department_fk'
  ) THEN
    ALTER TABLE student_groups
      ADD CONSTRAINT student_groups_department_fk
      FOREIGN KEY (department_id) REFERENCES departments(id) ON DELETE RESTRICT;
  END IF;
END $$;
-- +goose StatementEnd

UPDATE student_groups
SET group_code = UPPER(TRIM(COALESCE(code, '')))
WHERE group_code IS NULL;

UPDATE student_groups sg
SET department_id = d.id
FROM departments d
WHERE sg.department_id IS NULL
  AND d.tenant_id = sg.tenant_id
  AND d.faculty_id = sg.faculty_id
  AND d.code = (
    CASE UPPER(SPLIT_PART(COALESCE(sg.group_code, sg.code), '-', 1))
      WHEN 'CS' THEN 'CPI'
      WHEN 'CSE' THEN 'CPI'
      ELSE UPPER(SPLIT_PART(COALESCE(sg.group_code, sg.code), '-', 1))
    END
  );

UPDATE student_groups sg
SET department_id = d.id
FROM departments d
WHERE sg.department_id IS NULL
  AND d.tenant_id = sg.tenant_id
  AND d.faculty_id = sg.faculty_id
  AND d.code = 'CPI';

UPDATE student_groups sg
SET department_id = (
  SELECT d.id
  FROM departments d
  WHERE d.tenant_id = sg.tenant_id
    AND d.faculty_id = sg.faculty_id
  ORDER BY d.code ASC
  LIMIT 1
)
WHERE sg.department_id IS NULL;

UPDATE student_groups
SET group_number = (
  CASE
    WHEN REGEXP_REPLACE(COALESCE(group_code, code, ''), '^[^-]*-', '') ~ '^[0-9]+$'
      THEN REGEXP_REPLACE(COALESCE(group_code, code, ''), '^[^-]*-', '')::int
    ELSE 45
  END
)
WHERE group_number IS NULL;

WITH ranked AS (
  SELECT
    sg.id,
    ROW_NUMBER() OVER (
      PARTITION BY sg.tenant_id, sg.department_id, sg.group_number
      ORDER BY sg.created_at ASC, sg.id ASC
    ) AS rn
  FROM student_groups sg
)
UPDATE student_groups sg
SET group_number = sg.group_number * 1000 + ranked.rn
FROM ranked
WHERE sg.id = ranked.id
  AND ranked.rn > 1;

UPDATE student_groups sg
SET group_code = d.code || '-' || sg.group_number::text,
    code = d.code || '-' || sg.group_number::text,
    name = d.code || '-' || sg.group_number::text,
    updated_at = now()
FROM departments d
WHERE d.id = sg.department_id;

INSERT INTO student_groups(
  tenant_id,
  faculty_id,
  department_id,
  code,
  name,
  year,
  group_code,
  group_number,
  created_at,
  updated_at
)
SELECT
  d.tenant_id,
  d.faculty_id,
  d.id,
  d.code || '-45',
  d.code || '-45',
  NULL,
  d.code || '-45',
  45,
  now(),
  now()
FROM departments d
WHERE NOT EXISTS (
  SELECT 1
  FROM student_groups sg
  WHERE sg.tenant_id = d.tenant_id
    AND sg.department_id = d.id
);

ALTER TABLE student_groups
  ALTER COLUMN department_id SET NOT NULL,
  ALTER COLUMN group_code SET NOT NULL,
  ALTER COLUMN group_number SET NOT NULL;

ALTER TABLE student_groups
  DROP CONSTRAINT IF EXISTS student_groups_group_number_check;
ALTER TABLE student_groups
  ADD CONSTRAINT student_groups_group_number_check
  CHECK (group_number > 0);

CREATE UNIQUE INDEX IF NOT EXISTS uq_student_groups_tenant_group_code
  ON student_groups(tenant_id, group_code);

CREATE UNIQUE INDEX IF NOT EXISTS uq_student_groups_id_department
  ON student_groups(id, department_id);

CREATE UNIQUE INDEX IF NOT EXISTS uq_student_groups_department_number
  ON student_groups(department_id, group_number);

CREATE INDEX IF NOT EXISTS idx_student_groups_department
  ON student_groups(department_id);

ALTER TABLE user_profiles
  ADD COLUMN IF NOT EXISTS group_id UUID NULL;

UPDATE user_profiles up
SET group_id = (
  SELECT sg.id
  FROM student_groups sg
  WHERE sg.tenant_id = up.tenant_id
    AND sg.department_id = up.department_id
  ORDER BY sg.group_number ASC, sg.created_at ASC
  LIMIT 1
)
WHERE up.group_id IS NULL;

-- +goose StatementBegin
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'user_profiles_group_department_fk'
  ) THEN
    ALTER TABLE user_profiles
      ADD CONSTRAINT user_profiles_group_department_fk
      FOREIGN KEY (group_id, department_id)
      REFERENCES student_groups(id, department_id)
      ON DELETE RESTRICT;
  END IF;
END $$;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_user_profiles_group
  ON user_profiles(group_id);

CREATE TABLE IF NOT EXISTS group_change_requests (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE RESTRICT,
  student_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  current_group_id UUID NOT NULL REFERENCES student_groups(id) ON DELETE RESTRICT,
  requested_group_id UUID NOT NULL REFERENCES student_groups(id) ON DELETE RESTRICT,
  status TEXT NOT NULL DEFAULT 'PENDING',
  admin_comment TEXT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  reviewed_at TIMESTAMPTZ NULL,
  reviewed_by UUID NULL REFERENCES users(id) ON DELETE SET NULL,
  CONSTRAINT group_change_requests_status_check CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
  CONSTRAINT group_change_requests_groups_diff_check CHECK (current_group_id <> requested_group_id)
);

CREATE INDEX IF NOT EXISTS idx_group_change_requests_student
  ON group_change_requests(student_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_group_change_requests_status
  ON group_change_requests(status, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_group_change_requests_tenant
  ON group_change_requests(tenant_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS uq_group_change_requests_pending_student
  ON group_change_requests(tenant_id, student_id)
  WHERE status = 'PENDING';

-- +goose Down

DROP INDEX IF EXISTS uq_group_change_requests_pending_student;
DROP INDEX IF EXISTS idx_group_change_requests_tenant;
DROP INDEX IF EXISTS idx_group_change_requests_status;
DROP INDEX IF EXISTS idx_group_change_requests_student;
DROP TABLE IF EXISTS group_change_requests;

DROP INDEX IF EXISTS idx_user_profiles_group;
ALTER TABLE user_profiles DROP CONSTRAINT IF EXISTS user_profiles_group_department_fk;
ALTER TABLE user_profiles DROP COLUMN IF EXISTS group_id;

DROP INDEX IF EXISTS idx_student_groups_department;
DROP INDEX IF EXISTS uq_student_groups_department_number;
DROP INDEX IF EXISTS uq_student_groups_id_department;
DROP INDEX IF EXISTS uq_student_groups_tenant_group_code;

ALTER TABLE student_groups DROP CONSTRAINT IF EXISTS student_groups_group_number_check;
ALTER TABLE student_groups DROP CONSTRAINT IF EXISTS student_groups_department_fk;

ALTER TABLE student_groups
  DROP COLUMN IF EXISTS updated_at,
  DROP COLUMN IF EXISTS group_number,
  DROP COLUMN IF EXISTS group_code,
  DROP COLUMN IF EXISTS department_id;
