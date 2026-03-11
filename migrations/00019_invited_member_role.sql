-- +goose Up

INSERT INTO roles(code, name)
VALUES ('INVITED_MEMBER', 'Invited Project Member')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'INVITED_MEMBER'
  AND p.code = 'project.view'
ON CONFLICT DO NOTHING;

-- +goose Down

DELETE FROM roles WHERE code = 'INVITED_MEMBER';

