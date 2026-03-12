-- +goose Up

-- Demo credentials for seeded accounts:
-- password: DemoPass123!

WITH faculty_ctx AS (
  SELECT f.id AS faculty_id, f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), seed_users AS (
  SELECT *
  FROM (VALUES
    ('aibolat.student@idsai.local', 'Айболат Ермекбай',  'CPI', 'STUDENT'),
    ('aliya.student@idsai.local',   'Алия Нурланова',    'AI',  'STUDENT'),
    ('daniyar.student@idsai.local', 'Данияр Сагындыков', 'INF', 'STUDENT'),
    ('dinara.student@idsai.local',  'Динара Абдрахманова','IS', 'STUDENT'),
    ('marat.student@idsai.local',   'Марат Касенов',      'SAU','STUDENT'),
    ('nurzhan.prof@idsai.local',    'Нуржан Толеубаев',  'AI',  'PROFESSOR'),
    ('aidana.prof@idsai.local',     'Айдана Серикова',   'SEC', 'PROFESSOR'),
    ('sabina.prof@idsai.local',     'Сабина Омарова',    'CPI', 'PROFESSOR')
  ) AS v(email, full_name, department_code, role_code)
)
INSERT INTO users(tenant_id, email, password_hash, status)
SELECT
  fc.tenant_id,
  su.email,
  '$2a$10$Gsnz9k1a6acT.eGCCMDOwudDgrTBCGxM//7fI4PZzc0oYPyTNAokG',
  'ACTIVE'
FROM seed_users su
CROSS JOIN faculty_ctx fc
ON CONFLICT (email) DO UPDATE
SET status = EXCLUDED.status;

WITH faculty_ctx AS (
  SELECT f.id AS faculty_id, f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), seed_users AS (
  SELECT *
  FROM (VALUES
    ('aibolat.student@idsai.local', 'Айболат Ермекбай',  'CPI', 'STUDENT'),
    ('aliya.student@idsai.local',   'Алия Нурланова',    'AI',  'STUDENT'),
    ('daniyar.student@idsai.local', 'Данияр Сагындыков', 'INF', 'STUDENT'),
    ('dinara.student@idsai.local',  'Динара Абдрахманова','IS', 'STUDENT'),
    ('marat.student@idsai.local',   'Марат Касенов',      'SAU','STUDENT'),
    ('nurzhan.prof@idsai.local',    'Нуржан Толеубаев',  'AI',  'PROFESSOR'),
    ('aidana.prof@idsai.local',     'Айдана Серикова',   'SEC', 'PROFESSOR'),
    ('sabina.prof@idsai.local',     'Сабина Омарова',    'CPI', 'PROFESSOR')
  ) AS v(email, full_name, department_code, role_code)
)
INSERT INTO user_profiles(tenant_id, user_id, full_name, faculty_id, department_id)
SELECT
  fc.tenant_id,
  u.id,
  su.full_name,
  fc.faculty_id,
  d.id
FROM seed_users su
JOIN users u ON u.email = su.email
JOIN faculty_ctx fc ON TRUE
JOIN departments d ON d.faculty_id = fc.faculty_id AND d.code = su.department_code
ON CONFLICT (user_id) DO UPDATE
SET full_name = EXCLUDED.full_name,
    faculty_id = EXCLUDED.faculty_id,
    department_id = EXCLUDED.department_id,
    tenant_id = EXCLUDED.tenant_id;

WITH faculty_ctx AS (
  SELECT f.id AS faculty_id, f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), seed_users AS (
  SELECT *
  FROM (VALUES
    ('aibolat.student@idsai.local', 'STUDENT'),
    ('aliya.student@idsai.local',   'STUDENT'),
    ('daniyar.student@idsai.local', 'STUDENT'),
    ('dinara.student@idsai.local',  'STUDENT'),
    ('marat.student@idsai.local',   'STUDENT'),
    ('nurzhan.prof@idsai.local',    'PROFESSOR'),
    ('aidana.prof@idsai.local',     'PROFESSOR'),
    ('sabina.prof@idsai.local',     'PROFESSOR')
  ) AS v(email, role_code)
)
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT
  fc.tenant_id,
  u.id,
  r.id,
  'FACULTY',
  fc.faculty_id
