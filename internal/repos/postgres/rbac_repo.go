package postgres

import (
	"context"
	"time"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CacheInvalidatorFunc is called after role mutations to clear cached permissions.
type CacheInvalidatorFunc func(ctx context.Context, userID uuid.UUID)

type RBACRepo struct {
	db                 *pgxpool.Pool
	cacheInvalidatorFn CacheInvalidatorFunc
}

const resolvedScopesCTE = `
WITH resolved_scopes AS (
  SELECT 'SYSTEM'::text AS scope_type, NULL::uuid AS scope_id, u.tenant_id
  FROM users u
  WHERE $2 = 'SYSTEM'
    AND u.id = $1

  UNION ALL

  SELECT 'SYSTEM'::text, NULL::uuid, t.id
  FROM tenants t
  WHERE $2 = 'TENANT'
    AND t.id = $3

  UNION ALL

  SELECT 'TENANT'::text, t.id, t.id
  FROM tenants t
  WHERE $2 = 'TENANT'
    AND t.id = $3

  UNION ALL

  SELECT 'SYSTEM'::text, NULL::uuid, f.tenant_id
  FROM faculties f
  WHERE $2 = 'FACULTY'
    AND f.id = $3

  UNION ALL

  SELECT 'TENANT'::text, f.tenant_id, f.tenant_id
  FROM faculties f
  WHERE $2 = 'FACULTY'
    AND f.id = $3

  UNION ALL

  SELECT 'FACULTY'::text, f.id, f.tenant_id
  FROM faculties f
  WHERE $2 = 'FACULTY'
    AND f.id = $3

  UNION ALL

  SELECT 'SYSTEM'::text, NULL::uuid, d.tenant_id
  FROM departments d
  WHERE $2 = 'DEPARTMENT'
    AND d.id = $3

  UNION ALL

  SELECT 'TENANT'::text, d.tenant_id, d.tenant_id
  FROM departments d
  WHERE $2 = 'DEPARTMENT'
    AND d.id = $3

  UNION ALL

  SELECT 'FACULTY'::text, d.faculty_id, d.tenant_id
  FROM departments d
  WHERE $2 = 'DEPARTMENT'
    AND d.id = $3

  UNION ALL

  SELECT 'DEPARTMENT'::text, d.id, d.tenant_id
  FROM departments d
  WHERE $2 = 'DEPARTMENT'
    AND d.id = $3

  UNION ALL

  SELECT 'SYSTEM'::text, NULL::uuid, p.tenant_id
  FROM projects p
  WHERE $2 = 'PROJECT'
    AND p.id = $3

  UNION ALL

  SELECT 'TENANT'::text, p.tenant_id, p.tenant_id
  FROM projects p
  WHERE $2 = 'PROJECT'
    AND p.id = $3

  UNION ALL

  SELECT 'FACULTY'::text, p.faculty_id, p.tenant_id
  FROM projects p
  WHERE $2 = 'PROJECT'
    AND p.id = $3

  UNION ALL

  SELECT 'PROJECT'::text, p.id, p.tenant_id
  FROM projects p
  WHERE $2 = 'PROJECT'
    AND p.id = $3
)
`

func NewRBACRepo(db *pgxpool.Pool) *RBACRepo {
	return &RBACRepo{db: db}
}

// SetCacheInvalidator sets the function called after role mutations to invalidate
// the RBAC cache for the affected user. This is called from the module wiring layer.
func (r *RBACRepo) SetCacheInvalidator(fn CacheInvalidatorFunc) {
	r.cacheInvalidatorFn = fn
}

// invalidateCache calls the cache invalidator if one is set.
func (r *RBACRepo) invalidateCache(ctx context.Context, userID uuid.UUID) {
	if r.cacheInvalidatorFn != nil {
		r.cacheInvalidatorFn(ctx, userID)
	}
}

func (r *RBACRepo) HasPermission(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope, now time.Time) (bool, error) {
	const q = resolvedScopesCTE + `
SELECT EXISTS (
  SELECT 1
  FROM role_assignments ra
  JOIN users u ON u.id = ra.user_id
  JOIN resolved_scopes rs
    ON rs.scope_type = ra.scope_type
   AND rs.tenant_id = ra.tenant_id
   AND (
     (rs.scope_id IS NULL AND ra.scope_id IS NULL)
     OR (rs.scope_id = ra.scope_id)
   )
  JOIN role_permissions rp ON rp.role_id = ra.role_id
  JOIN permissions p ON p.id = rp.permission_id
  WHERE ra.user_id = $1
    AND u.tenant_id = ra.tenant_id
    AND (ra.expires_at IS NULL OR ra.expires_at > $4)
    AND p.code = $5
) AS ok;
`
	var ok bool
	err := r.db.QueryRow(ctx, q, userID, string(scope.Type), scope.ID, now, permissionCode).Scan(&ok)
	return ok, err
}

func (r *RBACRepo) ListPermissionCodes(ctx context.Context, userID uuid.UUID, scope rbac.Scope, now time.Time) ([]string, error) {
	const q = resolvedScopesCTE + `
SELECT DISTINCT p.code
FROM role_assignments ra
JOIN users u ON u.id = ra.user_id
JOIN resolved_scopes rs
  ON rs.scope_type = ra.scope_type
 AND rs.tenant_id = ra.tenant_id
 AND (
   (rs.scope_id IS NULL AND ra.scope_id IS NULL)
   OR (rs.scope_id = ra.scope_id)
 )
JOIN role_permissions rp ON rp.role_id = ra.role_id
JOIN permissions p ON p.id = rp.permission_id
WHERE ra.user_id = $1
  AND u.tenant_id = ra.tenant_id
  AND (ra.expires_at IS NULL OR ra.expires_at > $4)
ORDER BY p.code ASC;
`
	rows, err := r.db.Query(ctx, q, userID, string(scope.Type), scope.ID, now)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, 16)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, rows.Err()
}

func (r *RBACRepo) GrantRoleByCode(
	ctx context.Context,
	userID uuid.UUID,
	roleCode string,
	scope rbac.Scope,
	expiresAt *time.Time,
) error {
	// ensure role exists
	var roleID uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM roles WHERE code=$1`, roleCode).Scan(&roleID)
	if err != nil {
		return err
	}

	var tenantID uuid.UUID
	if err := r.db.QueryRow(ctx, `SELECT tenant_id FROM users WHERE id = $1`, userID).Scan(&tenantID); err != nil {
		return err
	}

	if scope.ID == nil {
		_, err = r.db.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id, expires_at)
VALUES ($1, $2, $3, $4, NULL, $5)
ON CONFLICT (tenant_id, user_id, role_id, scope_type) WHERE scope_id IS NULL
DO UPDATE SET expires_at = EXCLUDED.expires_at;
`, tenantID, userID, roleID, string(scope.Type), expiresAt)
		if err != nil {
			return err
		}
	} else {
		_, err = r.db.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id, expires_at)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (tenant_id, user_id, role_id, scope_type, scope_id) WHERE scope_id IS NOT NULL
DO UPDATE SET expires_at = EXCLUDED.expires_at;
`, tenantID, userID, roleID, string(scope.Type), scope.ID, expiresAt)
		if err != nil {
			return err
		}
	}

	// Atomically invalidate cache after role mutation
	r.invalidateCache(ctx, userID)
	return nil
}
