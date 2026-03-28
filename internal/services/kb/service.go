package kb

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"idsai-core-up/internal/repos/postgres"

	"github.com/google/uuid"
)

var (
	ErrNotFound         = postgres.ErrKBNotFound
	ErrInvalidInput     = errors.New("invalid input")
	ErrForbidden        = errors.New("forbidden")
	ErrCategoryNotEmpty = errors.New("category is not empty")
)

var slugRe = regexp.MustCompile(`[^a-z0-9-]+`)

type Service struct {
	repo *postgres.KBRepo
}

func NewService(repo *postgres.KBRepo) *Service {
	return &Service{repo: repo}
}

// ── helpers ─────────────────────────────────────────────────

func slugify(title string) string {
	s := strings.ToLower(strings.TrimSpace(title))
	runes := make([]rune, 0, len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			runes = append(runes, r)
		} else {
			runes = append(runes, '-')
		}
	}
	slug := slugRe.ReplaceAllString(string(runes), "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = uuid.New().String()[:8]
	}
	if len(slug) > 200 {
		slug = slug[:200]
	}
	return slug
}

func dedupeTagNames(raw []string) []string {
	seen := make(map[string]struct{}, len(raw))
	out := make([]string, 0, len(raw))
	for _, name := range raw {
		name = strings.TrimSpace(name)
		if name == "" || len(name) > 60 {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, name)
	}
	if len(out) > 20 {
		out = out[:20]
	}
	return out
}

// ── Category ────────────────────────────────────────────────

type CreateCategoryInput struct {
	TenantID  uuid.UUID
	ParentID  *uuid.UUID
	Title     string
	SortOrder int
}

func (s *Service) CreateCategory(ctx context.Context, in CreateCategoryInput) (postgres.KBCategory, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 200 {
		return postgres.KBCategory{}, fmt.Errorf("%w: title is required (max 200)", ErrInvalidInput)
	}
	slug := slugify(title)
	return s.repo.CreateCategory(ctx, in.TenantID, in.ParentID, title, slug, in.SortOrder)
}

type UpdateCategoryInput struct {
	TenantID   uuid.UUID
	CategoryID uuid.UUID
	Title      string
	SortOrder  int
}

func (s *Service) UpdateCategory(ctx context.Context, in UpdateCategoryInput) (postgres.KBCategory, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 200 {
		return postgres.KBCategory{}, fmt.Errorf("%w: title is required (max 200)", ErrInvalidInput)
	}
	slug := slugify(title)
	return s.repo.UpdateCategory(ctx, in.TenantID, in.CategoryID, title, slug, in.SortOrder)
}

func (s *Service) DeleteCategory(ctx context.Context, tenantID, categoryID uuid.UUID) error {
	hasChildren, err := s.repo.CategoryHasChildren(ctx, categoryID)
	if err != nil {
		return err
	}
	if hasChildren {
		return ErrCategoryNotEmpty
	}
	return s.repo.DeleteCategory(ctx, tenantID, categoryID)
}

func (s *Service) GetCategoryTree(ctx context.Context, tenantID uuid.UUID) ([]postgres.KBCategory, error) {
	return s.repo.ListCategoryTree(ctx, tenantID)
}

// ── Article ─────────────────────────────────────────────────

type CreateArticleInput struct {
	TenantID   uuid.UUID
	AuthorID   uuid.UUID
	CategoryID uuid.UUID
	Title      string
	Content    string
	Tags       []string
	Status     string
}

