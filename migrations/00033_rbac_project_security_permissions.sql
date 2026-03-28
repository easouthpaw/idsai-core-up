-- +goose Up
INSERT INTO permissions(code, description) VALUES
  ('project.delete', 'Delete owned project'),
  ('project.review.respond', 'Respond to professor review invitation')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'TEAM_LEAD'
  AND p.code = 'project.delete'
ON CONFLICT DO NOTHING;

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'PROFESSOR'
  AND p.code = 'project.review.respond'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions
WHERE permission_id IN (
  SELECT id
  FROM permissions
  WHERE code IN ('project.delete', 'project.review.respond')
);

DELETE FROM permissions
WHERE code IN ('project.delete', 'project.review.respond');
