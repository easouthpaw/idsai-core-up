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
	"github.com/stretchr/testify/require"
)

func TestRBACRepo_Integration_HasPermission_ProjectScope(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewRBACRepo(pool)

	userID := uuid.New()
	projectID := uuid.New()

	// Assign TEAM_LEAD role in PROJECT scope to user
	_, err = pool.Exec(ctx, `
INSERT INTO role_assignments(user_id, role_id, scope_type, scope_id)
VALUES (
  $1,
  (SELECT id FROM roles WHERE code='TEAM_LEAD'),
  'PROJECT',
  $2
);
`, userID, projectID)
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

	// Negative case: unknown permission
	ok, err = repo.HasPermission(ctx, userID, "task.delete", rbac.Scope{
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

	userID := uuid.New()
	projectID := uuid.New()

	// Expired assignment
	_, err = pool.Exec(ctx, `
INSERT INTO role_assignments(user_id, role_id, scope_type, scope_id, expires_at)
VALUES (
  $1,
  (SELECT id FROM roles WHERE code='MEMBER'),
  'PROJECT',
  $2,
  $3
);
`, userID, projectID, time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	ok, err := repo.HasPermission(ctx, userID, "task.close", rbac.Scope{
		Type: rbac.ScopeProject,
		ID:   &projectID,
	}, time.Now())
	require.NoError(t, err)
	require.False(t, ok)
}
