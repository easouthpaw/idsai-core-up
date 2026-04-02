-- +goose Up

INSERT INTO permissions(code, description) VALUES
  ('task.delete', 'Delete task from project board')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code IN ('TEAM_LEAD', 'TASK_MANAGER')
  AND p.code = 'task.delete'
ON CONFLICT DO NOTHING;

-- +goose Down

DELETE FROM role_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE code = 'task.delete')
  AND role_id IN (SELECT id FROM roles WHERE code IN ('TEAM_LEAD', 'TASK_MANAGER'));

DELETE FROM permissions WHERE code = 'task.delete';
