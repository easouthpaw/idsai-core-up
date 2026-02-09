-- +goose Up

INSERT INTO permissions(code, description) VALUES
('position.create','Create project positions'),
('member.apply','Apply to join project'),
('member.approve','Approve member for project'),
('task.claim','Claim a task')
ON CONFLICT (code) DO NOTHING;

-- TEAM_LEAD
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='TEAM_LEAD' AND p.code IN ('position.create','member.approve','task.claim')
ON CONFLICT DO NOTHING;

-- STUDENT
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='STUDENT' AND p.code IN ('member.apply')
ON CONFLICT DO NOTHING;

-- MEMBER
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='MEMBER' AND p.code IN ('task.claim')
ON CONFLICT DO NOTHING;

-- +goose Down

DELETE FROM role_permissions
WHERE permission_id IN (SELECT id FROM permissions WHERE code IN ('position.create','member.apply','member.approve','task.claim'));

DELETE FROM permissions WHERE code IN ('position.create','member.apply','member.approve','task.claim');