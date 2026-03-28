-- +goose Up

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON p.code = 'project.view'
WHERE r.code IN ('TEAM_LEAD', 'MEMBER', 'PROJECT_PROFESSOR')
ON CONFLICT DO NOTHING;

-- +goose Down

DELETE FROM role_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE code = 'project.view')
  AND role_id IN (
    SELECT id
    FROM roles
    WHERE code IN ('TEAM_LEAD', 'MEMBER', 'PROJECT_PROFESSOR')
  );
