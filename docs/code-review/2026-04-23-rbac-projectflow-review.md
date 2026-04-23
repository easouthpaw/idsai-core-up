# RBAC And Projectflow Review

Дата: `2026-04-23`

Scope:

- `internal/repos/postgres/projectflow_repo_access.go`
- `internal/services/projectflow/service_member_access.go`
- `internal/modules/rbac/module.go`
- `internal/http/router_projectflow.go`
- `migrations/00002_rbac.sql`

## Findings

1. High: custom project access roles используют глобальный namespace `roles.code`, хотя бизнес-смысл у них project-scoped.
Файлы: `internal/repos/postgres/projectflow_repo_access.go`, `migrations/00002_rbac.sql`.
Детали: `CreateProjectAccessRole` делает `INSERT INTO roles(code, name)`, а таблица `roles` имеет глобальный `UNIQUE` по `code`. Это значит, что два разных проекта или tenant не смогут независимо создать одинаковый код вроде `QA_REVIEWER`, даже если роль нужна только внутри своего проекта.
Риск: collision между проектами, невозможность переиспользовать понятные role codes, сложность multi-tenant эволюции.
Рекомендация: вынести custom project roles в отдельную сущность с уникальностью на `(tenant_id, project_id, code)` и не хранить их как глобальные записи в `roles`.

2. Medium: lifecycle custom project roles пока append-only.
Файлы: `internal/http/router_projectflow.go`, `internal/repos/postgres/projectflow_repo_access.go`, `internal/services/projectflow/service_member_access.go`.
Детали: есть создание роли и замена назначений участнику, но нет update/delete потока для самой роли и её `role_permissions`. После создания роль и её permission map остаются в системном каталоге навсегда.
Риск: накопление stale roles, неочевидные permission leftovers, осложнение администрирования.
Рекомендация: добавить операции archive/delete для custom role и явную очистку `project_access_roles`, `role_permissions`, а при необходимости и orphan `roles`.

3. Medium: встроенная ABAC-регистрация подключена к `task.edit`, а реальные HTTP permissions живут на `task.update`.
Файлы: `internal/modules/rbac/module.go`, `internal/http/router_projectflow.go`, `migrations/00002_rbac.sql`.
Детали: registry регистрирует условие для `task.edit`, но маршруты задач используют `task.update`. В текущем виде этот ABAC hook не участвует в реальных checks, если только какой-то внешний caller не начнет отдельно вызывать `CanWithAttributes(..., "task.edit", ...)`.
Риск: ложное ощущение, что author-based ABAC уже работает в runtime, хотя маршрутный слой его не использует.
Рекомендация: либо перевести условие на `task.update`, либо удалить/отложить ABAC hook до полноценного интегрированного сценария.

## Assumptions

- Ревью опирается на текущий runtime-контур и HTTP routes, а не на исторические миграции сами по себе.
- Custom delegated roles рассматриваются как продуктовая функция, а не как внутренний admin-only эксперимент.
