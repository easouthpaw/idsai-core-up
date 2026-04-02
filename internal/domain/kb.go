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
	ID        uuid.UUID
	TenantID  uuid.UUID
	ParentID  *uuid.UUID
	Title     string
	Slug      string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type KBArticle struct {
	ID           uuid.UUID
	TenantID     uuid.UUID
	CategoryID   uuid.UUID
	AuthorID     uuid.UUID
	Title        string
	Slug         string
	Content      string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	PublishedAt  *time.Time
	AuthorName   string
	AuthorAvatar string
	CategoryPath string
	Tags         []string
}

type KBArticleListItem struct {
	ID          uuid.UUID
	CategoryID  uuid.UUID
	AuthorID    uuid.UUID
	Title       string
	Slug        string
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	PublishedAt *time.Time
	AuthorName  string
	Tags        []string
}

type KBTag struct {
	ID   uuid.UUID
	Name string
}
