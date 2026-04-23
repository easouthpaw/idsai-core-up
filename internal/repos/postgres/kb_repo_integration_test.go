//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"idsai-core-up/internal/db"
	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/repos/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"
)

func TestKBRepo_Integration_CategoryArticleAndTagsFlow(t *testing.T) {
	dsn := os.Getenv("DATABASE_URL")
	require.NotEmpty(t, dsn)

	ctx := context.Background()
	pool, err := db.NewPool(ctx, dsn)
	require.NoError(t, err)
	defer pool.Close()

	repo := postgres.NewKBRepo(pool)
	tenantID, authorID := seedKBAuthor(t, ctx, pool)

	root, err := repo.CreateCategory(ctx, tenantID, nil, "Guides", "guides", 10)
	require.NoError(t, err)
	require.Equal(t, tenantID, root.TenantID)
	require.Nil(t, root.ParentID)

	_, err = repo.CreateCategory(ctx, tenantID, nil, "Duplicate", "guides", 20)
	require.ErrorIs(t, err, domain.ErrKBCategorySlugConflict)

	child, err := repo.CreateCategory(ctx, tenantID, &root.ID, "Go", "go", 5)
	require.NoError(t, err)
	require.NotNil(t, child.ParentID)
	require.Equal(t, root.ID, *child.ParentID)

	child, err = repo.UpdateCategory(ctx, tenantID, child.ID, "Go Basics", "go-basics", 3)
	require.NoError(t, err)
	require.Equal(t, "Go Basics", child.Title)
	require.Equal(t, "go-basics", child.Slug)
	require.Equal(t, 3, child.SortOrder)

	_, err = repo.UpdateCategory(ctx, tenantID, uuid.New(), "Missing", "missing", 0)
	require.ErrorIs(t, err, domain.ErrKBNotFound)

	tree, err := repo.ListCategoryTree(ctx, tenantID)
	require.NoError(t, err)
	require.Len(t, tree, 2)
	require.Equal(t, root.ID, tree[0].ID)
	require.Equal(t, child.ID, tree[1].ID)

	publishedAt := time.Now().UTC().Truncate(time.Second)
	articleID, err := repo.CreateArticle(ctx, tenantID, child.ID, authorID, "Go Testing", "go-testing", "Интеграционные тесты на Go", "PUBLISHED", &publishedAt)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, articleID)

	_, err = repo.CreateArticle(ctx, tenantID, child.ID, authorID, "Duplicate", "go-testing", "content", "DRAFT", nil)
	require.ErrorIs(t, err, domain.ErrKBArticleSlugConflict)

	require.NoError(t, repo.SyncArticleTags(ctx, tenantID, articleID, []string{"Go", "Testing", " ", "go"}))
	tags, err := repo.ListTagNamesByArticle(ctx, articleID)
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"go", "Testing"}, tags)

	article, err := repo.GetArticleByID(ctx, tenantID, articleID)
	require.NoError(t, err)
	require.Equal(t, "Go Testing", article.Title)
	require.Equal(t, "kb author", article.AuthorName)
	require.ElementsMatch(t, []string{"Testing", "go"}, article.Tags)

	hasChildren, err := repo.CategoryHasChildren(ctx, root.ID)
	require.NoError(t, err)
	require.True(t, hasChildren)
	hasChildren, err = repo.CategoryHasChildren(ctx, child.ID)
	require.NoError(t, err)
	require.True(t, hasChildren)

	items, total, err := repo.ListArticles(ctx, tenantID, &child.ID, "PUBLISHED", "тесты", "go", 0, 0)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, items, 1)
	require.Equal(t, articleID, items[0].ID)
	require.ElementsMatch(t, []string{"Testing", "go"}, items[0].Tags)

	items, total, err = repo.ListArticles(ctx, tenantID, nil, "", "", "", 200, 0)
	require.NoError(t, err)
	require.Equal(t, 1, total)
	require.Len(t, items, 1)

	popularTags, err := repo.ListPopularTags(ctx, tenantID, 0)
	require.NoError(t, err)
	require.Len(t, popularTags, 2)

	require.NoError(t, repo.UpdateArticle(ctx, tenantID, articleID, "Go Testing Updated", "go-testing-updated", "new content", "DRAFT", nil))
	article, err = repo.GetArticleByID(ctx, tenantID, articleID)
	require.NoError(t, err)
	require.Equal(t, "Go Testing Updated", article.Title)
	require.Equal(t, "DRAFT", article.Status)

	err = repo.UpdateArticle(ctx, tenantID, uuid.New(), "Missing", "missing", "", "DRAFT", nil)
	require.ErrorIs(t, err, domain.ErrKBNotFound)

	err = repo.DeleteArticle(ctx, tenantID, uuid.New())
	require.ErrorIs(t, err, domain.ErrKBNotFound)
	require.NoError(t, repo.DeleteArticle(ctx, tenantID, articleID))

	_, err = repo.GetArticleByID(ctx, tenantID, articleID)
	require.ErrorIs(t, err, domain.ErrKBNotFound)

	require.NoError(t, repo.DeleteCategory(ctx, tenantID, root.ID))
	err = repo.DeleteCategory(ctx, tenantID, root.ID)
	require.ErrorIs(t, err, domain.ErrKBNotFound)
}

func seedKBAuthor(t *testing.T, ctx context.Context, pool *pgxpool.Pool, prefix ...string) (uuid.UUID, uuid.UUID) {
	t.Helper()

	label := "kb"
	if len(prefix) > 0 && prefix[0] != "" {
		label = prefix[0]
	}

	tenantID := uuid.New()
	userID := uuid.New()
	facultyID := uuid.New()
	departmentID := uuid.New()

	_, err := pool.Exec(ctx, `
INSERT INTO tenants(id, code, name, status)
VALUES ($1, $2, $3, 'ACTIVE');
`, tenantID, "KB_T_"+tenantID.String()[:8], "KB Tenant")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO faculties(id, tenant_id, code, name)
VALUES ($1, $2, $3, $4);
`, facultyID, tenantID, "KB_F_"+facultyID.String()[:8], "KB Faculty")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO departments(id, tenant_id, faculty_id, code, name)
VALUES ($1, $2, $3, $4, $5);
`, departmentID, tenantID, facultyID, "KB_D_"+departmentID.String()[:8], "KB Department")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO users(id, tenant_id, email, password_hash, status, avatar_key)
VALUES ($1, $2, $3, 'integration-hash', 'ACTIVE', 'avatars/kb.jpg');
`, userID, tenantID, fmt.Sprintf("%s-author-%s@example.local", label, userID.String()))
	require.NoError(t, err)

	_, err = pool.Exec(ctx, `
INSERT INTO user_profiles(tenant_id, user_id, full_name, faculty_id, department_id)
VALUES ($1, $2, 'kb author', $3, $4);
`, tenantID, userID, facultyID, departmentID)
	require.NoError(t, err)

	return tenantID, userID
}
