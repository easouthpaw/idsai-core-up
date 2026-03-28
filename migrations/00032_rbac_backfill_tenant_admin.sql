-- +goose Up
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT ra.tenant_id,
       ra.user_id,
       (SELECT id FROM roles WHERE code = 'TENANT_ADMIN'),
       'TENANT',
       ra.tenant_id
FROM role_assignments ra
JOIN roles r ON r.id = ra.role_id
WHERE r.code = 'SUPER_ADMIN'
  AND ra.scope_type = 'SYSTEM'
  AND ra.scope_id IS NULL
ON CONFLICT (tenant_id, user_id, role_id, scope_type, scope_id) WHERE scope_id IS NOT NULL
DO NOTHING;

-- +goose Down
DELETE FROM role_assignments
WHERE role_id = (SELECT id FROM roles WHERE code = 'TENANT_ADMIN')
  AND scope_type = 'TENANT';
