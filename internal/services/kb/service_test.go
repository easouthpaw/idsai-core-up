package kb

import (
	"context"
	"testing"
	"time"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	createCategoryFn func(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error)
	updateCategoryFn func(ctx context.Context, tenantID, categoryID uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error)
	deleteCategoryFn func(ctx context.Context, tenantID, categoryID uuid.UUID) error
	categoryHasFn    func(ctx context.Context, categoryID uuid.UUID) (bool, error)
	listTreeFn       func(ctx context.Context, tenantID uuid.UUID) ([]domain.KBCategory, error)
	createArticleFn  func(ctx context.Context, tenantID, categoryID, authorID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) (uuid.UUID, error)
	updateArticleFn  func(ctx context.Context, tenantID, articleID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) error
	deleteArticleFn  func(ctx context.Context, tenantID, articleID uuid.UUID) error
	getArticleFn     func(ctx context.Context, tenantID, articleID uuid.UUID) (domain.KBArticle, error)
	listArticlesFn   func(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, status, search, tag string, limit, offset int) ([]domain.KBArticleListItem, int, error)
	syncTagsFn       func(ctx context.Context, tenantID, articleID uuid.UUID, tagNames []string) error
	listTagsFn       func(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.KBTag, error)
}

func (f *fakeRepo) CreateCategory(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error) {
	if f.createCategoryFn != nil {
		return f.createCategoryFn(ctx, tenantID, parentID, title, slug, sortOrder)
	}
	return domain.KBCategory{}, nil
}

func (f *fakeRepo) UpdateCategory(ctx context.Context, tenantID, categoryID uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error) {
	if f.updateCategoryFn != nil {
		return f.updateCategoryFn(ctx, tenantID, categoryID, title, slug, sortOrder)
	}
	return domain.KBCategory{}, nil
}

func (f *fakeRepo) DeleteCategory(ctx context.Context, tenantID, categoryID uuid.UUID) error {
	if f.deleteCategoryFn != nil {
		return f.deleteCategoryFn(ctx, tenantID, categoryID)
	}
	return nil
}

func (f *fakeRepo) CategoryHasChildren(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	if f.categoryHasFn != nil {
		return f.categoryHasFn(ctx, categoryID)
	}
	return false, nil
}

func (f *fakeRepo) ListCategoryTree(ctx context.Context, tenantID uuid.UUID) ([]domain.KBCategory, error) {
	if f.listTreeFn != nil {
		return f.listTreeFn(ctx, tenantID)
	}
	return nil, nil
}

func (f *fakeRepo) CreateArticle(ctx context.Context, tenantID, categoryID, authorID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) (uuid.UUID, error) {
	if f.createArticleFn != nil {
		return f.createArticleFn(ctx, tenantID, categoryID, authorID, title, slug, content, status, publishedAt)
	}
	return uuid.Nil, nil
}

func (f *fakeRepo) UpdateArticle(ctx context.Context, tenantID, articleID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) error {
	if f.updateArticleFn != nil {
		return f.updateArticleFn(ctx, tenantID, articleID, title, slug, content, status, publishedAt)
	}
	return nil
}

func (f *fakeRepo) DeleteArticle(ctx context.Context, tenantID, articleID uuid.UUID) error {
	if f.deleteArticleFn != nil {
		return f.deleteArticleFn(ctx, tenantID, articleID)
	}
	return nil
}

func (f *fakeRepo) GetArticleByID(ctx context.Context, tenantID, articleID uuid.UUID) (domain.KBArticle, error) {
	if f.getArticleFn != nil {
		return f.getArticleFn(ctx, tenantID, articleID)
	}
	return domain.KBArticle{}, nil
}

func (f *fakeRepo) ListArticles(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, status, search, tag string, limit, offset int) ([]domain.KBArticleListItem, int, error) {
	if f.listArticlesFn != nil {
		return f.listArticlesFn(ctx, tenantID, categoryID, status, search, tag, limit, offset)
	}
	return nil, 0, nil
}

func (f *fakeRepo) SyncArticleTags(ctx context.Context, tenantID, articleID uuid.UUID, tagNames []string) error {
	if f.syncTagsFn != nil {
		return f.syncTagsFn(ctx, tenantID, articleID, tagNames)
	}
	return nil
}

func (f *fakeRepo) ListPopularTags(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.KBTag, error) {
	if f.listTagsFn != nil {
		return f.listTagsFn(ctx, tenantID, limit)
	}
	return nil, nil
}

func TestCreateArticleRetriesWithUniqueSlug(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()
	categoryID := uuid.New()
	authorID := uuid.New()
	articleID := uuid.New()

	var gotSlugs []string
	repo := &fakeRepo{
		createArticleFn: func(ctx context.Context, tenantIDArg, categoryIDArg, authorIDArg uuid.UUID, title, slug, content, status string, publishedAt *time.Time) (uuid.UUID, error) {
			gotSlugs = append(gotSlugs, slug)
			if slug == "first" {
				return uuid.Nil, domain.ErrKBArticleSlugConflict
			}
			require.Equal(t, "first-2", slug)
			return articleID, nil
		},
		getArticleFn: func(ctx context.Context, tenantIDArg, articleIDArg uuid.UUID) (domain.KBArticle, error) {
			return domain.KBArticle{ID: articleIDArg, Title: "First", Slug: "first-2"}, nil
		},
	}

	svc := NewService(repo)
	article, err := svc.CreateArticle(ctx, CreateArticleInput{
		TenantID:   tenantID,
		AuthorID:   authorID,
		CategoryID: categoryID,
		Title:      "First",
		Content:    "body",
		Status:     "DRAFT",
	})

	require.NoError(t, err)
	require.Equal(t, articleID, article.ID)
	require.Equal(t, "first-2", article.Slug)
	require.Equal(t, []string{"first", "first-2"}, gotSlugs)
}

func TestCreateCategoryRetriesWithUniqueSlug(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New()

	var gotSlugs []string
	repo := &fakeRepo{
		createCategoryFn: func(ctx context.Context, tenantIDArg uuid.UUID, parentID *uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error) {
			gotSlugs = append(gotSlugs, slug)
			if slug == "guide" {
				return domain.KBCategory{}, domain.ErrKBCategorySlugConflict
			}
			require.Equal(t, "guide-2", slug)
			return domain.KBCategory{ID: uuid.New(), Title: title, Slug: slug}, nil
		},
	}

	svc := NewService(repo)
	cat, err := svc.CreateCategory(ctx, CreateCategoryInput{
		TenantID: tenantID,
		Title:    "Guide",
	})

	require.NoError(t, err)
	require.Equal(t, "guide-2", cat.Slug)
	require.Equal(t, []string{"guide", "guide-2"}, gotSlugs)
}
