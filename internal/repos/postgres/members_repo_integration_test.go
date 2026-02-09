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

func TestMembersRepo_Integration_ApplyApproveGet(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	projectsRepo := postgres.NewProjectsRepo(pool)
	positionsRepo := postgres.NewPositionsRepo(pool)
	membersRepo := postgres.NewMembersRepo(pool)

	// faculty id
	var facultyID uuid.UUID
	err = pool.QueryRow(ctx, `SELECT id FROM faculties WHERE code='IDSAI_ENU'`).Scan(&facultyID)
	require.NoError(t, err)

	// create project
	creator := uuid.New()
	projectID, err := projectsRepo.Create(ctx, "P", "D", facultyID, "FACULTY", nil, creator)
	require.NoError(t, err)

	// create a BACKEND position
	posID, err := positionsRepo.Create(ctx, projectID, "BACKEND", "Backend Dev", 1)
	require.NoError(t, err)

	// user applies
	user := uuid.New()
	_, err = membersRepo.Apply(ctx, projectID, user)
	require.NoError(t, err)

	m, err := membersRepo.GetByProjectAndUser(ctx, projectID, user)
	require.NoError(t, err)
	require.Equal(t, "APPLIED", string(m.Status))
	require.Nil(t, m.PositionID)

	// approve with position
	require.NoError(t, membersRepo.Approve(ctx, projectID, user, posID))

	m, err = membersRepo.GetByProjectAndUser(ctx, projectID, user)
	require.NoError(t, err)
	require.Equal(t, "ACTIVE", string(m.Status))
	require.NotNil(t, m.PositionID)
	require.Equal(t, posID, *m.PositionID)
	require.NotNil(t, m.JoinedAt)
}
