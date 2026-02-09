//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/repos/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestTasksRepo_Integration_ClaimOnlyMatchingPosition(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	projectsRepo := postgres.NewProjectsRepo(pool)
	positionsRepo := postgres.NewPositionsRepo(pool)
	membersRepo := postgres.NewMembersRepo(pool)
	tasksRepo := postgres.NewTasksRepo(pool)

	// faculty id
	var facultyID uuid.UUID
	err = pool.QueryRow(ctx, `SELECT id FROM faculties WHERE code='IDSAI_ENU'`).Scan(&facultyID)
	require.NoError(t, err)

	// create project
	creator := uuid.New()
	projectID, err := projectsRepo.Create(ctx, "P", "D", facultyID, "FACULTY", nil, creator)
	require.NoError(t, err)

	// positions
	backendPos, err := positionsRepo.Create(ctx, projectID, "BACKEND", "Backend", 1)
	require.NoError(t, err)
	frontendPos, err := positionsRepo.Create(ctx, projectID, "FRONTEND", "Frontend", 1)
	require.NoError(t, err)

	// create task for BACKEND
	taskID, err := tasksRepo.Create(ctx, projectID, "API", "Do API", backendPos, creator, nil)
	require.NoError(t, err)

	// user1 is FRONTEND member
	user1 := uuid.New()
	_, err = membersRepo.Apply(ctx, projectID, user1)
	require.NoError(t, err)
	require.NoError(t, membersRepo.Approve(ctx, projectID, user1, frontendPos))

	// user2 is BACKEND member
	user2 := uuid.New()
	_, err = membersRepo.Apply(ctx, projectID, user2)
	require.NoError(t, err)
	require.NoError(t, membersRepo.Approve(ctx, projectID, user2, backendPos))

	// frontend user cannot claim backend task
	err = tasksRepo.Claim(ctx, projectID, taskID, user1)
	require.Error(t, err)
	require.True(t, err == domain.ErrForbidden)

	// backend user can claim
	require.NoError(t, tasksRepo.Claim(ctx, projectID, taskID, user2))
}
