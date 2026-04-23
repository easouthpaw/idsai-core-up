package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/http/dto"
	"idsai-core-up/internal/services/kb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

type kbHandlerRepo struct {
	categories     []domain.KBCategory
	articles       []domain.KBArticleListItem
	total          int
	article        domain.KBArticle
	tags           []domain.KBTag
	lastListStatus string
}

func (f *kbHandlerRepo) CreateCategory(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error) {
	cat := domain.KBCategory{
		ID:        uuid.New(),
		TenantID:  tenantID,
		ParentID:  parentID,
		Title:     title,
		Slug:      slug,
		SortOrder: sortOrder,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	if len(f.categories) > 0 && f.categories[0].ID != uuid.Nil {
		cat.ID = f.categories[0].ID
	}
	f.categories = []domain.KBCategory{cat}
	return cat, nil
}

func (f *kbHandlerRepo) UpdateCategory(ctx context.Context, tenantID, categoryID uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error) {
	cat := domain.KBCategory{
		ID:        categoryID,
		TenantID:  tenantID,
		Title:     title,
		Slug:      slug,
		SortOrder: sortOrder,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}
	f.categories = []domain.KBCategory{cat}
	return cat, nil
}

func (f *kbHandlerRepo) DeleteCategory(ctx context.Context, tenantID, categoryID uuid.UUID) error {
	return nil
}

func (f *kbHandlerRepo) CategoryHasChildren(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	return false, nil
}

func (f *kbHandlerRepo) ListCategoryTree(ctx context.Context, tenantID uuid.UUID) ([]domain.KBCategory, error) {
	return f.categories, nil
}

func (f *kbHandlerRepo) CreateArticle(ctx context.Context, tenantID, categoryID, authorID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) (uuid.UUID, error) {
	if f.article.ID == uuid.Nil {
		f.article.ID = uuid.New()
	}
	f.article.TenantID = tenantID
	f.article.CategoryID = categoryID
	f.article.AuthorID = authorID
	f.article.Title = title
	f.article.Slug = slug
	f.article.Content = content
	f.article.Status = status
	f.article.PublishedAt = publishedAt
	return f.article.ID, nil
}

func (f *kbHandlerRepo) UpdateArticle(ctx context.Context, tenantID, articleID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) error {
	f.article.ID = articleID
	f.article.TenantID = tenantID
	f.article.Title = title
	f.article.Slug = slug
	f.article.Content = content
	f.article.Status = status
	f.article.PublishedAt = publishedAt
	return nil
}

func (f *kbHandlerRepo) DeleteArticle(ctx context.Context, tenantID, articleID uuid.UUID) error {
	return nil
}

func (f *kbHandlerRepo) GetArticleByID(ctx context.Context, tenantID, articleID uuid.UUID) (domain.KBArticle, error) {
	return f.article, nil
}

func (f *kbHandlerRepo) ListArticles(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, status, search, tag string, limit, offset int) ([]domain.KBArticleListItem, int, error) {
	f.lastListStatus = status
	return f.articles, f.total, nil
}

func (f *kbHandlerRepo) SyncArticleTags(ctx context.Context, tenantID, articleID uuid.UUID, tagNames []string) error {
	f.article.Tags = tagNames
	return nil
}

func (f *kbHandlerRepo) ListPopularTags(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.KBTag, error) {
	return f.tags, nil
}

func TestKBHandlerListCategories_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 1, 10, 0, 0, 0, time.UTC)
	repo := &kbHandlerRepo{
		categories: []domain.KBCategory{
			{
				ID:        uuid.MustParse("11111111-1111-1111-1111-111111111111"),
				TenantID:  uuid.MustParse("22222222-2222-2222-2222-222222222222"),
				Title:     "Backend",
				Slug:      "backend",
				SortOrder: 1,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
	}

	handler := NewKBHandler(kb.NewService(repo))
	router := gin.New()
	router.GET("/kb/categories", handler.ListCategories)

	req := httptest.NewRequest(http.MethodGet, "/kb/categories", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.KBCategoryResponsesFromDomain(repo.categories)), rec.Body.String())
}

func TestKBHandlerListArticles_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 1, 11, 0, 0, 0, time.UTC)
	repo := &kbHandlerRepo{
		articles: []domain.KBArticleListItem{
			{
				ID:         uuid.MustParse("33333333-3333-3333-3333-333333333333"),
				CategoryID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
				AuthorID:   uuid.MustParse("55555555-5555-5555-5555-555555555555"),
				Title:      "DTO Guide",
				Slug:       "dto-guide",
				Status:     "PUBLISHED",
				CreatedAt:  now,
				UpdatedAt:  now,
				AuthorName: "Doc Writer",
				Tags:       []string{"dto", "http"},
			},
		},
		total: 1,
	}

	handler := NewKBHandler(kb.NewService(repo))
	router := gin.New()
	router.GET("/kb/articles", handler.ListArticles)

	req := httptest.NewRequest(http.MethodGet, "/kb/articles", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.ListArticlesResponse{Items: dto.KBArticleListItemResponsesFromDomain(repo.articles), Total: 1}), rec.Body.String())
}

func TestKBHandlerGetArticle_UsesTransportDTO(t *testing.T) {
	gin.SetMode(gin.TestMode)

	now := time.Date(2026, 4, 1, 12, 0, 0, 0, time.UTC)
	articleID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	repo := &kbHandlerRepo{
		article: domain.KBArticle{
			ID:           articleID,
			TenantID:     uuid.MustParse("77777777-7777-7777-7777-777777777777"),
			CategoryID:   uuid.MustParse("88888888-8888-8888-8888-888888888888"),
			AuthorID:     uuid.MustParse("99999999-9999-9999-9999-999999999999"),
			Title:        "DTO Article",
			Slug:         "dto-article",
			Content:      "transport boundary",
			Status:       "PUBLISHED",
			CreatedAt:    now,
			UpdatedAt:    now,
			AuthorName:   "Author",
			AuthorAvatar: "/img/a.png",
			CategoryPath: "Docs / DTO",
			Tags:         []string{"dto"},
		},
	}

	handler := NewKBHandler(kb.NewService(repo))
	router := gin.New()
	router.GET("/kb/articles/:id", handler.GetArticle)

	req := httptest.NewRequest(http.MethodGet, "/kb/articles/"+articleID.String(), nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, mustJSON(t, dto.KBArticleResponseFromDomain(repo.article)), rec.Body.String())
}
