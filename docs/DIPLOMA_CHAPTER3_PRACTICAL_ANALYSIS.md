# Практический анализ проекта для 3-й главы диплома

## 3.1 Бағдарламалық қамтамасыз етуді іске асырудың қолданылатын технологиялары мен құралдары

Практически здесь можно писать не про "почему вообще RBAC делают так", а про то, как проект реально собран: `Go`-монолит с модульной сборкой, `Gin`-router, `PostgreSQL` через `pgx`, `goose`-миграции, встроенный frontend через `go:embed`, JWT-сессии, Redis-кэш для RBAC и S3-compatible storage для медиа. Архитектурно это лучше формулировать как `modular monolith + composition root + переход к ports/adapters`, потому что это прямо зафиксировано в `docs/architecture/ARCHITECTURE.md` и подтверждается `internal/app/wire_modules.go`, `internal/modules/*`, `internal/services/*`, `internal/repos/postgres/*`, `internal/http/router_*.go`.

| Технология | Назначение | Где используется |
| --- | --- | --- |
| `Go 1.25` | основной backend-код | `go.mod`, `cmd/api/main.go`, `cmd/migrate/main.go` |
| `Gin` | HTTP-layer, route groups, middleware chain | `internal/http/router.go`, `internal/http/router_*.go` |
| `PostgreSQL` | основная БД | `docker-compose.yml`, `internal/db/db.go`, `internal/repos/postgres/*` |
| `pgx/v5` | pool + driver | `internal/db/db.go`, `cmd/migrate/main.go`, `go.mod` |
| `Redis` | кэш permission-checks с graceful degradation | `internal/infra/cache/redis.go`, `internal/modules/rbac/module.go` |
| `MinIO / S3` | хранение медиа | `internal/infra/storage/minio.go`, `docker-compose.yml` |
| `local storage` fallback | локальное хранение медиа через `/media` | `internal/infra/storage/local.go`, `internal/app/app.go` |
| `goose` | SQL-миграции | `cmd/migrate/main.go`, `Makefile`, `migrations/*` |
| `Swagger/OpenAPI` | артефакты API-документации | `docs/swagger/swagger.yaml`, `docs/swagger/swagger.json`, `docs/swagger/docs.go` |
| встроенный frontend | страницы `/dev/*` без отдельного SPA-репозитория | `internal/http/frontend/*`, `internal/http/router_dev.go`, `internal/http/handlers/dev_tester_handler.go` |
| `JWT` | access/refresh auth | `internal/services/auth/service.go`, `internal/http/middleware/auth.go` |
| `Docker / Compose` | локальная инфраструктура и сборка | `Dockerfile`, `docker-compose.yml`, `Makefile` |

Что можно описать в тексте 3.1:

- HTTP-слой находится в `internal/http`; бизнес-логика в `internal/services`; PostgreSQL-adapters в `internal/repos/postgres`; инфраструктура в `internal/infra`; модульная сборка в `internal/modules`; composition root в `internal/app`.
- Router уже декомпозирован по модулям: `router_auth.go`, `router_admin.go`, `router_projects.go`, `router_projectflow.go`, `routes_kb.go`, `router_dev.go`.
- В проекте нет отдельного `package.json`/frontend build pipeline; UI shipped как embedded HTML/CSS/JS.
- Для RBAC выбраны собственные `Scope`, `Authorizer`, `middleware.RequirePermission*`, SQL-resolver в `rbac_repo.go` и миграции `00002/00011/00015/00038/00043`.
- Swagger-файлы в репозитории есть, но явный runtime-route для Swagger UI в Gin-коде явно не найден.
- Схема `TENANT_ADMIN` есть в миграциях, но основной admin web/API gate завязан на `is_admin`, который в `auth_repo.go` вычисляется по `SUPER_ADMIN`.

Что вставить как рисунки/скриншоты:

- дерево проекта: `internal/http`, `internal/services`, `internal/repos/postgres`, `internal/modules`, `migrations`;
- `docker-compose.yml`;
- `internal/app/wire_modules.go` как схема composition root;
- любой RBAC migration example: `00038_rbac_delegated_project_roles.sql` или `00043_project_custom_access_roles.sql`;
- `docs/swagger/swagger.yaml` как факт наличия OpenAPI-артефактов.

