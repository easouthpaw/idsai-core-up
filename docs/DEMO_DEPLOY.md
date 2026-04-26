# Demo Deploy

Для показа комиссии рекомендованный путь такой:

1. Задеплоить приложение на Render через [`render.yaml`](/home/aibolat/diploma/idsai-core-up/render.yaml).
2. Использовать seeded demo-данные из миграций.
3. Держать Cloudflare Quick Tunnel как запасной вариант, если нужен показ прямо с ноутбука без деплоя.

## Вариант 1. Render

Почему это лучший путь для текущего проекта:

- приложение уже устроено как обычный long-running Go сервер;
- встроенный frontend уже входит в бинарник;
- Docker-образ сам прогоняет миграции перед стартом;
- комиссия получает стабильную публичную ссылку.

Что уже подготовлено в репозитории:

- [`Dockerfile`](/home/aibolat/diploma/idsai-core-up/Dockerfile)
- [`docker/entrypoint.sh`](/home/aibolat/diploma/idsai-core-up/docker/entrypoint.sh)
- [`render.yaml`](/home/aibolat/diploma/idsai-core-up/render.yaml)
- [`cmd/migrate/main.go`](/home/aibolat/diploma/idsai-core-up/cmd/migrate/main.go)

### Как запустить

1. Запушить актуальную ветку в GitHub.
2. В Render выбрать `New +` -> `Blueprint`.
3. Указать этот репозиторий.
4. Подтвердить создание `web service` и `postgres database`.
5. Дождаться первого деплоя.

После деплоя приложение будет доступно на домене вида `https://<service-name>.onrender.com`.

### Важные замечания

- В текущем `render.yaml` `EMAIL_ENABLED=true`, но для `free` web service на Render SMTP-порты `25`, `465` и `587` заблокированы, поэтому email-рассылку лучше не считать рабочей без paid plan, альтернативного порта вроде `2525` или API-based mail provider.
- Новые регистрации на Render автоматически подтверждаются через `AUTH_AUTO_VERIFY_REGISTRATIONS=true`, поэтому пользователь может зарегистрироваться и сразу войти без письма.
- В текущем `render.yaml` фоновые задачи и health monitor включены (`BACKGROUND_JOBS_ENABLED=true`, `HEALTH_MONITOR_ENABLED=true`), так что email outbox и фоновые проверки запускаются вместе с приложением.
- Redis отключен, поэтому кэш RBAC просто не используется.
- Если `PUBLIC_BASE_URL` не задан вручную, приложение на Render автоматически возьмёт `RENDER_EXTERNAL_URL`, поэтому публичные ссылки будут корректными без дополнительной настройки.
- Если `MINIO_*` не заданы, аватары и обложки проектов на демо складываются в локальную папку `LOCAL_STORAGE_DIR=/tmp/idsai-media`; при заполненных `MINIO_*` приложение использует MinIO, а локальная папка остаётся fallback-вариантом.
- Это файловое хранилище на Render эфемерное: после redeploy или рестарта загруженные картинки могут исчезнуть, но для защиты и демо этого обычно достаточно.
- Форма контактов на landing page может доставлять сообщения через Telegram или напрямую на email. Для email fallback можно задать `CONTACT_EMAIL_TO`, либо сервис возьмёт адрес из `SMTP_USER`.
- У free web service на Render есть cold start: если сервис простаивает, перед показом лучше открыть ссылку самому за 1-2 минуты, чтобы комиссия не ждала первый прогрев.

### Демо-логины

Основной администратор:

- `admin@idsai.local`

Seeded demo accounts:

- пароль для demo-пользователей: `DemoPass123!`
- `aibolat.student@idsai.local`
- `aliya.student@idsai.local`
- `daniyar.student@idsai.local`
- `dinara.student@idsai.local`
- `marat.student@idsai.local`
- `nurzhan.prof@idsai.local`
- `aidana.prof@idsai.local`
- `sabina.prof@idsai.local`
- aibolat.smartedu@idsai.local
- asylkhan.prof@idsai.local
## Вариант 2. Cloudflare Quick Tunnel

Подходит как аварийный запасной вариант, если нужно показать систему сегодня без деплоя.

1. Поднять локально Postgres и MinIO как обычно.
2. Запустить приложение локально на `:8080`.
3. Выполнить:

```bash
cloudflared tunnel --url http://localhost:8080
```

Cloudflare выдаст публичный URL вида `https://random-name.trycloudflare.com`, который можно отправить комиссии.

Ограничения:

- это вариант только для теста и разработки;
- Cloudflare прямо пишет, что Quick Tunnels не для production;
- не стоит использовать его как единственный план на важную защиту, если есть время на normal deploy.
