-- +goose Up
-- +goose StatementBegin

-- 1. Добавляем само разрешение в каталог
INSERT INTO permissions(code, description)
VALUES ('project.view', 'View project details')
ON CONFLICT (code) DO NOTHING;

-- 2. Назначаем это право роли TEAM_LEAD
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='TEAM_LEAD' AND p.code='project.view'
ON CONFLICT DO NOTHING;

-- 3. Назначаем это право роли MEMBER
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='MEMBER' AND p.code='project.view'
ON CONFLICT DO NOTHING;

-- 4. Назначаем это право роли PROJECT_PROFESSOR
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.code='PROJECT_PROFESSOR' AND p.code='project.view'
ON CONFLICT DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

-- Удаляем связи ролей с этим разрешением
DELETE FROM role_permissions
WHERE permission_id = (SELECT id FROM permissions WHERE code='project.view');

-- Удаляем само разрешение из каталога
DELETE FROM permissions WHERE code='project.view';

-- +goose StatementEnd