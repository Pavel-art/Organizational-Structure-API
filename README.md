# Organizational-Structure-API

API для управления оргструктурой: подразделения (дерево) и сотрудники.

## Стек

- Go + `net/http`
- PostgreSQL
- GORM
- goose migrations
- Docker / docker-compose
- Zerolog

## Запуск

1) Создай/проверь `.env` (в репозитории уже есть пример со значениями по умолчанию).
   - Для локального запуска (без Docker для API) используй `.env.local` (см. `.env.local.example`) — он подгружается поверх `.env`.

2) Выбери режим:

### Вариант A: всё в Docker (рекомендуется)

```bash
docker compose up --build
```

API поднимется на `http://localhost:${HTTP_PORT}` (по умолчанию `http://localhost:8082`).

PostgreSQL будет доступен на хосте по порту `5435` (контейнерный порт `5432`).

### Вариант B: Postgres в Docker, API локально (удобно для дебага)

1) Подними Postgres:

```bash
docker compose up -d postgres
```

2) Останови контейнер API (чтобы не было конфликта порта):

```bash
docker compose stop api
```

3) Создай `.env.local`:

```bash
cp .env.local.example .env.local
```

4) Запусти API локально:

```bash
go run ./cmd
```

## Swagger

- UI: `http://localhost:${HTTP_PORT}/swagger/` (по умолчанию `http://localhost:8082/swagger/`)
- OpenAPI: `http://localhost:${HTTP_PORT}/openapi.json` (по умолчанию `http://localhost:8082/openapi.json`)

## Healthcheck

`GET /healthz`

## Основные эндпоинты

- `POST /departments/`
- `POST /departments/{id}/employees/`
- `GET /departments/{id}?depth=1..5&include_employees=true|false`
- `PATCH /departments/{id}`
- `DELETE /departments/{id}?mode=cascade|reassign&reassign_to_department_id=...`
