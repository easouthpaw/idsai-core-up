package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrKBNotFound             = errors.New("kb: not found")
	ErrKBCategorySlugConflict = errors.New("kb: category slug conflict")
	ErrKBArticleSlugConflict  = errors.New("kb: article slug conflict")
)

type KBCategory struct {
	ID        uuid.UUID  `json:"id"`
	TenantID  uuid.UUID  `json:"tenant_id"`
	ParentID  *uuid.UUID `json:"parent_id"`
	Title     string     `json:"title"`
	Slug      string     `json:"slug"`
	SortOrder int        `json:"sort_order"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type KBArticle struct {
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

type KBArticleListItem struct {
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

type KBTag struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
}
