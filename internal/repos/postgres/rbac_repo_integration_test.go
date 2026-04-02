//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func seedProjectScopeFixture(t *testing.T, ctx context.Context, pool *pgxpool.Pool, emailPrefix string) (uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID) {
	t.Helper()

	tenantID := uuid.New()
	facultyID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()

	_, err := pool.Exec(ctx, `
INSERT INTO tenants(id, code, name, status)
VALUES ($1, $2, $3, 'ACTIVE');
`, tenantID, "RBAC_T_"+tenantID.String()[:8], "RBAC Tenant")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO faculties(id, tenant_id, code, name)
VALUES ($1, $2, $3, $4);
`, facultyID, tenantID, "RBAC_F_"+facultyID.String()[:8], "RBAC Faculty")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO users(id, tenant_id, email, password_hash, status)
VALUES ($1, $2, $3, 'integration-hash', 'ACTIVE');
`, userID, tenantID, emailPrefix+"-"+userID.String()+"@example.local")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO projects(id, tenant_id, title, description, status, is_public, created_by, faculty_id, visibility)
VALUES ($1, $2, 'RBAC Integration Project', 'demo', 'ACTIVE', FALSE, $3, $4, 'FACULTY');
`, projectID, tenantID, userID, facultyID)
	require.NoError(t, err)

	return tenantID, facultyID, userID, projectID
}

func TestRBACRepo_Integration_HasPermission_ProjectScope(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewRBACRepo(pool)

	tenantID, _, userID, projectID := seedProjectScopeFixture(t, ctx, pool, "rbac-project")

	// Assign TEAM_LEAD role in PROJECT scope to user
	_, err = pool.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
VALUES (
  $1,
  $2,
  (SELECT id FROM roles WHERE code='TEAM_LEAD'),
  'PROJECT',
  $3
);
`, tenantID, userID, projectID)
	require.NoError(t, err)

	now := time.Now()

	ok, err := repo.HasPermission(ctx, userID, "task.create", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	}, now)
	require.NoError(t, err)
	require.True(t, ok)

	// Negative case: wrong project scope id
	otherProjectID := uuid.New()
	ok, err = repo.HasPermission(ctx, userID, "task.create", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &otherProjectID,
	}, now)
	require.NoError(t, err)
	require.False(t, ok)

	// Negative case: permission outside the role grant
	ok, err = repo.HasPermission(ctx, userID, "admin.manage_rbac", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	}, now)
	require.NoError(t, err)
	require.False(t, ok)
}

func TestRBACRepo_Integration_ExpiredAssignmentDenied(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewRBACRepo(pool)

	tenantID, _, userID, projectID := seedProjectScopeFixture(t, ctx, pool, "rbac-expired")

	// Expired assignment
	_, err = pool.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id, expires_at)
VALUES (
  $1,
  $2,
  (SELECT id FROM roles WHERE code='MEMBER'),
  'PROJECT',
  $3,
  $4
);
`, tenantID, userID, projectID, time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	ok, err := repo.HasPermission(ctx, userID, "task.close", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	}, time.Now())
	require.NoError(t, err)
	require.False(t, ok)
}

func TestRBACRepo_Integration_GrantRoleByCode_AllowsPermission(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewRBACRepo(pool)

	_, _, userID, projectID := seedProjectScopeFixture(t, ctx, pool, "rbac-grant")
	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}

	// grant TEAM_LEAD in this project
	require.NoError(t, repo.GrantRoleByCode(ctx, userID, "TEAM_LEAD", scope, nil))

	// TEAM_LEAD should have task.create (per seeds)
	ok, err := repo.HasPermission(ctx, userID, "task.create", scope, time.Now())
	require.NoError(t, err)
	require.True(t, ok)
}

func TestRBACRepo_Integration_HasPermission_ProjectScopeInheritsFacultyScope(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewRBACRepo(pool)

	tenantID := uuid.New()
	facultyID := uuid.New()
	userID := uuid.New()
	projectID := uuid.New()

	_, err = pool.Exec(ctx, `