func (s *Service) CreateArticle(ctx context.Context, in CreateArticleInput) (postgres.KBArticle, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 300 {
		return postgres.KBArticle{}, fmt.Errorf("%w: title is required (max 300)", ErrInvalidInput)
	}
	content := strings.TrimSpace(in.Content)
	status := strings.ToUpper(strings.TrimSpace(in.Status))
	if status != "DRAFT" && status != "PUBLISHED" {
		status = "DRAFT"
	}

	slug := slugify(title)
	tags := dedupeTagNames(in.Tags)

	var publishedAt *time.Time
	if status == "PUBLISHED" {
		now := time.Now().UTC()
		publishedAt = &now
	}

	id, err := s.repo.CreateArticle(ctx, in.TenantID, in.CategoryID, in.AuthorID, title, slug, content, status, publishedAt)
	if err != nil {
		return postgres.KBArticle{}, err
	}

	if len(tags) > 0 {
		_ = s.repo.SyncArticleTags(ctx, in.TenantID, id, tags)
	}

	return s.repo.GetArticleByID(ctx, in.TenantID, id)
}

type UpdateArticleInput struct {
	TenantID  uuid.UUID
	ArticleID uuid.UUID
	Title     string
	Content   string
	Tags      []string
	Status    string
}

func (s *Service) UpdateArticle(ctx context.Context, in UpdateArticleInput) (postgres.KBArticle, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 300 {
		return postgres.KBArticle{}, fmt.Errorf("%w: title is required (max 300)", ErrInvalidInput)
	}
	content := strings.TrimSpace(in.Content)
	status := strings.ToUpper(strings.TrimSpace(in.Status))
	if status != "DRAFT" && status != "PUBLISHED" {
		status = "DRAFT"
	}

	slug := slugify(title)
	tags := dedupeTagNames(in.Tags)

	var publishedAt *time.Time
	if status == "PUBLISHED" {
		now := time.Now().UTC()
		publishedAt = &now
	}

	if err := s.repo.UpdateArticle(ctx, in.TenantID, in.ArticleID, title, slug, content, status, publishedAt); err != nil {
		return postgres.KBArticle{}, err
	}

	if err := s.repo.SyncArticleTags(ctx, in.TenantID, in.ArticleID, tags); err != nil {
		return postgres.KBArticle{}, err
	}

	return s.repo.GetArticleByID(ctx, in.TenantID, in.ArticleID)
}

func (s *Service) DeleteArticle(ctx context.Context, tenantID, articleID uuid.UUID) error {
	return s.repo.DeleteArticle(ctx, tenantID, articleID)
}

func (s *Service) GetArticle(ctx context.Context, tenantID, articleID uuid.UUID) (postgres.KBArticle, error) {
	return s.repo.GetArticleByID(ctx, tenantID, articleID)
}

func (s *Service) ListArticles(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, status, search, tag string, limit, offset int) ([]postgres.KBArticleListItem, int, error) {
	return s.repo.ListArticles(ctx, tenantID, categoryID, status, search, tag, limit, offset)
}

// ── Tags ────────────────────────────────────────────────────

func (s *Service) ListPopularTags(ctx context.Context, tenantID uuid.UUID) ([]postgres.KBTag, error) {
	return s.repo.ListPopularTags(ctx, tenantID, 30)
}

// ── Upload .md file ─────────────────────────────────────────

type UploadMarkdownInput struct {
	TenantID   uuid.UUID
	AuthorID   uuid.UUID
	CategoryID uuid.UUID
	Filename   string
	Content    []byte
	Status     string
	Tags       []string
}

func (s *Service) UploadMarkdownFile(ctx context.Context, in UploadMarkdownInput) (postgres.KBArticle, error) {
	body := string(in.Content)

	title := strings.TrimSpace(in.Filename)
	title = strings.TrimSuffix(title, ".md")
	title = strings.TrimSuffix(title, ".MD")
	title = strings.TrimSuffix(title, ".markdown")
	if title == "" {
		title = "Untitled"
	}

	// Extract title from first heading if present
	for _, line := range strings.SplitN(body, "\n", 20) {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "# ") {
			title = strings.TrimSpace(strings.TrimPrefix(trimmed, "# "))
			break
		}
	}

	return s.CreateArticle(ctx, CreateArticleInput{
		TenantID:   in.TenantID,
		AuthorID:   in.AuthorID,
		CategoryID: in.CategoryID,
		Title:      title,
		Content:    body,
		Tags:       in.Tags,
		Status:     in.Status,
	})
}
