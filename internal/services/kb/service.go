package kb

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
)

var (
	ErrNotFound         = domain.ErrKBNotFound
	ErrInvalidInput     = errors.New("invalid input")
	ErrForbidden        = errors.New("forbidden")
	ErrCategoryNotEmpty = errors.New("category is not empty")
)

var slugRe = regexp.MustCompile(`[^a-z0-9-]+`)

type Service struct {
	repo repository
}

type repository interface {
	CreateCategory(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error)
	UpdateCategory(ctx context.Context, tenantID, categoryID uuid.UUID, title, slug string, sortOrder int) (domain.KBCategory, error)
	DeleteCategory(ctx context.Context, tenantID, categoryID uuid.UUID) error
	CategoryHasChildren(ctx context.Context, categoryID uuid.UUID) (bool, error)
	ListCategoryTree(ctx context.Context, tenantID uuid.UUID) ([]domain.KBCategory, error)
	CreateArticle(ctx context.Context, tenantID, categoryID, authorID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) (uuid.UUID, error)
	UpdateArticle(ctx context.Context, tenantID, articleID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) error
	DeleteArticle(ctx context.Context, tenantID, articleID uuid.UUID) error
	GetArticleByID(ctx context.Context, tenantID, articleID uuid.UUID) (domain.KBArticle, error)
	ListArticles(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, status, search, tag string, limit, offset int) ([]domain.KBArticleListItem, int, error)
	SyncArticleTags(ctx context.Context, tenantID, articleID uuid.UUID, tagNames []string) error
	ListPopularTags(ctx context.Context, tenantID uuid.UUID, limit int) ([]domain.KBTag, error)
}

func NewService(repo repository) *Service {
	return &Service{repo: repo}
}

// ── helpers ─────────────────────────────────────────────────

func slugify(title string, maxLen int) string {
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
	if maxLen > 0 && len(slug) > maxLen {
		slug = slug[:maxLen]
	}
	return slug
}

