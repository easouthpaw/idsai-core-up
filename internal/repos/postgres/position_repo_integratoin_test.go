//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/repos/postgres"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestPositionsRepo_Integration_CreateAndList(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	projectsRepo := postgres.NewProjectsRepo(pool)
	positionsRepo := postgres.NewPositionsRepo(pool)

	// faculty id
	var facultyID uuid.UUID
	err = pool.QueryRow(ctx, `SELECT id FROM faculties WHERE code='IDSAI_ENU'`).Scan(&facultyID)
	require.NoError(t, err)

	// create project
	createdBy := uuid.New()
	projectID, err := projectsRepo.Create(ctx, "P", "D", facultyID, "FACULTY", nil, createdBy)
	require.NoError(t, err)

	// create positions
	_, err = positionsRepo.Create(ctx, projectID, "BACKEND", "Backend Developer", 2)
	require.NoError(t, err)
	_, err = positionsRepo.Create(ctx, projectID, "FRONTEND", "Frontend Developer", 1)
	require.NoError(t, err)

	list, err := positionsRepo.ListByProject(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, list, 2)
	require.Equal(t, "BACKEND", list[0].Code)
}
