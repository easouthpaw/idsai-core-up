-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE c RECORD;
BEGIN
  FOR c IN
    SELECT conname
    FROM pg_constraint
    WHERE conrelid = 'role_assignments'::regclass
      AND contype = 'c'
      AND pg_get_constraintdef(oid) ILIKE '%scope_type%'
  LOOP
    EXECUTE format('ALTER TABLE role_assignments DROP CONSTRAINT %I', c.conname);
  END LOOP;
END $$;
-- +goose StatementEnd

ALTER TABLE role_assignments
ADD CONSTRAINT role_assignments_scope_type_check
CHECK (scope_type IN ('SYSTEM','FACULTY','DEPARTMENT','PROJECT'));

-- +goose Down
ALTER TABLE role_assignments DROP CONSTRAINT IF EXISTS role_assignments_scope_type_check;
ALTER TABLE role_assignments
ADD CONSTRAINT role_assignments_scope_type_check
CHECK (scope_type IN ('SYSTEM','FACULTY','PROJECT'));