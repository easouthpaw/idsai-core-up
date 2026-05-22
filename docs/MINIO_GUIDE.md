# MinIO — Object Storage

## Что это и зачем

MinIO используется как S3-совместимое объектное хранилище для медиафайлов:

- Обложки проектов (`/projects/{id}/image`)
- Аватары пользователей (`/settings/avatar`)
- Вложения к статьям базы знаний
- Вложения к задачам (task submissions)

## Архитектура

```
Handler → services/projects (или auth, kb) → infra/storage.ObjectStorage
                                                     ↓
                                          minioStorage  (если MINIO_ENDPOINT задан)
                                          localStorage  (если LOCAL_STORAGE_DIR задан)
                                          nopStorage    (если ничего не задано)
```

`ObjectStorage` — интерфейс. Реализаций три:

| Реализация | Когда включается | Поведение при недоступности |
|---|---|---|
| `minioStorage` | `MINIO_ENDPOINT` + ключи заданы | возвращает ошибку |
| `localStorage` | `LOCAL_STORAGE_DIR` задан | пишет в папку на диске |
| `nopStorage` | ничего не задано | всегда возвращает ошибку |

## Запуск через Docker Compose

```bash
docker compose up -d minio
```

После запуска доступно:
- **API** — `http://localhost:9000`
- **Web-консоль** — `http://localhost:9001`
- Логин/пароль по умолчанию: `minioadmin` / `minioadmin`

Бакет создаётся автоматически при первом обращении (через `sync.Once`). Политика бакета — public-read (файлы доступны по прямой ссылке без авторизации).

## Переменные окружения

| Переменная | Обязательна | По умолчанию | Описание |
|---|---|---|---|
| `MINIO_ENDPOINT` | да (для MinIO) | — | Адрес MinIO, например `localhost:9000` |
| `MINIO_ACCESS_KEY` | да (для MinIO) | — | Access key (root user) |
| `MINIO_SECRET_KEY` | да (для MinIO) | — | Secret key (root password) |
| `MINIO_BUCKET` | нет | `idsai-media` | Имя бакета |
| `MINIO_USE_SSL` | нет | `false` | `true` если MinIO за HTTPS |
| `MINIO_PUBLIC_BASE_URL` | нет | — | Публичный URL для отдачи файлов (CDN, nginx-прокси). Если не задан — строится из endpoint |
| `LOCAL_STORAGE_DIR` | нет | — | Папка для fallback на диск (если MinIO не настроен) |

### Пример `.env` для локальной разработки

```env
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=idsai-media
MINIO_USE_SSL=false
MINIO_PUBLIC_BASE_URL=http://localhost:9000
```

### Пример `.env` для продакшна (MinIO за nginx)

```env
MINIO_ENDPOINT=minio:9000
MINIO_ACCESS_KEY=<strong-key>
MINIO_SECRET_KEY=<strong-secret>
MINIO_BUCKET=idsai-media
MINIO_USE_SSL=false
MINIO_PUBLIC_BASE_URL=https://media.example.com
```

## Ограничения на загрузку

Проверяются на уровне хендлеров до передачи в storage:

| Тип файла | Лимит | Проверка формата |
|---|---|---|
| Обложка проекта | `images.MaxProjectCoverBytes` | По сигнатуре байт (не по расширению) |
| Аватар пользователя | `images.MaxAvatarBytes` | По сигнатуре, минимум 400×400 px |

Неподходящий файл → HTTP 400/413, до MinIO не доходит.

## Fallback на локальное хранилище

Если `MINIO_ENDPOINT` не задан, но задан `LOCAL_STORAGE_DIR`:

```env
LOCAL_STORAGE_DIR=/var/lib/idsai/media
PUBLIC_BASE_URL=http://localhost:8080
```

Файлы будут сохраняться в `LOCAL_STORAGE_DIR`, а публичные URL строиться как `/media/{key}` относительно `PUBLIC_BASE_URL`. Запись атомарная — через временный файл с `os.Rename`.

## Проверка работоспособности

```bash
# Проверить что MinIO запущен
curl http://localhost:9000/minio/health/live

# Открыть консоль
open http://localhost:9001
```

При старте приложения в логах будет:
```
[INFO] storage: using MinIO at localhost:9000, bucket=idsai-media
```
или
```
[INFO] storage: using local storage at /var/lib/idsai/media
```
