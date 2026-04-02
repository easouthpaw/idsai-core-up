package dto

import (
	"time"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
)

type CreateCategoryRequest struct {
	ParentID  *string `json:"parent_id"`
	Title     string  `json:"title"`
	SortOrder int     `json:"sort_order"`
}

type UpdateCategoryRequest struct {
	Title     string `json:"title"`
	SortOrder int    `json:"sort_order"`
}

type CreateArticleRequest struct {
	CategoryID string   `json:"category_id"`
	Title      string   `json:"title"`
	Content    string   `json:"content"`
	Tags       []string `json:"tags"`
	Status     string   `json:"status"`
}

type UpdateArticleRequest struct {
	Title   string   `json:"title"`
	Content string   `json:"content"`
	Tags    []string `json:"tags"`
	Status  string   `json:"status"`
}

type KBCategoryResponse struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Title     string     `json:"title"`
	Slug      string     `json:"slug"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type KBArticleResponse struct {
	ID           uuid.UUID  `json:"id"`
	TenantID     uuid.UUID  `json:"tenant_id"`
	CategoryID   uuid.UUID  `json:"category_id"`
	AuthorID     uuid.UUID  `json:"author_id"`
	Title        string     `json:"title"`
	Slug         string     `json:"slug"`
	Content      string     `json:"content"`
	Status       string     `json:"status"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	PublishedAt  *time.Time `json:"published_at"`
	AuthorName   string     `json:"author_name"`
	AuthorAvatar string     `json:"author_avatar"`
	CategoryPath string     `json:"category_path"`
	Tags         []string   `json:"tags"`
}

type KBArticleListItemResponse struct {
	ID          uuid.UUID  `json:"id"`
	CategoryID  uuid.UUID  `json:"category_id"`
	AuthorID    uuid.UUID  `json:"author_id"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	PublishedAt *time.Time `json:"published_at"`
	AuthorName  string     `json:"author_name"`
	Tags        []string   `json:"tags"`
}

type ListArticlesResponse struct {
	Items []KBArticleListItemResponse `json:"items"`
	Total int                         `json:"total"`
}

type KBTagResponse struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}

func KBCategoryResponseFromDomain(item domain.KBCategory) KBCategoryResponse {
	return KBCategoryResponse{
		ID:        item.ID,
		TenantID:  item.TenantID,
		ParentID:  item.ParentID,
		Title:     item.Title,
		Slug:      item.Slug,
		SortOrder: item.SortOrder,
		CreatedAt: item.CreatedAt,
		UpdatedAt: item.UpdatedAt,
	}
}

func KBCategoryResponsesFromDomain(items []domain.KBCategory) []KBCategoryResponse {
	if items == nil {
		return nil
	}
	out := make([]KBCategoryResponse, 0, len(items))
	for _, item := range items {
		out = append(out, KBCategoryResponseFromDomain(item))
	}
	return out
}

func KBArticleResponseFromDomain(item domain.KBArticle) KBArticleResponse {
	return KBArticleResponse{
		ID:           item.ID,
		TenantID:     item.TenantID,
		CategoryID:   item.CategoryID,
		AuthorID:     item.AuthorID,
		Title:        item.Title,
		Slug:         item.Slug,
		Content:      item.Content,
		Status:       item.Status,
		CreatedAt:    item.CreatedAt,
		UpdatedAt:    item.UpdatedAt,
		PublishedAt:  item.PublishedAt,
		AuthorName:   item.AuthorName,
		AuthorAvatar: item.AuthorAvatar,
		CategoryPath: item.CategoryPath,
		Tags:         item.Tags,
	}
}

func KBArticleListItemResponsesFromDomain(items []domain.KBArticleListItem) []KBArticleListItemResponse {
	if items == nil {
		return nil
	}
	out := make([]KBArticleListItemResponse, 0, len(items))
	for _, item := range items {
		out = append(out, KBArticleListItemResponse{
			ID:          item.ID,
			CategoryID:  item.CategoryID,
			AuthorID:    item.AuthorID,
			Title:       item.Title,
			Slug:        item.Slug,
			Status:      item.Status,
			CreatedAt:   item.CreatedAt,
			UpdatedAt:   item.UpdatedAt,
			PublishedAt: item.PublishedAt,
			AuthorName:  item.AuthorName,
			Tags:        item.Tags,
		})
	}
	return out
}

func KBTagResponsesFromDomain(items []domain.KBTag) []KBTagResponse {
	if items == nil {
		return nil
	}
	out := make([]KBTagResponse, 0, len(items))
	for _, item := range items {
		out = append(out, KBTagResponse{
			ID:   item.ID,
			Name: item.Name,
		})
	}
	return out
}
