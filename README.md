# IDSAI Core

IDSAI Core - backend и встроенный web-интерфейс для управления учебными проектами. Приложение объединяет REST API `/v2`, страницы `/dev/*`, RBAC, проектный lifecycle, кабинет преподавателя, уведомления и knowledge base.

## Актуальное состояние

- Миграция `00044_remove_legacy_project_files_and_moderator.sql` удаляет legacy-роль `MODERATOR` и неиспользуемую таблицу `project_files`.
- Административный контур поддерживает смену базовой faculty-роли только между `STUDENT` и `PROFESSOR`.
- Удалены мёртвые заглушки: legacy route `/dev/tester` и неиспользуемый `TaskDemoHandler`.
- Встроенный frontend по умолчанию открывается на казахском языке; пользователь может переключиться на русский или английский, и выбор сохраняется в `localStorage`.

## Что есть в проекте

- JWT-аутентификация, профиль пользователя и настройки аккаунта.
- RBAC с доступом на уровнях tenant, faculty, department и project.
- Полный поток работы с проектом: создание, набор, участники, роли, задачи, readiness, grading и archive.
- Отдельные сценарии для администратора и преподавателя.
- In-app уведомления, email outbox и Telegram alerts.
- Встроенный frontend без отдельного SPA-репозитория.

## Стек

- Go `1.25`
- Gin
- PostgreSQL
- Redis для кэша RBAC с graceful degradation
- MinIO / S3-compatible storage для медиа
- Docker Compose

## Быстрый старт

Проект автоматически читает `.env`, если файл лежит в корне репозитория.

1. Поднимите локальную инфраструктуру:

```bash
docker compose up -d postgres minio
```

При желании можно сразу поднять и Redis:

```bash
docker compose up -d redis
```

2. Создайте `.env` с минимальными переменными:

```env
ADDR=:8080
DATABASE_URL=postgres://postgres:postgres@localhost:5433/idsai?sslmode=disable
JWT_SECRET=replace-this-with-at-least-32-characters
PUBLIC_BASE_URL=http://localhost:8080
```

3. Примените миграции:

```bash
make migrate
```

4. Запустите приложение:

```bash
make run
```

После запуска будут доступны:

- `http://localhost:8080/`
- `http://localhost:8080/dev/login`
- `http://localhost:8080/health`

## Полезные команды

```bash
make up
make migrate
make migrate-status
make run
make test
make test-integration
make bench
make coverage
make report-artifacts
go test ./...
```

`make up` поднимает `postgres` и `minio`. Если нужен Redis, запускайте его отдельно через `docker compose`.

## Переменные окружения

Обязательные:

- `JWT_SECRET` - минимум 32 символа.
- `PUBLIC_BASE_URL` - базовый публичный URL приложения.

Основные:

- `ADDR` - адрес HTTP-сервера, по умолчанию `:8080`.
- `DATABASE_URL` - строка подключения к PostgreSQL.

Опциональные:

- `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM`, `SMTP_USE_SSL`, `EMAIL_ENABLED`
- `TELEGRAM_BOT_TOKEN`, `TELEGRAM_SUPERADMIN_CHAT_ID`
- `MINIO_ENDPOINT`, `MINIO_ACCESS_KEY`, `MINIO_SECRET_KEY`, `MINIO_BUCKET`, `MINIO_USE_SSL`, `MINIO_PUBLIC_BASE_URL`
- `PHOTON_BASE_URL`, `PHOTON_LANG`, `PHOTON_COUNTRY_CODE`, `PHOTON_DEFAULT_LON`, `PHOTON_DEFAULT_LAT`, `PHOTON_REQUEST_TIMEOUT_SECONDS` (`Photon` работает без API key)
- `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB`

Если optional-переменные не заданы, приложение запускается без соответствующих интеграций.

## Структура репозитория

- `cmd/api` - точка входа сервера.
- `internal/http` - router, handlers и встроенный frontend.
- `internal/services` - бизнес-логика и use cases.
- `internal/repos/postgres` - PostgreSQL adapters.
- `migrations` - SQL-миграции.
- `docs` - архитектурная и предметная документация.

## Документация

- `docs/PROJECT_LIFECYCLE.md` - текущая логика жизненного цикла проекта.
- `docs/USER_GUIDE.md` - обзорный вход в пользовательскую документацию.
- `docs/STUDENT_GUIDE.md` - подробная инструкция для студентов.
- `docs/PROFESSOR_GUIDE.md` - подробная инструкция для преподавателей.
- `docs/ADMIN_GUIDE.md` - подробная инструкция для администратора.
- `docs/architecture/ARCHITECTURE.md` - принятая архитектура и правила зависимостей.
- `docs/code-review/README.md` - индекс code review заметок по текущему состоянию репозитория.
- `docs/TESTING_REPORT.md` - результаты тестирования, coverage и benchmark-метрики.
