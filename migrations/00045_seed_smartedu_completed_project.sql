-- +goose Up

-- Demo credentials for this seeded scenario:
-- password: DemoPass123!

WITH faculty_ctx AS (
  SELECT f.id AS faculty_id, f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), department_ctx AS (
  SELECT d.id AS department_id
  FROM departments d
  JOIN faculty_ctx fc ON fc.faculty_id = d.faculty_id
  WHERE d.code = 'CPI'
  LIMIT 1
)
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
  fc.tenant_id,
  fc.faculty_id,
  dc.department_id,
  'CS-21-07',
  'CS-21-07',
  4,
  'CS-21-07',
  2107,
  now() - interval '45 day',
  now()
FROM faculty_ctx fc
JOIN department_ctx dc ON TRUE
ON CONFLICT (faculty_id, code) DO UPDATE
SET department_id = EXCLUDED.department_id,
    name = EXCLUDED.name,
    year = EXCLUDED.year,
    group_code = EXCLUDED.group_code,
    group_number = EXCLUDED.group_number,
    updated_at = EXCLUDED.updated_at,
    tenant_id = EXCLUDED.tenant_id;

WITH faculty_ctx AS (
  SELECT f.id AS faculty_id, f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), seed_users AS (
  SELECT *
  FROM (VALUES
    ('aibolat.smartedu@idsai.local',   'Айболат Ермекбай',      'CPI', 'STUDENT'),
    ('dana.smartedu@idsai.local',      'Дана Сериккызы',        'CPI', 'STUDENT'),
    ('nursultan.smartedu@idsai.local', 'Нұрсұлтан Тлеуберген',  'CPI', 'STUDENT'),
    ('aruzhan.smartedu@idsai.local',   'Аружан Бекенова',       'CPI', 'STUDENT'),
    ('madiyar.smartedu@idsai.local',   'Мадияр Сапар',          'CPI', 'STUDENT'),
    ('asylkhan.prof@idsai.local',      'Асылхан Жумабеков',     'CPI', 'PROFESSOR')
  ) AS v(email, full_name, department_code, role_code)
)
INSERT INTO users(tenant_id, email, password_hash, status, email_verified_at, password_changed_at)
SELECT
  fc.tenant_id,
  su.email,
  '$2a$10$Gsnz9k1a6acT.eGCCMDOwudDgrTBCGxM//7fI4PZzc0oYPyTNAokG',
  'ACTIVE',
  now() - interval '45 day',
  now() - interval '45 day'
FROM seed_users su
CROSS JOIN faculty_ctx fc
ON CONFLICT (email) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    password_hash = EXCLUDED.password_hash,
    status = EXCLUDED.status,
    email_verified_at = COALESCE(users.email_verified_at, EXCLUDED.email_verified_at),
    password_changed_at = EXCLUDED.password_changed_at;

WITH faculty_ctx AS (
  SELECT f.id AS faculty_id, f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), group_ctx AS (
  SELECT sg.id AS group_id
  FROM student_groups sg
  JOIN faculty_ctx fc ON fc.faculty_id = sg.faculty_id
  WHERE sg.code = 'CS-21-07'
  LIMIT 1
), seed_users AS (
  SELECT *
  FROM (VALUES
    ('aibolat.smartedu@idsai.local',   'Айболат Ермекбай',      'CPI', 'STUDENT'),
    ('dana.smartedu@idsai.local',      'Дана Сериккызы',        'CPI', 'STUDENT'),
    ('nursultan.smartedu@idsai.local', 'Нұрсұлтан Тлеуберген',  'CPI', 'STUDENT'),
    ('aruzhan.smartedu@idsai.local',   'Аружан Бекенова',       'CPI', 'STUDENT'),
    ('madiyar.smartedu@idsai.local',   'Мадияр Сапар',          'CPI', 'STUDENT'),
    ('asylkhan.prof@idsai.local',      'Асылхан Жумабеков',     'CPI', 'PROFESSOR')
  ) AS v(email, full_name, department_code, role_code)
)
INSERT INTO user_profiles(tenant_id, user_id, full_name, faculty_id, department_id, group_id)
SELECT
  fc.tenant_id,
  u.id,
  su.full_name,
  fc.faculty_id,
  d.id,
  CASE WHEN su.role_code = 'STUDENT' THEN gc.group_id ELSE NULL END
