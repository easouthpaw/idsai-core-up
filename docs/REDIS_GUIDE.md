# Redis — RBAC Permission Cache

## Что это и зачем

Redis используется как кэш для проверок прав доступа (RBAC). Каждый раз, когда система проверяет, может ли пользователь выполнить действие (создать задачу, открыть проект, изменить роль и т.д.), результат кэшируется в Redis.

**Без Redis:** каждая проверка — запрос в PostgreSQL.  
**С Redis:** повторные проверки отдаются за микросекунды из памяти.

Redis **не является обязательным**. При недоступности приложение продолжает работу, просто каждый запрос идёт в БД (graceful degradation).

## Архитектура

```
Handler → requireProjectPermission() → CachedAuthorizer
                                              ↓
                                    Redis.Get(key)
                                    ├── hit  → вернуть результат
                                    └── miss → Authorizer.Can() → PostgreSQL
                                                     ↓
                                              Redis.Set(key, result, TTL)
```

Ключ кэша:
```
rbac:user:{user_id}:perm:{permission}:scope:{scope_type}:{scope_id}
```

Пример:
```
rbac:user:550e8400-...:perm:task.create:scope:project:7f3a1b2c-...
```

Инвалидация — при изменении ролей пользователя удаляются все ключи по паттерну `rbac:user:{user_id}:*` (через `SCAN`, не `KEYS` — безопасно для prod).

## Запуск через Docker Compose

```bash
docker compose up -d redis
```

Redis слушает на `localhost:6380` (хостовый порт смаплен с контейнерного 6379).

```bash
# Проверить подключение
redis-cli -p 6380 ping
# → PONG
```

## Переменные окружения

| Переменная | Обязательна | По умолчанию | Описание |
|---|---|---|---|
| `REDIS_ADDR` | нет | — | Адрес Redis, например `localhost:6380`. Если пусто — кэш отключён |
| `REDIS_PASSWORD` | нет | — | Пароль (если Redis запущен с `requirepass`) |
| `REDIS_DB` | нет | `0` | Номер базы данных Redis (0–15) |

### Пример `.env` для локальной разработки

```env
REDIS_ADDR=localhost:6380
REDIS_PASSWORD=
REDIS_DB=0
```

### Пример `.env` для продакшна

```env
REDIS_ADDR=redis:6379
REDIS_PASSWORD=<strong-password>
REDIS_DB=0
```

> В docker-compose для продакшна контейнер `redis` доступен по имени `redis:6379` изнутри сети.

## Поведение при недоступности

Если Redis недоступен при старте или во время работы:

- При старте в лог пишется предупреждение, приложение **не падает**:
  ```
  [WARN] redis: ping failed (addr=localhost:6380): ... — RBAC cache disabled, falling through to DB
  ```
- Все проверки прав идут напрямую в PostgreSQL.
- При восстановлении Redis кэш начинает заполняться снова автоматически.

## TTL кэша

TTL задаётся в коде при инициализации `CachedAuthorizer` (в `wire_modules.go`). По умолчанию — разумное значение (обычно 5–15 минут). Изменить можно там же без переменных окружения.

## Что НЕ кэшируется

- `CanWithAttributes` — атрибутные проверки (зависят от runtime-контекста запроса)
- `ListPermissionCodes` — список всех прав пользователя

Эти методы всегда идут в БД.

## Проверка содержимого кэша

```bash
# Подключиться к Redis
redis-cli -p 6380

# Посмотреть все RBAC ключи
SCAN 0 MATCH "rbac:*" COUNT 100

# Посмотреть значение конкретного ключа
GET "rbac:user:550e8400-...:perm:task.create:scope:project:7f3a1b2c-..."
# → "1" (разрешено) или "0" (запрещено)

# Сбросить кэш конкретного пользователя вручную
redis-cli -p 6380 --scan --pattern "rbac:user:<user_id>:*" | xargs redis-cli -p 6380 DEL
```
