-- +goose Up

ALTER TABLE users
  ADD COLUMN IF NOT EXISTS email_verified_at TIMESTAMPTZ NULL,
  ADD COLUMN IF NOT EXISTS password_changed_at TIMESTAMPTZ NOT NULL DEFAULT now();

UPDATE users
SET email_verified_at = COALESCE(email_verified_at, created_at)
WHERE status = 'ACTIVE'
  AND email_verified_at IS NULL;

UPDATE users
SET password_changed_at = COALESCE(password_changed_at, created_at, now())
WHERE password_changed_at IS NULL;

CREATE TABLE IF NOT EXISTS auth_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  tenant_id UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  purpose TEXT NOT NULL,
  token_hash TEXT NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  consumed_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  CONSTRAINT auth_tokens_purpose_check CHECK (purpose IN ('EMAIL_VERIFICATION', 'PASSWORD_RESET'))
);

CREATE INDEX IF NOT EXISTS idx_auth_tokens_lookup
  ON auth_tokens(purpose, token_hash);

CREATE INDEX IF NOT EXISTS idx_auth_tokens_user_purpose
  ON auth_tokens(user_id, purpose, expires_at DESC);

-- +goose Down

DROP INDEX IF EXISTS idx_auth_tokens_user_purpose;
DROP INDEX IF EXISTS idx_auth_tokens_lookup;
DROP TABLE IF EXISTS auth_tokens;

ALTER TABLE users
  DROP COLUMN IF EXISTS password_changed_at,
  DROP COLUMN IF EXISTS email_verified_at;
