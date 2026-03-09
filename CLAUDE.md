# Price Tracker

Product price tracking API. Go + Gin + PostgreSQL.

## Commands

- **Run locally**: `docker compose up` (API on port 8080, PostgreSQL on 5432)
- **Build**: `go build -buildvcs=false -o ./tmp/main ./cmd/api`
- **Lint**: `golangci-lint run ./...`
- **Migrations**: run automatically by the `migrate` container in docker-compose

## Structure

```
cmd/api/          → API entrypoint
internal/
  handler/        → HTTP handlers (Gin)
  models/         → model structs
  repository/     → data access layer (pgx)
migrations/       → SQL migrations (up/down)
```

## Conventions

- Repository pattern: each model has its own repository with `NewXRepository(db *pgxpool.Pool)`
- JSON tags in snake_case, structs in PascalCase
- Optional fields use pointers (`*float64`, `*int`)
- Handlers return appropriate HTTP status codes (200, 201, 204, 400, 404, 500)
- Errors are always handled explicitly, no panics
- Default currency: BRL
- Timezone: America/Sao_Paulo

## Database

PostgreSQL 16. Tables: products, prices, notifications.
Use GENERATED ALWAYS AS IDENTITY for PKs.
Foreign keys with CASCADE DELETE.

## Rules

- Always run `golangci-lint run ./...` before committing
- Never commit .env — use .env.example as reference
- Commit messages and PRs in English, using Conventional Commits format (feat, fix, docs, etc.)
- Write unit tests for all handlers and services
- Handle all errors explicitly, no panics allowed
- Follow Go idioms and best practices for code style and structure
