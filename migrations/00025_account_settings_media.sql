-- +goose Up

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS pending_email TEXT NULL,
  ADD COLUMN IF NOT EXISTS pending_email_requested_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS avatar_key TEXT NULL,
  ADD COLUMN IF NOT EXISTS avatar_updated_at TIMESTAMPTZ NULL;

CREATE INDEX IF NOT EXISTS idx_users_pending_email
  ON users(pending_email)
  WHERE pending_email IS NOT NULL;

ALTER TABLE projects
  ADD COLUMN IF NOT EXISTS image_key TEXT NULL,
  ADD COLUMN IF NOT EXISTS image_updated_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS default_cover_variant SMALLINT NOT NULL DEFAULT 1;

ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_default_cover_variant_check;
ALTER TABLE projects
  ADD CONSTRAINT projects_default_cover_variant_check
  CHECK (default_cover_variant BETWEEN 1 AND 6);

ALTER TABLE auth_tokens DROP CONSTRAINT IF EXISTS auth_tokens_purpose_check;
ALTER TABLE auth_tokens
  ADD CONSTRAINT auth_tokens_purpose_check
  CHECK (purpose IN ('EMAIL_VERIFICATION', 'PASSWORD_RESET', 'EMAIL_CHANGE'));

-- +goose Down

ALTER TABLE auth_tokens DROP CONSTRAINT IF EXISTS auth_tokens_purpose_check;
ALTER TABLE auth_tokens
  ADD CONSTRAINT auth_tokens_purpose_check
  CHECK (purpose IN ('EMAIL_VERIFICATION', 'PASSWORD_RESET'));

ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_default_cover_variant_check;
ALTER TABLE projects
  DROP COLUMN IF EXISTS default_cover_variant,
  DROP COLUMN IF EXISTS image_updated_at,
  DROP COLUMN IF EXISTS image_key;

DROP INDEX IF EXISTS idx_users_pending_email;

ALTER TABLE users
  DROP COLUMN IF EXISTS avatar_updated_at,
  DROP COLUMN IF EXISTS avatar_key,
  DROP COLUMN IF EXISTS pending_email_requested_at,
  DROP COLUMN IF EXISTS pending_email;
