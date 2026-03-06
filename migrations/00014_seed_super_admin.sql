-- +goose Up

WITH target_department AS (
  SELECT d.id AS department_id, d.faculty_id
  FROM departments d
  JOIN faculties f ON f.id = d.faculty_id
  WHERE f.code = 'IDSAI_ENU'
    AND d.code = 'CPI'
  LIMIT 1
), upsert_user AS (
  INSERT INTO users(email, password_hash, status)
  VALUES (
    'admin@idsai.local',
    '$2a$10$155UfCkHn6HUSylmh.Wn5O.oC9EVk1twAiHodURoy6dZwSqOZDLkS',
    'ACTIVE'
  )
  ON CONFLICT (email) DO UPDATE
  SET status = EXCLUDED.status
  RETURNING id
), resolved_user AS (
  SELECT id FROM upsert_user
  UNION ALL
  SELECT u.id FROM users u WHERE u.email = 'admin@idsai.local'
  LIMIT 1
)
INSERT INTO user_profiles(user_id, full_name, faculty_id, department_id)
SELECT ru.id, 'Главный администратор', td.faculty_id, td.department_id
FROM resolved_user ru
CROSS JOIN target_department td
ON CONFLICT (user_id) DO UPDATE
SET full_name = EXCLUDED.full_name,
    faculty_id = EXCLUDED.faculty_id,
    department_id = EXCLUDED.department_id;

INSERT INTO role_assignments(user_id, role_id, scope_type, scope_id)
SELECT u.id, r.id, 'SYSTEM', NULL
FROM users u
JOIN roles r ON r.code = 'SUPER_ADMIN'
WHERE u.email = 'admin@idsai.local'
  AND NOT EXISTS (
    SELECT 1
    FROM role_assignments ra
    WHERE ra.user_id = u.id
      AND ra.role_id = r.id
      AND ra.scope_type = 'SYSTEM'
      AND ra.scope_id IS NULL
      AND ra.expires_at IS NULL
  );

-- +goose Down

DELETE FROM role_assignments ra
USING users u, roles r
WHERE ra.user_id = u.id
  AND ra.role_id = r.id
  AND u.email = 'admin@idsai.local'
  AND r.code = 'SUPER_ADMIN'
  AND ra.scope_type = 'SYSTEM'
  AND ra.scope_id IS NULL;

DELETE FROM user_profiles
WHERE user_id IN (
  SELECT id
  FROM users
  WHERE email = 'admin@idsai.local'
);

DELETE FROM users
WHERE email = 'admin@idsai.local';