FROM seed_users su
JOIN users u ON u.email = su.email
JOIN roles r ON r.code = su.role_code
JOIN faculty_ctx fc ON TRUE
WHERE NOT EXISTS (
  SELECT 1
  FROM role_assignments ra
  WHERE ra.user_id = u.id
    AND ra.role_id = r.id
    AND ra.scope_type = 'FACULTY'
    AND ra.scope_id = fc.faculty_id
    AND ra.expires_at IS NULL
);

WITH faculty_ctx AS (
  SELECT f.id AS faculty_id, f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), project_seed AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000001'::uuid, 'Smart Campus Attendance',      'DRAFT',       'PUBLIC',   NULL::text,   'aibolat.student@idsai.local', NULL::text,                'NONE'),
    ('11000000-0000-0000-0000-000000000002'::uuid, 'AI Schedule Optimizer',         'REVIEW',      'GROUP',    'AI-46',      'aliya.student@idsai.local',   'nurzhan.prof@idsai.local','PENDING'),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'Secure Student ID',             'RECRUITMENT', 'FACULTY',  NULL::text,   'daniyar.student@idsai.local', 'aidana.prof@idsai.local', 'ACCEPTED'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'Digital Dean Office',           'ACTIVE',      'GROUP',    'IS-46',      'aibolat.student@idsai.local', 'nurzhan.prof@idsai.local','ACCEPTED'),
    ('11000000-0000-0000-0000-000000000005'::uuid, 'LLM Study Assistant',           'GRADING',     'PUBLIC',   NULL::text,   'aliya.student@idsai.local',   'aidana.prof@idsai.local', 'ACCEPTED'),
    ('11000000-0000-0000-0000-000000000006'::uuid, 'Archive: Legacy ERP Connector', 'ARCHIVE',     'FACULTY',  NULL::text,   'daniyar.student@idsai.local', 'nurzhan.prof@idsai.local','ACCEPTED')
  ) AS v(id, title, status, visibility, group_code, created_by_email, professor_email, professor_review_status)
)
INSERT INTO projects(
  id,
  tenant_id,
  title,
  description,
  status,
  is_public,
  created_by,
  professor_id,
  professor_review_status,
  professor_invited_at,
  professor_responded_at,
  faculty_id,
  visibility,
  group_id
)
SELECT
  ps.id,
  fc.tenant_id,
  ps.title,
  'Seeded demo project with realistic state for local development.',
  ps.status,
  (ps.visibility = 'PUBLIC'),
  creator.id,
  professor.id,
  ps.professor_review_status,
  CASE WHEN ps.professor_email IS NULL THEN NULL ELSE now() - interval '2 day' END,
  CASE WHEN ps.professor_review_status IN ('ACCEPTED', 'REJECTED') THEN now() - interval '1 day' ELSE NULL END,
  fc.faculty_id,
  ps.visibility,
  sg.id
FROM project_seed ps
JOIN faculty_ctx fc ON TRUE
JOIN users creator ON creator.email = ps.created_by_email
LEFT JOIN users professor ON professor.email = ps.professor_email
LEFT JOIN student_groups sg
  ON sg.faculty_id = fc.faculty_id
 AND sg.code = ps.group_code
ON CONFLICT (id) DO NOTHING;

WITH faculty_ctx AS (
  SELECT f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), project_leads AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000001'::uuid, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000002'::uuid, 'aliya.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'daniyar.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000005'::uuid, 'aliya.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000006'::uuid, 'daniyar.student@idsai.local')
  ) AS v(project_id, email)
)
INSERT INTO project_members(tenant_id, project_id, user_id, status, joined_at)
SELECT fc.tenant_id, pl.project_id, u.id, 'ACTIVE', now() - interval '5 day'
FROM project_leads pl
JOIN users u ON u.email = pl.email
JOIN faculty_ctx fc ON TRUE
ON CONFLICT (project_id, user_id) DO UPDATE
SET status = EXCLUDED.status,
    joined_at = COALESCE(project_members.joined_at, EXCLUDED.joined_at);

WITH faculty_ctx AS (
  SELECT f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), project_leads AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000001'::uuid, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000002'::uuid, 'aliya.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'daniyar.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000005'::uuid, 'aliya.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000006'::uuid, 'daniyar.student@idsai.local')
  ) AS v(project_id, email)
)
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT fc.tenant_id, u.id, r.id, 'PROJECT', pl.project_id
FROM project_leads pl
JOIN users u ON u.email = pl.email
JOIN roles r ON r.code = 'TEAM_LEAD'
JOIN faculty_ctx fc ON TRUE
WHERE NOT EXISTS (
  SELECT 1
  FROM role_assignments ra
  WHERE ra.user_id = u.id
    AND ra.role_id = r.id
    AND ra.scope_type = 'PROJECT'
    AND ra.scope_id = pl.project_id
    AND ra.expires_at IS NULL
);

