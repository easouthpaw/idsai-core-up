//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/requestctx"
	"idsai-core-up/internal/services/projects"

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
	var tenantID uuid.UUID
	err = pool.QueryRow(ctx, `SELECT id, tenant_id FROM faculties WHERE code='IDSAI_ENU'`).Scan(&facultyID, &tenantID)
	require.NoError(t, err)

	createdBy := uuid.New()
	ctx = requestctx.WithIdentity(ctx, createdBy, tenantID, facultyID, uuid.New())

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

func TestProjectsRepo_Integration_GetByID_DeniesCrossTenant(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	baseCtx := context.Background()
	pool, err := db.NewPool(baseCtx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewProjectsRepo(pool)

	tenantA := uuid.New()
	tenantB := uuid.New()
	facultyA := uuid.New()
	facultyB := uuid.New()
	userA := uuid.New()
	userB := uuid.New()

	_, err = pool.Exec(baseCtx, `
INSERT INTO tenants(id, code, name, status)
VALUES ($1, $2, $3, 'ACTIVE'), ($4, $5, $6, 'ACTIVE');
`,
		tenantA, "TENANT_A_"+tenantA.String()[:8], "Tenant A",
		tenantB, "TENANT_B_"+tenantB.String()[:8], "Tenant B",
	)
	require.NoError(t, err)

	_, err = pool.Exec(baseCtx, `
INSERT INTO faculties(id, tenant_id, code, name)
VALUES ($1, $2, $3, $4), ($5, $6, $7, $8);
`,
		facultyA, tenantA, "FAC_A_"+facultyA.String()[:8], "Faculty A",
		facultyB, tenantB, "FAC_B_"+facultyB.String()[:8], "Faculty B",
	)
	require.NoError(t, err)

	_, err = pool.Exec(baseCtx, `
INSERT INTO users(id, tenant_id, email, password_hash, status)
VALUES
  ($1, $2, $3, 'integration-hash', 'ACTIVE'),
  ($4, $5, $6, 'integration-hash', 'ACTIVE');
`,
		userA, tenantA, fmt.Sprintf("projects-a-%s@example.local", userA.String()),
		userB, tenantB, fmt.Sprintf("projects-b-%s@example.local", userB.String()),
	)
	require.NoError(t, err)

	ctxA := requestctx.WithIdentity(baseCtx, userA, tenantA, facultyA, uuid.New())
	ctxB := requestctx.WithIdentity(baseCtx, userB, tenantB, facultyB, uuid.New())

	projectA, err := repo.Create(ctxA, "Tenant A Project", "A", facultyA, "FACULTY", nil, userA)
	require.NoError(t, err)
	projectB, err := repo.Create(ctxB, "Tenant B Project", "B", facultyB, "FACULTY", nil, userB)
	require.NoError(t, err)
	require.NotEqual(t, projectA, projectB)

	_, err = repo.GetByID(ctxA, projectB)
	require.ErrorIs(t, err, projects.ErrNotFound)
}

func TestProjectsRepo_Integration_ListByCreator_IncludesActiveMemberProjects(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewProjectsRepo(pool)

	var facultyID uuid.UUID
	var tenantID uuid.UUID
	err = pool.QueryRow(ctx, `SELECT id, tenant_id FROM faculties WHERE code='IDSAI_ENU'`).Scan(&facultyID, &tenantID)
	require.NoError(t, err)

	creatorID := uuid.New()
	memberID := uuid.New()

	_, err = pool.Exec(ctx, `
INSERT INTO users(id, tenant_id, email, password_hash, status)
VALUES
  ($1, $2, $3, 'integration-hash', 'ACTIVE'),
  ($4, $2, $5, 'integration-hash', 'ACTIVE');
`,
		creatorID, tenantID, fmt.Sprintf("creator-%s@example.local", creatorID.String()),
		memberID, fmt.Sprintf("member-%s@example.local", memberID.String()),
	)
	require.NoError(t, err)

	creatorCtx := requestctx.WithIdentity(ctx, creatorID, tenantID, facultyID, uuid.New())
	projectID, err := repo.Create(creatorCtx, "Shared Project", "team flow", facultyID, "FACULTY", nil, creatorID)
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO project_members(tenant_id, project_id, user_id, status, joined_at)
VALUES ($1, $2, $3, 'ACTIVE', now())
ON CONFLICT (project_id, user_id)
DO UPDATE SET status = 'ACTIVE', joined_at = now();
`, tenantID, projectID, memberID)
	require.NoError(t, err)

	memberCtx := requestctx.WithIdentity(ctx, memberID, tenantID, facultyID, uuid.New())
	items, err := repo.ListByCreator(memberCtx, memberID)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, projectID, items[0].ID)

	_, err = pool.Exec(ctx, `
UPDATE project_members
SET status = 'REMOVED', joined_at = NULL
WHERE tenant_id = $1
  AND project_id = $2
  AND user_id = $3;
`, tenantID, projectID, memberID)
	require.NoError(t, err)

	items, err = repo.ListByCreator(memberCtx, memberID)
	require.NoError(t, err)
	require.Empty(t, items)
}
