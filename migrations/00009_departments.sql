-- +goose Up

CREATE TABLE IF NOT EXISTS departments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  faculty_id UUID NOT NULL REFERENCES faculties(id) ON DELETE CASCADE,
  code TEXT NOT NULL,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (faculty_id, code)
);

-- Seed 6 кафедр внутри IDSAI_ENU
INSERT INTO departments(faculty_id, code, name)
SELECT f.id, v.code, v.name
FROM faculties f
JOIN (VALUES
  ('CPI',  'Кафедра Компьютерной и программной инженерии'),
  ('INF',  'Кафедра Информатики'),
  ('SEC',  'Кафедра Информационная безопасность'),
  ('IS',   'Кафедра Информационных систем'),
  ('SAU',  'Кафедра Системного анализа и управления'),
  ('AI',   'Кафедра Технологии искусственного интеллекта')
) AS v(code, name) ON true
WHERE f.code = 'IDSAI_ENU'
ON CONFLICT (faculty_id, code) DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS departments;