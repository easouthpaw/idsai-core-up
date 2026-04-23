package handlers

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"idsai-core-up/internal/domain"
	"idsai-core-up/internal/services/kb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestKBHandlerRequireEditor_AllowsProfessorAndAdmin(t *testing.T) {
	t.Run("professor", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("isProfessor", true)

		h := &KBHandler{}
		require.True(t, h.requireEditor(c))
	})

	t.Run("admin", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Set("isAdmin", true)

		h := &KBHandler{}
		require.True(t, h.requireEditor(c))
	})
}

func TestKBHandlerContextIdentityUsesMiddlewareKeys(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	tenantID := uuid.New()
	userID := uuid.New()
	c.Set("tenantID", tenantID)
	c.Set("userID", userID)

	h := &KBHandler{}
	require.Equal(t, tenantID, h.tenantID(c))
	require.Equal(t, userID, h.userID(c))
}

func TestKBHandlerCRUDAndUploadRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tenantID := uuid.New()
	userID := uuid.New()
	categoryID := uuid.New()
	articleID := uuid.New()
	now := time.Now().UTC()
	repo := &kbHandlerRepo{
		categories: []domain.KBCategory{
			{
				ID:        categoryID,
				TenantID:  tenantID,
				Title:     "Guides",
				Slug:      "guides",
				SortOrder: 1,
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		article: domain.KBArticle{
			ID:          articleID,
			TenantID:    tenantID,
			CategoryID:  categoryID,
			AuthorID:    userID,
			Title:       "Testing",
			Slug:        "testing",
			Content:     "# Testing",
			Status:      "PUBLISHED",
			CreatedAt:   now,
			UpdatedAt:   now,
			AuthorName:  "Professor",
			Tags:        []string{"go"},
			PublishedAt: &now,
		},
		articles: []domain.KBArticleListItem{
			{
				ID:          articleID,
				CategoryID:  categoryID,
				AuthorID:    userID,
				Title:       "Testing",
				Slug:        "testing",
				Status:      "PUBLISHED",
				CreatedAt:   now,
				UpdatedAt:   now,
				PublishedAt: &now,
				AuthorName:  "Professor",
				Tags:        []string{"go"},
			},
		},
		total: 1,
		tags:  []domain.KBTag{{ID: uuid.New(), Name: "go"}},
	}
	handler := NewKBHandler(kb.NewService(repo))

	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenantID", tenantID)
		c.Set("userID", userID)
		c.Set("isProfessor", true)
		c.Next()
	})
	router.GET("/categories", handler.ListCategories)
	router.POST("/categories", handler.CreateCategory)
	router.PUT("/categories/:id", handler.UpdateCategory)
	router.DELETE("/categories/:id", handler.DeleteCategory)
	router.GET("/articles", handler.ListArticles)
	router.POST("/articles", handler.CreateArticle)
	router.GET("/articles/:id", handler.GetArticle)
	router.PUT("/articles/:id", handler.UpdateArticle)
	router.DELETE("/articles/:id", handler.DeleteArticle)
	router.POST("/articles/upload", handler.UploadArticle)
	router.GET("/tags", handler.ListTags)

	requireStatus(t, router, http.MethodGet, "/categories", "", http.StatusOK)
	requireStatus(t, router, http.MethodPost, "/categories", `{"title":"Guides","sort_order":1}`, http.StatusCreated)
	requireStatus(t, router, http.MethodPut, "/categories/"+categoryID.String(), `{"title":"Guides Updated","sort_order":2}`, http.StatusOK)
	requireStatus(t, router, http.MethodGet, "/articles?category_id="+categoryID.String()+"&status=draft&search=test&tag=go&limit=5&offset=1", "", http.StatusOK)
	require.Equal(t, "DRAFT", repo.lastListStatus)
	requireStatus(t, router, http.MethodPost, "/articles", `{"category_id":"`+categoryID.String()+`","title":"Testing","content":"# Testing","status":"published","tags":["go"]}`, http.StatusCreated)
	requireStatus(t, router, http.MethodGet, "/articles/"+articleID.String(), "", http.StatusOK)
	requireStatus(t, router, http.MethodPut, "/articles/"+articleID.String(), `{"title":"Testing 2","content":"body","status":"draft","tags":["go"]}`, http.StatusOK)
	requireStatus(t, router, http.MethodDelete, "/articles/"+articleID.String(), "", http.StatusNoContent)
	requireStatus(t, router, http.MethodDelete, "/categories/"+categoryID.String(), "", http.StatusNoContent)
	requireStatus(t, router, http.MethodGet, "/tags", "", http.StatusOK)

	var uploadBody bytes.Buffer
	writer := multipart.NewWriter(&uploadBody)
	require.NoError(t, writer.WriteField("category_id", categoryID.String()))
	require.NoError(t, writer.WriteField("status", "published"))
	require.NoError(t, writer.WriteField("tags", "go, testing"))
	file, err := writer.CreateFormFile("file", "article.md")
	require.NoError(t, err)
	_, err = file.Write([]byte("# Uploaded\ncontent"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	req := httptest.NewRequest(http.MethodPost, "/articles/upload", &uploadBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func TestKBHandlerRejectsInvalidAndNonEditorRequests(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewKBHandler(kb.NewService(&kbHandlerRepo{}))
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set("tenantID", uuid.New())
		c.Set("userID", uuid.New())
		c.Next()
	})
	router.POST("/categories", handler.CreateCategory)
	router.GET("/articles", handler.ListArticles)
	router.GET("/articles/:id", handler.GetArticle)

	requireStatus(t, router, http.MethodPost, "/categories", `{"title":"No access"}`, http.StatusForbidden)
	requireStatus(t, router, http.MethodGet, "/articles?category_id=not-a-uuid", "", http.StatusBadRequest)

	draftRepo := &kbHandlerRepo{article: domain.KBArticle{ID: uuid.New(), Status: "DRAFT"}}
	draftHandler := NewKBHandler(kb.NewService(draftRepo))
	draftRouter := gin.New()
	draftRouter.Use(func(c *gin.Context) {
		c.Set("tenantID", uuid.New())
		c.Set("userID", uuid.New())
		c.Next()
	})
	draftRouter.GET("/articles/:id", draftHandler.GetArticle)
	requireStatus(t, draftRouter, http.MethodGet, "/articles/"+draftRepo.article.ID.String(), "", http.StatusNotFound)
}

func requireStatus(t *testing.T, router *gin.Engine, method, target, body string, want int) {
	t.Helper()
	req := httptest.NewRequest(method, target, bytes.NewBufferString(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	require.Equal(t, want, rec.Code, rec.Body.String())
}