INSERT INTO tenants(id, code, name, status)
VALUES ($1, $2, $3, 'ACTIVE');
`, tenantID, "RBAC_T_"+tenantID.String()[:8], "RBAC Tenant")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO faculties(id, tenant_id, code, name)
VALUES ($1, $2, $3, $4);
`, facultyID, tenantID, "RBAC_F_"+facultyID.String()[:8], "RBAC Faculty")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO users(id, tenant_id, email, password_hash, status)
VALUES ($1, $2, $3, 'integration-hash', 'ACTIVE');
`, userID, tenantID, "rbac-hierarchy-"+userID.String()+"@example.local")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO projects(id, tenant_id, title, description, status, is_public, created_by, faculty_id, visibility)
VALUES ($1, $2, 'Hierarchy Project', 'demo', 'RECRUITMENT', FALSE, $3, $4, 'FACULTY');
`, projectID, tenantID, userID, facultyID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
VALUES (
  $1,
  $2,
  (SELECT id FROM roles WHERE code='STUDENT'),
  'FACULTY',
  $3
);
`, tenantID, userID, facultyID)
	require.NoError(t, err)

	ok, err := repo.HasPermission(ctx, userID, "member.apply", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	}, time.Now())
	require.NoError(t, err)
	require.True(t, ok)
}

func TestRBACRepo_Integration_GrantRoleByCode_UpsertDoesNotDuplicate(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewRBACRepo(pool)

	_, _, userID, projectID := seedProjectScopeFixture(t, ctx, pool, "rbac-upsert")

	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}
	firstExpiry := time.Now().Add(2 * time.Hour).UTC()
	secondExpiry := firstExpiry.Add(3 * time.Hour)

	require.NoError(t, repo.GrantRoleByCode(ctx, userID, "TEAM_LEAD", scope, &firstExpiry))
	require.NoError(t, repo.GrantRoleByCode(ctx, userID, "TEAM_LEAD", scope, &secondExpiry))

	var (
		count     int
		expiresAt time.Time
	)
	err = pool.QueryRow(ctx, `
SELECT COUNT(*), COALESCE(MAX(expires_at), 'epoch'::timestamptz)
FROM role_assignments
WHERE user_id = $1
  AND role_id = (SELECT id FROM roles WHERE code = 'TEAM_LEAD')
  AND scope_type = 'PROJECT'
  AND scope_id = $2;
	`, userID, projectID).Scan(&count, &expiresAt)
	require.NoError(t, err)
	require.Equal(t, 1, count)
	require.WithinDuration(t, secondExpiry, expiresAt, time.Millisecond)
}

func TestRBACRepo_Integration_ProjectRolesHaveProjectView(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewRBACRepo(pool)

	now := time.Now()

	cases := []struct {
		name     string
		roleCode string
	}{
		{name: "team lead", roleCode: "TEAM_LEAD"},
		{name: "member", roleCode: "MEMBER"},
		{name: "project professor", roleCode: "PROJECT_PROFESSOR"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, userID, seededProjectID := seedProjectScopeFixture(t, ctx, pool, "rbac-project-view-"+tc.roleCode)
			scope := rbac.Scope{Type: rbac.ScopeProject, ID: &seededProjectID}
			require.NoError(t, repo.GrantRoleByCode(ctx, userID, tc.roleCode, scope, nil))

			ok, err := repo.HasPermission(ctx, userID, "project.view", scope, now)
			require.NoError(t, err)
			require.True(t, ok)
		})
	}
}

func TestRBACRepo_Integration_ProfessorFacultyScopeCanReadProjectDetails(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewRBACRepo(pool)

	tenantID := uuid.New()
	facultyID := uuid.New()
	professorID := uuid.New()
	projectID := uuid.New()

	_, err = pool.Exec(ctx, `
INSERT INTO tenants(id, code, name, status)
VALUES ($1, $2, $3, 'ACTIVE');
`, tenantID, "RBAC_PF_T_"+tenantID.String()[:8], "Professor Faculty Tenant")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO faculties(id, tenant_id, code, name)
VALUES ($1, $2, $3, $4);
`, facultyID, tenantID, "RBAC_PF_F_"+facultyID.String()[:8], "Professor Faculty")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO users(id, tenant_id, email, password_hash, status)
VALUES ($1, $2, $3, 'integration-hash', 'ACTIVE');
`, professorID, tenantID, "rbac-prof-faculty-"+professorID.String()+"@example.local")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO projects(id, tenant_id, title, description, status, is_public, created_by, faculty_id, visibility)
VALUES ($1, $2, 'Professor Faculty Project', 'demo', 'ACTIVE', FALSE, $3, $4, 'FACULTY');
`, projectID, tenantID, professorID, facultyID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO role_assignments(tenant_id, user_id, role_id, scope_type, scope_id)
VALUES (
  $1,
  $2,
  (SELECT id FROM roles WHERE code='PROFESSOR'),
  'FACULTY',
  $3
);
`, tenantID, professorID, facultyID)
	require.NoError(t, err)

	for _, permission := range []string{"project.view", "task.view", "grading.view"} {
		ok, err := repo.HasPermission(ctx, professorID, permission, rbac.Scope{
			Type: rbac.ScopeProject,
			ID:   &projectID,
		}, time.Now())
		require.NoError(t, err, permission)
		require.True(t, ok, permission)
	}
}