Что не повторять из 2-й главы:

- общую теорию выбора `Go/Gin/PostgreSQL/JWT`;
- абстрактное сравнение monolith vs microservices;
- общую теорию RBAC и scope-моделей.

Опора:

- `go.mod`
- `docker-compose.yml`
- `internal/app/wire_modules.go`
- `docs/architecture/ARCHITECTURE.md`
- `internal/http/frontend/frontend.go`
- `cmd/migrate/main.go`

## 3.2 Артықшылықтарды тексерудің бағдарламалық логикасын дайындау және іске асыру

Фактическая цепочка доступа такая: route group -> `AuthRequired` -> при необходимости `AdminRequired` -> `RequirePermission`/`RequirePermissionIf` -> handler -> service-level guard -> repo/SQL checks. Пользователь берется из cookie `idsai_access` или `Authorization: Bearer ...`; JWT claims содержат `tenant_id`, `faculty_id`, `department_id`, `is_admin`, `is_professor`, `pwd_at`; потом это кладется в Gin context и `requestctx.WithIdentity(...)`. Scope-уровни в коде: `SYSTEM`, `TENANT`, `FACULTY`, `DEPARTMENT`, `PROJECT`; resolver'ы найдены: `SystemScope`, `TenantScopeFromCtx`, `FacultyScopeFromCtx`, `DepartmentScopeFromCtx`, `ProjectScopeFromParam`, `FacultyScopeFromHeader`. Других runtime-resolver'ов явно не найдено.

Важные практические факты:

- В `internal/repos/postgres/rbac_repo.go` проверка `PROJECT` разворачивает scope в `SYSTEM -> TENANT -> FACULTY -> PROJECT`; `DEPARTMENT` для project-наследования не участвует.
- `RequireAllPermissions`, `RequirePermissionWithAttrs`, `CanAll`, `CanWithAttributes` реализованы и покрыты тестами/benchmark, но реальные route bindings для них в runtime явно не найдены.
- ABAC registry есть в `internal/modules/rbac/module.go`, но он зарегистрирован на `task.edit`; реальные projectflow-routes используют `task.update`, поэтому фактическое runtime-применение этого ABAC hook явно не найдено.
- `RBAC_ENFORCE_ADMIN_PERMS`, `RBAC_ENFORCE_PROJECTS_GET`, `RBAC_ENFORCE_PROJECTFLOW` позволяют включать/выключать enforcement на router-слое.

