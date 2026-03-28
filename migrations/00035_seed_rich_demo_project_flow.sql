-- +goose Up

WITH faculty_ctx AS (
  SELECT f.id AS faculty_id, f.tenant_id
  FROM faculties f
  WHERE f.code = 'IDSAI_ENU'
  LIMIT 1
), project_seed AS (
  SELECT *
  FROM (VALUES
    (
      '11000000-0000-0000-0000-000000000007'::uuid,
      'Research Pulse Platform',
      'Единая панель для исследовательских проектов кафедры: дедлайны, этапы, статусы защиты и сигналы по рискам.',
      'REVIEW',
      'FACULTY',
      NULL::text,
      'dinara.student@idsai.local',
      'sabina.prof@idsai.local',
      7,
      30
    ),
    (
      '11000000-0000-0000-0000-000000000008'::uuid,
      'Smart Dorm Assistant',
      'Сервис для заявок на заселение, статусов комнат и автоматических подсказок студентам общежития.',
      'RECRUITMENT',
      'GROUP',
      'CPI-45',
      'aibolat.student@idsai.local',
      'aidana.prof@idsai.local',
      6,
      12
    ),
    (
      '11000000-0000-0000-0000-000000000009'::uuid,
      'Exam Integrity Hub',
      'Контур мониторинга прокторинга: события, антифрод-сигналы, история инцидентов и панель преподавателя.',
      'ACTIVE',
      'FACULTY',
      NULL::text,
      'marat.student@idsai.local',
      'nurzhan.prof@idsai.local',
      10,
      6
    ),
    (
      '11000000-0000-0000-0000-000000000010'::uuid,
      'Mentor Match AI',
      'Рекомендательный сервис для подбора наставников по стеку, темпу команды и истории завершенных проектов.',
      'GRADING',
      'PUBLIC',
      NULL::text,
      'aliya.student@idsai.local',
      'aidana.prof@idsai.local',
      8,
      3
    )
  ) AS v(
    id,
    title,
    description,
    status,
    visibility,
    group_code,
    created_by_email,
    professor_email,
    created_days_ago,
    updated_hours_ago
  )
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
  created_at,
  updated_at
)
SELECT
  ps.id,
  fc.tenant_id,
  ps.title,
  ps.description,
  ps.status,
  (ps.visibility = 'PUBLIC'),
  creator.id,
  professor.id,
  'ACCEPTED',
  now() - interval '5 day',
  now() - interval '4 day',
  fc.faculty_id,
  ps.visibility,
  sg.id,
  now() - (ps.created_days_ago || ' day')::interval,
  now() - (ps.updated_hours_ago || ' hour')::interval
FROM project_seed ps
JOIN faculty_ctx fc ON TRUE
JOIN users creator ON creator.email = ps.created_by_email
JOIN users professor ON professor.email = ps.professor_email
LEFT JOIN student_groups sg
  ON sg.faculty_id = fc.faculty_id
 AND sg.code = ps.group_code
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
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
    created_at = EXCLUDED.created_at,
    updated_at = EXCLUDED.updated_at;

