-- +goose Up

INSERT INTO faculties(tenant_id, code, name)
SELECT
  t.id,
  t.code || '_SCHOOL',
  'Школьное направление'
FROM tenants t
WHERE NOT EXISTS (
  SELECT 1
  FROM faculties f
  WHERE f.tenant_id = t.id
    AND f.code = t.code || '_SCHOOL'
);

INSERT INTO departments(tenant_id, faculty_id, code, name)
SELECT
  t.id,
  f.id,
  'CLASS',
  'Школьные классы'
FROM tenants t
JOIN faculties f
  ON f.tenant_id = t.id
 AND f.code = t.code || '_SCHOOL'
WHERE NOT EXISTS (
  SELECT 1
  FROM departments d
  WHERE d.tenant_id = t.id
    AND d.faculty_id = f.id
    AND d.code = 'CLASS'
);

-- +goose Down

DELETE FROM departments d
USING tenants t, faculties f
WHERE d.tenant_id = t.id
  AND f.tenant_id = t.id
  AND f.id = d.faculty_id
  AND f.code = t.code || '_SCHOOL'
  AND d.code = 'CLASS'
  AND NOT EXISTS (
    SELECT 1
    FROM user_profiles up
    WHERE up.department_id = d.id
  )
  AND NOT EXISTS (
    SELECT 1
    FROM student_groups sg
    WHERE sg.department_id = d.id
  );

DELETE FROM faculties f
USING tenants t
WHERE f.tenant_id = t.id
  AND f.code = t.code || '_SCHOOL'
  AND NOT EXISTS (
    SELECT 1
    FROM user_profiles up
    WHERE up.faculty_id = f.id
  )
  AND NOT EXISTS (
    SELECT 1
    FROM departments d
    WHERE d.faculty_id = f.id
  )
  AND NOT EXISTS (
    SELECT 1
    FROM student_groups sg
    WHERE sg.faculty_id = f.id
  );