| Операция | Permission | Scope | Кто может по текущим migrations | Где проверяется |
| --- | --- | --- | --- | --- |
| `POST /v2/projects` | `project.create` | `FACULTY` | `STUDENT` | `router_projects.go`, `projects.Service.CreateProject`, `00002_rbac.sql` |
| `POST /v2/projects/:id/recruitment/open` | `project.edit` | `PROJECT` | `TEAM_LEAD`, `CO_LEAD`, custom-role с `project.edit` | `router_projectflow.go`, `service_members.go:OpenRecruitment`, `00038/00043` |
| `POST /members/apply` | `member.apply` | route check идет на `PROJECT`, но срабатывает faculty-grant через hierarchy | `STUDENT` | `router_projectflow.go`, `service_members.go:ApplyMember`, `00008_rbac` |
| `POST /members/:user_id/approve` | `member.approve` | `PROJECT` | `TEAM_LEAD`, `CO_LEAD`, `RECRUITER`, custom-role | `router_projectflow.go`, `service_members.go:ApproveMember`, `00008/00038/00043` |
| `POST /professor` | `project.invite_professor` | `PROJECT` | `TEAM_LEAD`, `CO_LEAD`, custom-role | `router_projectflow.go`, `service_professors.go:AssignProfessor`, `00002/00038/00043` |
| `POST /professor/respond`, `GET /v2/professor/review-invites` | `project.review.respond` | `FACULTY` | `PROFESSOR` | `router_projectflow.go`, `service_professors.go:RespondProfessorInvite`, `00033_rbac_project_security_permissions.sql` |
| `POST /criteria` | `project.set_criteria` | `PROJECT` | по текущим seeds найден `PROJECT_PROFESSOR`; отдельный grant для `TEAM_LEAD` явно не найден | `router_projectflow.go`, `service_grading.go:CreateCriterion`, `00002_rbac.sql` |
| `POST /approve` | `project.approve` | `PROJECT` | по текущим seeds найден `PROJECT_PROFESSOR`; затем еще проверяется `readiness.can_activate` | `router_projectflow.go`, `service_grading.go:ApproveProject`, `00002_rbac.sql` |
| `POST /grading/submit` | `project.submit_for_review` | `PROJECT` | `TEAM_LEAD`, `MEMBER`, `CO_LEAD`, custom-role; service допускает fallback для active member | `router_projectflow.go`, `service_access.go`, `service_grading.go`, `00002/00022/00038/00043` |
| `POST /grading/publish`, `POST /grading/return` | `grading.publish` | `PROJECT` | `PROJECT_PROFESSOR`; service дополнительно требует `project.professor_id == caller` | `router_projectflow.go`, `service_grading.go:PublishGrading/ReturnProjectForRetake`, `00002_rbac.sql` |
| `DELETE /v2/projects/:id` | `project.delete` + owner check | `PROJECT` | middleware: `TEAM_LEAD`; repo: только `created_by` | `router_projectflow.go`, `projectflow_repo_lifecycle.go:DeleteOwnedProject`, `00033_rbac_project_security_permissions.sql` |
| `GET/POST /access/*`, `GET/PUT /members/:user_id/access` | `member.access.manage` | `PROJECT` | `TEAM_LEAD` и custom-role, которому вручную дали этот код | `router_projectflow.go`, `service_member_access.go`, `00038/00043` |
| `POST/PATCH/DELETE /tasks*` | `task.create/update/assign/delete/claim` | `PROJECT` | `TEAM_LEAD`, `MEMBER` частично, `TASK_MANAGER`, `CO_LEAD`, custom-role по каталогу прав | `router_projectflow.go`, `service_tasks.go`, `00002/00038/00040/00043` |

Лучшие сценарии для диплома:

- создание проекта: создается `project_members` row и сразу выдается `TEAM_LEAD`;
- recruitment: `member.apply` -> `ApproveMember` -> выдача `MEMBER`;
- professor flow: приглашение -> `project.review.respond` -> выдача `PROJECT_PROFESSOR`;
- launch/grading/publish: readiness-check, submit, grading, retake, publish;
- delegated access: `CO_LEAD` / `RECRUITER` / `TASK_MANAGER` и custom project-roles через `project_access_roles`;
- удаление проекта: route-level permission + repo-level owner-check;
- BOLA-сценарий на `GET /members/:user_id/access`.

Миграции, которые стоит цитировать:

- `00002_rbac.sql` — базовые роли, permissions, `role_assignments`;
- `00011_rbac_department_scope.sql` — `DEPARTMENT`;
- `00015_multitenant_notifications_docs.sql` — `TENANT` + `tenant_id`;
- `00030_rbac_role_assignments_unique.sql` — уникальность assignment tuple;
- `00031/00032` — `TENANT_ADMIN` и backfill;
- `00033` — `project.delete`, `project.review.respond`;
- `00038` — `member.access.manage`, `CO_LEAD`, `RECRUITER`, `TASK_MANAGER`;
- `00040` — `task.delete`;
- `00043` — `project_access_roles`;
- `00044` — удаление `MODERATOR`.

Тесты и метрики, которые можно использовать:

- BOLA: `.tmp/report/security-bola.txt` и `project_flow_handler_transport_test.go` — `403 Forbidden` для подмены `user_id`;
- lifecycle: `internal/http/handlers/project_lifecycle_e2e_test.go`;
- middleware: `internal/http/middleware/rbac_test.go`, `rbac_additional_test.go`, `auth_additional_test.go`, `auth_cookie_test.go`;
- RBAC service: `internal/services/rbac/service_test.go`, `cached_authorizer_test.go`, `service_benchmark_test.go`;
- coverage из `.tmp/report/coverage-func.txt`: `internal/services/rbac 94.7%`, `internal/http/middleware 86.8%`, `internal/services/projectflow 64.5%`, `whole project 34.7%`;
- benchmarks из `.tmp/report/bench-rbac.txt` и `.tmp/report/bench-middleware.txt`: `Can(ProjectScope) 61.45 ns/op`, `CanWithAttributes 112.5 ns/op`, `CanAll x5 298.1 ns/op`, `RequirePermission 3735 ns/op`, `RequireAllPermissions 3782 ns/op`, `RequirePermissionWithAttrs 3936 ns/op`.

