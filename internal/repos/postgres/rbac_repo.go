package postgres

import (
	"context"
	"time"

	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RBACRepo struct {
	db *pgxpool.Pool
}

func NewRBACRepo(db *pgxpool.Pool) *RBACRepo {
	return &RBACRepo{db: db}
}

func (r *RBACRepo) HasPermission(ctx context.Context, userID uuid.UUID, permissionCode string, scope rbac.Scope, now time.Time) (bool, error) {
	const q = `
SELECT EXISTS (
  SELECT 1
  FROM role_assignments ra
  JOIN users u ON u.id = ra.user_id
  JOIN role_permissions rp ON rp.role_id = ra.role_id
  JOIN permissions p ON p.id = rp.permission_id
  WHERE ra.user_id = $1
    AND u.tenant_id = ra.tenant_id
    AND ra.scope_type = $2
    AND (
      ($3::uuid IS NULL AND ra.scope_id IS NULL)
      OR (ra.scope_id = $3::uuid)
    )
    AND (ra.expires_at IS NULL OR ra.expires_at > $4)
    AND p.code = $5
) AS ok;
`
	var ok bool
	err := r.db.QueryRow(ctx, q, userID, string(scope.Type), scope.ID, now, permissionCode).Scan(&ok)
	return ok, err
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

	_, err = r.db.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id, expires_at)
VALUES ((SELECT tenant_id FROM users WHERE id = $1), $1, $2, $3, $4, $5);
`, userID, roleID, string(scope.Type), scope.ID, expiresAt)

	return err
}