WITH faculty_ctx AS (
  SELECT f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), project_professors AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000002'::uuid, 'nurzhan.prof@idsai.local'),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'aidana.prof@idsai.local'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'nurzhan.prof@idsai.local'),
    ('11000000-0000-0000-0000-000000000005'::uuid, 'aidana.prof@idsai.local'),
    ('11000000-0000-0000-0000-000000000006'::uuid, 'nurzhan.prof@idsai.local')
  ) AS v(project_id, email)
)
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT fc.tenant_id, u.id, r.id, 'PROJECT', pp.project_id
FROM project_professors pp
JOIN users u ON u.email = pp.email
JOIN roles r ON r.code = 'PROJECT_PROFESSOR'
JOIN faculty_ctx fc ON TRUE
WHERE NOT EXISTS (
  SELECT 1
  FROM role_assignments ra
  WHERE ra.user_id = u.id
    AND ra.role_id = r.id
    AND ra.scope_type = 'PROJECT'
    AND ra.scope_id = pp.project_id
    AND ra.expires_at IS NULL
);

WITH stack_seed AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000001'::uuid, 'GO'),
    ('11000000-0000-0000-0000-000000000001'::uuid, 'REACT'),
    ('11000000-0000-0000-0000-000000000001'::uuid, 'POSTGRESQL'),
    ('11000000-0000-0000-0000-000000000002'::uuid, 'PYTHON'),
    ('11000000-0000-0000-0000-000000000002'::uuid, 'FASTAPI'),
    ('11000000-0000-0000-0000-000000000002'::uuid, 'POSTGRESQL'),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'GO'),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'FLUTTER'),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'REDIS'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'TYPESCRIPT'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'NESTJS'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'POSTGRESQL'),
    ('11000000-0000-0000-0000-000000000005'::uuid, 'PYTHON'),
    ('11000000-0000-0000-0000-000000000005'::uuid, 'LLM'),
    ('11000000-0000-0000-0000-000000000005'::uuid, 'VECTORDB'),
    ('11000000-0000-0000-0000-000000000006'::uuid, 'JAVA'),
    ('11000000-0000-0000-0000-000000000006'::uuid, 'KAFKA'),
    ('11000000-0000-0000-0000-000000000006'::uuid, 'POSTGRESQL')
  ) AS v(project_id, stack_code)
)
INSERT INTO project_stacks(tenant_id, project_id, stack_code)
SELECT p.tenant_id, ss.project_id, ss.stack_code
FROM stack_seed ss
JOIN projects p ON p.id = ss.project_id
ON CONFLICT (project_id, stack_code) DO NOTHING;

