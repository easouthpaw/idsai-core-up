# IDSAI Core API

Бэкенд-сервис на Go для платформы IDSAI (Gin + PostgreSQL + Goose migrations).

## Требования

- Go `1.24+`
- Docker + Docker Compose
- GNU Make

## Быстрый старт

1. Поднять PostgreSQL:
   ```bash
   make up
   ```
2. Запустить API:
   ```bash
   make run
   ```

После запуска:

- API: `http://localhost:8080`
- Swagger: `http://localhost:8080/swagger/index.html`

## Основные команды

```bash
make up                # поднять postgres (docker compose)
make run               # запустить API
```

## Переменные окружения

Приложение читает `.env` (если есть) и переменные окружения:

- `ADDR` (по умолчанию `:8080`)
- `DATABASE_URL`
- `JWT_SECRET` (по умолчанию `dev-jwt-secret`)

Для `Makefile` используется переменная `DB_URL` (по умолчанию `postgres://postgres:postgres@localhost:5433/idsai?sslmode=disable`).

Пример:

```bash
make run DB_URL=postgres://postgres:postgres@localhost:5432/idsai?sslmode=disable
```