func slugWithSuffix(base string, maxLen, attempt int) string {
	if attempt <= 0 {
		return base
	}
	suffix := fmt.Sprintf("-%d", attempt+1)
	if len(suffix) >= maxLen {
		return suffix[len(suffix)-maxLen:]
	}
	cut := maxLen - len(suffix)
	if cut < len(base) {
		base = base[:cut]
	}
	base = strings.TrimRight(base, "-")
	if base == "" {
		return strings.TrimLeft(suffix, "-")
	}
	return base + suffix
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

func (s *Service) CreateCategory(ctx context.Context, in CreateCategoryInput) (domain.KBCategory, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 200 {
		return domain.KBCategory{}, fmt.Errorf("%w: title is required (max 200)", ErrInvalidInput)
	}
	baseSlug := slugify(title, 200)
	for attempt := 0; attempt < 32; attempt++ {
		slug := slugWithSuffix(baseSlug, 200, attempt)
		cat, err := s.repo.CreateCategory(ctx, in.TenantID, in.ParentID, title, slug, in.SortOrder)
		if errors.Is(err, domain.ErrKBCategorySlugConflict) {
			continue
		}
		return cat, err
	}
	return domain.KBCategory{}, fmt.Errorf("%w: could not generate unique category slug", ErrInvalidInput)
}

type UpdateCategoryInput struct {
	TenantID   uuid.UUID
	CategoryID uuid.UUID
	Title      string
	SortOrder  int
}

func (s *Service) UpdateCategory(ctx context.Context, in UpdateCategoryInput) (domain.KBCategory, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 200 {
		return domain.KBCategory{}, fmt.Errorf("%w: title is required (max 200)", ErrInvalidInput)
	}
	baseSlug := slugify(title, 200)
	for attempt := 0; attempt < 32; attempt++ {
		slug := slugWithSuffix(baseSlug, 200, attempt)
		cat, err := s.repo.UpdateCategory(ctx, in.TenantID, in.CategoryID, title, slug, in.SortOrder)
		if errors.Is(err, domain.ErrKBCategorySlugConflict) {
			continue
		}
		return cat, err
	}
	return domain.KBCategory{}, fmt.Errorf("%w: could not generate unique category slug", ErrInvalidInput)
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

func (s *Service) GetCategoryTree(ctx context.Context, tenantID uuid.UUID) ([]domain.KBCategory, error) {
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

const maxKBContentBytes = 512 * 1024 // 512 KB

func (s *Service) CreateArticle(ctx context.Context, in CreateArticleInput) (domain.KBArticle, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 300 {
		return domain.KBArticle{}, fmt.Errorf("%w: title is required (max 300)", ErrInvalidInput)
	}
	content := strings.TrimSpace(in.Content)
	if len(content) > maxKBContentBytes {
		return domain.KBArticle{}, fmt.Errorf("%w: content exceeds maximum size (512 KB)", ErrInvalidInput)
	}
	status := strings.ToUpper(strings.TrimSpace(in.Status))
	if status != "DRAFT" && status != "PUBLISHED" {
		status = "DRAFT"
	}

	baseSlug := slugify(title, 300)
	tags := dedupeTagNames(in.Tags)

	var publishedAt *time.Time
	if status == "PUBLISHED" {
		now := time.Now().UTC()
		publishedAt = &now
	}

	var (
		id  uuid.UUID
		err error
	)
	for attempt := 0; attempt < 32; attempt++ {
		slug := slugWithSuffix(baseSlug, 300, attempt)
		id, err = s.repo.CreateArticle(ctx, in.TenantID, in.CategoryID, in.AuthorID, title, slug, content, status, publishedAt)
		if errors.Is(err, domain.ErrKBArticleSlugConflict) {
			continue
		}
		break
	}
	if err != nil {
		if errors.Is(err, domain.ErrKBArticleSlugConflict) {
			return domain.KBArticle{}, fmt.Errorf("%w: could not generate unique article slug", ErrInvalidInput)
		}
		return domain.KBArticle{}, err
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

func (s *Service) UpdateArticle(ctx context.Context, in UpdateArticleInput) (domain.KBArticle, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" || len(title) > 300 {
		return domain.KBArticle{}, fmt.Errorf("%w: title is required (max 300)", ErrInvalidInput)
	}
	content := strings.TrimSpace(in.Content)
	if len(content) > maxKBContentBytes {
		return domain.KBArticle{}, fmt.Errorf("%w: content exceeds maximum size (512 KB)", ErrInvalidInput)
	}
	status := strings.ToUpper(strings.TrimSpace(in.Status))
	if status != "DRAFT" && status != "PUBLISHED" {
		status = "DRAFT"
	}

	baseSlug := slugify(title, 300)
	tags := dedupeTagNames(in.Tags)

	var publishedAt *time.Time
	if status == "PUBLISHED" {
		now := time.Now().UTC()
		publishedAt = &now
	}

	var err error
	for attempt := 0; attempt < 32; attempt++ {
		slug := slugWithSuffix(baseSlug, 300, attempt)
		err = s.repo.UpdateArticle(ctx, in.TenantID, in.ArticleID, title, slug, content, status, publishedAt)
		if errors.Is(err, domain.ErrKBArticleSlugConflict) {
			continue
		}
		break
	}
	if err != nil {
		if errors.Is(err, domain.ErrKBArticleSlugConflict) {
			return domain.KBArticle{}, fmt.Errorf("%w: could not generate unique article slug", ErrInvalidInput)
		}
		return domain.KBArticle{}, err
	}

	if err := s.repo.SyncArticleTags(ctx, in.TenantID, in.ArticleID, tags); err != nil {
		return domain.KBArticle{}, err
	}

	return s.repo.GetArticleByID(ctx, in.TenantID, in.ArticleID)
}

func (s *Service) DeleteArticle(ctx context.Context, tenantID, articleID uuid.UUID) error {
	return s.repo.DeleteArticle(ctx, tenantID, articleID)
}

func (s *Service) GetArticle(ctx context.Context, tenantID, articleID uuid.UUID) (domain.KBArticle, error) {
	return s.repo.GetArticleByID(ctx, tenantID, articleID)
}

func (s *Service) ListArticles(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, status, search, tag string, limit, offset int) ([]domain.KBArticleListItem, int, error) {
	return s.repo.ListArticles(ctx, tenantID, categoryID, status, search, tag, limit, offset)
}

// ── Tags ────────────────────────────────────────────────────

func (s *Service) ListPopularTags(ctx context.Context, tenantID uuid.UUID) ([]domain.KBTag, error) {
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

func (s *Service) UploadMarkdownFile(ctx context.Context, in UploadMarkdownInput) (domain.KBArticle, error) {
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