WITH position_seed AS (
  SELECT *
  FROM (VALUES
    ('21000000-0000-0000-0000-000000000101'::uuid, '11000000-0000-0000-0000-000000000001'::uuid, 'BACKEND',     'Backend Developer', 2),
    ('21000000-0000-0000-0000-000000000102'::uuid, '11000000-0000-0000-0000-000000000001'::uuid, 'FRONTEND',    'Frontend Developer', 2),
    ('21000000-0000-0000-0000-000000000103'::uuid, '11000000-0000-0000-0000-000000000001'::uuid, 'UX',          'UX Designer', 1),
    ('21000000-0000-0000-0000-000000000201'::uuid, '11000000-0000-0000-0000-000000000002'::uuid, 'ML',          'ML Engineer', 2),
    ('21000000-0000-0000-0000-000000000202'::uuid, '11000000-0000-0000-0000-000000000002'::uuid, 'BACKEND',     'Backend Developer', 2),
    ('21000000-0000-0000-0000-000000000203'::uuid, '11000000-0000-0000-0000-000000000002'::uuid, 'ANALYST',     'Data Analyst', 1),
    ('21000000-0000-0000-0000-000000000301'::uuid, '11000000-0000-0000-0000-000000000003'::uuid, 'BACKEND',     'Backend Developer', 2),
    ('21000000-0000-0000-0000-000000000302'::uuid, '11000000-0000-0000-0000-000000000003'::uuid, 'MOBILE',      'Mobile Developer', 2),
    ('21000000-0000-0000-0000-000000000303'::uuid, '11000000-0000-0000-0000-000000000003'::uuid, 'QA',          'QA Engineer', 1),
    ('21000000-0000-0000-0000-000000000401'::uuid, '11000000-0000-0000-0000-000000000004'::uuid, 'BACKEND',     'Backend Developer', 2),
    ('21000000-0000-0000-0000-000000000402'::uuid, '11000000-0000-0000-0000-000000000004'::uuid, 'FRONTEND',    'Frontend Developer', 2),
    ('21000000-0000-0000-0000-000000000403'::uuid, '11000000-0000-0000-0000-000000000004'::uuid, 'DEVOPS',      'DevOps Engineer', 1),
    ('21000000-0000-0000-0000-000000000501'::uuid, '11000000-0000-0000-0000-000000000005'::uuid, 'ML',          'ML Engineer', 2),
    ('21000000-0000-0000-0000-000000000502'::uuid, '11000000-0000-0000-0000-000000000005'::uuid, 'DATA',        'Data Engineer', 2),
    ('21000000-0000-0000-0000-000000000503'::uuid, '11000000-0000-0000-0000-000000000005'::uuid, 'BACKEND',     'Backend Developer', 1),
    ('21000000-0000-0000-0000-000000000601'::uuid, '11000000-0000-0000-0000-000000000006'::uuid, 'BACKEND',     'Backend Developer', 2),
    ('21000000-0000-0000-0000-000000000602'::uuid, '11000000-0000-0000-0000-000000000006'::uuid, 'INTEGRATION', 'Integration Engineer', 1),
    ('21000000-0000-0000-0000-000000000603'::uuid, '11000000-0000-0000-0000-000000000006'::uuid, 'QA',          'QA Engineer', 1)
  ) AS v(id, project_id, code, name, capacity)
)
INSERT INTO project_positions(id, tenant_id, project_id, code, name, capacity, created_at)
SELECT ps.id, p.tenant_id, ps.project_id, ps.code, ps.name, ps.capacity, now() - interval '6 day'
FROM position_seed ps
JOIN projects p ON p.id = ps.project_id
ON CONFLICT (id) DO NOTHING;

WITH member_seed AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000001'::uuid, 'dinara.student@idsai.local',  '21000000-0000-0000-0000-000000000102'::uuid, 5, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000001'::uuid, 'marat.student@idsai.local',   '21000000-0000-0000-0000-000000000101'::uuid, 4, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000002'::uuid, 'marat.student@idsai.local',   '21000000-0000-0000-0000-000000000203'::uuid, 5, 'aliya.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'aliya.student@idsai.local',   '21000000-0000-0000-0000-000000000302'::uuid, 4, 'daniyar.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'daniyar.student@idsai.local', '21000000-0000-0000-0000-000000000403'::uuid, 6, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'marat.student@idsai.local',   '21000000-0000-0000-0000-000000000401'::uuid, 6, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000004'::uuid, 'dinara.student@idsai.local',  '21000000-0000-0000-0000-000000000402'::uuid, 5, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000005'::uuid, 'aibolat.student@idsai.local', '21000000-0000-0000-0000-000000000503'::uuid, 4, 'aliya.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000005'::uuid, 'marat.student@idsai.local',   '21000000-0000-0000-0000-000000000502'::uuid, 4, 'aliya.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000006'::uuid, 'aibolat.student@idsai.local', '21000000-0000-0000-0000-000000000601'::uuid, 7, 'daniyar.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000006'::uuid, 'aliya.student@idsai.local',   '21000000-0000-0000-0000-000000000603'::uuid, 7, 'daniyar.student@idsai.local')
  ) AS v(project_id, email, position_id, joined_days_ago, invited_by_email)
)
INSERT INTO project_members(tenant_id, project_id, user_id, position_id, status, joined_at, invite_comment, invited_by, responded_at)
SELECT
  p.tenant_id,
  ms.project_id,
  u.id,
  ms.position_id,
  'ACTIVE',
  now() - (ms.joined_days_ago || ' day')::interval,
  '',
  inviter.id,
  now() - interval '2 day'
FROM member_seed ms
JOIN projects p ON p.id = ms.project_id
JOIN users u ON u.email = ms.email
LEFT JOIN users inviter ON inviter.email = ms.invited_by_email
ON CONFLICT (project_id, user_id) DO UPDATE
SET status = EXCLUDED.status,
    position_id = EXCLUDED.position_id,
    joined_at = COALESCE(project_members.joined_at, EXCLUDED.joined_at),
    invited_by = EXCLUDED.invited_by,
    responded_at = COALESCE(project_members.responded_at, EXCLUDED.responded_at);

