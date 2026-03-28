package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"idsai-core-up/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrKBNotFound             = domain.ErrKBNotFound
	ErrKBCategorySlugConflict = domain.ErrKBCategorySlugConflict
	ErrKBArticleSlugConflict  = domain.ErrKBArticleSlugConflict
)

func isUniqueViolation(err error, names ...string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		return false
	}
	for _, name := range names {
		if pgErr.ConstraintName == name {
			return true
		}
	}
	return false
}

// KBRepo is the Knowledge Base repository.
type KBRepo struct {
	db *pgxpool.Pool
}

func NewKBRepo(pool *pgxpool.Pool) *KBRepo {
	return &KBRepo{db: pool}
}

// ── Category types ──────────────────────────────────────────

type KBCategory = domain.KBCategory

// ── Article types ───────────────────────────────────────────

type KBArticle = domain.KBArticle

type KBArticleListItem = domain.KBArticleListItem

type KBTag = domain.KBTag

// ── Category CRUD ───────────────────────────────────────────

func (r *KBRepo) CreateCategory(ctx context.Context, tenantID uuid.UUID, parentID *uuid.UUID, title, slug string, sortOrder int) (KBCategory, error) {
	now := time.Now().UTC()
	var cat KBCategory
	err := r.db.QueryRow(ctx, `
		INSERT INTO kb_categories (tenant_id, parent_id, title, slug, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
		RETURNING id, tenant_id, parent_id, title, slug, sort_order, created_at, updated_at`,
		tenantID, parentID, title, slug, sortOrder, now,
	).Scan(&cat.ID, &cat.TenantID, &cat.ParentID, &cat.Title, &cat.Slug, &cat.SortOrder, &cat.CreatedAt, &cat.UpdatedAt)
	if isUniqueViolation(err, "kb_categories_tenant_id_parent_id_slug_key", "idx_kb_categories_root_slug") {
		return KBCategory{}, ErrKBCategorySlugConflict
	}
	return cat, err
}

func (r *KBRepo) UpdateCategory(ctx context.Context, tenantID, categoryID uuid.UUID, title, slug string, sortOrder int) (KBCategory, error) {
	var cat KBCategory
	err := r.db.QueryRow(ctx, `
		UPDATE kb_categories
		SET title = $3, slug = $4, sort_order = $5, updated_at = now()
		WHERE id = $2 AND tenant_id = $1
		RETURNING id, tenant_id, parent_id, title, slug, sort_order, created_at, updated_at`,
		tenantID, categoryID, title, slug, sortOrder,
	).Scan(&cat.ID, &cat.TenantID, &cat.ParentID, &cat.Title, &cat.Slug, &cat.SortOrder, &cat.CreatedAt, &cat.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return KBCategory{}, ErrKBNotFound
	}
	if isUniqueViolation(err, "kb_categories_tenant_id_parent_id_slug_key", "idx_kb_categories_root_slug") {
		return KBCategory{}, ErrKBCategorySlugConflict
	}
	return cat, err
}

func (r *KBRepo) DeleteCategory(ctx context.Context, tenantID, categoryID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM kb_categories WHERE id = $1 AND tenant_id = $2`, categoryID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrKBNotFound
	}
	return nil
}

func (r *KBRepo) CategoryHasChildren(ctx context.Context, categoryID uuid.UUID) (bool, error) {
	var hasArticles, hasChildren bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(SELECT 1 FROM kb_articles WHERE category_id = $1),
		        EXISTS(SELECT 1 FROM kb_categories WHERE parent_id = $1)`, categoryID,
	).Scan(&hasArticles, &hasChildren)
	return hasArticles || hasChildren, err
}

