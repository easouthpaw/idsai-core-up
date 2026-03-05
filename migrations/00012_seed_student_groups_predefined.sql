-- +goose Up

INSERT INTO student_groups(faculty_id, code, name, year)
SELECT f.id, v.code, v.name, v.year
FROM faculties f
JOIN (VALUES
  ('IS-45',  'IS-45',  4),
  ('IS-46',  'IS-46',  3),
  ('IS-47',  'IS-47',  2),
  ('AI-45',  'AI-45',  4),
  ('AI-46',  'AI-46',  3),
  ('CPI-45', 'CPI-45', 4),
  ('INF-45', 'INF-45', 4),
  ('SEC-45', 'SEC-45', 4),
  ('SAU-45', 'SAU-45', 4)
) AS v(code, name, year) ON true
WHERE f.code = 'IDSAI_ENU'
ON CONFLICT (faculty_id, code) DO NOTHING;

-- +goose Down

DELETE FROM student_groups
WHERE faculty_id = (SELECT id FROM faculties WHERE code = 'IDSAI_ENU')
  AND code IN ('IS-45','IS-46','IS-47','AI-45','AI-46','CPI-45','INF-45','SEC-45','SAU-45');
