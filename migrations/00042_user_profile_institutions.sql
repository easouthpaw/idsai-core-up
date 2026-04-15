-- +goose Up

ALTER TABLE user_profiles
  ADD COLUMN IF NOT EXISTS institution_provider TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS institution_external_id TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS institution_name TEXT NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS institution_address TEXT NOT NULL DEFAULT '';

-- +goose Down

ALTER TABLE user_profiles
  DROP COLUMN IF EXISTS institution_address,
  DROP COLUMN IF EXISTS institution_name,
  DROP COLUMN IF EXISTS institution_external_id,
  DROP COLUMN IF EXISTS institution_provider;
