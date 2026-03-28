-- +goose Up

-- 1. New permission for managing delegated project roles.
INSERT INTO permissions(code, description) VALUES
  ('member.access.manage', 'Manage delegated project roles for members')
ON CONFLICT (code) DO NOTHING;

-- 2. New delegated project roles.
INSERT INTO roles(code, name) VALUES
  ('CO_LEAD', 'Co-Lead'),
  ('RECRUITER', 'Recruiter'),
  ('TASK_MANAGER', 'Task Manager')
ON CONFLICT (code) DO NOTHING;

-- 3. Grant member.access.manage to TEAM_LEAD only.
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'TEAM_LEAD'
  AND p.code = 'member.access.manage'
ON CONFLICT DO NOTHING;

-- 4. CO_LEAD permissions.
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'CO_LEAD'
  AND p.code IN (
    'project.view',
    'project.edit',
    'project.invite_professor',
    'project.submit_for_review',
    'position.create',
    'member.approve',
    'task.create',
    'task.assign'
  )
ON CONFLICT DO NOTHING;

-- 5. RECRUITER permissions.
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'RECRUITER'
  AND p.code IN (
    'project.view',
    'member.approve'
  )
ON CONFLICT DO NOTHING;

-- 6. TASK_MANAGER permissions.
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r, permissions p
WHERE r.code = 'TASK_MANAGER'
  AND p.code IN (
    'project.view',
    'task.create',
    'task.assign'
  )
ON CONFLICT DO NOTHING;

-- +goose Down

-- Remove role_permissions for delegated roles.
DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE code IN ('CO_LEAD', 'RECRUITER', 'TASK_MANAGER'));

-- Remove member.access.manage from TEAM_LEAD.
DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE code = 'TEAM_LEAD')
  AND permission_id = (SELECT id FROM permissions WHERE code = 'member.access.manage');

-- Remove delegated roles.
DELETE FROM roles WHERE code IN ('CO_LEAD', 'RECRUITER', 'TASK_MANAGER');

-- Remove the permission.
DELETE FROM permissions WHERE code = 'member.access.manage';