func (r *KBRepo) ListCategoryTree(ctx context.Context, tenantID uuid.UUID) ([]KBCategory, error) {
	rows, err := r.db.Query(ctx, `
		WITH RECURSIVE tree AS (
			SELECT id, tenant_id, parent_id, title, slug, sort_order, created_at, updated_at, 0 AS depth
			FROM kb_categories WHERE tenant_id = $1 AND parent_id IS NULL
			UNION ALL
			SELECT c.id, c.tenant_id, c.parent_id, c.title, c.slug, c.sort_order, c.created_at, c.updated_at, t.depth + 1
			FROM kb_categories c JOIN tree t ON c.parent_id = t.id
		)
		SELECT id, tenant_id, parent_id, title, slug, sort_order, created_at, updated_at
		FROM tree ORDER BY depth, sort_order, title`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cats []KBCategory
	for rows.Next() {
		var c KBCategory
		if err := rows.Scan(&c.ID, &c.TenantID, &c.ParentID, &c.Title, &c.Slug, &c.SortOrder, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

// ── Article CRUD ────────────────────────────────────────────

func (r *KBRepo) CreateArticle(ctx context.Context, tenantID, categoryID, authorID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) (uuid.UUID, error) {
	now := time.Now().UTC()
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO kb_articles (tenant_id, category_id, author_id, title, slug, content, status, created_at, updated_at, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $8, $9)
		RETURNING id`,
		tenantID, categoryID, authorID, title, slug, content, status, now, publishedAt,
	).Scan(&id)
	if isUniqueViolation(err, "kb_articles_tenant_id_category_id_slug_key") {
		return uuid.Nil, ErrKBArticleSlugConflict
	}
	return id, err
}

func (r *KBRepo) UpdateArticle(ctx context.Context, tenantID, articleID uuid.UUID, title, slug, content, status string, publishedAt *time.Time) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE kb_articles
		SET title = $3, slug = $4, content = $5, status = $6, updated_at = now(), published_at = COALESCE($7, published_at)
		WHERE id = $2 AND tenant_id = $1`,
		tenantID, articleID, title, slug, content, status, publishedAt,
	)
	if err != nil {
		if isUniqueViolation(err, "kb_articles_tenant_id_category_id_slug_key") {
			return ErrKBArticleSlugConflict
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrKBNotFound
	}
	return nil
}

func (r *KBRepo) DeleteArticle(ctx context.Context, tenantID, articleID uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM kb_articles WHERE id = $1 AND tenant_id = $2`, articleID, tenantID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrKBNotFound
	}
	return nil
}

func (r *KBRepo) GetArticleByID(ctx context.Context, tenantID, articleID uuid.UUID) (KBArticle, error) {
	var a KBArticle
	err := r.db.QueryRow(ctx, `
		SELECT a.id, a.tenant_id, a.category_id, a.author_id, a.title, a.slug, a.content, a.status,
		       a.created_at, a.updated_at, a.published_at,
		       COALESCE(p.full_name, u.email, '') AS author_name,
		       COALESCE(u.avatar_key, '') AS author_avatar
		FROM kb_articles a
		JOIN users u ON u.id = a.author_id
		LEFT JOIN user_profiles p ON p.user_id = a.author_id
		WHERE a.id = $1 AND a.tenant_id = $2`,
		articleID, tenantID,
	).Scan(&a.ID, &a.TenantID, &a.CategoryID, &a.AuthorID, &a.Title, &a.Slug, &a.Content, &a.Status,
		&a.CreatedAt, &a.UpdatedAt, &a.PublishedAt, &a.AuthorName, &a.AuthorAvatar)
	if errors.Is(err, pgx.ErrNoRows) {
		return KBArticle{}, ErrKBNotFound
	}
	if err != nil {
		return KBArticle{}, err
	}
	a.Tags, _ = r.ListTagNamesByArticle(ctx, a.ID)
	return a, nil
}

func (r *KBRepo) ListArticles(ctx context.Context, tenantID uuid.UUID, categoryID *uuid.UUID, status, search, tag string, limit, offset int) ([]KBArticleListItem, int, error) {
	var args []any
	argN := 0
	nextArg := func(v any) string {
		argN++
		args = append(args, v)
		return "$" + strings.Repeat("", 0) + itoa(argN)
	}

	where := "a.tenant_id = " + nextArg(tenantID)
	if categoryID != nil {
		where += " AND a.category_id = " + nextArg(*categoryID)
	}
	if status != "" {
		where += " AND a.status = " + nextArg(status)
	}
	if search != "" {
		where += " AND a.search_vector @@ plainto_tsquery('russian', " + nextArg(search) + ")"
	}
	if tag != "" {
		where += " AND EXISTS (SELECT 1 FROM kb_article_tags at2 JOIN kb_tags t ON t.id = at2.tag_id WHERE at2.article_id = a.id AND lower(t.name) = lower(" + nextArg(tag) + "))"
	}

	var total int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM kb_articles a WHERE "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	query := `
		SELECT a.id, a.category_id, a.author_id, a.title, a.slug, a.status,
		       a.created_at, a.updated_at, a.published_at,
		       COALESCE(p.full_name, u.email, '')
		FROM kb_articles a
		JOIN users u ON u.id = a.author_id
		LEFT JOIN user_profiles p ON p.user_id = a.author_id
		WHERE ` + where + `
		ORDER BY a.updated_at DESC
		LIMIT ` + nextArg(limit) + ` OFFSET ` + nextArg(offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var items []KBArticleListItem
	for rows.Next() {
		var item KBArticleListItem
		if err := rows.Scan(&item.ID, &item.CategoryID, &item.AuthorID, &item.Title, &item.Slug, &item.Status,
			&item.CreatedAt, &item.UpdatedAt, &item.PublishedAt, &item.AuthorName); err != nil {
			return nil, 0, err
		}
		item.Tags, _ = r.ListTagNamesByArticle(ctx, item.ID)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

// ── Tags ────────────────────────────────────────────────────

func (r *KBRepo) CreateTagIfNotExists(ctx context.Context, tenantID uuid.UUID, name string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.db.QueryRow(ctx, `
		INSERT INTO kb_tags (tenant_id, name) VALUES ($1, $2)
		ON CONFLICT (tenant_id, lower(name)) DO UPDATE SET name = EXCLUDED.name
		RETURNING id`, tenantID, name).Scan(&id)
	return id, err
}

func (r *KBRepo) SyncArticleTags(ctx context.Context, tenantID, articleID uuid.UUID, tagNames []string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM kb_article_tags WHERE article_id = $1`, articleID)
	if err != nil {
		return err
	}
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		tagID, err := r.CreateTagIfNotExists(ctx, tenantID, name)
		if err != nil {
			return err
		}
		_, err = r.db.Exec(ctx, `INSERT INTO kb_article_tags (article_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, articleID, tagID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *KBRepo) ListTagNamesByArticle(ctx context.Context, articleID uuid.UUID) ([]string, error) {
	rows, err := r.db.Query(ctx, `
		SELECT t.name FROM kb_tags t
		JOIN kb_article_tags at2 ON at2.tag_id = t.id
		WHERE at2.article_id = $1
		ORDER BY t.name`, articleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		tags = append(tags, name)
	}
	return tags, rows.Err()
}

func (r *KBRepo) ListPopularTags(ctx context.Context, tenantID uuid.UUID, limit int) ([]KBTag, error) {
	if limit <= 0 {
		limit = 30
	}
	rows, err := r.db.Query(ctx, `
		SELECT t.id, t.name FROM kb_tags t
		JOIN kb_article_tags at2 ON at2.tag_id = t.id
		JOIN kb_articles a ON a.id = at2.article_id
		WHERE a.tenant_id = $1 AND a.status = 'PUBLISHED'
		GROUP BY t.id, t.name
		ORDER BY COUNT(*) DESC, t.name
		LIMIT $2`, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []KBTag
	for rows.Next() {
		var t KBTag
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			return nil, err
		}
		tags = append(tags, t)
	}
	return tags, rows.Err()
}

// ── helpers ─────────────────────────────────────────────────

func itoa(n int) string {
	if n < 10 {
		return string(rune('0' + n))
	}
	return itoa(n/10) + string(rune('0'+n%10))
}
