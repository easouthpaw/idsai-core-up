//go:build integration

package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/repos/postgres"
	"idsai-core-up/internal/requestctx"
	"idsai-core-up/internal/services/projects"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestProjectsRepo_Integration_ProjectListsPermissionsMediaAndSummary(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	baseCtx := context.Background()
	pool, err := db.NewPool(baseCtx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	authRepo := postgres.NewAuthRepo(pool)
	repo := postgres.NewProjectsRepo(pool)
	scope := seedAuthScope(t, baseCtx, pool)
	groupID, err := authRepo.CreateGroupInDepartment(baseCtx, scope.tenantID, scope.facultyID, scope.departmentID, scope.groupCode1, 101)
	require.NoError(t, err)

	ownerID := seedAuthUserProfile(t, baseCtx, pool, scope, "projects owner", groupID)
	memberID := seedAuthUserProfile(t, baseCtx, pool, scope, "projects member", groupID)
	professorID := seedAuthUserProfile(t, baseCtx, pool, scope, "projects professor", groupID)
	ctx := requestctx.WithIdentity(baseCtx, ownerID, scope.tenantID, scope.facultyID, scope.departmentID)

	projectID, err := repo.CreateWithLeadRole(ctx, "Repo Project", "Integration coverage", scope.facultyID, "GROUP", &groupID, ownerID)
	require.NoError(t, err)
	publicID, err := repo.Create(ctx, "Public Repo Project", "Public coverage", scope.facultyID, "PUBLIC", nil, ownerID)
	require.NoError(t, err)

	project, err := repo.GetByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, projectID, project.ID)
	require.Equal(t, ownerID, project.CreatedBy)
	require.Equal(t, scope.facultyID, project.FacultyID)
	require.False(t, project.IsPublic)
	require.NotNil(t, project.GroupID)
	require.Equal(t, groupID, *project.GroupID)
	require.GreaterOrEqual(t, project.DefaultCoverVariant, 1)
	require.LessOrEqual(t, project.DefaultCoverVariant, 6)

	_, err = repo.GetByID(ctx, uuid.New())
	require.ErrorIs(t, err, projects.ErrNotFound)

	ok, err := repo.HasProjectPermission(ctx, ownerID, projectID, "grading.view")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = repo.HasResolvedProjectPermission(ctx, ownerID, projectID, "project.view")
	require.NoError(t, err)
	require.True(t, ok)
	ok, err = repo.HasProjectPermission(ctx, memberID, projectID, "project.edit")
	require.NoError(t, err)
	require.False(t, ok)

	imageUpdatedAt := time.Date(2026, 4, 21, 12, 0, 0, 0, time.UTC)
	require.NoError(t, repo.SetProjectImage(ctx, projectID, "projects/covers/repo.jpg", imageUpdatedAt))
	project, err = repo.GetByID(ctx, projectID)
	require.NoError(t, err)
	require.Equal(t, "projects/covers/repo.jpg", project.ImageKey)
	require.NotNil(t, project.ImageUpdatedAt)
	require.NoError(t, repo.ClearProjectImage(ctx, projectID))
	project, err = repo.GetByID(ctx, projectID)
	require.NoError(t, err)
	require.Empty(t, project.ImageKey)
	require.NotNil(t, project.ImageUpdatedAt)
	err = repo.SetProjectImage(ctx, uuid.New(), "missing.jpg", imageUpdatedAt)
	require.ErrorIs(t, err, projects.ErrNotFound)
	err = repo.ClearProjectImage(ctx, uuid.New())
	require.ErrorIs(t, err, projects.ErrNotFound)

	foundGroupID, err := repo.FindGroupIDByCode(ctx, scope.facultyID, scope.groupCode1)
	require.NoError(t, err)
	require.Equal(t, groupID, foundGroupID)
	_, err = repo.FindGroupIDByCode(ctx, scope.facultyID, "NOPE-404")
	require.ErrorIs(t, err, projects.ErrGroupNotFound)

	groups, err := repo.ListGroupsByFaculty(ctx, scope.facultyID)
	require.NoError(t, err)
	require.Contains(t, groupIDs(groups), groupID)

	_, err = pool.Exec(baseCtx, `
INSERT INTO project_members(tenant_id, project_id, user_id, status, joined_at)
VALUES ($1, $2, $3, 'ACTIVE', now())
ON CONFLICT (project_id, user_id)
DO UPDATE SET status = 'ACTIVE', joined_at = now();
`, scope.tenantID, projectID, memberID)
	require.NoError(t, err)
	require.NoError(t, grantProjectFlowRole(baseCtx, pool, scope.tenantID, professorID, "PROJECT_PROFESSOR", "PROJECT", &projectID))

	ownerProjects, err := repo.ListByCreator(ctx, ownerID)
	require.NoError(t, err)
	require.Contains(t, projectIDs(ownerProjects), projectID)
	require.Contains(t, projectIDs(ownerProjects), publicID)

	memberProjects, err := repo.ListByCreator(ctx, memberID)
	require.NoError(t, err)
	require.Contains(t, projectIDs(memberProjects), projectID)

	professorProjects, err := repo.ListByCreator(ctx, professorID)
	require.NoError(t, err)
	require.Contains(t, projectIDs(professorProjects), projectID)

	facultyProjects, err := repo.ListByFaculty(ctx, scope.facultyID, memberID)
	require.NoError(t, err)
	require.Contains(t, projectIDs(facultyProjects), projectID)
	require.Contains(t, projectIDs(facultyProjects), publicID)

	publicProjects, err := repo.ListPublic(ctx, uuid.Nil)
	require.NoError(t, err)
	require.Contains(t, projectIDs(publicProjects), publicID)
	require.NotContains(t, projectIDs(publicProjects), projectID)

	reviewedAt := time.Date(2026, 4, 21, 13, 0, 0, 0, time.UTC)
	criterionMetID := uuid.New()
	criterionMissID := uuid.New()
	_, err = pool.Exec(baseCtx, `
UPDATE projects
SET professor_id = $1,
    status = 'COMPLETED',
    retake_count = 1
WHERE id = $2;
`, professorID, projectID)
	require.NoError(t, err)
	_, err = pool.Exec(baseCtx, `
INSERT INTO project_criteria(id, tenant_id, project_id, title, description, weight, created_by)
VALUES
  ($1, $3, $4, 'Core works', 'main path', 80, $5),
  ($2, $3, $4, 'Docs ready', 'docs path', 20, $5);
`, criterionMetID, criterionMissID, scope.tenantID, projectID, ownerID)
	require.NoError(t, err)
	_, err = pool.Exec(baseCtx, `
INSERT INTO project_criterion_reviews(tenant_id, project_id, criterion_id, professor_id, is_met, comment, created_at, updated_at)
VALUES
  ($1, $2, $3, $5, TRUE, 'done', $6, $6),
  ($1, $2, $4, $5, FALSE, 'missing docs', $6, $6);
`, scope.tenantID, projectID, criterionMetID, criterionMissID, professorID, reviewedAt)
	require.NoError(t, err)

	summary, err := repo.GetProjectReviewSummary(ctx, projectID)
	require.NoError(t, err)
	require.NotNil(t, summary)
	require.Equal(t, 2, summary.Total)
	require.Equal(t, 1, summary.Met)
	require.Equal(t, 75, summary.PassPercent)
	require.Equal(t, "3.8", summary.Score)
	require.Equal(t, "Auth projects professor", summary.Reviewer)
	require.NotNil(t, summary.ReviewedAt)

	missingSummary, err := repo.GetProjectReviewSummary(ctx, uuid.New())
	require.NoError(t, err)
	require.Nil(t, missingSummary)
}

func projectIDs(items []domain.Project) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		out = append(out, item.ID)
	}
	return out
}

func groupIDs(items []projects.Group) []uuid.UUID {
	out := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		out = append(out, item.ID)
	}
	return out
}