WITH invite_seed AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000001'::uuid, 'daniyar.student@idsai.local', 'INVITED', 'Нужен QA-инженер для MVP', 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000002'::uuid, 'dinara.student@idsai.local',  'INVITED', 'Подключайся к аналитике расписаний', 'aliya.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000002'::uuid, 'aibolat.student@idsai.local', 'APPLIED', 'Хочу помочь с архитектурой backend', NULL),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'marat.student@idsai.local',   'INVITED', 'Нужен опыт с системной интеграцией', 'daniyar.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000003'::uuid, 'dinara.student@idsai.local',  'APPLIED', 'Готова взять мобильную часть', NULL)
  ) AS v(project_id, email, status, invite_comment, invited_by_email)
)
INSERT INTO project_members(tenant_id, project_id, user_id, status, invite_comment, invited_by, responded_at)
SELECT
  p.tenant_id,
  isd.project_id,
  u.id,
  isd.status,
  isd.invite_comment,
  inviter.id,
  CASE WHEN isd.status = 'APPLIED' THEN now() - interval '12 hour' ELSE NULL END
FROM invite_seed isd
JOIN projects p ON p.id = isd.project_id
JOIN users u ON u.email = isd.email
LEFT JOIN users inviter ON inviter.email = isd.invited_by_email
ON CONFLICT (project_id, user_id) DO UPDATE
SET status = EXCLUDED.status,
    invite_comment = EXCLUDED.invite_comment,
    invited_by = EXCLUDED.invited_by,
    responded_at = EXCLUDED.responded_at;

INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT p.tenant_id, pm.user_id, r.id, 'PROJECT', pm.project_id
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
JOIN roles r ON r.code = 'MEMBER'
WHERE pm.status = 'ACTIVE'
  AND pm.user_id <> p.created_by
  AND NOT EXISTS (
    SELECT 1
    FROM role_assignments ra
    WHERE ra.user_id = pm.user_id
      AND ra.role_id = r.id
      AND ra.scope_type = 'PROJECT'
      AND ra.scope_id = pm.project_id
      AND ra.expires_at IS NULL
  );

INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT p.tenant_id, pm.user_id, r.id, 'PROJECT', pm.project_id
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
JOIN roles r ON r.code = 'INVITED_MEMBER'
WHERE pm.status = 'INVITED'
  AND NOT EXISTS (
    SELECT 1
    FROM role_assignments ra
    WHERE ra.user_id = pm.user_id
      AND ra.role_id = r.id
      AND ra.scope_type = 'PROJECT'
      AND ra.scope_id = pm.project_id
      AND ra.expires_at IS NULL
  );