Что вставить как рисунки/схемы:

- chain `AuthRequired -> RBAC middleware -> handler -> service -> repo`;
- scope hierarchy с `SYSTEM/TENANT/FACULTY/DEPARTMENT/PROJECT`;
- route-to-permission matrix по `projectflow`;
- lifecycle проекта с точками `role check`.

Что не повторять из 2-й главы:

- теорию hierarchical RBAC как модели;
- абстрактные таблицы "роль/право" без привязки к route и migration;
- определения scope inheritance без SQL-реализации `resolved_scopes`.

Опора:

- `internal/http/middleware/auth.go`
- `internal/http/middleware/rbac.go`
- `internal/http/router_projectflow.go`
- `internal/services/projectflow/service_access.go`
- `internal/services/projectflow/service_grading.go`
- `internal/repos/postgres/rbac_repo.go`
- `migrations/00002_rbac.sql`
- `migrations/00038_rbac_delegated_project_roles.sql`
- `migrations/00043_project_custom_access_roles.sql`

## 3.3 Пайдаланушы интерфейсін қорғау және рұқсатсыз кіру әрекеттерін анықтау

Во frontend реально есть страницы `landing`, `author`, `login`, `admin`, `projects`, `project`, `invites`, `professor`, `professor-reviews`, `professor-criteria`, `professor-grading`, `settings`, `profile`, `groups`, `kb`, `kb-article`, `404`. UI определяет роль и доступ через `/v2/auth/me`, default-scope capabilities через `/v2/auth/capabilities`, а project-specific кнопки — через `/v2/projects/:id/my-permissions`. Это важно: скрытие кнопки само по себе не считается защитой, потому что backend почти везде повторяет проверку через middleware или service/repo guard.

| Тип попытки | Механизм защиты | Где реализовано | Результат |
| --- | --- | --- | --- |
| нет токена | `AuthRequired` | `middleware/auth.go` | `401 {"error":"missing auth token"}` |
| битый/подмененный JWT | parse + issuer/method checks | `middleware/auth.go` | `401 {"error":"invalid token"}` |
| токен после смены пароля, blocked/inactive, unverified | stateReader + `pwd_at` + user status/email verification | `middleware/auth.go`, `auth_repo.GetUserAuthState` | `401 {"error":"invalid token"}` |
| cross-site POST с cookie | `CSRFProtection` | `middleware/csrf.go`, `csrf_test.go` | `403 {"error":"csrf check failed"}` |
| студент идет в `/dev/admin` или `/dev/professor` | `ensureSession(...)` + redirect | `frontend/js/admin.js`, `professor*.js`, `auth-session.js` | redirect на свой workspace |
| скрытые UI-действия вызываются напрямую через API | route middleware + service/repo guard | `router_projectflow.go`, `service_grading.go`, `projectflow_repo_lifecycle.go`, `kb_handler.go` | `403` / `400` / `404` |
| подмена чужого `user_id` в project access | permission `member.access.manage` | `project_flow_handler_transport_test.go`, `service_member_access.go` | BOLA test дает `403 Forbidden` |
| invalid scope / bad UUID route param | scope resolvers + UUID parse | `middleware/rbac.go`, `project_flow_handler.go` | `400 {"error":"invalid scope"}` / `invalid project_id` |

Практические факты для текста 3.3:

- `auth-session.js` кэширует `user_id`, `faculty_id`, `is_admin`, `is_professor` в `localStorage`, а затем `role-sidebar.js` скрывает admin menu items по `auth.canCached("admin.manage_rbac")`.
- `project-detail.js` скрывает/отключает кнопки по `state.myPermissions`: `member.access.manage`, `project.invite_professor`, `task.create`, `task.assign`, `task.update`, `task.delete`, `project.approve`.
- `admin.js` требует `ensureSession("admin")`; `professor.js`/`professor-reviews.js`/`professor-criteria.js`/`professor-grading.js` требуют `ensureSession("professor")`; `projects.js` и `invites.js` не пускают admin/professor в student workspace.
- `groups.js` UI пускает только admin/professor, и backend это повторно проверяет в `AuthHandler.ListDepartmentGroupsTree`.
- `kb.js` и `kb-article.js` показывают editor controls только если `profile.is_admin || profile.is_professor`; backend повторно режет create/update/delete через `KBHandler.requireEditor`.
- `GET /v2/projects/:project_id` не опирается на UI-флаги: сервис сам решает, что можно видеть (`public`, creator, `grading.view`, resolved `project.view`).
- В `requestJSON(...)` на `403` frontend показывает user-facing alert, а на `401` чистит client state и кидает на `/dev/login`.

