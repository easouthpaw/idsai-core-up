# Admin And User Lifecycle Review

Дата: `2026-04-23`

Scope:

- `internal/repos/postgres/admin_repo.go`
- `internal/services/admin/service.go`
- `internal/http/router_admin.go`
- `internal/repos/postgres/auth_repo.go`
- `migrations/00018_project_criteria_grading.sql`
- `migrations/00020_task_activity_and_submissions.sql`
- `migrations/00031_rbac_tenant_admin.sql`

## Findings

1. High: удаление пользователя не очищает все `ON DELETE RESTRICT` зависимости.
Файлы: `internal/repos/postgres/admin_repo.go`, `migrations/00018_project_criteria_grading.sql`, `migrations/00020_task_activity_and_submissions.sql`.
Детали: `DeleteUser` чистит `role_assignments`, `project_members` и отвязку `projects.professor_id`, но не удаляет `task_submissions.user_id` и `project_criterion_reviews.professor_id`. Оба FK имеют `ON DELETE RESTRICT`.
Риск: админ сможет удалить только “простых” пользователей; пользователи с отправленными task submissions или выставленными criterion reviews могут стать фактически неудаляемыми.
Рекомендация: перед `DELETE FROM users` либо явно чистить эти таблицы, либо перевести FK/архивную стратегию на более мягкую семантику.

2. Medium: роль `TENANT_ADMIN` существует в схеме, но HTTP admin-контур фактически ориентирован только на `SUPER_ADMIN`.
Файлы: `internal/http/router_admin.go`, `internal/http/middleware/auth.go`, `internal/repos/postgres/auth_repo.go`, `migrations/00031_rbac_tenant_admin.sql`.
Детали: admin routes требуют `AdminRequired()`, а `is_admin` в access claims вычисляется только из `SUPER_ADMIN`. При этом `TENANT_ADMIN` получает отдельные permissions (`tenant.manage_users`, `tenant.manage_rbac`), но не проходит через текущий web/API gate.
Риск: схема и runtime-контур расходятся; роль выглядит поддержанной, но в HTTP-слое почти не даёт эффекта.
Рекомендация: либо полноценно провести `TENANT_ADMIN` через claims/middleware/routes, либо явно зафиксировать, что это reserved schema role без UI/API контракта.

3. Medium: создание студента привязано к “первой попавшейся” группе кафедры.
Файлы: `internal/repos/postgres/admin_repo.go`.
Детали: `CreateUser` для студентов автоматически выбирает первую группу через `ORDER BY sg.group_number ASC, sg.created_at ASC LIMIT 1`. Входной API при этом не позволяет администратору явно указать группу.
Риск: новые студенты могут оказываться в неверной группе просто из-за порядка записей в БД.
Рекомендация: передавать `group_id` или `group_code` в admin create API, а fallback на “первую группу” убрать.

## Assumptions

- Под “удалением пользователя” понимается именно административное hard-delete поведение, а не soft-delete с архивом.
- `TENANT_ADMIN` рассматривается как живая продуктовая роль, потому что она добавлена отдельной миграцией и имеет свои permissions.
