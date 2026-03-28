-- +goose Up
INSERT INTO roles(code, name)
VALUES ('TENANT_ADMIN', 'Tenant Admin')
ON CONFLICT (code) DO NOTHING;

INSERT INTO permissions(code, description) VALUES
  ('tenant.manage_users', 'Manage tenant users'),
  ('tenant.manage_rbac', 'Manage tenant RBAC')
ON CONFLICT (code) DO NOTHING;

INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'TENANT_ADMIN'
  AND p.code IN ('tenant.manage_users', 'tenant.manage_rbac')
ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE code = 'TENANT_ADMIN');

DELETE FROM permissions
WHERE code IN ('tenant.manage_users', 'tenant.manage_rbac');

DELETE FROM roles
WHERE code = 'TENANT_ADMIN';