WITH position_seed AS (
  SELECT *
  FROM (VALUES
    ('21000000-0000-0000-0000-000000000701'::uuid, '11000000-0000-0000-0000-000000000007'::uuid, 'BACKEND',  'Backend Engineer',   1),
    ('21000000-0000-0000-0000-000000000702'::uuid, '11000000-0000-0000-0000-000000000007'::uuid, 'FRONTEND', 'Frontend Engineer',  1),
    ('21000000-0000-0000-0000-000000000703'::uuid, '11000000-0000-0000-0000-000000000007'::uuid, 'ANALYST',  'Research Analyst',   1),
    ('21000000-0000-0000-0000-000000000801'::uuid, '11000000-0000-0000-0000-000000000008'::uuid, 'BACKEND',  'Backend Engineer',   1),
    ('21000000-0000-0000-0000-000000000802'::uuid, '11000000-0000-0000-0000-000000000008'::uuid, 'MOBILE',   'Mobile Engineer',    1),
    ('21000000-0000-0000-0000-000000000803'::uuid, '11000000-0000-0000-0000-000000000008'::uuid, 'QA',       'QA Engineer',        1),
    ('21000000-0000-0000-0000-000000000901'::uuid, '11000000-0000-0000-0000-000000000009'::uuid, 'BACKEND',  'Backend Engineer',   1),
    ('21000000-0000-0000-0000-000000000902'::uuid, '11000000-0000-0000-0000-000000000009'::uuid, 'FRONTEND', 'Frontend Engineer',  1),
    ('21000000-0000-0000-0000-000000000903'::uuid, '11000000-0000-0000-0000-000000000009'::uuid, 'QA',       'QA Engineer',        1),
    ('21000000-0000-0000-0000-000000001001'::uuid, '11000000-0000-0000-0000-000000000010'::uuid, 'ML',       'ML Engineer',        1),
    ('21000000-0000-0000-0000-000000001002'::uuid, '11000000-0000-0000-0000-000000000010'::uuid, 'BACKEND',  'Backend Engineer',   1),
    ('21000000-0000-0000-0000-000000001003'::uuid, '11000000-0000-0000-0000-000000000010'::uuid, 'DATA',     'Data Engineer',      1)
  ) AS v(id, project_id, code, name, capacity)
)
INSERT INTO project_positions(id, tenant_id, project_id, code, name, capacity, created_at)
SELECT
  ps.id,
  p.tenant_id,
  ps.project_id,
  ps.code,
  ps.name,
  ps.capacity,
  now() - interval '6 day'
FROM position_seed ps
JOIN projects p ON p.id = ps.project_id
ON CONFLICT (id) DO UPDATE
SET code = EXCLUDED.code,
    name = EXCLUDED.name,
    capacity = EXCLUDED.capacity;

WITH member_seed AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000007'::uuid, 'dinara.student@idsai.local',  '21000000-0000-0000-0000-000000000702'::uuid, 6, NULL::text),
    ('11000000-0000-0000-0000-000000000007'::uuid, 'aibolat.student@idsai.local', '21000000-0000-0000-0000-000000000701'::uuid, 5, 'dinara.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000007'::uuid, 'marat.student@idsai.local',   '21000000-0000-0000-0000-000000000703'::uuid, 4, 'dinara.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000008'::uuid, 'aibolat.student@idsai.local', '21000000-0000-0000-0000-000000000801'::uuid, 5, NULL::text),
    ('11000000-0000-0000-0000-000000000008'::uuid, 'dinara.student@idsai.local',  '21000000-0000-0000-0000-000000000802'::uuid, 4, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000008'::uuid, 'daniyar.student@idsai.local', '21000000-0000-0000-0000-000000000803'::uuid, 4, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000009'::uuid, 'marat.student@idsai.local',   '21000000-0000-0000-0000-000000000901'::uuid, 9, NULL::text),
    ('11000000-0000-0000-0000-000000000009'::uuid, 'dinara.student@idsai.local',  '21000000-0000-0000-0000-000000000902'::uuid, 8, 'marat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000009'::uuid, 'aliya.student@idsai.local',   '21000000-0000-0000-0000-000000000903'::uuid, 8, 'marat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000010'::uuid, 'aliya.student@idsai.local',   '21000000-0000-0000-0000-000000001001'::uuid, 7, NULL::text),
    ('11000000-0000-0000-0000-000000000010'::uuid, 'aibolat.student@idsai.local', '21000000-0000-0000-0000-000000001002'::uuid, 6, 'aliya.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000010'::uuid, 'daniyar.student@idsai.local', '21000000-0000-0000-0000-000000001003'::uuid, 6, 'aliya.student@idsai.local')
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
  CASE
    WHEN ms.invited_by_email IS NULL THEN now() - interval '6 day'
    ELSE now() - interval '2 day'
  END
FROM member_seed ms
JOIN projects p ON p.id = ms.project_id
JOIN users u ON u.email = ms.email
LEFT JOIN users inviter ON inviter.email = ms.invited_by_email
ON CONFLICT (project_id, user_id) DO UPDATE
SET status = EXCLUDED.status,
    position_id = EXCLUDED.position_id,
    joined_at = COALESCE(project_members.joined_at, EXCLUDED.joined_at),
    invite_comment = EXCLUDED.invite_comment,
    invited_by = EXCLUDED.invited_by,
    responded_at = COALESCE(project_members.responded_at, EXCLUDED.responded_at);