WITH task_seed AS (
  SELECT *
  FROM (VALUES
    ('22000000-0000-0000-0000-000000000101'::uuid, '11000000-0000-0000-0000-000000000001'::uuid, 'Собрать требования по пилоту',          'Сформировать customer journey и требования к MVP.',            '21000000-0000-0000-0000-000000000103'::uuid, 'aibolat.student@idsai.local', 'OPEN',        'aibolat.student@idsai.local', 2),
    ('22000000-0000-0000-0000-000000000201'::uuid, '11000000-0000-0000-0000-000000000002'::uuid, 'Проверить модель ограничений',         'Валидировать модели оптимизации расписаний на исторических данных.', '21000000-0000-0000-0000-000000000201'::uuid, 'aliya.student@idsai.local',   'IN_PROGRESS', 'aliya.student@idsai.local',   1),
    ('22000000-0000-0000-0000-000000000301'::uuid, '11000000-0000-0000-0000-000000000003'::uuid, 'Прототип secure token flow',           'Сделать прототип цепочки регистрации и выдачи токена.',      '21000000-0000-0000-0000-000000000301'::uuid, 'daniyar.student@idsai.local', 'OPEN',        'daniyar.student@idsai.local', 3),
    ('22000000-0000-0000-0000-000000000302'::uuid, '11000000-0000-0000-0000-000000000003'::uuid, 'Экран мобильной верификации',          'Собрать flow мобильной верификации студента.',                '21000000-0000-0000-0000-000000000302'::uuid, 'aliya.student@idsai.local',   'IN_PROGRESS', 'daniyar.student@idsai.local', 4),
    ('22000000-0000-0000-0000-000000000401'::uuid, '11000000-0000-0000-0000-000000000004'::uuid, 'Refactor API gateway',                'Рефакторинг API gateway и нормализация ответов.',             '21000000-0000-0000-0000-000000000401'::uuid, 'marat.student@idsai.local',   'IN_PROGRESS', 'aibolat.student@idsai.local', 2),
    ('22000000-0000-0000-0000-000000000402'::uuid, '11000000-0000-0000-0000-000000000004'::uuid, 'UI экран дашборда деканата',           'Собрать production-ready макет дашборда и таблиц.',           '21000000-0000-0000-0000-000000000402'::uuid, 'dinara.student@idsai.local',  'OPEN',        'aibolat.student@idsai.local', 5),
    ('22000000-0000-0000-0000-000000000403'::uuid, '11000000-0000-0000-0000-000000000004'::uuid, 'Hardening CI/CD пайплайна',            'Добавить quality gates и уведомления по релизам.',            '21000000-0000-0000-0000-000000000403'::uuid, 'daniyar.student@idsai.local', 'DONE',        'aibolat.student@idsai.local', -1),
    ('22000000-0000-0000-0000-000000000501'::uuid, '11000000-0000-0000-0000-000000000005'::uuid, 'Benchmark prompt-стратегий',           'Сравнить baseline и chain-of-thought по quality метрикам.',   '21000000-0000-0000-0000-000000000501'::uuid, 'aliya.student@idsai.local',   'DONE',        'aliya.student@idsai.local',   -2),
    ('22000000-0000-0000-0000-000000000502'::uuid, '11000000-0000-0000-0000-000000000005'::uuid, 'Pipeline разметки датасета',           'Стабилизировать data labeling pipeline и валидацию.',         '21000000-0000-0000-0000-000000000502'::uuid, 'marat.student@idsai.local',   'DONE',        'aliya.student@idsai.local',   -1),
    ('22000000-0000-0000-0000-000000000601'::uuid, '11000000-0000-0000-0000-000000000006'::uuid, 'Smoke test legacy-синхронизации',      'Финальный smoke test интеграции ERP-коннектора.',             '21000000-0000-0000-0000-000000000603'::uuid, 'aliya.student@idsai.local',   'DONE',        'daniyar.student@idsai.local', -10)
  ) AS v(id, project_id, title, description, position_id, assignee_email, status, created_by_email, due_days_offset)
)
INSERT INTO tasks(id, tenant_id, project_id, title, description, position_id, assignee_user_id, status, created_by, due_at, created_at, updated_at)
SELECT
  ts.id,
  p.tenant_id,
  ts.project_id,
  ts.title,
  ts.description,
  ts.position_id,
  assignee.id,
  ts.status,
  creator.id,
  now() + (ts.due_days_offset || ' day')::interval,
  now() - interval '4 day',
  now() - interval '1 day'
FROM task_seed ts
JOIN projects p ON p.id = ts.project_id
JOIN users creator ON creator.email = ts.created_by_email
LEFT JOIN users assignee ON assignee.email = ts.assignee_email
ON CONFLICT (id) DO NOTHING;

WITH submission_seed AS (
  SELECT *
  FROM (VALUES
    ('22000000-0000-0000-0000-000000000403'::uuid, 'daniyar.student@idsai.local', 'CI/CD hardening завершён, pipeline стабилен.', '["https://git.example.local/ci/report-403"]'),
    ('22000000-0000-0000-0000-000000000501'::uuid, 'aliya.student@idsai.local',   'Подготовлены метрики и графики по prompt benchmark.', '["https://git.example.local/ml/benchmark-501"]'),
    ('22000000-0000-0000-0000-000000000502'::uuid, 'marat.student@idsai.local',   'Pipeline разметки выведен в стабильный режим.', '["https://git.example.local/data/pipeline-502"]'),
    ('22000000-0000-0000-0000-000000000601'::uuid, 'aliya.student@idsai.local',   'Smoke test пройден, интеграция стабильна.', '["https://git.example.local/legacy/smoke-601"]')
  ) AS v(task_id, user_email, comment, attachments_json)
)
INSERT INTO task_submissions(tenant_id, project_id, task_id, user_id, comment, attachments, submitted_at, updated_at)
SELECT
  p.tenant_id,
  t.project_id,
  t.id,
  u.id,
  ss.comment,
  ss.attachments_json::jsonb,
  now() - interval '12 hour',
  now() - interval '12 hour'
