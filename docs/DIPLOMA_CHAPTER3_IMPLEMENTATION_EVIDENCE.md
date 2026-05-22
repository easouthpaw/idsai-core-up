# Chapter 3 Implementation Evidence

Бұл құжат 3.1, 3.2, 3.3 бөлімдері үшін репозиторийдегі нақты файлдарды, олардың рөлін және дипломда қолдануға болатын негізгі техникалық ақпаратты бір жерге жинайды.

## 3.1. Жоба Құрылымы, Тәуелділіктер Және Іске Қосу

### `go.mod` — тәуелділіктер мен нақты нұсқалар

Файл: `go.mod`

Жоба Go модулі ретінде анықталған:

```text
module idsai-core-up
go 1.25.0
```

Негізгі runtime және infrastructure тәуелділіктері:

| Тәуелділік | Нұсқа | Қолданылуы |
| --- | --- | --- |
| `github.com/gin-gonic/gin` | `v1.10.0` | HTTP router және middleware |
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` | JWT access token тексеруі |
| `github.com/google/uuid` | `v1.6.0` | UUID идентификаторлары |
| `github.com/jackc/pgx/v5` | `v5.8.0` | PostgreSQL driver/pool |
| `github.com/pressly/goose/v3` | `v3.27.0` | SQL migration орындау |
| `github.com/redis/go-redis/v9` | `v9.18.0` | Redis cache |
| `github.com/minio/minio-go/v7` | `v7.0.99` | Object storage |
| `github.com/stretchr/testify` | `v1.11.1` | Unit/integration test assertions |
| `golang.org/x/crypto` | `v0.48.0` | Қауіпсіздік және crypto helper-лері |

### `docker-compose.yml` — сервистер құрылымы

Файл: `docker-compose.yml`

Жергілікті орта үш негізгі сервиспен көтеріледі:

| Сервис | Image | Порттар | Мақсаты |
| --- | --- | --- | --- |
| `postgres` | `postgres:16` | `5433:5432` | Негізгі дерекқор, DB аты `idsai` |
| `minio` | `minio/minio:latest` | `9000:9000`, `9001:9001` | Файл және media storage |
| `redis` | `redis:7-alpine` | `6379:6379` | RBAC permission cache және басқа cache сценарийлері |

Persistent volume-дар: `pgdata`, `minio-data`, `redis-data`.

### `Makefile` — командалар тізімі

Файл: `Makefile`

Негізгі командалар:

| Команда | Мақсаты |
| --- | --- |
| `make run` | `.env` жүктеп, `cmd/api/main.go` арқылы API іске қосу |
| `make up` | `postgres` және `minio` контейнерлерін көтеру |
| `make migrate` | `migrations/` ішіндегі goose migration-дарды орындау |
| `make migrate-status` | Migration status көру |
| `make test` | Барлық unit/transport test-терді орындау |
| `make test-integration` | `integration` tag-пен PostgreSQL integration test-терді орындау |
| `make bench` | RBAC service және HTTP middleware benchmark-тарын орындау |
| `make coverage` | Coverage profile және function coverage шығару |
| `make coverage-html` | HTML coverage report жасау |
| `make report-artifacts` | Unit, integration, benchmark, coverage нәтижелерін `.tmp/report/` ішіне сақтау |

Default database URL:

```text
postgres://postgres:postgres@localhost:5433/idsai?sslmode=disable
```

### `cmd/api/main.go` — қолданба кіру нүктесі

Файл: `cmd/api/main.go`

Бұл файл API қосымшасының entrypoint-і. Негізгі міндеттері:

- `config.Load()` арқылы конфигурацияны оқу;
- Telegram alert notifier құру;
- `app.New(rootCtx, cfg)` арқылы application composition root-ты іске қосу;
- `http.Server` құру және Gin router-ді handler ретінде беру;
- health monitor-ды бөлек goroutine ретінде қосу;
- `SIGINT/SIGTERM` кезінде graceful shutdown орындау;
- panic және critical error жағдайларын notifier арқылы жіберу.

Server timeout параметрлері:

| Параметр | Мәні |
| --- | --- |
| `ReadHeaderTimeout` | `5s` |
| `ReadTimeout` | `30s` |
| `WriteTimeout` | `60s` |
| `IdleTimeout` | `120s` |

### `internal/app/` және `wire_modules.go` — dependency injection, composition root

Файлдар:

- `internal/app/app.go`
- `internal/app/wire_modules.go`

`internal/app/app.go` application composition root ретінде жұмыс істейді:

- config validation;
- PostgreSQL pool құру;
- module wiring;
- email outbox dispatcher іске қосу;
- public contact handler құру;
- HTTP router жинау;
- local media static route қосу;
- Redis және DB resource-тарын жабу.

`internal/app/wire_modules.go` модульдерді бір жерге жинайды:

| Модуль | Құралатын компоненттер |
| --- | --- |
| `auth` | auth repo/service/handler |
| `admin` | admin service/handler |
| `rbac` | RBAC repo/service/authorizer |
| `projects` | projects service/handler dependencies |
| `projectflow` | project lifecycle handler/service/repo |
| `notifications` | notifications service/handler/repo |
| `kb` | knowledge base handler |

RBAC module Redis қолжетімді болса `CachedAuthorizer` қолданады, ал Redis болмаса base `rbac.Service` арқылы graceful fallback жасайды.

## 3.2. Authentication, RBAC Және Access Control

### `internal/http/middleware/auth.go` — AuthRequired middleware

Файл: `internal/http/middleware/auth.go`

`AuthRequired(jwtSecret, stateReader)` middleware-і:

- access token-ді алдымен cookie-ден (`AccessCookieName`), кейін `Authorization: Bearer ...` header-ден оқиды;
- JWT issuer және HS256 signing method тексереді;
- claims ішінен `userID`, `tenantID`, `facultyID`, `departmentID`, `isAdmin`, `isProfessor` шығарады;
- `stateReader` берілсе, user status, email verification және password change уақытын тексереді;
- Gin context және request context ішіне identity жазады.

Қосымша helper-лер:

- `UserIDFromCtx`
- `TenantIDFromCtx`
- `FacultyIDFromCtx`
- `DepartmentIDFromCtx`
- `IsAdminFromCtx`
- `IsProfessorFromCtx`
- `AdminRequired`

### `internal/http/middleware/rbac.go` — RequirePermission middleware

Файл: `internal/http/middleware/rbac.go`

Негізгі middleware функциялары:

| Функция | Мақсаты |
| --- | --- |
| `RequirePermission` | Бір permission бойынша RBAC тексеру |
| `RequirePermissionIf` | Feature flag/condition арқылы RBAC тексеруді қосу/өшіру |
| `RequireAllPermissions` | Бірнеше permission-ның бәрі бар екенін тексеру |
| `RequirePermissionWithAttrs` | RBAC + runtime attribute conditions, яғни ABAC тексеру |

Қате сценарийлері:

- user context жоқ болса: `401 Unauthorized`;
- scope resolve болмай қалса: `400 Bad Request`;
- permission жоқ немесе authorizer error болса: `403 Forbidden`.

Middleware RBAC denial оқиғаларын structured log форматында шығарады: method, path, user_id, tenant_id, permission, scope және reason.

### Scope resolver функциялары

Файл: `internal/http/middleware/rbac.go`

Қолданылатын resolver-лер:

| Resolver | Scope |
| --- | --- |
| `SystemScope()` | `SYSTEM`, `scope_id = nil` |
| `TenantScopeFromCtx()` | context ішіндегі tenant ID |
| `FacultyScopeFromCtx()` | context ішіндегі faculty ID |
| `DepartmentScopeFromCtx()` | context ішіндегі department ID |
| `FacultyScopeFromHeader(header)` | header ішіндегі faculty UUID |
| `ProjectScopeFromParam(param)` | route param ішіндегі project UUID |

### `internal/services/rbac/` — Authorizer интерфейсі және іске асырылуы

Папка: `internal/services/rbac/`

Негізгі файлдар:

| Файл | Рөлі |
| --- | --- |
| `authorizer.go` | `Authorizer` интерфейсі |
| `service.go` | Core RBAC service |
| `cached_authorizer.go` | Redis cache decorator |
| `scope.go` | Scope type және validation |
| `condition.go` | ABAC condition registry |
| `repository.go` | RBAC repository port |

`Authorizer` интерфейсі:

```go
Can(ctx, userID, permissionCode, scope) (bool, error)
CanAll(ctx, userID, permissions, scope) (bool, error)
CanWithAttributes(ctx, userID, permissionCode, scope, attrs) (bool, error)
ListPermissionCodes(ctx, userID, scope) ([]string, error)
```

`Service` implementation:

- scope validation жасайды;
- permission тексеруді repository-ге делегирлейді;
- `CanAll` кезінде барлық permission-ды тексеріп, бірінші denied permission-да тоқтайды;
- `CanWithAttributes` кезінде алдымен RBAC, кейін ABAC condition тексереді;
- deterministic test үшін `SetNow` арқылы clock override қолдайды.

`CachedAuthorizer`:

- Redis key форматы: `rbac:user:{user_id}:perm:{permission}:scope:{scope_type}:{scope_id}`;
- `Can` және `CanAll` нәтижелерін cache-ке жазады;
- Redis error кезінде inner authorizer-ге fallback жасайды;
- `CanWithAttributes` cache қолданбайды, себебі attribute context request сайын өзгереді;
- role mutation кейін `InvalidateUser` арқылы cache тазаланады.

### `internal/repos/postgres/rbac_repo.go` — RBAC repository сұраныстары

Нақты файл: `internal/repos/postgres/rbac_repo.go`

Ескерту: тапсырмада `internal/repos/postgres/rbac.go` деп көрсетілген, репода нақты атауы `rbac_repo.go`.

Негізгі функциялар:

| Функция | Мақсаты |
| --- | --- |
| `HasPermission` | User permission бар-жоғын SQL арқылы тексеру |
| `ListPermissionCodes` | Scope бойынша user permission кодтарын шығару |
| `GrantRoleByCode` | Role assignment беру немесе upsert жасау |
| `SetCacheInvalidator` | Cache invalidation callback қосу |

Сердце логики — `resolved_scopes` CTE. Ол сұралған scope үшін parent scope-тарды есептейді:

- `SYSTEM` үшін: `SYSTEM`;
- `TENANT` үшін: `SYSTEM`, `TENANT`;
- `FACULTY` үшін: `SYSTEM`, `TENANT`, `FACULTY`;
- `DEPARTMENT` үшін: `SYSTEM`, `TENANT`, `FACULTY`, `DEPARTMENT`;
- `PROJECT` үшін: `SYSTEM`, `TENANT`, `FACULTY`, `PROJECT`.

Осылайша жүйе hierarchical RBAC моделін іске асырады: permission нақты project scope-та болмаса да, parent scope-тағы role арқылы рұқсат берілуі мүмкін.

### `migrations/00002_rbac.sql` — базалық схема

Файл: `migrations/00002_rbac.sql`

Базалық RBAC кестелері:

| Кесте | Мақсаты |
| --- | --- |
| `roles` | Role catalog |
| `permissions` | Permission catalog |
| `role_permissions` | Role -> permission байланысы |
| `role_assignments` | User -> role -> scope assignment |

Бастапқы role-дар:

- `SUPER_ADMIN`
- `STUDENT`
- `PROFESSOR`
- `MODERATOR`
- `TEAM_LEAD`
- `MEMBER`
- `PROJECT_PROFESSOR`

Бастапқы permission топтары:

- project lifecycle: `project.create`, `project.edit`, `project.approve`, т.б.;
- team/recruitment: `team.apply`, `team.accept_applicant`, т.б.;
- tasks: `task.view`, `task.create`, `task.assign`, т.б.;
- docs/lab: `doc.view`, `doc.upload`, `lab.link_repo`;
- grading: `grading.view`, `grading.mark_criteria`, `grading.publish`;
- admin/moderation: `admin.manage_rbac`, `audit.view_system`, т.б.

### `migrations/00038_rbac_delegated_project_roles.sql` — делегирленген рөлдер

Файл: `migrations/00038_rbac_delegated_project_roles.sql`

Бұл migration project ішіндегі delegated access моделін қосады:

- жаңа permission: `member.access.manage`;
- жаңа project role-дар: `CO_LEAD`, `RECRUITER`, `TASK_MANAGER`;
- `TEAM_LEAD` role-ына `member.access.manage` беріледі.

Delegated role permission мысалдары:

| Role | Permission мысалдары |
| --- | --- |
| `CO_LEAD` | `project.view`, `project.edit`, `project.invite_professor`, `member.approve`, `task.create`, `task.assign` |
| `RECRUITER` | `project.view`, `member.approve` |
| `TASK_MANAGER` | `project.view`, `task.create`, `task.assign` |

### `migrations/00043_project_custom_access_roles.sql` — custom рөлдер

Файл: `migrations/00043_project_custom_access_roles.sql`

Бұл migration project-specific custom access role моделін қосады.

Кесте: `project_access_roles`

Негізгі өрістер:

- `tenant_id`
- `project_id`
- `role_id`
- `code`
- `name`
- `description`
- `created_by`
- `created_at`

Constraint-тер:

- `UNIQUE (tenant_id, project_id, code)`;
- `UNIQUE (role_id)`.

Индекс:

```sql
idx_project_access_roles_project ON project_access_roles(tenant_id, project_id)
```

## 3.3. Тестілеу, Coverage Және Benchmark

### `docs/TESTING_REPORT.md` — coverage + benchmark нәтижелері

Файл: `docs/TESTING_REPORT.md`

Фиксация күні: `2026-04-10`.

Сводный нәтиже:

| Категория | Команда | Нәтиже |
| --- | --- | --- |
| Unit / transport | `make test` | `PASS` |
| Integration | `make test-integration` | `PASS` |
| Security / BOLA | `go test ./internal/http/handlers -run 'TestProjectFlowHandlerGetMemberAccess' -v` | `PASS` |
| Coverage | `make coverage` | `PASS` |
| Benchmarks | `make bench` | `PASS` |

Coverage summary:

| Область | Coverage |
| --- | --- |
| `internal/services/rbac` | `94.7%` |
| `internal/http/middleware` | `86.8%` |
| `internal/http/dto` | `99.6%` |
| `internal/services/admin` | `78.9%` |
| `internal/services/auth` | `58.8%` |
| `internal/services/kb` | `82.9%` |
| `internal/services/notifications` | `69.3%` |
| `internal/services/projectflow` | `64.5%` |
| `internal/app` | `36.5%` |
| `internal/domain` | `100.0%` |
| Бүкіл project | `34.7%` |

### RBAC test сценарийлері

Файлдар:

- `internal/services/rbac/service_test.go`
- `internal/services/rbac/service_additional_test.go`
- `internal/services/rbac/cached_authorizer_test.go`
- `internal/http/middleware/rbac_test.go`
- `internal/http/middleware/rbac_additional_test.go`
- `internal/repos/postgres/rbac_repo_integration_test.go`

Негізгі сценарийлер:

- invalid scope denial;
- `Can` repository delegation;
- `CanAll` allow/deny/error behavior;
- `CanWithAttributes` RBAC + ABAC behavior;
- condition registry register/evaluate/has;
- permission list delegation;
- cache key format;
- positive/negative cache;
- error кезінде cache жазылмауы;
- cache invalidation;
- middleware unauthorized/forbidden/invalid scope/allowed cases;
- integration: project scope permission, expired assignment denial, role grant upsert, faculty scope inheritance.

### `csrf_test.go` — CSRF тесттері

Файл: `internal/http/middleware/csrf_test.go`

Тексерілетін сценарийлер:

- cross-site cookie mutation request reject болады;
- same-origin cookie mutation request allow болады;
- cookie жоқ Bearer-style mutation request allow болады.

### `auth_cookie_test.go` — cookie/token тесттері

Файл: `internal/http/middleware/auth_cookie_test.go`

Тексерілетін сценарийлер:

- valid cookie invalid header-ден басым болады;
- password өзгергеннен бұрын берілген token reject болады.

Қосымша auth тесттер `internal/http/middleware/auth_additional_test.go` ішінде:

- missing token;
- invalid bearer token;
- invalid UUID claims;
- inactive/unverified/stale auth state;
- context және request identity дұрыс жазылуы;
- `AdminRequired` reject/allow behavior.

### Benchmark шығысы — `make bench`

Artifact файлдары:

- `.tmp/report/bench-rbac.txt`
- `.tmp/report/bench-middleware.txt`

RBAC service benchmark:

| Benchmark | Нәтиже | Memory | Allocs |
| --- | --- | --- | --- |
| `BenchmarkServiceCan_ProjectScope` | `61.45 ns/op` | `0 B/op` | `0 allocs/op` |
| `BenchmarkServiceCanAll_FivePermissions` | `298.1 ns/op` | `0 B/op` | `0 allocs/op` |
| `BenchmarkServiceCanWithAttributes_ThreeConditions` | `112.5 ns/op` | `0 B/op` | `0 allocs/op` |

HTTP middleware benchmark:

| Benchmark | Нәтиже | Memory | Allocs |
| --- | --- | --- | --- |
| `BenchmarkRequirePermission_Allowed` | `3735 ns/op` | `5666 B/op` | `16 allocs/op` |
| `BenchmarkRequireAllPermissions_Allowed` | `3782 ns/op` | `5666 B/op` | `16 allocs/op` |
| `BenchmarkRequirePermissionWithAttrs_Allowed` | `3936 ns/op` | `6003 B/op` | `18 allocs/op` |

Benchmark ортасы:

```text
goos: linux
goarch: amd64
cpu: 12th Gen Intel(R) Core(TM) i5-12450H
```

## Жалпы Архитектура

### `docs/architecture/ARCHITECTURE.md` — архитектуралық шешімдер

Нақты файл: `docs/architecture/ARCHITECTURE.md`

Ескерту: тапсырмада `docs/ARCHITECTURE.md` деп көрсетілген, репода нақты файл `docs/architecture/ARCHITECTURE.md`.

Қабылданған архитектуралық стиль:

```text
Modular Monolith + Clean (Ports & Adapters)
```

Негізгі себептері:

- бір deploy және бір database арқылы эксплуатацияны қарапайым сақтау;
- `auth`, `admin`, `projects`, `projectflow`, `notifications`, `rbac` сияқты bounded context-терді модуль ретінде бөлу;
- кейінгі migration-ды big-bang rewrite жасамай, біртіндеп орындау.

Dependency rules:

- `domain` HTTP/DB/infra туралы білмейді;
- SQL тек adapter қабатында (`repos/postgres`) орналасады;
- `cmd/api` және `internal/app` composition root болып қалады;
- `internal/modules/*` модульдік composition қабаты ретінде қолданылады;
- `internal/http/router.go` орталық router болып, route registration бөлек `router_*` файлдарға декомпозицияланған.

### `docs/PROJECT_LIFECYCLE.md` — жоба lifecycle логикасы

Файл: `docs/PROJECT_LIFECYCLE.md`

Негізгі project status-тар:

- `DRAFT`
- `REVIEW`
- `RECRUITMENT`
- `ACTIVE`
- `GRADING`
- `ARCHIVE`

Lifecycle қысқаша:

```text
DRAFT -> RECRUITMENT -> ACTIVE -> GRADING -> ARCHIVE/COMPLETED
```

Негізгі blocker-лер:

- project activation үшін positions, active members, accepted professor және criteria болуы керек;
- grading submit үшін project `ACTIVE`, professor accepted, task бар және барлық task `DONE` болуы керек;
- grading publish үшін project `GRADING`, барлық criteria бағаланған және әрекетті assigned professor орындауы керек.

Admin override:

```text
PATCH /v2/admin/projects/:id/status
```

Бұл endpoint lifecycle-ды bypass етіп, project status-ты қолмен өзгертуге мүмкіндік береді.

### `internal/http/router.go` — маршруттар картасы

Файл: `internal/http/router.go`

Router global middleware-лері:

- `RequestLogger`;
- `gin.Recovery`;
- `CSRFProtection`;
- `/dev` route-тары үшін no-cache headers.

Route groups:

| Group | Файл | Мақсаты |
| --- | --- | --- |
| `/v2/auth` | `router_auth.go` | register/login/logout/refresh/profile/settings |
| `/v2/admin` | `router_admin.go` | user және project administration |
| `/v2/projects` | `router_projects.go`, `router_projectflow.go` | project CRUD, lifecycle, members, tasks, grading, access |
| `/v2/notifications` | `router_notifications.go` | notifications list/read/delete |
| `/v2/kb` | `router_kb.go` | knowledge base |
| `/v2/contact` | `router_public.go` | public contact form |
| `/health`, `/dev/*`, `/swagger/*` | `router_dev.go` | healthcheck, dev frontend, docs |

Project flow route-тары RBAC middleware арқылы permission-based қорғалады. Мысалы:

| Endpoint | Permission |
| --- | --- |
| `PATCH /v2/projects/:project_id` | `project.edit` |
| `POST /v2/projects/:project_id/positions` | `position.create` |
| `POST /v2/projects/:project_id/members/:user_id/approve` | `member.approve` |
| `POST /v2/projects/:project_id/tasks` | `task.create` |
| `POST /v2/projects/:project_id/grading/publish` | `grading.publish` |
| `GET /v2/projects/:project_id/members/:user_id/access` | `member.access.manage` |
| `PUT /v2/projects/:project_id/members/:user_id/access` | `member.access.manage` |

## Қорытынды

3-тараудың практикалық дәлелдемелері төрт қабатқа бөлінеді:

1. Configuration және startup: `go.mod`, `docker-compose.yml`, `Makefile`, `cmd/api/main.go`, `internal/app/*`.
2. Security және authorization: `auth.go`, `rbac.go`, `internal/services/rbac/*`, `internal/repos/postgres/rbac_repo.go`, RBAC migration-дары.
3. Quality assurance: `docs/TESTING_REPORT.md`, RBAC/auth/CSRF test-тері, coverage және benchmark artifact-тері.
4. Architecture және lifecycle: `docs/architecture/ARCHITECTURE.md`, `docs/PROJECT_LIFECYCLE.md`, `internal/http/router.go`.

Бұл файлдар бірге жүйенің dependency management, local infrastructure, composition root, authentication, hierarchical RBAC, delegated roles, custom project roles, testing және performance baseline бөліктерін дәлелдейді.
