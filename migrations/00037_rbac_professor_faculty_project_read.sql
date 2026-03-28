-- +goose Up
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code IN ('project.view', 'task.view', 'grading.view')
WHERE r.code = 'PROFESSOR'
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE code = 'PROFESSOR')
  AND permission_id IN (
    SELECT id
    FROM permissions
    WHERE code IN ('project.view', 'task.view', 'grading.view')
  );
