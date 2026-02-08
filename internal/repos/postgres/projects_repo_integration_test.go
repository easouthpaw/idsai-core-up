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

func TestProjectsRepo_Integration_CreateAndGet(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewProjectsRepo(pool)

	// ✅ ВОТ СЮДА добавляем: достаем facultyID из таблицы faculties
	var facultyID uuid.UUID
	err = pool.QueryRow(ctx, `SELECT id FROM faculties WHERE code='IDSAI_ENU'`).Scan(&facultyID)
	require.NoError(t, err)

	createdBy := uuid.New()

	// ✅ Create теперь принимает facultyID + visibility + groupID
	id, err := repo.Create(ctx, "My Project", "Demo", facultyID, "FACULTY", nil, createdBy)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, id)

	p, err := repo.GetByID(ctx, id)
	require.NoError(t, err)

	require.Equal(t, id, p.ID)
	require.Equal(t, "My Project", p.Title)
	require.Equal(t, "Demo", p.Description)
	require.Equal(t, createdBy, p.CreatedBy)
	require.Equal(t, "DRAFT", string(p.Status))
	require.False(t, p.IsPublic)
	require.Nil(t, p.ProfessorID)

	require.Equal(t, facultyID, p.FacultyID)
	require.Equal(t, "FACULTY", p.Visibility)
	require.Nil(t, p.GroupID)
}