WITH project_leads AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000007'::uuid, 'dinara.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000008'::uuid, 'aibolat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000009'::uuid, 'marat.student@idsai.local'),
    ('11000000-0000-0000-0000-000000000010'::uuid, 'aliya.student@idsai.local')
  ) AS v(project_id, email)
)
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT
  p.tenant_id,
  u.id,
  r.id,
  'PROJECT',
  pl.project_id
FROM project_leads pl
JOIN projects p ON p.id = pl.project_id
JOIN users u ON u.email = pl.email
JOIN roles r ON r.code = 'TEAM_LEAD'
WHERE NOT EXISTS (
  SELECT 1
  FROM role_assignments ra
  WHERE ra.tenant_id = p.tenant_id
    AND ra.user_id = u.id
    AND ra.role_id = r.id
    AND ra.scope_type = 'PROJECT'
    AND ra.scope_id = pl.project_id
    AND ra.expires_at IS NULL
);

WITH project_professors AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000007'::uuid, 'sabina.prof@idsai.local'),
    ('11000000-0000-0000-0000-000000000008'::uuid, 'aidana.prof@idsai.local'),
    ('11000000-0000-0000-0000-000000000009'::uuid, 'nurzhan.prof@idsai.local'),
    ('11000000-0000-0000-0000-000000000010'::uuid, 'aidana.prof@idsai.local')
  ) AS v(project_id, email)
)
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT
  p.tenant_id,
  u.id,
  r.id,
  'PROJECT',
  pp.project_id
FROM project_professors pp
JOIN projects p ON p.id = pp.project_id
JOIN users u ON u.email = pp.email
JOIN roles r ON r.code = 'PROJECT_PROFESSOR'
WHERE NOT EXISTS (
  SELECT 1
  FROM role_assignments ra
  WHERE ra.tenant_id = p.tenant_id
    AND ra.user_id = u.id
    AND ra.role_id = r.id
    AND ra.scope_type = 'PROJECT'
    AND ra.scope_id = pp.project_id
    AND ra.expires_at IS NULL
);

INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT
  p.tenant_id,
  pm.user_id,
  r.id,
  'PROJECT',
  pm.project_id
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
JOIN roles r ON r.code = 'MEMBER'
WHERE pm.project_id IN (
  '11000000-0000-0000-0000-000000000007'::uuid,
  '11000000-0000-0000-0000-000000000008'::uuid,
  '11000000-0000-0000-0000-000000000009'::uuid,
  '11000000-0000-0000-0000-000000000010'::uuid
)
  AND pm.status = 'ACTIVE'
  AND pm.user_id <> p.created_by
  AND NOT EXISTS (
    SELECT 1
    FROM role_assignments ra
    WHERE ra.tenant_id = p.tenant_id
      AND ra.user_id = pm.user_id
      AND ra.role_id = r.id
      AND ra.scope_type = 'PROJECT'
      AND ra.scope_id = pm.project_id
      AND ra.expires_at IS NULL
  );

WITH stack_seed AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000007'::uuid, 'GO'),
    ('11000000-0000-0000-0000-000000000007'::uuid, 'REACT'),
    ('11000000-0000-0000-0000-000000000007'::uuid, 'POSTGRESQL'),
    ('11000000-0000-0000-0000-000000000008'::uuid, 'PYTHON'),
    ('11000000-0000-0000-0000-000000000008'::uuid, 'TELEGRAM'),
    ('11000000-0000-0000-0000-000000000008'::uuid, 'POSTGRESQL'),
    ('11000000-0000-0000-0000-000000000009'::uuid, 'GO'),
    ('11000000-0000-0000-0000-000000000009'::uuid, 'KAFKA'),
    ('11000000-0000-0000-0000-000000000009'::uuid, 'POSTGRESQL'),
    ('11000000-0000-0000-0000-000000000010'::uuid, 'PYTHON'),
    ('11000000-0000-0000-0000-000000000010'::uuid, 'FASTAPI'),
    ('11000000-0000-0000-0000-000000000010'::uuid, 'LLM')
  ) AS v(project_id, stack_code)
)
INSERT INTO project_stacks(tenant_id, project_id, stack_code)
SELECT
  p.tenant_id,
  ss.project_id,
  ss.stack_code
FROM stack_seed ss
JOIN projects p ON p.id = ss.project_id
ON CONFLICT (project_id, stack_code) DO NOTHING;

