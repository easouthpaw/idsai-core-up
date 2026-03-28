-- +goose Up

-- =============================================
-- Knowledge Base: Categories (tree via parent_id)
-- =============================================
CREATE TABLE IF NOT EXISTS kb_categories (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    parent_id    UUID REFERENCES kb_categories(id) ON DELETE CASCADE,
    title        VARCHAR(200) NOT NULL,
    slug         VARCHAR(200) NOT NULL,
    sort_order   INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(tenant_id, parent_id, slug)
);

-- Partial unique for root-level categories (parent_id IS NULL)
CREATE UNIQUE INDEX IF NOT EXISTS idx_kb_categories_root_slug
    ON kb_categories(tenant_id, slug)
    WHERE parent_id IS NULL;

CREATE INDEX IF NOT EXISTS idx_kb_categories_tenant ON kb_categories(tenant_id);
CREATE INDEX IF NOT EXISTS idx_kb_categories_parent ON kb_categories(parent_id);

-- =============================================
-- Knowledge Base: Articles
-- =============================================
CREATE TABLE IF NOT EXISTS kb_articles (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id     UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    category_id   UUID NOT NULL REFERENCES kb_categories(id) ON DELETE CASCADE,
    author_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title         VARCHAR(300) NOT NULL,
    slug          VARCHAR(300) NOT NULL,
    content       TEXT NOT NULL DEFAULT '',
    status        VARCHAR(20) NOT NULL DEFAULT 'DRAFT',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    published_at  TIMESTAMPTZ,
    search_vector tsvector,
    UNIQUE(tenant_id, category_id, slug),
    CONSTRAINT kb_articles_status_check CHECK (status IN ('DRAFT', 'PUBLISHED'))
);

CREATE INDEX IF NOT EXISTS idx_kb_articles_tenant     ON kb_articles(tenant_id);
CREATE INDEX IF NOT EXISTS idx_kb_articles_category   ON kb_articles(category_id, status);
CREATE INDEX IF NOT EXISTS idx_kb_articles_author     ON kb_articles(author_id);
CREATE INDEX IF NOT EXISTS idx_kb_articles_search     ON kb_articles USING GIN(search_vector);
CREATE INDEX IF NOT EXISTS idx_kb_articles_published  ON kb_articles(tenant_id, status, published_at DESC)
    WHERE status = 'PUBLISHED';

-- Auto-update search_vector on insert/update
CREATE OR REPLACE FUNCTION kb_articles_search_trigger() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('russian', coalesce(NEW.title, '')), 'A') ||
        setweight(to_tsvector('russian', coalesce(NEW.content, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_kb_articles_search ON kb_articles;
CREATE TRIGGER trg_kb_articles_search
    BEFORE INSERT OR UPDATE OF title, content ON kb_articles
    FOR EACH ROW EXECUTE FUNCTION kb_articles_search_trigger();

-- =============================================
-- Knowledge Base: Tags
-- =============================================
CREATE TABLE IF NOT EXISTS kb_tags (
    id        UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name      VARCHAR(60) NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_kb_tags_unique_name
    ON kb_tags(tenant_id, lower(name));

-- =============================================
-- Knowledge Base: Article ↔ Tag junction
-- =============================================
CREATE TABLE IF NOT EXISTS kb_article_tags (
    article_id UUID NOT NULL REFERENCES kb_articles(id) ON DELETE CASCADE,
    tag_id     UUID NOT NULL REFERENCES kb_tags(id) ON DELETE CASCADE,
    PRIMARY KEY(article_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_kb_article_tags_tag ON kb_article_tags(tag_id);

-- +goose Down

DROP TABLE IF EXISTS kb_article_tags;
DROP TABLE IF EXISTS kb_tags;
DROP TRIGGER IF EXISTS trg_kb_articles_search ON kb_articles;
DROP FUNCTION IF EXISTS kb_articles_search_trigger;
DROP TABLE IF EXISTS kb_articles;
DROP INDEX IF EXISTS idx_kb_categories_root_slug;
DROP TABLE IF EXISTS kb_categories;
