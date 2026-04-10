package kb

import (
	"context"
	"testing"
	"time"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestSlugHelpersAndTagDeduplication(t *testing.T) {
	require.Equal(t, "guide-to-rbac", slugify(" Guide to RBAC ", 200))
	require.NotEmpty(t, slugify("!!!", 8))
	require.Equal(t, "guide-2", slugWithSuffix("guide", 200, 1))
	require.Equal(t, "7891", slugWithSuffix("guide", 4, 1234567890))

	tags := dedupeTagNames([]string{
		" Go ",
		"go",
		"RBAC",
		"",
		"this-tag-is-way-too-long-to-be-kept-because-it-exceeds-the-sixty-character-maximum",
	})
	require.Equal(t, []string{"Go", "RBAC"}, tags)
}

func TestUpdateCategoryAndArticle_RetrySlugConflict(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	categoryID := uuid.New()
	articleID := uuid.New()

	var categorySlugs []string
	var articleSlugs []string
	var syncedTags []string

	repo := &fakeRepo{
		updateCategoryFn: func(ctx context.Context, tenantIDArg, categoryIDArg uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error) {
			categorySlugs = append(categorySlugs, slug)
			if slug == "guides" {
				return domain.KBCategory{}, domain.ErrKBCategorySlugConflict
			}
			return domain.KBCategory{ID: categoryIDArg, Title: title, Slug: slug}, nil
		},
		updateArticleFn: func(ctx context.Context, tenantIDArg, articleIDArg uuid.UUID, title, slug, content, status string, publishedAt *time.Time) error {
			articleSlugs = append(articleSlugs, slug)
			if slug == "deep-dive" {
				return domain.ErrKBArticleSlugConflict
			}
			require.Equal(t, "PUBLISHED", status)
			require.NotNil(t, publishedAt)
			return nil
		},
		syncTagsFn: func(ctx context.Context, tenantIDArg, articleIDArg uuid.UUID, tagNames []string) error {
			syncedTags = append([]string(nil), tagNames...)
			return nil
		},
		getArticleFn: func(ctx context.Context, tenantIDArg, articleIDArg uuid.UUID) (domain.KBArticle, error) {
			return domain.KBArticle{ID: articleIDArg, Title: "Deep Dive", Slug: "deep-dive-2"}, nil
		},
	}

	svc := NewService(repo)
	category, err := svc.UpdateCategory(ctx, UpdateCategoryInput{
		TenantID:   tenantID,
		CategoryID: categoryID,
		Title:      "Guides",
		SortOrder:  10,
	})
	require.NoError(t, err)
	require.Equal(t, "guides-2", category.Slug)
	require.Equal(t, []string{"guides", "guides-2"}, categorySlugs)

	article, err := svc.UpdateArticle(ctx, UpdateArticleInput{
		TenantID:  tenantID,
		ArticleID: articleID,
		Title:     "Deep Dive",
		Content:   " content ",
		Status:    "published",
		Tags:      []string{" Go ", "go", "RBAC"},
	})
	require.NoError(t, err)
	require.Equal(t, "deep-dive-2", article.Slug)
	require.Equal(t, []string{"deep-dive", "deep-dive-2"}, articleSlugs)
	require.Equal(t, []string{"Go", "RBAC"}, syncedTags)
}

func TestDeleteAndListPassthroughMethods(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	categoryID := uuid.New()
	articleID := uuid.New()
	category := domain.KBCategory{ID: categoryID, Title: "Guides"}
	article := domain.KBArticle{ID: articleID, Title: "First"}
	tags := []domain.KBTag{{ID: uuid.New(), Name: "rbac"}}

	repo := &fakeRepo{
		categoryHasFn: func(ctx context.Context, categoryIDArg uuid.UUID) (bool, error) {
			return true, nil
		},
		listTreeFn: func(ctx context.Context, tenantIDArg uuid.UUID) ([]domain.KBCategory, error) {
			return []domain.KBCategory{category}, nil
		},
		deleteArticleFn: func(ctx context.Context, tenantIDArg, articleIDArg uuid.UUID) error {
			require.Equal(t, articleID, articleIDArg)
			return nil
		},
		getArticleFn: func(ctx context.Context, tenantIDArg, articleIDArg uuid.UUID) (domain.KBArticle, error) {
			return article, nil
		},
		listArticlesFn: func(ctx context.Context, tenantIDArg uuid.UUID, categoryIDArg *uuid.UUID, status, search, tag string, limit, offset int) ([]domain.KBArticleListItem, int, error) {
			require.Equal(t, "PUBLISHED", status)
			require.Equal(t, "rbac", tag)
			return []domain.KBArticleListItem{{ID: articleID, Title: "First"}}, 1, nil
		},
		listTagsFn: func(ctx context.Context, tenantIDArg uuid.UUID, limit int) ([]domain.KBTag, error) {
			require.Equal(t, 30, limit)
			return tags, nil
		},
	}

	svc := NewService(repo)
	err := svc.DeleteCategory(ctx, tenantID, categoryID)
	require.ErrorIs(t, err, ErrCategoryNotEmpty)

	tree, err := svc.GetCategoryTree(ctx, tenantID)
	require.NoError(t, err)
	require.Len(t, tree, 1)
	require.Equal(t, categoryID, tree[0].ID)

	err = svc.DeleteArticle(ctx, tenantID, articleID)
	require.NoError(t, err)

	gotArticle, err := svc.GetArticle(ctx, tenantID, articleID)
	require.NoError(t, err)
	require.Equal(t, articleID, gotArticle.ID)

	items, total, err := svc.ListArticles(ctx, tenantID, &categoryID, "PUBLISHED", "guide", "rbac", 20, 0)
	require.NoError(t, err)
	require.Len(t, items, 1)
	require.Equal(t, 1, total)

	popularTags, err := svc.ListPopularTags(ctx, tenantID)
	require.NoError(t, err)
	require.Equal(t, tags, popularTags)
}

func TestUploadMarkdownFile_UsesHeadingAsTitle(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	categoryID := uuid.New()
	authorID := uuid.New()
	articleID := uuid.New()

	repo := &fakeRepo{
		createArticleFn: func(ctx context.Context, tenantIDArg, categoryIDArg, authorIDArg uuid.UUID, title, slug, content, status string, publishedAt *time.Time) (uuid.UUID, error) {
			require.Equal(t, "RBAC Benchmarking", title)
			require.Equal(t, "# RBAC Benchmarking\n\nDetails", content)
			require.Equal(t, "DRAFT", status)
			return articleID, nil
		},
		getArticleFn: func(ctx context.Context, tenantIDArg, articleIDArg uuid.UUID) (domain.KBArticle, error) {
			return domain.KBArticle{ID: articleIDArg, Title: "RBAC Benchmarking"}, nil
		},
	}

	svc := NewService(repo)
	article, err := svc.UploadMarkdownFile(ctx, UploadMarkdownInput{
		TenantID:   tenantID,
		AuthorID:   authorID,
		CategoryID: categoryID,
		Filename:   "draft.md",
		Content:    []byte("# RBAC Benchmarking\n\nDetails"),
		Status:     "DRAFT",
		Tags:       []string{"rbac"},
	})

	require.NoError(t, err)
	require.Equal(t, articleID, article.ID)
	require.Equal(t, "RBAC Benchmarking", article.Title)
}