WITH task_seed AS (
  SELECT *
  FROM (VALUES
    ('22000000-0000-0000-0000-000000000701'::uuid, '11000000-0000-0000-0000-000000000007'::uuid, 'Собрать карту исследовательских сценариев', 'Собрать пользовательские сценарии для лабораторий, дедлайнов и статусов защиты.', '21000000-0000-0000-0000-000000000703'::uuid, 'marat.student@idsai.local',   'DONE',        'dinara.student@idsai.local',  -1, 18),
    ('22000000-0000-0000-0000-000000000702'::uuid, '11000000-0000-0000-0000-000000000007'::uuid, 'API календаря грантов',                 'Поднять endpoint-ы для дедлайнов, этапов и событий по проектам.',                  '21000000-0000-0000-0000-000000000701'::uuid, 'aibolat.student@idsai.local', 'IN_PROGRESS', 'dinara.student@idsai.local',   3, 8),
    ('22000000-0000-0000-0000-000000000703'::uuid, '11000000-0000-0000-0000-000000000007'::uuid, 'Лента активности кафедры',             'Сверстать экран с таймлайном исследований и рисками по дедлайнам.',                '21000000-0000-0000-0000-000000000702'::uuid, NULL::text,                    'OPEN',        'dinara.student@idsai.local',   5, 4),
    ('22000000-0000-0000-0000-000000000801'::uuid, '11000000-0000-0000-0000-000000000008'::uuid, 'Архитектура чат-бота общежития',       'Собрать backend-флоу заявок, уведомлений и статусов комнат.',                      '21000000-0000-0000-0000-000000000801'::uuid, 'aibolat.student@idsai.local', 'DONE',        'aibolat.student@idsai.local',  1, 16),
    ('22000000-0000-0000-0000-000000000802'::uuid, '11000000-0000-0000-0000-000000000008'::uuid, 'Мобильный экран заселения',            'Сделать мобильный сценарий подачи заявки и проверки статуса.',                     '21000000-0000-0000-0000-000000000802'::uuid, 'dinara.student@idsai.local',  'IN_PROGRESS', 'aibolat.student@idsai.local',  4, 7),
    ('22000000-0000-0000-0000-000000000803'::uuid, '11000000-0000-0000-0000-000000000008'::uuid, 'Smoke-набор тестов для релиза',        'Подготовить регрессионный чек-лист для релизного прогона.',                        '21000000-0000-0000-0000-000000000803'::uuid, NULL::text,                    'OPEN',        'aibolat.student@idsai.local',  6, 5),
    ('22000000-0000-0000-0000-000000000901'::uuid, '11000000-0000-0000-0000-000000000009'::uuid, 'Схема событий прокторинга',            'Описать модель событий, флаги антифрода и связи с экзаменационными сессиями.',     '21000000-0000-0000-0000-000000000901'::uuid, 'marat.student@idsai.local',   'DONE',        'marat.student@idsai.local',    -2, 10),
    ('22000000-0000-0000-0000-000000000902'::uuid, '11000000-0000-0000-0000-000000000009'::uuid, 'Панель инцидентов преподавателя',      'Собрать интерфейс приоритизации событий и разметки инцидентов.',                   '21000000-0000-0000-0000-000000000902'::uuid, 'dinara.student@idsai.local',  'IN_PROGRESS', 'marat.student@idsai.local',    2, 6),
    ('22000000-0000-0000-0000-000000000903'::uuid, '11000000-0000-0000-0000-000000000009'::uuid, 'Набор регрессионных кейсов',           'Подготовить QA-кейсы для цепочки обнаружения подозрительных паттернов.',           '21000000-0000-0000-0000-000000000903'::uuid, NULL::text,                    'OPEN',        'marat.student@idsai.local',    3, 4),
    ('22000000-0000-0000-0000-000000001001'::uuid, '11000000-0000-0000-0000-000000000010'::uuid, 'Очистить датасет профилей наставников','Унифицировать признаки, стек и доступность по historical данным.',                 '21000000-0000-0000-0000-000000001003'::uuid, 'daniyar.student@idsai.local', 'DONE',        'aliya.student@idsai.local',   -3, 14),
    ('22000000-0000-0000-0000-000000001002'::uuid, '11000000-0000-0000-0000-000000000010'::uuid, 'Сервис ранжирования кандидатов',       'Поднять API для выдачи shortlist по проекту и выбранным критериям.',               '21000000-0000-0000-0000-000000001002'::uuid, 'aibolat.student@idsai.local', 'DONE',        'aliya.student@idsai.local',   -2, 12),
    ('22000000-0000-0000-0000-000000001003'::uuid, '11000000-0000-0000-0000-000000000010'::uuid, 'Эксперименты с matching score',        'Зафиксировать quality метрики и сравнить варианты ранжирования по cold-start.',    '21000000-0000-0000-0000-000000001001'::uuid, 'aliya.student@idsai.local',   'DONE',        'aliya.student@idsai.local',   -1, 9)
  ) AS v(id, project_id, title, description, position_id, assignee_email, status, created_by_email, due_days_offset, updated_hours_ago)
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
  now() - interval '5 day',
  now() - (ts.updated_hours_ago || ' hour')::interval
FROM task_seed ts
JOIN projects p ON p.id = ts.project_id
JOIN users creator ON creator.email = ts.created_by_email
LEFT JOIN users assignee ON assignee.email = ts.assignee_email
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
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
    ('22000000-0000-0000-0000-000000000701'::uuid, 'marat.student@idsai.local',   'Сценарии по лабораториям и дедлайнам собраны, риски описаны.', '["https://git.example.local/research/scenarios-701"]'),
    ('22000000-0000-0000-0000-000000000801'::uuid, 'aibolat.student@idsai.local', 'Backend-флоу заявок и уведомлений готов к показу.',             '["https://git.example.local/dorm/backend-801"]'),
    ('22000000-0000-0000-0000-000000000901'::uuid, 'marat.student@idsai.local',   'Схема событий и антифрод-сигналов согласована с командой.',     '["https://git.example.local/exam/schema-901"]'),
    ('22000000-0000-0000-0000-000000001001'::uuid, 'daniyar.student@idsai.local', 'Датасет очищен, пропуски и дубликаты убраны.',                  '["https://git.example.local/mentor/data-1001"]'),
    ('22000000-0000-0000-0000-000000001002'::uuid, 'aibolat.student@idsai.local', 'Ранжирование кандидатов поднято в API и готово к ревью.',       '["https://git.example.local/mentor/api-1002"]'),
    ('22000000-0000-0000-0000-000000001003'::uuid, 'aliya.student@idsai.local',   'Матчинг-score сравнен на baseline и tuned-модели.',             '["https://git.example.local/mentor/metrics-1003"]')
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
  now() - interval '10 hour',
  now() - interval '10 hour'
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
    ('22000000-0000-0000-0000-000000000702'::uuid, 'aibolat.student@idsai.local', 'STATUS_CHANGED', 'OPEN',        'IN_PROGRESS', 'API взят в работу',                 'Начал поднимать endpoint-ы для календаря.', '[]', 8),
    ('22000000-0000-0000-0000-000000000801'::uuid, 'aibolat.student@idsai.local', 'COMPLETED',      'IN_PROGRESS', 'DONE',        'Задача завершена',                  'Backend-флоу готов к ревью.',               '["https://git.example.local/dorm/backend-801"]', 16),
    ('22000000-0000-0000-0000-000000000901'::uuid, 'marat.student@idsai.local',   'COMPLETED',      'IN_PROGRESS', 'DONE',        'Задача завершена',                  'Схема событий готова и документирована.',   '["https://git.example.local/exam/schema-901"]', 10),
    ('22000000-0000-0000-0000-000000001001'::uuid, 'daniyar.student@idsai.local', 'COMPLETED',      'IN_PROGRESS', 'DONE',        'Задача завершена',                  'Подготовлен чистый датасет для оценки.',    '["https://git.example.local/mentor/data-1001"]', 14),
    ('22000000-0000-0000-0000-000000001002'::uuid, 'aibolat.student@idsai.local', 'COMPLETED',      'IN_PROGRESS', 'DONE',        'Задача завершена',                  'Ranking API стабилен и покрыт тестами.',    '["https://git.example.local/mentor/api-1002"]', 12),
    ('22000000-0000-0000-0000-000000001003'::uuid, 'aliya.student@idsai.local',   'COMPLETED',      'IN_PROGRESS', 'DONE',        'Задача завершена',                  'Эксперименты сведены в отчет и графики.',   '["https://git.example.local/mentor/metrics-1003"]', 9)
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
    ('31000000-0000-0000-0000-000000000701'::uuid, '11000000-0000-0000-0000-000000000007'::uuid, 'Методология и сценарии',         'Насколько полно покрыты сценарии лабораторий, этапов и дедлайнов.', 35, 'sabina.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000702'::uuid, '11000000-0000-0000-0000-000000000007'::uuid, 'Архитектура панели и API',     'Чистота модели данных, API и пригодность к расширению.',            35, 'sabina.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000703'::uuid, '11000000-0000-0000-0000-000000000007'::uuid, 'Презентация и читаемость',     'Насколько понятно отображаются статусы, риски и история проекта.',  30, 'sabina.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000801'::uuid, '11000000-0000-0000-0000-000000000008'::uuid, 'Полнота пользовательского пути','Покрытие сценариев подачи заявки, проверки статуса и обратной связи.', 30, 'aidana.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000802'::uuid, '11000000-0000-0000-0000-000000000008'::uuid, 'Инженерное качество решения',  'Структура backend, стабильность интеграций и читаемость кода.',      35, 'aidana.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000803'::uuid, '11000000-0000-0000-0000-000000000008'::uuid, 'Тестирование и эксплуатация',  'Наличие smoke/regression чек-листа и критериев релизной готовности.', 35, 'aidana.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000901'::uuid, '11000000-0000-0000-0000-000000000009'::uuid, 'Надежность антифрод-логики',   'Корректность сигналов и качество связки событий прокторинга.',       40, 'nurzhan.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000902'::uuid, '11000000-0000-0000-0000-000000000009'::uuid, 'UX и оперативность панели',    'Насколько быстро и понятно преподаватель может обработать инцидент.', 30, 'nurzhan.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000000903'::uuid, '11000000-0000-0000-0000-000000000009'::uuid, 'Документация и метрики',       'Достаточность аналитики, журналов событий и продуктовых метрик.',    30, 'nurzhan.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000001001'::uuid, '11000000-0000-0000-0000-000000000010'::uuid, 'Качество matching-модели',     'Качество рекомендаций и устойчивость подбора на реальных сценариях.', 40, 'aidana.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000001002'::uuid, '11000000-0000-0000-0000-000000000010'::uuid, 'Production readiness сервиса', 'Готовность backend и data pipeline к эксплуатации в контуре проекта.', 35, 'aidana.prof@idsai.local'),
    ('31000000-0000-0000-0000-000000001003'::uuid, '11000000-0000-0000-0000-000000000010'::uuid, 'Демонстрация и отчет',         'Качество презентации, читаемость выводов и ясность next steps.',     25, 'aidana.prof@idsai.local')
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
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title,
    description = EXCLUDED.description,
    weight = EXCLUDED.weight,
    created_by = EXCLUDED.created_by;

WITH review_seed AS (
  SELECT *
  FROM (VALUES
    ('11000000-0000-0000-0000-000000000010'::uuid, '31000000-0000-0000-0000-000000001001'::uuid, 'aidana.prof@idsai.local', TRUE,  'Модель хорошо учитывает стек и доступность наставника.'),
    ('11000000-0000-0000-0000-000000000010'::uuid, '31000000-0000-0000-0000-000000001002'::uuid, 'aidana.prof@idsai.local', TRUE,  'API и data pipeline выглядят устойчиво и готовы к дальнейшему росту.'),
    ('11000000-0000-0000-0000-000000000010'::uuid, '31000000-0000-0000-0000-000000001003'::uuid, 'aidana.prof@idsai.local', FALSE, 'Нужно усилить story вокруг cold-start кейсов и итоговой презентации.')
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
    '11000000-0000-0000-0000-000000000007'::uuid,
    '11000000-0000-0000-0000-000000000008'::uuid,
    '11000000-0000-0000-0000-000000000009'::uuid,
    '11000000-0000-0000-0000-000000000010'::uuid
  );

DELETE FROM projects
WHERE id IN (
  '11000000-0000-0000-0000-000000000007'::uuid,
  '11000000-0000-0000-0000-000000000008'::uuid,
  '11000000-0000-0000-0000-000000000009'::uuid,
  '11000000-0000-0000-0000-000000000010'::uuid
);
