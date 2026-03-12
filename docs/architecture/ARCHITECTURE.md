# Архитектура IDSAI Core

Дата фиксации: 2026-03-12

## Выбранный подход

Для проекта принят подход `Modular Monolith + Clean (Ports & Adapters)`.

Почему это подходит:
- один деплой и одна БД сохраняют простоту эксплуатации;
- домен уже широкий (auth/admin/projects/projectflow/notifications/rbac), поэтому нужна изоляция по модулям;
- можно делать миграцию постепенно без "big-bang" переписывания.

## Базовые принципы

1. Модуль = bounded context (например, `projectflow`), а не техническая папка.
2. Зависимости направлены внутрь:
   - `domain` не знает ни о БД, ни об HTTP;
   - `app` (use cases) зависит от `ports`;
   - `adapters` (postgres/http/email/telegram) реализуют `ports`.
3. SQL живет только в adapter-слое (`postgres`), не в use-case сервисах.
4. `cmd/api` и `internal/app` остаются composition root (сборка зависимостей).

## Целевая структура

```text
internal/
  modules/
    auth/
      domain/
      app/
      ports/
      adapters/
        postgres/
        http/
    projectflow/
      domain/
      app/
      ports/
      adapters/
        postgres/
        http/
  platform/
    db/
    http/
    observability/
```

Текущие папки (`services`, `repos/postgres`, `http/handlers`) считаются переходным состоянием и будут переноситься по модулям.

## Текущие архитектурные проблемы

- Код модулей всё ещё физически распределён по переходным папкам `services/repos/http`; `internal/modules/*` сейчас используется как явный модульный composition-слой.
- HTTP composition остается централизован в `internal/http/router.go`, хотя регистрация маршрутов уже декомпозирована по модульным `register*Routes` файлам.

## Правила зависимостей

В проект добавлен архитектурный тест (`internal/architecture/dependency_rules_test.go`), который проверяет:

- `domain` не импортирует `http/repos/infra/app`;
- `services` не импортируют `http/app/repos/postgres`;
- `services` не должны тащить transport/DB-фреймворки напрямую (`gin`, `pgx`) кроме временных legacy-исключений;
- `http` не импортирует `repos/postgres`;
- `repos` не импортируют `http`.

## Legacy-исключения (временные)

- На текущем этапе legacy-исключения для service-слоя убраны.

## Прогресс миграции

- 2026-03-12: вынесены `stacks` и `positions` операции из `projectflow.Service` в postgres-adapter
  (`internal/repos/postgres/projectflow_repo.go`) через порты
  (`internal/services/projectflow/ports.go`).
- 2026-03-12: вынесены `members/invites` операции (`invite/apply/respond/list/approve/set-position`)
  из `projectflow.Service` в тот же postgres-adapter через `MembersRepository`.
- 2026-03-12: вынесен `professor` блок (`search/assign/get/respond/list-review-invites`)
  из `projectflow.Service` в тот же postgres-adapter через `ProfessorsRepository`.
- 2026-03-12: вынесен `criteria/grading` блок (`create/list criteria`, `get/upsert grading`,
  `criteria counts`) из `projectflow.Service` в postgres-adapter через `CriteriaRepository`.
- 2026-03-12: вынесен `project lifecycle` блок (`approve`, `submit/publish grading`, `delete`,
  `task summary for grading`) в postgres-adapter через `LifecycleRepository`.
- 2026-03-12: вынесен `tasks` блок (`create/list/update/assign/claim/complete`,
  `activity log`, `submissions`) в postgres-adapter через `TasksRepository`.
- 2026-03-12: вынесены оставшиеся `project-level` операции (`project by id`, `project update`,
  `open recruitment`, `active-member check`, `project-role check/revoke`,
  `student candidates`) в postgres-adapter через `ProjectsRepository`.
- `internal/services/projectflow/service.go` больше не содержит прямых SQL-запросов и не зависит от `pgxpool`.
- 2026-03-12: для `projectflow` убран прямой импорт `pgx/pgconn` из service-слоя, `ErrNoRows`/`undefined relation` изолированы в adapter-слое через доменные ошибки.
- 2026-03-12: `projectflow` service декомпозирован по use-case файлам (`service_access`, `service_members`, `service_professors`, `service_grading`, `service_tasks`, `service_utils`) без изменения контрактов.
- 2026-03-12: `projectflow` postgres-adapter декомпозирован по use-case файлам (`projectflow_repo_projects`, `projectflow_repo_members`, `projectflow_repo_professors`, `projectflow_repo_criteria`, `projectflow_repo_lifecycle`, `projectflow_repo_tasks`, `projectflow_repo_helpers`) без изменения SQL-контрактов.
- 2026-03-12: `internal/http/router.go` декомпозирован на модульные регистраторы маршрутов (`router_auth`, `router_admin`, `router_projects`, `router_notifications`, `router_projectflow`, `router_dev`) без изменения URL-контрактов и middleware-цепочек.
- 2026-03-12: `internal/app/app.go` упрощён через модульные wire-функции (`wireModules`, `startEmailOutboxDispatcher`) без изменения сборки зависимостей и runtime-поведения.
- 2026-03-12: `internal/http/handlers/project_flow_handler.go` декомпозирован на use-case файлы (`project_flow_handler_project`, `project_flow_handler_members`, `project_flow_handler_professors`, `project_flow_handler_grading`, `project_flow_handler_tasks`) без изменения HTTP-контрактов.
- 2026-03-12: `internal/http/handlers/admin_handler.go` декомпозирован на `admin_handler_users` и `admin_handler_projects` без изменения контрактов admin API.
- 2026-03-12: `internal/http/handlers/projects_handler.go` декомпозирован на `projects_handler_create` и `projects_handler_read`, при этом общие DTO/mapper-хелперы сохранены в базовом файле для переиспользования в `projectflow`.
- 2026-03-12: для `admin/projects/notifications` перенесено маппирование `not found` ошибок в repo/service слой; HTTP handlers больше не импортируют `pgx/pgconn` и не зависят от DB-ошибок напрямую.
- 2026-03-12: в `internal/modules/*` добавлены модульные `New(...)` конструкторы (`auth/admin/rbac/projects/projectflow/notifications`), и `internal/app/wire_modules.go` переведён на модульную сборку.
- 2026-03-12: `health` handler переведён на абстракцию `DBPinger`, после чего в `dependency_rules_test` добавлен guard: `internal/http/handlers` не должны импортировать `pgx/pgconn`.

## План миграции (без остановки разработки)

1. Зафиксировать правила и документ (сделано).
2. Разбить `projectflow` на use-cases и вынести SQL в `projectflow/adapters/postgres`.
3. Переносить остальные модули по одному: `auth`, `admin`, `projects`, `notifications`, `rbac` (выполнено на уровне composition и границ).
4. Упростить `internal/app` через модульные `Wire`-функции (выполнено).
5. После переноса убрать legacy-исключения из архитектурного теста (выполнено для service-слоя; дополнительные guards добавлены для handlers).
