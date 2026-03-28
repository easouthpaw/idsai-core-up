-- +goose Up
WITH ranked AS (
  SELECT id,
         ROW_NUMBER() OVER (
           PARTITION BY tenant_id, user_id, role_id, scope_type, COALESCE(scope_id, '00000000-0000-0000-0000-000000000000'::uuid)
           ORDER BY (expires_at IS NULL) DESC, COALESCE(expires_at, 'infinity'::timestamptz) DESC, created_at DESC
         ) AS rn
  FROM role_assignments
)
DELETE FROM role_assignments
WHERE id IN (SELECT id FROM ranked WHERE rn > 1);

CREATE UNIQUE INDEX IF NOT EXISTS ux_role_assignments_tuple_scoped
ON role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
WHERE scope_id IS NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS ux_role_assignments_tuple_system
ON role_assignments(tenant_id, user_id, role_id, scope_type)
WHERE scope_id IS NULL;

-- +goose Down
DROP INDEX IF EXISTS ux_role_assignments_tuple_system;
DROP INDEX IF EXISTS ux_role_assignments_tuple_scoped;
