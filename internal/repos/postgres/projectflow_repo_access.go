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

func (r *ProjectFlowRepo) ListProjectAccessRoles(ctx context.Context, projectID uuid.UUID) ([]projectflow.AccessCatalogItem, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	const q = `
SELECT r.code,
       par.code AS display_code,
       par.name,
       par.description,
       COALESCE(array_agg(p.code ORDER BY p.code) FILTER (WHERE p.code IS NOT NULL), '{}') AS permission_codes
FROM project_access_roles par
JOIN roles r ON r.id = par.role_id
LEFT JOIN role_permissions rp ON rp.role_id = r.id
LEFT JOIN permissions p ON p.id = rp.permission_id
WHERE par.tenant_id = $1
  AND par.project_id = $2
GROUP BY r.code, par.code, par.name, par.description, par.created_at
ORDER BY par.created_at ASC, par.name ASC;
`
	rows, err := r.db.Query(ctx, q, tenantID, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]projectflow.AccessCatalogItem, 0, 8)
	for rows.Next() {
		var item projectflow.AccessCatalogItem
		if err := rows.Scan(&item.Code, &item.DisplayCode, &item.Name, &item.Description, &item.PermissionCodes); err != nil {
			return nil, err
		}
		item.Custom = true
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *ProjectFlowRepo) CreateProjectAccessRole(
	ctx context.Context,
	projectID, createdBy uuid.UUID,
	roleCode, displayCode, name, description string,
	permissionCodes []string,
) (projectflow.AccessCatalogItem, error) {
	tenantID, err := tenantIDFromContext(ctx)
	if err != nil {
		return projectflow.AccessCatalogItem{}, err
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return projectflow.AccessCatalogItem{}, err
	}
	defer tx.Rollback(ctx)

	var roleID uuid.UUID
	if err := tx.QueryRow(ctx, `
INSERT INTO roles(code, name)
VALUES ($1, $2)
RETURNING id;
`, roleCode, name).Scan(&roleID); err != nil {
		return projectflow.AccessCatalogItem{}, err
	}

	if _, err := tx.Exec(ctx, `
INSERT INTO project_access_roles(tenant_id, project_id, role_id, code, name, description, created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7);
`, tenantID, projectID, roleID, displayCode, name, description, createdBy); err != nil {
		return projectflow.AccessCatalogItem{}, err
	}

	if len(permissionCodes) > 0 {
		if _, err := tx.Exec(ctx, `
INSERT INTO role_permissions(role_id, permission_id)
SELECT $1, p.id
FROM permissions p
WHERE p.code = ANY($2)
ON CONFLICT DO NOTHING;
`, roleID, permissionCodes); err != nil {
			return projectflow.AccessCatalogItem{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return projectflow.AccessCatalogItem{}, err
	}

	return projectflow.AccessCatalogItem{
		Code:            roleCode,
		DisplayCode:     displayCode,
		Name:            name,
		Description:     description,
		PermissionCodes: append([]string(nil), permissionCodes...),
		Custom:          true,
	}, nil
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
