package postgres

import (
	"context"

	"idsai-core-up/internal/services/projectflow"

	"github.com/google/uuid"
)

func (r *ProjectFlowRepo) ListProjectRoleCodes(ctx context.Context, userID, projectID uuid.UUID) ([]string, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT r.code
FROM role_assignments ra
JOIN roles r ON r.id = ra.role_id
WHERE ra.tenant_id = $1
  AND ra.user_id = $2
  AND ra.scope_type = 'PROJECT'
  AND ra.scope_id = $3
ORDER BY r.code ASC;
`
	rows, err := r.db.Query(ctx, q, tenantID, userID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]string, 0, 8)
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) ReplaceAssignableRoles(
	ctx context.Context,
	userID, projectID uuid.UUID,
	assignableCodes []string,
	wantCodes []string,
) error {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Delete all current assignable roles for this user in this project.
	if len(assignableCodes) > 0 {
		_, err = tx.Exec(ctx, `
DELETE FROM role_assignments
WHERE tenant_id = $1
  AND user_id = $2
  AND scope_type = 'PROJECT'
  AND scope_id = $3
  AND role_id IN (
    SELECT id FROM roles WHERE code = ANY($4)
  );
`, tenantID, userID, projectID, assignableCodes)
		if err != nil {
			return err
		}
	}

	// Insert wanted roles.
	for _, code := range wantCodes {
		_, err = tx.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
SELECT $1, $2, r.id, 'PROJECT', $3
FROM roles r
WHERE r.code = $4
ON CONFLICT DO NOTHING;
`, tenantID, userID, projectID, code)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ProjectFlowRepo) GetMemberStatusAndCreator(ctx context.Context, userID, projectID uuid.UUID) (string, uuid.UUID, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return "", uuid.Nil, err
	}

	const q = `
SELECT COALESCE(pm.status, ''), p.created_by
FROM projects p
LEFT JOIN project_members pm
  ON pm.project_id = p.id
  AND pm.tenant_id = p.tenant_id
  AND pm.user_id = $2
WHERE p.tenant_id = $1
  AND p.id = $3;
`
	var status string
	var creatorID uuid.UUID
	err = r.db.QueryRow(ctx, q, tenantID, userID, projectID).Scan(&status, &creatorID)
	if err != nil {
		return "", uuid.Nil, mapProjectFlowErr(err)
	}

	// If the user is the creator, they are always ACTIVE even without a project_members row.
	if status == "" && userID == creatorID {
		status = "ACTIVE"
	}

	if status == "" {
		return "", uuid.Nil, projectflow.ErrNotFound
	}

	return status, creatorID, nil
}
