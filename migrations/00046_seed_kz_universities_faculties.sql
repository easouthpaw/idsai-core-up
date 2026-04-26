-- +goose Up

-- ─────────────────────────────────────────────────────────────────────────────
-- КазНУ им. аль-Фараби
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO faculties(code, name)
VALUES
  ('IT_KAZNU',   'Факультет информационных технологий (КазНУ им. аль-Фараби)'),
  ('MATH_KAZNU', 'Механико-математический факультет (КазНУ им. аль-Фараби)'),
  ('PHYS_KAZNU', 'Физико-технический факультет (КазНУ им. аль-Фараби)'),
  ('ECON_KAZNU', 'Экономический факультет (КазНУ им. аль-Фараби)'),
  ('LAW_KAZNU',  'Юридический факультет (КазНУ им. аль-Фараби)')
ON CONFLICT (code) DO NOTHING;

INSERT INTO departments(faculty_id, code, name)
SELECT f.id, v.code, v.name
FROM faculties f
JOIN (VALUES
  ('IT_KAZNU', 'SE',   'Кафедра программной инженерии'),
  ('IT_KAZNU', 'CS',   'Кафедра компьютерных наук'),
  ('IT_KAZNU', 'IS',   'Кафедра информационных систем'),
  ('IT_KAZNU', 'AI',   'Кафедра интеллектуальных систем'),
  ('MATH_KAZNU', 'MM', 'Кафедра математики и механики'),
  ('MATH_KAZNU', 'AM', 'Кафедра прикладной математики'),
  ('PHYS_KAZNU', 'PH', 'Кафедра физики'),
  ('PHYS_KAZNU', 'EE', 'Кафедра электроники и робототехники'),
  ('ECON_KAZNU', 'EC', 'Кафедра экономики'),
  ('ECON_KAZNU', 'FIN','Кафедра финансов и учёта'),
  ('LAW_KAZNU',  'CL', 'Кафедра гражданского права'),
  ('LAW_KAZNU',  'CRL','Кафедра уголовного права')
) AS v(fcode, code, name) ON f.code = v.fcode
ON CONFLICT (faculty_id, code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- КБТУ (Казахстанско-Британский технический университет)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO faculties(code, name)
VALUES
  ('IT_KBTU',  'Факультет информационных технологий (КБТУ)'),
  ('OGE_KBTU', 'Факультет нефтяной и газовой инженерии (КБТУ)'),
  ('BUS_KBTU', 'Школа бизнеса (КБТУ)')
ON CONFLICT (code) DO NOTHING;

INSERT INTO departments(faculty_id, code, name)
SELECT f.id, v.code, v.name
FROM faculties f
JOIN (VALUES
  ('IT_KBTU', 'CS',  'Кафедра компьютерных наук'),
  ('IT_KBTU', 'SE',  'Кафедра программной инженерии'),
  ('IT_KBTU', 'CY',  'Кафедра кибербезопасности'),
  ('IT_KBTU', 'DS',  'Кафедра науки о данных'),
  ('OGE_KBTU','PE',  'Кафедра нефтяной инженерии'),
  ('OGE_KBTU','GE',  'Кафедра геологии'),
  ('BUS_KBTU', 'MG', 'Кафедра менеджмента'),
  ('BUS_KBTU', 'FIN','Кафедра финансов')
) AS v(fcode, code, name) ON f.code = v.fcode
ON CONFLICT (faculty_id, code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- Назарбаев Университет (НУ / NU)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO faculties(code, name)
VALUES
  ('SEDS_NU', 'Школа инженерии и цифровых наук (Назарбаев Университет)'),
  ('GSB_NU',  'Высшая школа бизнеса (Назарбаев Университет)'),
  ('MED_NU',  'Медицинская школа (Назарбаев Университет)'),
  ('LAW_NU',  'Школа права (Назарбаев Университет)'),
  ('SHSS_NU', 'Школа гуманитарных и социальных наук (Назарбаев Университет)')
ON CONFLICT (code) DO NOTHING;

INSERT INTO departments(faculty_id, code, name)
SELECT f.id, v.code, v.name
FROM faculties f
JOIN (VALUES
  ('SEDS_NU', 'CS',  'Компьютерные науки'),
  ('SEDS_NU', 'EE',  'Электротехника и робототехника'),
  ('SEDS_NU', 'CE',  'Гражданское и экологическое строительство'),
  ('SEDS_NU', 'ME',  'Машиностроение'),
  ('SEDS_NU', 'AI',  'Искусственный интеллект'),
  ('GSB_NU',  'MBA', 'Программа MBA'),
  ('GSB_NU',  'FIN', 'Финансы'),
  ('MED_NU',  'MD',  'Медицина'),
  ('LAW_NU',  'LW',  'Право'),
  ('SHSS_NU', 'POL', 'Политология'),
  ('SHSS_NU', 'ECO', 'Экономика')
) AS v(fcode, code, name) ON f.code = v.fcode
ON CONFLICT (faculty_id, code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- МУИТ (Международный университет информационных технологий)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO faculties(code, name)
VALUES
  ('FITE_MUIT', 'Факультет информационных технологий и инженерии (МУИТ)'),
  ('FMB_MUIT',  'Факультет менеджмента и бизнеса (МУИТ)')
ON CONFLICT (code) DO NOTHING;

INSERT INTO departments(faculty_id, code, name)
SELECT f.id, v.code, v.name
FROM faculties f
JOIN (VALUES
  ('FITE_MUIT', 'SE',  'Программная инженерия'),
  ('FITE_MUIT', 'CS',  'Компьютерные науки'),
  ('FITE_MUIT', 'CY',  'Кибербезопасность'),
  ('FITE_MUIT', 'DS',  'Наука о данных и ИИ'),
  ('FITE_MUIT', 'NET', 'Сетевые технологии'),
  ('FMB_MUIT',  'MG',  'Менеджмент'),
  ('FMB_MUIT',  'FIN', 'Финансы и аналитика')
) AS v(fcode, code, name) ON f.code = v.fcode
ON CONFLICT (faculty_id, code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- AITU (Astana IT University)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO faculties(code, name)
VALUES
  ('FIT_AITU',  'Факультет информационных технологий (AITU)'),
  ('FAI_AITU',  'Факультет искусственного интеллекта (AITU)'),
  ('FCYB_AITU', 'Факультет кибербезопасности (AITU)'),
  ('FDM_AITU',  'Факультет цифровых медиа и дизайна (AITU)')
ON CONFLICT (code) DO NOTHING;

INSERT INTO departments(faculty_id, code, name)
SELECT f.id, v.code, v.name
FROM faculties f
JOIN (VALUES
  ('FIT_AITU',  'SE',  'Программная инженерия'),
  ('FIT_AITU',  'CS',  'Компьютерные науки'),
  ('FIT_AITU',  'BD',  'Большие данные'),
  ('FAI_AITU',  'ML',  'Машинное обучение'),
  ('FAI_AITU',  'AI',  'Искусственный интеллект'),
  ('FCYB_AITU', 'CY',  'Кибербезопасность'),
  ('FCYB_AITU', 'IS',  'Информационная безопасность'),
  ('FDM_AITU',  'DM',  'Цифровые медиа'),
  ('FDM_AITU',  'UX',  'UI/UX дизайн')
) AS v(fcode, code, name) ON f.code = v.fcode
ON CONFLICT (faculty_id, code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- Satbayev University (КазНТУ им. К.И. Сатпаева)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO faculties(code, name)
VALUES
  ('IICT_SAT', 'Институт информационных и вычислительных технологий (Satbayev University)'),
  ('IOG_SAT',  'Институт нефти и газа (Satbayev University)'),
  ('IME_SAT',  'Институт горного дела (Satbayev University)'),
  ('IEB_SAT',  'Институт экономики и бизнеса (Satbayev University)'),
  ('ASI_SAT',  'Архитектурно-строительный институт (Satbayev University)')
ON CONFLICT (code) DO NOTHING;

INSERT INTO departments(faculty_id, code, name)
SELECT f.id, v.code, v.name
FROM faculties f
JOIN (VALUES
  ('IICT_SAT', 'CS',  'Компьютерные науки'),
  ('IICT_SAT', 'SE',  'Программная инженерия'),
  ('IICT_SAT', 'AI',  'Искусственный интеллект'),
  ('IICT_SAT', 'IS',  'Информационные системы'),
  ('IOG_SAT',  'PE',  'Нефтяная инженерия'),
  ('IOG_SAT',  'GE',  'Геология'),
  ('IME_SAT',  'ME',  'Горное дело'),
  ('IEB_SAT',  'EC',  'Экономика'),
  ('IEB_SAT',  'MG',  'Менеджмент'),
  ('ASI_SAT',  'AR',  'Архитектура'),
  ('ASI_SAT',  'CE',  'Строительство')
) AS v(fcode, code, name) ON f.code = v.fcode
ON CONFLICT (faculty_id, code) DO NOTHING;

-- ─────────────────────────────────────────────────────────────────────────────
-- SDU (Университет им. Сулеймана Демиреля)
-- ─────────────────────────────────────────────────────────────────────────────
INSERT INTO faculties(code, name)
VALUES
  ('FIT_SDU',  'Факультет информационных технологий (SDU University)'),
  ('FBE_SDU',  'Факультет бизнеса и экономики (SDU University)'),
  ('FLA_SDU',  'Факультет права и администрирования (SDU University)'),
  ('FNS_SDU',  'Факультет естественных наук (SDU University)')
ON CONFLICT (code) DO NOTHING;

INSERT INTO departments(faculty_id, code, name)
SELECT f.id, v.code, v.name
FROM faculties f
JOIN (VALUES
  ('FIT_SDU', 'CS',  'Компьютерные науки'),
  ('FIT_SDU', 'SE',  'Программная инженерия'),
  ('FIT_SDU', 'IS',  'Информационные системы'),
  ('FBE_SDU', 'EC',  'Экономика'),
  ('FBE_SDU', 'MG',  'Менеджмент и маркетинг'),
  ('FLA_SDU', 'LW',  'Право'),
  ('FNS_SDU', 'BIO', 'Биология'),
  ('FNS_SDU', 'CH',  'Химия')
) AS v(fcode, code, name) ON f.code = v.fcode
ON CONFLICT (faculty_id, code) DO NOTHING;

-- +goose Down

DELETE FROM departments
WHERE faculty_id IN (
  SELECT id FROM faculties
  WHERE code IN (
    'IT_KAZNU','MATH_KAZNU','PHYS_KAZNU','ECON_KAZNU','LAW_KAZNU',
    'IT_KBTU','OGE_KBTU','BUS_KBTU',
    'SEDS_NU','GSB_NU','MED_NU','LAW_NU','SHSS_NU',
    'FITE_MUIT','FMB_MUIT',
    'FIT_AITU','FAI_AITU','FCYB_AITU','FDM_AITU',
    'IICT_SAT','IOG_SAT','IME_SAT','IEB_SAT','ASI_SAT',
    'FIT_SDU','FBE_SDU','FLA_SDU','FNS_SDU'
  )
);

DELETE FROM faculties
WHERE code IN (
  'IT_KAZNU','MATH_KAZNU','PHYS_KAZNU','ECON_KAZNU','LAW_KAZNU',
  'IT_KBTU','OGE_KBTU','BUS_KBTU',
  'SEDS_NU','GSB_NU','MED_NU','LAW_NU','SHSS_NU',
  'FITE_MUIT','FMB_MUIT',
  'FIT_AITU','FAI_AITU','FCYB_AITU','FDM_AITU',
  'IICT_SAT','IOG_SAT','IME_SAT','IEB_SAT','ASI_SAT',
  'FIT_SDU','FBE_SDU','FLA_SDU','FNS_SDU'
);