FROM submission_seed ss
JOIN tasks t ON t.id = ss.task_id
JOIN projects p ON p.id = t.project_id
JOIN users u ON u.email = ss.user_email
ON CONFLICT (task_id) DO UPDATE
SET comment = EXCLUDED.comment,
    attachments = EXCLUDED.attachments,
    updated_at = EXCLUDED.updated_at;

WITH activity_seed AS (
  SELECT *
  FROM (VALUES
    ('22000000-0000-0000-0000-000000000401'::uuid, 'aibolat.student@idsai.local', 'CREATED',        NULL::text,       NULL::text,      'Создана задача', 'Новый backend блок.', '[]', 96),
    ('22000000-0000-0000-0000-000000000401'::uuid, 'marat.student@idsai.local',   'STATUS_CHANGED', 'OPEN',           'IN_PROGRESS',   'Задача взята в работу', 'Начал реализацию.', '[]', 48),
    ('22000000-0000-0000-0000-000000000403'::uuid, 'daniyar.student@idsai.local', 'COMPLETED',      'IN_PROGRESS',    'DONE',          'Задача завершена', 'Пайплайн стабилизирован.', '["https://git.example.local/ci/report-403"]', 16),
    ('22000000-0000-0000-0000-000000000501'::uuid, 'aliya.student@idsai.local',   'COMPLETED',      'IN_PROGRESS',    'DONE',          'Задача завершена', 'Бенчмарк готов к ревью.', '["https://git.example.local/ml/benchmark-501"]', 14),
    ('22000000-0000-0000-0000-000000000502'::uuid, 'marat.student@idsai.local',   'COMPLETED',      'IN_PROGRESS',    'DONE',          'Задача завершена', 'Data pipeline стабилен.', '["https://git.example.local/data/pipeline-502"]', 12),
    ('22000000-0000-0000-0000-000000000601'::uuid, 'aliya.student@idsai.local',   'COMPLETED',      'IN_PROGRESS',    'DONE',          'Задача завершена', 'Smoke test отчёт приложен.', '["https://git.example.local/legacy/smoke-601"]', 10)
  ) AS v(task_id, actor_email, event_type, from_status, to_status, title, comment, attachments_json, hours_ago)
)
INSERT INTO task_activity_logs(tenant_id, project_id, task_id, actor_user_id, event_type, from_status, to_status, title, comment, attachments, created_at)
SELECT
  p.tenant_id,
  t.project_id,
  t.id,
  u.id,
  ac.event_type,
  ac.from_status,
  ac.to_status,
  ac.title,
  ac.comment,
  ac.attachments_json::jsonb,
  now() - (ac.hours_ago || ' hour')::interval
FROM activity_seed ac
JOIN tasks t ON t.id = ac.task_id
JOIN projects p ON p.id = t.project_id
LEFT JOIN users u ON u.email = ac.actor_email;

WITH criteria_seed AS (
  SELECT *
  FROM (VALUES
    ('31000000-0000-0000-0000-000000000501'::uuid, '11000000-0000-0000-0000-000000000005'::uuid, 'Качество итогового решения',  'Насколько решение соответствует целям проекта.', 40, 'aidana.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000502'::uuid, '11000000-0000-0000-0000-000000000005'::uuid, 'Техническая реализация',      'Архитектура, стабильность и производительность.', 35, 'aidana.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000503'::uuid, '11000000-0000-0000-0000-000000000005'::uuid, 'Документация и презентация',  'Полнота документации и ясность демонстрации.', 25, 'aidana.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000601'::uuid, '11000000-0000-0000-0000-000000000006'::uuid, 'Стабильность интеграции',      'Надежность работы под нагрузкой.', 45, 'nurzhan.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000602'::uuid, '11000000-0000-0000-0000-000000000006'::uuid, 'Код и архитектура',           'Поддерживаемость и модульность решения.', 35, 'nurzhan.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000603'::uuid, '11000000-0000-0000-0000-000000000006'::uuid, 'Отчёт и защита',              'Качество отчета и презентации результата.', 20, 'nurzhan.prof@idsai.local')
  ) AS v(id, project_id, title, description, weight, created_by_email)
)
INSERT INTO project_criteria(id, tenant_id, project_id, title, description, weight, created_by, created_at)
SELECT
  cs.id,
  p.tenant_id,
  cs.project_id,
  cs.title,
  cs.description,
  cs.weight,
  u.id,
  now() - interval '3 day'
