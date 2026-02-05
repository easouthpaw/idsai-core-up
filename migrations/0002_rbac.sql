-- +goose Up

-- Roles catalog
CREATE TABLE IF NOT EXISTS roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Permissions catalog
CREATE TABLE IF NOT EXISTS permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Role -> permissions mapping
CREATE TABLE IF NOT EXISTS role_permissions (
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);

-- User role assignments with scope
-- scope_type: SYSTEM | FACULTY | PROJECT
CREATE TABLE IF NOT EXISTS role_assignments (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  scope_type TEXT NOT NULL,
  scope_id UUID NULL,
  expires_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  CONSTRAINT role_assignments_scope_check
    CHECK (scope_type IN ('SYSTEM', 'FACULTY', 'PROJECT')),

  CONSTRAINT role_assignments_scope_id_check
    CHECK (
      (scope_type = 'SYSTEM' AND scope_id IS NULL)
      OR (scope_type IN ('FACULTY','PROJECT') AND scope_id IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS idx_role_assignments_user_scope
  ON role_assignments(user_id, scope_type, scope_id);

-- --------------------
-- Seeds: roles
-- --------------------
INSERT INTO roles(code, name) VALUES
  ('SUPER_ADMIN', 'Super Admin'),
  ('STUDENT', 'Student'),
  ('PROFESSOR', 'Professor'),
  ('MODERATOR', 'Moderator'),
  ('TEAM_LEAD', 'Team Lead'),
  ('MEMBER', 'Project Member'),
  ('PROJECT_PROFESSOR', 'Project Professor')
ON CONFLICT (code) DO NOTHING;

-- --------------------
-- Seeds: permissions (minimal but useful)
-- --------------------
INSERT INTO permissions(code, description) VALUES
  -- Projects lifecycle
  ('project.create', 'Create a project (draft)'),
  ('project.edit', 'Edit project card/settings'),
  ('project.change_visibility', 'Change project visibility'),
  ('project.invite_professor', 'Invite professor to project'),
  ('project.submit_for_review', 'Submit project for review'),
  ('project.approve', 'Approve project at review stage'),
  ('project.reject', 'Reject project at review stage'),
  ('project.set_criteria', 'Set grading criteria'),
  ('project.set_deadline', 'Set project deadline'),

  -- Team / recruitment
  ('team.apply', 'Apply to join a project'),
  ('team.withdraw_application', 'Withdraw own application'),
  ('team.accept_applicant', 'Accept applicant'),
  ('team.reject_applicant', 'Reject applicant'),
  ('team.submit_roster', 'Submit roster for professor approval'),
  ('team.approve_roster', 'Approve roster'),
  ('team.reject_roster', 'Reject roster'),

  -- Tasks
  ('task.view', 'View tasks in project'),
  ('task.create', 'Create task'),
  ('task.update', 'Update task'),
  ('task.assign', 'Assign task'),
  ('task.claim', 'Claim task for yourself'),
  ('task.close', 'Close task'),
  ('task.attach_artifact', 'Attach artifact to task'),

  -- Docs / Lab
  ('doc.view', 'View project documents'),
  ('doc.upload', 'Upload project documents'),
  ('doc.delete', 'Delete project documents'),
  ('lab.link_repo', 'Link repository to project'),

  -- Grading
  ('grading.view', 'View grading criteria and score'),
  ('grading.mark_criteria', 'Mark criteria as done/not done'),
  ('grading.publish', 'Publish final score'),

  -- Moderation / Admin
  ('moderation.approve_free_project', 'Approve free project'),
  ('moderation.reject_free_project', 'Reject free project'),
  ('admin.approve_professor', 'Approve professor account'),
  ('admin.manage_rbac', 'Manage roles and permissions'),
  ('audit.view_system', 'View system audit')
ON CONFLICT (code) DO NOTHING;

-- --------------------
-- Seeds: role_permissions (MVP mapping)
-- --------------------

-- STUDENT (FACULTY)
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='STUDENT' AND p.code IN (
  'project.create',
  'team.apply',
  'team.withdraw_application'
)
ON CONFLICT DO NOTHING;

-- TEAM_LEAD (PROJECT)
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='TEAM_LEAD' AND p.code IN (
  'project.edit','project.change_visibility','project.invite_professor','project.submit_for_review',
  'team.accept_applicant','team.reject_applicant','team.submit_roster',
  'task.view','task.create','task.update','task.assign',
  'doc.view','doc.upload','doc.delete',
  'lab.link_repo',
  'grading.view'
)
ON CONFLICT DO NOTHING;

-- MEMBER (PROJECT)
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='MEMBER' AND p.code IN (
  'task.view','task.claim','task.update','task.close','task.attach_artifact',
  'doc.view','doc.upload',
  'grading.view'
)
ON CONFLICT DO NOTHING;

-- PROFESSOR (FACULTY) - base capabilities; project-specific actions should be via PROJECT_PROFESSOR
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='PROFESSOR' AND p.code IN (
  'audit.view_system'
)
ON CONFLICT DO NOTHING;

-- PROJECT_PROFESSOR (PROJECT)
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='PROJECT_PROFESSOR' AND p.code IN (
  'project.approve','project.reject','project.set_criteria','project.set_deadline',
  'team.approve_roster','team.reject_roster',
  'grading.view','grading.mark_criteria','grading.publish',
  'task.view'
)
ON CONFLICT DO NOTHING;

-- MODERATOR (FACULTY)
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='MODERATOR' AND p.code IN (
  'moderation.approve_free_project','moderation.reject_free_project'
)
ON CONFLICT DO NOTHING;

-- SUPER_ADMIN (SYSTEM)
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='SUPER_ADMIN' AND p.code IN (
  'admin.approve_professor','admin.manage_rbac','audit.view_system'
)
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS role_assignments;
DROP TABLE IF EXISTS role_permissions;
DROP TABLE IF EXISTS permissions;
DROP TABLE IF EXISTS roles;