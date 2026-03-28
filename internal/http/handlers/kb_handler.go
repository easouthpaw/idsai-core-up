package handlers

import (
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"idsai-core-up/internal/services/kb"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type KBHandler struct {
	svc *kb.Service
}

func NewKBHandler(svc *kb.Service) *KBHandler {
	return &KBHandler{svc: svc}
}

func (h *KBHandler) requireEditor(c *gin.Context) bool {
	isAdmin, _ := c.Get("is_admin")
	isProfessor, _ := c.Get("is_professor")
	if isAdmin == true || isProfessor == true {
		return true
	}
	c.JSON(http.StatusForbidden, gin.H{"error": "only professors and admins can manage the knowledge base"})
	return false
}

func (h *KBHandler) tenantID(c *gin.Context) uuid.UUID {
	raw, _ := c.Get("tenant_id")
	if id, ok := raw.(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			return parsed
		}
	}
	return uuid.Nil
}

func (h *KBHandler) userID(c *gin.Context) uuid.UUID {
	raw, _ := c.Get("sub")
	if id, ok := raw.(string); ok {
		if parsed, err := uuid.Parse(id); err == nil {
			return parsed
		}
	}
	return uuid.Nil
}

// ── Categories ──────────────────────────────────────────────

func (h *KBHandler) ListCategories(c *gin.Context) {
	tenantID := h.tenantID(c)
	cats, err := h.svc.GetCategoryTree(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cats == nil {
		c.JSON(http.StatusOK, []struct{}{})
		return
	}
	c.JSON(http.StatusOK, cats)
}

func (h *KBHandler) CreateCategory(c *gin.Context) {
	if !h.requireEditor(c) {
		return
	}
	var body struct {
		ParentID  *string `json:"parent_id"`
		Title     string  `json:"title"`
		SortOrder int     `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var parentID *uuid.UUID
	if body.ParentID != nil && *body.ParentID != "" {
		parsed, err := uuid.Parse(*body.ParentID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid parent_id"})
			return
		}
		parentID = &parsed
	}

	cat, err := h.svc.CreateCategory(c.Request.Context(), kb.CreateCategoryInput{
		TenantID:  h.tenantID(c),
		ParentID:  parentID,
		Title:     body.Title,
		SortOrder: body.SortOrder,
	})
	if err != nil {
		if errors.Is(err, kb.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, cat)
}

func (h *KBHandler) UpdateCategory(c *gin.Context) {
	if !h.requireEditor(c) {
		return
	}
	catID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}
	var body struct {
		Title     string `json:"title"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	cat, err := h.svc.UpdateCategory(c.Request.Context(), kb.UpdateCategoryInput{
		TenantID:   h.tenantID(c),
		CategoryID: catID,
		Title:      body.Title,
		SortOrder:  body.SortOrder,
	})
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, kb.ErrInvalidInput) {
			status = http.StatusBadRequest
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, cat)
}

func (h *KBHandler) DeleteCategory(c *gin.Context) {
	if !h.requireEditor(c) {
		return
	}
	catID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category id"})
		return
	}
	if err := h.svc.DeleteCategory(c.Request.Context(), h.tenantID(c), catID); err != nil {
		if errors.Is(err, kb.ErrCategoryNotEmpty) {
			c.JSON(http.StatusConflict, gin.H{"error": "category is not empty; delete articles and subcategories first"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ── Articles ────────────────────────────────────────────────

func (h *KBHandler) ListArticles(c *gin.Context) {
	tenantID := h.tenantID(c)
	var categoryID *uuid.UUID
	if raw := c.Query("category_id"); raw != "" {
		parsed, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
			return
		}
		categoryID = &parsed
	}

	status := strings.ToUpper(strings.TrimSpace(c.Query("status")))
	search := strings.TrimSpace(c.Query("search"))
	tag := strings.TrimSpace(c.Query("tag"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// Non-editors only see published
	isAdmin, _ := c.Get("is_admin")
	isProfessor, _ := c.Get("is_professor")
	if isAdmin != true && isProfessor != true {
		status = "PUBLISHED"
	}

	items, total, err := h.svc.ListArticles(c.Request.Context(), tenantID, categoryID, status, search, tag, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if items == nil {
		c.JSON(http.StatusOK, gin.H{"items": []struct{}{}, "total": total})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

func (h *KBHandler) CreateArticle(c *gin.Context) {
	if !h.requireEditor(c) {
		return
	}
	var body struct {
		CategoryID string   `json:"category_id"`
		Title      string   `json:"title"`
		Content    string   `json:"content"`
		Tags       []string `json:"tags"`
		Status     string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	catID, err := uuid.Parse(body.CategoryID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}
	article, err := h.svc.CreateArticle(c.Request.Context(), kb.CreateArticleInput{
		TenantID:   h.tenantID(c),
		AuthorID:   h.userID(c),
		CategoryID: catID,
		Title:      body.Title,
		Content:    body.Content,
		Tags:       body.Tags,
		Status:     body.Status,
	})
	if err != nil {
		if errors.Is(err, kb.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, article)
}

func (h *KBHandler) GetArticle(c *gin.Context) {
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	article, err := h.svc.GetArticle(c.Request.Context(), h.tenantID(c), articleID)
	if err != nil {
		if errors.Is(err, kb.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Non-editors can't see drafts
	if article.Status != "PUBLISHED" {
		isAdmin, _ := c.Get("is_admin")
		isProfessor, _ := c.Get("is_professor")
		if isAdmin != true && isProfessor != true {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
	}

	c.JSON(http.StatusOK, article)
}

func (h *KBHandler) UpdateArticle(c *gin.Context) {
	if !h.requireEditor(c) {
		return
	}
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	var body struct {
		Title   string   `json:"title"`
		Content string   `json:"content"`
		Tags    []string `json:"tags"`
		Status  string   `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	article, err := h.svc.UpdateArticle(c.Request.Context(), kb.UpdateArticleInput{
		TenantID:  h.tenantID(c),
		ArticleID: articleID,
		Title:     body.Title,
		Content:   body.Content,
		Tags:      body.Tags,
		Status:    body.Status,
	})
	if err != nil {
		if errors.Is(err, kb.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, article)
}

func (h *KBHandler) DeleteArticle(c *gin.Context) {
	if !h.requireEditor(c) {
		return
	}
	articleID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
		return
	}
	if err := h.svc.DeleteArticle(c.Request.Context(), h.tenantID(c), articleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *KBHandler) UploadArticle(c *gin.Context) {
	if !h.requireEditor(c) {
		return
	}
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	if header.Size > 2*1024*1024 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file too large (max 2MB)"})
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read file"})
		return
	}

	catIDStr := c.PostForm("category_id")
	catID, err := uuid.Parse(catIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid category_id"})
		return
	}

	status := c.DefaultPostForm("status", "DRAFT")
	tagsRaw := c.PostForm("tags")
	var tags []string
	if tagsRaw != "" {
		for _, t := range strings.Split(tagsRaw, ",") {
			t = strings.TrimSpace(t)
			if t != "" {
				tags = append(tags, t)
			}
		}
	}

	article, err := h.svc.UploadMarkdownFile(c.Request.Context(), kb.UploadMarkdownInput{
		TenantID:   h.tenantID(c),
		AuthorID:   h.userID(c),
		CategoryID: catID,
		Filename:   header.Filename,
		Content:    data,
		Status:     status,
		Tags:       tags,
	})
	if err != nil {
		if errors.Is(err, kb.ErrInvalidInput) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, article)
}

// ── Tags ────────────────────────────────────────────────────

func (h *KBHandler) ListTags(c *gin.Context) {
	tags, err := h.svc.ListPopularTags(c.Request.Context(), h.tenantID(c))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tags == nil {
		c.JSON(http.StatusOK, []struct{}{})
		return
	}
	c.JSON(http.StatusOK, tags)
}