Что по логированию и выявлению несанкционированных действий:

- denial-лог есть: `rbac_deny status=... method=... path=... user_id=... tenant_id=... permission=... scope_type=... scope_id=... reason=...`;
- причины, реально зашитые в код: `missing_user`, `invalid_scope`, `denied`, `authorizer_error`, `abac_denied`;
- общий request log пишет `status/method/path/latency/ip`;
- отдельной audit-table для unauthorized access, Prometheus metrics endpoint или security-alert pipeline по RBAC denial в репозитории явно не найдено;
- внутренние counters `rbacUnauthorizedCount`, `rbacBadScopeCount`, `rbacDeniedCount` есть, но наружу не экспортируются.

Тесты/доказательства для защиты интерфейса:

- `auth_additional_test.go` и `auth_cookie_test.go` — missing token, invalid token, invalid claims, blocked/unverified state, cookie-over-header;
- `csrf_test.go` — защита от cross-site cookie mutation;
- `rbac_test.go` и `rbac_additional_test.go` — missing user, invalid scope, forbidden, `RequireAllPermissions`, `RequirePermissionWithAttrs`;
- `project_flow_handler_transport_test.go` — BOLA `403`;
- `.tmp/report/security-bola.txt`, `.tmp/report/bench-*.txt`, `.tmp/report/coverage-func.txt` — готовые артефакты для приложений;
- coverage для security-critical частей уже высокая именно там, где нужна аргументация: `RBAC 94.7%`, `middleware 86.8%`.

Что вставить как скриншоты/схемы:

- `/dev/admin`, `/dev/professor`, `/dev/projects`, `/dev/projects/:id`, `/dev/groups`, `/dev/kb`;
- пример `401 missing auth token` и `403 forbidden`;
- project page до/после загрузки `my-permissions`;
- BOLA terminal result из `.tmp/report/security-bola.txt`;
- схема `/auth/me` + `/auth/capabilities` + `my-permissions` -> UI gating.

Что не повторять:

- общую теорию BOLA, CSRF, JWT и unauthorized access;
- рассуждения "security by obscurity плоха" без кода;
- общие UX-фразы про роли без конкретных frontend/backend checks.

Опора:

- `internal/http/router_dev.go`
- `internal/http/frontend/frontend.go`
- `internal/http/frontend/js/auth-session.js`
- `internal/http/frontend/js/role-sidebar.js`
- `internal/http/frontend/js/project-detail.js`
- `internal/http/handlers/kb_handler.go`
- `docs/BOLA_TESTING.md`
- `.tmp/report/security-bola.txt`

## Итог

