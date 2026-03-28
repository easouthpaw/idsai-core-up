-- +goose Up
UPDATE projects
SET status = 'COMPLETED',
    updated_at = now()
WHERE id = '11000000-0000-0000-0000-000000000005'::uuid
  AND status = 'GRADING';

-- +goose Down
UPDATE projects
SET status = 'GRADING',
    updated_at = now()
WHERE id = '11000000-0000-0000-0000-000000000005'::uuid
  AND status = 'COMPLETED';
