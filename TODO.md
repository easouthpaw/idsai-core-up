# TODO: IDSAI Core (Kanban + TDD)

Обновлено: 2026-03-07  
Источник правды по прогрессу: только этот файл.

## Done
- [x] (2026-03-07) Введён единый API namespace `v2` + JWT middleware (`AuthRequired`) для защищённых маршрутов.
- [x] (2026-03-07) Выполнена миграция multi-tenant: таблица `tenants`, `tenant_id` в tenant-bound таблицах, backfill в tenant `CORE`.
- [x] (2026-03-07) Реализованы уведомления `In-app` + `Email` через `notifications` + `notification_outbox` (retry/backoff/dead).
- [x] (2026-03-07) Обновлён встроенный UI (landing/login/author/admin/projects/project) под текущие потоки.
- [x] (2026-03-07) Добавлены Telegram-алерты состояния сервера (startup/failure/recovery/heartbeat).

## In Progress
- [ ] (начато) Полная tenant-isolation во всех репозиториях и SQL-запросах (tenant-aware доступ end-to-end).
- [ ] (начато) Полный lifecycle refresh-токенов: rotation + revoke + logout endpoint.
- [ ] (начато) Полная state-machine проектов и задач: строгие переходы + idempotency + race-safe гарантии.

## Backlog
- [ ] `project-files` API + S3 presigned URL + ACL.
- [ ] Super-admin CRUD: tenant/faculty/department/group/admin (API + UI).
- [ ] `notification-preferences` API и правила доставки.
- [ ] Очистка Swagger от legacy заголовков `X-User-ID` / `X-Faculty-ID`, выравнивание с JWT-only контрактом.
- [ ] Structured JSON logging + correlation/request/tenant/user IDs.
- [ ] CI/CD alignment: Go `1.24` + миграционные проверки + release gates.

## TDD Policy
1. Для каждой новой задачи сначала пишем тесты (unit/integration), затем код.
2. Базовый цикл обязателен: `RED -> GREEN -> REFACTOR`.
3. Задача может быть перенесена в `Done` только при зелёных релевантных тестах.
4. Минимум для backend-задач: unit-тесты бизнес-логики + integration для критичных SQL/API сценариев.
5. Для frontend-задач применяем TDD там, где поведение тестируемо (format/validation/state/HTTP contracts).
6. Каждый task должен явно фиксировать:
   - цель,
   - тесты до кода,
   - реализацию,
   - критерий приёмки.

## Task Template
```md
- [ ] <Короткое название задачи>
  Дата: YYYY-MM-DD
  Цель: <что должно работать>
  Tests First (RED):
  - [ ] Unit: <сценарий>
  - [ ] Integration/E2E: <сценарий>
  Реализация (GREEN):
  - [ ] <изменение 1>
  - [ ] <изменение 2>
  Refactor:
  - [ ] <рефакторинг без смены поведения>
  Критерий приёмки:
  - [ ] Все релевантные тесты зелёные
  - [ ] Обновлены docs/контракты (если менялись)
```

## Test Plan для этого файла
- [x] Проверена структура `TODO.md`: есть `Done / In Progress / Backlog / TDD Policy / Task Template`.
- [x] Задачи разнесены по статусам и готовы для отметки прогресса чекбоксами.
- [ ] На первой новой задаче подтверждаем TDD-процесс: сначала тесты (RED), затем реализация (GREEN), затем перенос в `Done`.
