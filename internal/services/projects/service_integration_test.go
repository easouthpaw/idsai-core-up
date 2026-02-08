//go:build integration

package projects_test

import (
	"context"
	"os"
	"testing"
	"time"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/services/projects"
	"idsai-core-up/internal/services/rbac"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestProjectsService_Integration_CreateProject_GrantsTeamLeadPermissions(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	projectsRepo := postgres.NewProjectsRepo(pool)
	rbacRepo := postgres.NewRBACRepo(pool)

	svc := projects.NewService(projectsRepo, rbacRepo)

	var facultyID uuid.UUID
	err = pool.QueryRow(ctx, `SELECT id FROM faculties WHERE code='IDSAI_ENU'`).Scan(&facultyID)
	require.NoError(t, err)

	creator := uuid.New()

	projectID, err := svc.CreateProject(ctx, "RBAC Demo Project", "desc", facultyID, "FACULTY", nil, creator)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, projectID)

	scope := rbac.Scope{Type: rbac.ScopeProject, ID: &projectID}

	ok, err := rbacRepo.HasPermission(ctx, creator, "task.create", scope, time.Now())
	require.NoError(t, err)
	require.True(t, ok, "creator should have TEAM_LEAD permissions in the created project")
}