FROM criteria_seed cs
JOIN projects p ON p.id = cs.project_id
JOIN users u ON u.email = cs.created_by_email
ON CONFLICT (id) DO NOTHING;

WITH review_seed AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000005'::uuid, '31000000-0000-0000-0000-000000000501'::uuid, 'aidana.prof@idsai.local', TRUE,  'Решение хорошо покрывает пользовательские сценарии.'),
    ('11000000-0000-0000-0000-000000000005'::uuid, '31000000-0000-0000-0000-000000000502'::uuid, 'aidana.prof@idsai.local', TRUE,  'Архитектура стабильная, есть незначительные замечания.'),
    ('11000000-0000-0000-0000-000000000005'::uuid, '31000000-0000-0000-0000-000000000503'::uuid, 'aidana.prof@idsai.local', FALSE, 'Нужно усилить демонстрационную часть презентации.'),
    ('11000000-0000-0000-0000-000000000006'::uuid, '31000000-0000-0000-0000-000000000601'::uuid, 'nurzhan.prof@idsai.local', TRUE,  'Интеграция показала высокую стабильность.'),
    ('11000000-0000-0000-0000-000000000006'::uuid, '31000000-0000-0000-0000-000000000602'::uuid, 'nurzhan.prof@idsai.local', TRUE,  'Код чистый и поддерживаемый.'),
    ('11000000-0000-0000-0000-000000000006'::uuid, '31000000-0000-0000-0000-000000000603'::uuid, 'nurzhan.prof@idsai.local', TRUE,  'Отчет и защита выполнены на высоком уровне.')
  ) AS v(project_id, criterion_id, professor_email, is_met, comment)
)
INSERT INTO project_criterion_reviews(tenant_id, project_id, criterion_id, professor_id, is_met, comment, created_at, updated_at)
SELECT
  p.tenant_id,
  rs.project_id,
  rs.criterion_id,
  u.id,
  rs.is_met,
  rs.comment,
  now() - interval '1 day',
  now() - interval '1 day'
FROM review_seed rs
JOIN projects p ON p.id = rs.project_id
JOIN users u ON u.email = rs.professor_email
ON CONFLICT (project_id, criterion_id, professor_id) DO UPDATE
SET is_met = EXCLUDED.is_met,
    comment = EXCLUDED.comment,
    updated_at = EXCLUDED.updated_at;

-- +goose Down

DELETE FROM role_assignments
WHERE scope_type = 'PROJECT'
  AND scope_id IN (
    '11000000-0000-0000-0000-000000000001',
    '11000000-0000-0000-0000-000000000002',
    '11000000-0000-0000-0000-000000000003',
    '11000000-0000-0000-0000-000000000004',
    '11000000-0000-0000-0000-000000000005',
    '11000000-0000-0000-0000-000000000006'
  );

DELETE FROM projects
WHERE id IN (
  '11000000-0000-0000-0000-000000000001',
  '11000000-0000-0000-0000-000000000002',
  '11000000-0000-0000-0000-000000000003',
  '11000000-0000-0000-0000-000000000004',
  '11000000-0000-0000-0000-000000000005',
  '11000000-0000-0000-0000-000000000006'
);

DELETE FROM role_assignments
WHERE user_id IN (
  SELECT id
  FROM users
  WHERE email IN (
    'aibolat.student@idsai.local',
    'aliya.student@idsai.local',
    'daniyar.student@idsai.local',
    'dinara.student@idsai.local',
    'marat.student@idsai.local',
    'nurzhan.prof@idsai.local',
    'aidana.prof@idsai.local',
    'sabina.prof@idsai.local'
  )
)
AND scope_type = 'FACULTY';

DELETE FROM user_profiles
WHERE user_id IN (
  SELECT id
  FROM users
  WHERE email IN (
    'aibolat.student@idsai.local',
    'aliya.student@idsai.local',
    'daniyar.student@idsai.local',
    'dinara.student@idsai.local',
    'marat.student@idsai.local',
    'nurzhan.prof@idsai.local',
    'aidana.prof@idsai.local',
    'sabina.prof@idsai.local'
  )
);

DELETE FROM users
WHERE email IN (
  'aibolat.student@idsai.local',
  'aliya.student@idsai.local',
  'daniyar.student@idsai.local',
  'dinara.student@idsai.local',
  'marat.student@idsai.local',
  'nurzhan.prof@idsai.local',
  'aidana.prof@idsai.local',
  'sabina.prof@idsai.local'
);