FROM seed_users su
JOIN users u ON u.email = su.email
JOIN faculty_ctx fc ON TRUE
JOIN departments d ON d.faculty_id = fc.faculty_id AND d.code = su.department_code
LEFT JOIN group_ctx gc ON TRUE
ON CONFLICT (user_id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    full_name = EXCLUDED.full_name,
    faculty_id = EXCLUDED.faculty_id,
    department_id = EXCLUDED.department_id,
    group_id = EXCLUDED.group_id;

WITH faculty_ctx AS (
  SELECT f.id AS faculty_id, f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), seed_users AS (
  SELECT *
  FROM (VALUES
    ('aibolat.smartedu@idsai.local',   'STUDENT'),
    ('dana.smartedu@idsai.local',      'STUDENT'),
    ('nursultan.smartedu@idsai.local', 'STUDENT'),
    ('aruzhan.smartedu@idsai.local',   'STUDENT'),
    ('madiyar.smartedu@idsai.local',   'STUDENT'),
    ('asylkhan.prof@idsai.local',      'PROFESSOR')
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
  WHERE ra.tenant_id = fc.tenant_id
    AND ra.user_id = u.id
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
), group_ctx AS (
  SELECT sg.id AS group_id
  FROM student_groups sg
  JOIN faculty_ctx fc ON fc.faculty_id = sg.faculty_id
  WHERE sg.code = 'CS-21-07'
  LIMIT 1
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
  group_id,
  retake_count,
  default_cover_variant,
  created_at,
  updated_at
)
SELECT
  '11000000-0000-0000-0000-000000000011'::uuid,
  fc.tenant_id,
  'SmartEdu Analytics',
  'Интеллектуальная платформа раннего выявления академических рисков студентов. Система анализирует посещаемость, сроки сдачи задач, учебную активность и формирует аналитическую панель для преподавателя и команды проекта.',
  'COMPLETED',
  FALSE,
  lead_user.id,
  professor.id,
  'ACCEPTED',
  now() - interval '12 day',
  now() - interval '11 day',
  fc.faculty_id,
  'GROUP',
  gc.group_id,
  0,
  4,
  now() - interval '16 day',
  now() - interval '2 hour'
FROM faculty_ctx fc
JOIN group_ctx gc ON TRUE
JOIN users lead_user ON lead_user.email = 'aibolat.smartedu@idsai.local'
JOIN users professor ON professor.email = 'asylkhan.prof@idsai.local'
ON CONFLICT (id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    status = EXCLUDED.status,
    is_public = EXCLUDED.is_public,
    created_by = EXCLUDED.created_by,
    professor_id = EXCLUDED.professor_id,
    professor_review_status = EXCLUDED.professor_review_status,
    professor_invited_at = EXCLUDED.professor_invited_at,
    professor_responded_at = EXCLUDED.professor_responded_at,
    faculty_id = EXCLUDED.faculty_id,
    visibility = EXCLUDED.visibility,
    group_id = EXCLUDED.group_id,
    retake_count = EXCLUDED.retake_count,
    default_cover_variant = EXCLUDED.default_cover_variant,
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at;

WITH stack_seed AS (
  SELECT *
  FROM (VALUES
    ('GO'),
    ('GIN'),
    ('POSTGRESQL'),
    ('REDIS'),
    ('DOCKER'),
    ('VANILLA_JS'),
    ('CHART_JS')
  ) AS v(stack_code)
)
INSERT INTO project_stacks(tenant_id, project_id, stack_code)
SELECT
  p.tenant_id,
  p.id,
  ss.stack_code
FROM stack_seed ss
JOIN projects p ON p.id = '11000000-0000-0000-0000-000000000011'::uuid
ON CONFLICT (project_id, stack_code) DO NOTHING;

WITH position_seed AS (
  SELECT *
  FROM (VALUES
    ('21000000-0000-0000-0000-000000001101'::uuid, 'ARCHITECT',  'Backend Engineer / Solution Architect', 1),
    ('21000000-0000-0000-0000-000000001102'::uuid, 'FRONTEND',   'Frontend Engineer / UI Designer',       1),
    ('21000000-0000-0000-0000-000000001103'::uuid, 'DEVOPS',     'DevOps Engineer / Task Coordinator',    1),
    ('21000000-0000-0000-0000-000000001104'::uuid, 'QA_ANALYST', 'QA Engineer / Business Analyst',        1),
    ('21000000-0000-0000-0000-000000001105'::uuid, 'DATA',       'Data Analyst / ML Assistant',           1)
  ) AS v(id, code, name, capacity)
)
INSERT INTO project_positions(id, tenant_id, project_id, code, name, capacity, created_at)
SELECT
  ps.id,
  p.tenant_id,
  p.id,
  ps.code,
  ps.name,
  ps.capacity,
  now() - interval '15 day'
FROM position_seed ps
JOIN projects p ON p.id = '11000000-0000-0000-0000-000000000011'::uuid
ON CONFLICT (id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    project_id = EXCLUDED.project_id,
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    capacity = EXCLUDED.capacity;

WITH member_seed AS (
  SELECT *
  FROM (VALUES
    ('aibolat.smartedu@idsai.local',   '21000000-0000-0000-0000-000000001101'::uuid, 'ACTIVE', NULL::text,                        NULL::text,                     15),
    ('dana.smartedu@idsai.local',      '21000000-0000-0000-0000-000000001102'::uuid, 'ACTIVE', 'Айболат подключил к роли CO_LEAD', 'aibolat.smartedu@idsai.local', 14),
    ('nursultan.smartedu@idsai.local', '21000000-0000-0000-0000-000000001103'::uuid, 'ACTIVE', 'Назначен менеджером задач проекта', 'aibolat.smartedu@idsai.local', 14),
    ('aruzhan.smartedu@idsai.local',   '21000000-0000-0000-0000-000000001104'::uuid, 'ACTIVE', 'Отвечает за QA и приемку',          'aibolat.smartedu@idsai.local', 13),
    ('madiyar.smartedu@idsai.local',   '21000000-0000-0000-0000-000000001105'::uuid, 'ACTIVE', 'Ведет аналитический блок данных',   'aibolat.smartedu@idsai.local', 13)
  ) AS v(email, position_id, status, invite_comment, invited_by_email, joined_days_ago)
)
INSERT INTO project_members(tenant_id, project_id, user_id, position_id, status, joined_at, invite_comment, invited_by, responded_at, created_at)
SELECT
  p.tenant_id,
  p.id,
  u.id,
  ms.position_id,
  ms.status,
  now() - (ms.joined_days_ago || ' day')::interval,
  COALESCE(ms.invite_comment, ''),
  inviter.id,
  CASE WHEN ms.invited_by_email IS NULL THEN now() - interval '15 day' ELSE now() - interval '13 day' END,
  now() - interval '15 day'
FROM member_seed ms
JOIN projects p ON p.id = '11000000-0000-0000-0000-000000000011'::uuid
JOIN users u ON u.email = ms.email
LEFT JOIN users inviter ON inviter.email = ms.invited_by_email
ON CONFLICT (project_id, user_id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    position_id = EXCLUDED.position_id,
    status = EXCLUDED.status,
    joined_at = EXCLUDED.joined_at,
    invite_comment = EXCLUDED.invite_comment,
    invited_by = EXCLUDED.invited_by,
    responded_at = EXCLUDED.responded_at;

INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT
  p.tenant_id,
  u.id,
  r.id,
  'PROJECT',
  p.id
FROM projects p
JOIN users u ON u.email = 'aibolat.smartedu@idsai.local'
JOIN roles r ON r.code = 'TEAM_LEAD'
WHERE p.id = '11000000-0000-0000-0000-000000000011'::uuid
  AND NOT EXISTS (
    SELECT 1
    FROM role_assignments ra
    WHERE ra.tenant_id = p.tenant_id
      AND ra.user_id = u.id
      AND ra.role_id = r.id
      AND ra.scope_type = 'PROJECT'
      AND ra.scope_id = p.id
      AND ra.expires_at IS NULL
  );

INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT
  p.tenant_id,
  u.id,
  r.id,
  'PROJECT',
  p.id
FROM projects p
JOIN users u ON u.email = 'asylkhan.prof@idsai.local'
JOIN roles r ON r.code = 'PROJECT_PROFESSOR'
WHERE p.id = '11000000-0000-0000-0000-000000000011'::uuid
  AND NOT EXISTS (
    SELECT 1
    FROM role_assignments ra
    WHERE ra.tenant_id = p.tenant_id
      AND ra.user_id = u.id
      AND ra.role_id = r.id
      AND ra.scope_type = 'PROJECT'
      AND ra.scope_id = p.id
      AND ra.expires_at IS NULL
  );

INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT
  p.tenant_id,
  pm.user_id,
  r.id,
  'PROJECT',
  p.id
FROM projects p
JOIN project_members pm ON pm.project_id = p.id
JOIN roles r ON r.code = 'MEMBER'
WHERE p.id = '11000000-0000-0000-0000-000000000011'::uuid
  AND pm.status = 'ACTIVE'
  AND pm.user_id <> p.created_by
  AND NOT EXISTS (
    SELECT 1
    FROM role_assignments ra
    WHERE ra.tenant_id = p.tenant_id
      AND ra.user_id = pm.user_id
      AND ra.role_id = r.id
      AND ra.scope_type = 'PROJECT'
      AND ra.scope_id = p.id
      AND ra.expires_at IS NULL
  );

WITH delegated_seed AS (
  SELECT *
  FROM (VALUES
    ('dana.smartedu@idsai.local',      'CO_LEAD'),
    ('nursultan.smartedu@idsai.local', 'TASK_MANAGER'),
    ('aruzhan.smartedu@idsai.local',   'RECRUITER')
  ) AS v(email, role_code)
)
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT
  p.tenant_id,
  u.id,
  r.id,
  'PROJECT',
  p.id
FROM delegated_seed ds
JOIN projects p ON p.id = '11000000-0000-0000-0000-000000000011'::uuid
JOIN users u ON u.email = ds.email
JOIN roles r ON r.code = ds.role_code
WHERE NOT EXISTS (
  SELECT 1
  FROM role_assignments ra
  WHERE ra.tenant_id = p.tenant_id
    AND ra.user_id = u.id
    AND ra.role_id = r.id
    AND ra.scope_type = 'PROJECT'
    AND ra.scope_id = p.id
    AND ra.expires_at IS NULL
);

WITH task_seed AS (
  SELECT *
  FROM (VALUES
    ('22000000-0000-0000-0000-000000001101'::uuid, 'Спроектировать ER-модель проекта и доступа',            'Подготовить структуру данных проекта, ролей, прав доступа и связей между сущностями.',                          '21000000-0000-0000-0000-000000001101'::uuid, 'aibolat.smartedu@idsai.local',   'DONE', 'aibolat.smartedu@idsai.local',   -14, 72),
    ('22000000-0000-0000-0000-000000001102'::uuid, 'Реализовать scoped RBAC для проектного модуля',         'Поднять project-scope роли, проверки permission и цепочку авторизации для командных действий.',                 '21000000-0000-0000-0000-000000001101'::uuid, 'aibolat.smartedu@idsai.local',   'DONE', 'aibolat.smartedu@idsai.local',   -13, 68),
    ('22000000-0000-0000-0000-000000001103'::uuid, 'Собрать страницу проекта и рабочее пространство команды','Сверстать главную страницу проекта с описанием, метаданными, составом команды и статусными блоками.',           '21000000-0000-0000-0000-000000001102'::uuid, 'dana.smartedu@idsai.local',      'DONE', 'aibolat.smartedu@idsai.local',   -11, 60),
    ('22000000-0000-0000-0000-000000001104'::uuid, 'Реализовать отображение ролей и карточек участников',   'Показать в UI роли TEAM_LEAD, MEMBER и delegated access roles для каждого участника команды.',                  '21000000-0000-0000-0000-000000001102'::uuid, 'dana.smartedu@idsai.local',      'DONE', 'dana.smartedu@idsai.local',      -10, 58),
    ('22000000-0000-0000-0000-000000001105'::uuid, 'Настроить task board и переходы статусов',               'Добавить создание задач, назначение исполнителей, перевод в IN_PROGRESS и завершение с журналом действий.',     '21000000-0000-0000-0000-000000001103'::uuid, 'nursultan.smartedu@idsai.local', 'DONE', 'nursultan.smartedu@idsai.local',  -9,  54),
    ('22000000-0000-0000-0000-000000001106'::uuid, 'Подготовить Docker-инфраструктуру проекта',              'Настроить локальный compose-контур приложения, БД, кэша и вспомогательных сервисов для стабильного demo.',      '21000000-0000-0000-0000-000000001103'::uuid, 'nursultan.smartedu@idsai.local', 'DONE', 'aibolat.smartedu@idsai.local',   -8,  50),
    ('22000000-0000-0000-0000-000000001107'::uuid, 'Составить тест-кейсы и сценарии приемки',                'Подготовить QA-чек-лист по ролям, проектному флоу, критериям и жизненному циклу проекта.',                      '21000000-0000-0000-0000-000000001104'::uuid, 'aruzhan.smartedu@idsai.local',   'DONE', 'aruzhan.smartedu@idsai.local',   -7,  42),
    ('22000000-0000-0000-0000-000000001108'::uuid, 'Оформить документацию по RBAC и lifecycle',              'Собрать описание архитектуры, ролей, прав, критериев и сценариев демонстрации завершенного проекта.',          '21000000-0000-0000-0000-000000001104'::uuid, 'aruzhan.smartedu@idsai.local',   'DONE', 'aibolat.smartedu@idsai.local',   -6,  39),
    ('22000000-0000-0000-0000-000000001109'::uuid, 'Подготовить demo-датасет по академической активности',   'Собрать и нормализовать данные студентов, посещаемости, дедлайнов и активности для аналитической панели.',      '21000000-0000-0000-0000-000000001105'::uuid, 'madiyar.smartedu@idsai.local',   'DONE', 'madiyar.smartedu@idsai.local',   -5,  32),
    ('22000000-0000-0000-0000-000000001110'::uuid, 'Построить risk-dashboard и визуализацию метрик',         'Подготовить блоки с индексом риска, сравнительными графиками и аналитическими карточками.',                     '21000000-0000-0000-0000-000000001105'::uuid, 'madiyar.smartedu@idsai.local',   'DONE', 'dana.smartedu@idsai.local',      -4,  28),
    ('22000000-0000-0000-0000-000000001111'::uuid, 'Провести финальную интеграцию модулей',                  'Проверить совместную работу backend, frontend, RBAC, grading и итогового сценария completed-проекта.',          '21000000-0000-0000-0000-000000001101'::uuid, 'aibolat.smartedu@idsai.local',   'DONE', 'aibolat.smartedu@idsai.local',   -3,  18),
    ('22000000-0000-0000-0000-000000001112'::uuid, 'Подготовить демонстрационный сценарий защиты',           'Сформировать последовательность показа интерфейса, критериев, задач, оценки и итогового отчета для комиссии.',  '21000000-0000-0000-0000-000000001104'::uuid, 'aruzhan.smartedu@idsai.local',   'DONE', 'aibolat.smartedu@idsai.local',   -2,  6)
  ) AS v(id, title, description, position_id, assignee_email, status, created_by_email, due_days_offset, updated_hours_ago)
)
INSERT INTO tasks(id, tenant_id, project_id, title, description, position_id, assignee_user_id, status, created_by, due_at, created_at, updated_at)
SELECT
  ts.id,
  p.tenant_id,
  p.id,
  ts.title,
  ts.description,
  ts.position_id,
  assignee.id,
  ts.status,
  creator.id,
  now() + (ts.due_days_offset || ' day')::interval,
  now() - interval '14 day',
  now() - (ts.updated_hours_ago || ' hour')::interval
FROM task_seed ts
JOIN projects p ON p.id = '11000000-0000-0000-0000-000000000011'::uuid
JOIN users creator ON creator.email = ts.created_by_email
LEFT JOIN users assignee ON assignee.email = ts.assignee_email
ON CONFLICT (id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    project_id = EXCLUDED.project_id,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    position_id = EXCLUDED.position_id,
    assignee_user_id = EXCLUDED.assignee_user_id,
    status = EXCLUDED.status,
    created_by = EXCLUDED.created_by,
    due_at = EXCLUDED.due_at,
    updated_at = EXCLUDED.updated_at;

WITH submission_seed AS (
  SELECT *
  FROM (VALUES
    ('22000000-0000-0000-0000-000000001101'::uuid, 'aibolat.smartedu@idsai.local',   'ER-модель и связи ролей согласованы, подготовлена схема данных для демонстрации.',           '["https://git.example.local/smartedu/erd-1101"]'),
    ('22000000-0000-0000-0000-000000001102'::uuid, 'aibolat.smartedu@idsai.local',   'Scoped RBAC реализован, project permissions и role assignments работают стабильно.',         '["https://git.example.local/smartedu/rbac-1102"]'),
    ('22000000-0000-0000-0000-000000001103'::uuid, 'dana.smartedu@idsai.local',      'Главная страница проекта готова: шапка, карточки, команда и аналитические блоки.',           '["https://git.example.local/smartedu/ui-project-1103"]'),
    ('22000000-0000-0000-0000-000000001104'::uuid, 'dana.smartedu@idsai.local',      'Отображение TEAM_LEAD, MEMBER и delegated roles добавлено во все ключевые блоки интерфейса.','["https://git.example.local/smartedu/ui-members-1104"]'),
    ('22000000-0000-0000-0000-000000001105'::uuid, 'nursultan.smartedu@idsai.local', 'Task board завершен, переходы статусов и журнал событий готовы к показу.',                     '["https://git.example.local/smartedu/tasks-1105"]'),
    ('22000000-0000-0000-0000-000000001106'::uuid, 'nursultan.smartedu@idsai.local', 'Docker-контур поднят, локальная среда воспроизводится без ручных доработок.',                 '["https://git.example.local/smartedu/docker-1106"]'),
    ('22000000-0000-0000-0000-000000001107'::uuid, 'aruzhan.smartedu@idsai.local',   'QA-чек-лист покрывает роли, permissions, grading и сценарий completed проекта.',              '["https://git.example.local/smartedu/qa-1107"]'),
    ('22000000-0000-0000-0000-000000001108'::uuid, 'aruzhan.smartedu@idsai.local',   'Документация по RBAC и lifecycle оформлена и синхронизирована с системой.',                    '["https://git.example.local/smartedu/docs-1108"]'),
    ('22000000-0000-0000-0000-000000001109'::uuid, 'madiyar.smartedu@idsai.local',   'Demo-датасет очищен и нормализован для панели академического риска.',                           '["https://git.example.local/smartedu/data-1109"]'),
    ('22000000-0000-0000-0000-000000001110'::uuid, 'madiyar.smartedu@idsai.local',   'Risk-dashboard и визуальные метрики готовы к защите и демонстрации.',                           '["https://git.example.local/smartedu/dashboard-1110"]'),
    ('22000000-0000-0000-0000-000000001111'::uuid, 'aibolat.smartedu@idsai.local',   'Интеграция модулей проверена, completed-flow закрыт без замечаний по критическим блокам.',    '["https://git.example.local/smartedu/integration-1111"]'),
    ('22000000-0000-0000-0000-000000001112'::uuid, 'aruzhan.smartedu@idsai.local',   'Сценарий защиты структурирован: команда, задачи, критерии, оценка и итоговый результат.',     '["https://git.example.local/smartedu/demo-1112"]')
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
  now() - interval '5 hour',
  now() - interval '5 hour'
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
    ('22000000-0000-0000-0000-000000001101'::uuid, 'aibolat.smartedu@idsai.local',   'COMPLETED', 'IN_PROGRESS', 'DONE', 'ER-диаграмма готова',                'Архитектурная схема и связи ролей утверждены командой.',                                    '["https://git.example.local/smartedu/erd-1101"]',         70),
    ('22000000-0000-0000-0000-000000001102'::uuid, 'aibolat.smartedu@idsai.local',   'COMPLETED', 'IN_PROGRESS', 'DONE', 'RBAC модуль завершен',               'Project-scope роли и permissions протестированы на основных сценариях.',                    '["https://git.example.local/smartedu/rbac-1102"]',        66),
    ('22000000-0000-0000-0000-000000001103'::uuid, 'dana.smartedu@idsai.local',      'COMPLETED', 'IN_PROGRESS', 'DONE', 'UI страницы проекта завершены',      'Карточка проекта и командное рабочее пространство готовы.',                                 '["https://git.example.local/smartedu/ui-project-1103"]',  58),
    ('22000000-0000-0000-0000-000000001104'::uuid, 'dana.smartedu@idsai.local',      'COMPLETED', 'IN_PROGRESS', 'DONE', 'Роли участников отображаются',      'В интерфейсе показаны TEAM_LEAD, MEMBER и delegated roles.',                                '["https://git.example.local/smartedu/ui-members-1104"]',  56),
    ('22000000-0000-0000-0000-000000001105'::uuid, 'nursultan.smartedu@idsai.local', 'COMPLETED', 'IN_PROGRESS', 'DONE', 'Task board стабилизирован',          'Статусы задач и журнал событий работают согласованно.',                                     '["https://git.example.local/smartedu/tasks-1105"]',       52),
    ('22000000-0000-0000-0000-000000001106'::uuid, 'nursultan.smartedu@idsai.local', 'COMPLETED', 'IN_PROGRESS', 'DONE', 'Docker-среда готова',                'Локальный контур воспроизводится и подходит для демо.',                                     '["https://git.example.local/smartedu/docker-1106"]',      48),
    ('22000000-0000-0000-0000-000000001107'::uuid, 'aruzhan.smartedu@idsai.local',   'COMPLETED', 'IN_PROGRESS', 'DONE', 'QA-сценарии зафиксированы',         'Покрыты проверки ролей, access control и completed-flow.',                                  '["https://git.example.local/smartedu/qa-1107"]',          40),
    ('22000000-0000-0000-0000-000000001108'::uuid, 'aruzhan.smartedu@idsai.local',   'COMPLETED', 'IN_PROGRESS', 'DONE', 'Документация приведена в порядок',  'Lifecycle, RBAC и grading описаны в едином наборе документов.',                             '["https://git.example.local/smartedu/docs-1108"]',        36),
    ('22000000-0000-0000-0000-000000001109'::uuid, 'madiyar.smartedu@idsai.local',   'COMPLETED', 'IN_PROGRESS', 'DONE', 'Демо-данные подготовлены',          'Набор данных очищен и пригоден для аналитических сценариев.',                               '["https://git.example.local/smartedu/data-1109"]',        30),
    ('22000000-0000-0000-0000-000000001110'::uuid, 'madiyar.smartedu@idsai.local',   'COMPLETED', 'IN_PROGRESS', 'DONE', 'Аналитическая панель завершена',    'Панель риска студентов и графики подключены к данным.',                                     '["https://git.example.local/smartedu/dashboard-1110"]',   24),
    ('22000000-0000-0000-0000-000000001111'::uuid, 'aibolat.smartedu@idsai.local',   'COMPLETED', 'IN_PROGRESS', 'DONE', 'Финальная интеграция закрыта',      'Backend, frontend, критерии и итоговая оценка согласованы для демонстрации.',               '["https://git.example.local/smartedu/integration-1111"]', 16),
    ('22000000-0000-0000-0000-000000001112'::uuid, 'aruzhan.smartedu@idsai.local',   'COMPLETED', 'IN_PROGRESS', 'DONE', 'Сценарий защиты готов',             'Подготовлена последовательность показа completed-проекта для комиссии и демо.',              '["https://git.example.local/smartedu/demo-1112"]',        4)
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
LEFT JOIN users u ON u.email = ac.actor_email
WHERE NOT EXISTS (
  SELECT 1
  FROM task_activity_logs l
  WHERE l.task_id = t.id
    AND l.event_type = ac.event_type
    AND l.title = ac.title
);

WITH criteria_seed AS (
  SELECT *
  FROM (VALUES
    ('31000000-0000-0000-0000-000000001101'::uuid, 'Актуальность темы',                    'Решение ориентировано на реальную проблему раннего выявления академических рисков студентов.',                 10),
    ('31000000-0000-0000-0000-000000001102'::uuid, 'Полнота предметной области',           'Проект охватывает роли пользователей, данные, аналитику и основные рабочие сценарии платформы.',              10),
    ('31000000-0000-0000-0000-000000001103'::uuid, 'Архитектура решения',                  'Модульная backend-архитектура и структура данных позволяют масштабировать проект и поддерживать его.',         10),
    ('31000000-0000-0000-0000-000000001104'::uuid, 'Реализация backend',                   'API, бизнес-логика, хранение данных и обработка проектных действий реализованы корректно.',                   10),
    ('31000000-0000-0000-0000-000000001105'::uuid, 'Реализация frontend',                  'Пользовательский интерфейс проекта обеспечивает понятную навигацию и доступ к основным разделам.',            10),
    ('31000000-0000-0000-0000-000000001106'::uuid, 'Ролевая модель и безопасность',        'RBAC и project-scope permissions корректно разграничивают действия участников и преподавателя.',               10),
    ('31000000-0000-0000-0000-000000001107'::uuid, 'Аналитический модуль',                 'Панель аналитики показывает риск-профили, сводные метрики и визуальные индикаторы.',                          10),
    ('31000000-0000-0000-0000-000000001108'::uuid, 'Тестирование и стабильность',          'Ключевые сценарии проверены, демонстрационный контур стабилен и пригоден для повторного запуска.',            10),
    ('31000000-0000-0000-0000-000000001109'::uuid, 'Документация проекта',                 'Подготовлены описания архитектуры, lifecycle, RBAC и сценария демонстрации завершенного проекта.',            10),
    ('31000000-0000-0000-0000-000000001110'::uuid, 'Качество UX и визуальной части',       'Интерфейс цельный и читаемый, но часть аналитических экранов еще можно усилить по визуальной иерархии.',      10)
  ) AS v(id, title, description, weight)
)
INSERT INTO project_criteria(id, tenant_id, project_id, title, description, weight, created_by, created_at)
SELECT
  cs.id,
  p.tenant_id,
  p.id,
  cs.title,
  cs.description,
  cs.weight,
  professor.id,
  now() - interval '1 day'
FROM criteria_seed cs
JOIN projects p ON p.id = '11000000-0000-0000-0000-000000000011'::uuid
JOIN users professor ON professor.email = 'asylkhan.prof@idsai.local'
ON CONFLICT (id) DO UPDATE
SET tenant_id = EXCLUDED.tenant_id,
    project_id = EXCLUDED.project_id,
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    weight = EXCLUDED.weight,
    created_by = EXCLUDED.created_by;

WITH review_seed AS (
  SELECT *
  FROM (VALUES
    ('31000000-0000-0000-0000-000000001101'::uuid, TRUE,  'Тема проекта имеет высокую прикладную ценность для образовательной среды.'),
    ('31000000-0000-0000-0000-000000001102'::uuid, TRUE,  'Основные сущности, роли и сценарии платформы отражены полно и последовательно.'),
    ('31000000-0000-0000-0000-000000001103'::uuid, TRUE,  'Архитектура аккуратная, проект хорошо разделен по модулям и зонам ответственности.'),
    ('31000000-0000-0000-0000-000000001104'::uuid, TRUE,  'Backend реализован качественно, ключевые бизнес-сценарии работают корректно.'),
    ('31000000-0000-0000-0000-000000001105'::uuid, TRUE,  'Frontend хорошо поддерживает рабочий сценарий проекта и его демонстрацию.'),
    ('31000000-0000-0000-0000-000000001106'::uuid, TRUE,  'Ролевая модель доступа и проектные разрешения реализованы убедительно.'),
    ('31000000-0000-0000-0000-000000001107'::uuid, TRUE,  'Аналитический модуль полезен и хорошо интегрирован в общий сценарий системы.'),
    ('31000000-0000-0000-0000-000000001108'::uuid, TRUE,  'Система стабильна, ключевые проверки и демонстрационные сценарии отработаны.'),
    ('31000000-0000-0000-0000-000000001109'::uuid, TRUE,  'Документация проекта подробная и пригодна для сопровождения и презентации.'),
    ('31000000-0000-0000-0000-000000001110'::uuid, FALSE, 'Интерфейс хороший, но аналитические экраны еще можно усилить по визуальной иерархии и акцентам.')
  ) AS v(criterion_id, is_met, comment)
)
INSERT INTO project_criterion_reviews(tenant_id, project_id, criterion_id, professor_id, is_met, comment, created_at, updated_at)
SELECT
  p.tenant_id,
  p.id,
  rs.criterion_id,
  professor.id,
  rs.is_met,
  rs.comment,
  now() - interval '3 hour',
  now() - interval '3 hour'
FROM review_seed rs
JOIN projects p ON p.id = '11000000-0000-0000-0000-000000000011'::uuid
JOIN users professor ON professor.email = 'asylkhan.prof@idsai.local'
ON CONFLICT (project_id, criterion_id, professor_id) DO UPDATE
SET is_met = EXCLUDED.is_met,
    comment = EXCLUDED.comment,
    updated_at = EXCLUDED.updated_at;

-- +goose Down

DELETE FROM role_assignments
WHERE scope_type = 'PROJECT'
  AND scope_id = '11000000-0000-0000-0000-000000000011'::uuid;

DELETE FROM projects
WHERE id = '11000000-0000-0000-0000-000000000011'::uuid;

DELETE FROM role_assignments
WHERE scope_type = 'FACULTY'
  AND user_id IN (
    SELECT id
    FROM users
    WHERE email IN (
      'aibolat.smartedu@idsai.local',
      'dana.smartedu@idsai.local',
      'nursultan.smartedu@idsai.local',
      'aruzhan.smartedu@idsai.local',
      'madiyar.smartedu@idsai.local',
      'asylkhan.prof@idsai.local'
    )
  );

DELETE FROM user_profiles
WHERE user_id IN (
  SELECT id
  FROM users
  WHERE email IN (
    'aibolat.smartedu@idsai.local',
    'dana.smartedu@idsai.local',
    'nursultan.smartedu@idsai.local',
    'aruzhan.smartedu@idsai.local',
    'madiyar.smartedu@idsai.local',
    'asylkhan.prof@idsai.local'
  )
);

DELETE FROM users
WHERE email IN (
  'aibolat.smartedu@idsai.local',
  'dana.smartedu@idsai.local',
  'nursultan.smartedu@idsai.local',
  'aruzhan.smartedu@idsai.local',
  'madiyar.smartedu@idsai.local',
  'asylkhan.prof@idsai.local'
);