1. Для `3.1` уже достаточно собрано: стек, runtime-архитектура, composition root, модульная маршрутизация, встроенный frontend, миграции, Docker/Compose, Redis/MinIO/JWT, OpenAPI-артефакты и точные файлы-опоры.
2. Для `3.2` уже достаточно собрано: полная цепочка auth/RBAC, scope types/resolvers, route-to-permission mapping по `projects`, `projectflow`, `admin`, `auth`, сервисные и repo-level защитные проверки, роль-модель из миграций, delegated/custom project roles, BOLA test, benchmarks и coverage.
3. Для `3.3` уже достаточно собрано: список frontend-страниц, role/UI gating через `/auth/me`, `/auth/capabilities`, `/my-permissions`, обязательные backend-checks, реакции на `401/403/400`, RBAC deny logging, CSRF protection, security tests и готовые report-artifacts.
4. Чего еще не хватает и что лучше снять вручную:
   - скриншоты: дерево проекта, `docker-compose.yml`, `wire_modules.go`, `00038` или `00043` migration, `/dev/admin`, `/dev/projects`, `/dev/projects/:id`, `/dev/professor`, `/dev/groups`, `/dev/kb`;
   - команды: `go test ./internal/http/handlers -run 'TestProjectFlowHandlerGetMemberAccess' -v`, `go test ./internal/services/rbac -run '^$' -bench . -benchmem -benchtime=2s`, `go test ./internal/http/middleware -run '^$' -bench . -benchmem -benchtime=2s`, `make coverage`, `make report-artifacts`;
   - таблицы: `технология -> назначение -> где`, `операция -> permission -> scope -> кто может -> где проверяется`, `угроза -> механизм -> файл -> результат`;
   - артефакты: сохранить `.tmp/report/security-bola.txt`, `.tmp/report/bench-rbac.txt`, `.tmp/report/bench-middleware.txt`, `.tmp/report/coverage-func.txt`;
   - страницы для открытия: `/dev/login`, `/dev/admin`, `/dev/projects`, `/dev/projects/{id}`, `/dev/professor`, `/dev/professor/reviews`, `/dev/groups`, `/dev/kb`.
5. Отдельно: минимальный набор артефактов из проекта для написания 3-й главы.

### Минимальный набор артефактов из проекта для написания 3-й главы

- [ ] `README.md`
- [ ] `go.mod`
- [ ] `docker-compose.yml`
- [ ] `Dockerfile`
- [ ] `Makefile`
- [ ] `internal/app/wire_modules.go`
- [ ] `internal/http/router.go` и `internal/http/router_*.go`
- [ ] `internal/http/middleware/auth.go`
- [ ] `internal/http/middleware/rbac.go`
- [ ] `internal/services/rbac/*`
- [ ] `internal/repos/postgres/rbac_repo.go`
- [ ] `internal/services/projectflow/service_access.go`
- [ ] `internal/services/projectflow/service_members.go`
- [ ] `internal/services/projectflow/service_professors.go`
- [ ] `internal/services/projectflow/service_grading.go`
- [ ] `internal/services/projectflow/service_tasks.go`
- [ ] `internal/services/projectflow/service_member_access.go`
- [ ] `internal/services/projects/service.go`
- [ ] `internal/repos/postgres/projectflow_repo_access.go`
- [ ] `internal/repos/postgres/projectflow_repo_lifecycle.go`
- [ ] `migrations/00002_rbac.sql`
- [ ] `migrations/00011_rbac_department_scope.sql`
- [ ] `migrations/00015_multitenant_notifications_docs.sql`
- [ ] `migrations/00033_rbac_project_security_permissions.sql`
- [ ] `migrations/00038_rbac_delegated_project_roles.sql`
- [ ] `migrations/00040_task_delete_permission.sql`
- [ ] `migrations/00043_project_custom_access_roles.sql`
- [ ] `migrations/00044_remove_legacy_project_files_and_moderator.sql`
- [ ] `internal/http/frontend/frontend.go`
- [ ] `internal/http/frontend/js/auth-session.js`
- [ ] `internal/http/frontend/js/role-sidebar.js`
- [ ] `internal/http/frontend/js/project-detail.js`
- [ ] `internal/http/frontend/js/admin.js`
- [ ] `internal/http/frontend/js/groups.js`
- [ ] `internal/http/frontend/js/kb.js`
- [ ] `internal/http/handlers/project_flow_handler_transport_test.go`
- [ ] `internal/http/handlers/project_lifecycle_e2e_test.go`
- [ ] `internal/http/middleware/auth_additional_test.go`
- [ ] `internal/http/middleware/rbac_test.go`
- [ ] `internal/http/middleware/rbac_benchmark_test.go`
- [ ] `internal/services/rbac/service_benchmark_test.go`
- [ ] `docs/architecture/ARCHITECTURE.md`
- [ ] `docs/architecture/RBAC_HIERARCHY.md`
- [ ] `docs/PROJECT_LIFECYCLE.md`
- [ ] `docs/BOLA_TESTING.md`
- [ ] `.tmp/report/security-bola.txt`
- [ ] `.tmp/report/bench-rbac.txt`
- [ ] `.tmp/report/bench-middleware.txt`
- [ ] `.tmp/report/coverage-func.txt`
