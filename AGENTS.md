# AGENTS.md

## Project
Pulse: Go 1.24 service for live polls & reactions. Serve REST CRUD, SSE streaming results, and signed webhooks. Generate & host OpenAPI/Swagger.

## Design (follow strictly)
- Architecture: Hexagonal (ports/adapters). Packages:
  - `domain/` (entities, pure logic), `app/` (use cases/services), `adapters/http` (Gin handlers, request/response DTOs), `adapters/persistence` (GORM), `internal/webhook` (dispatcher), `cmd/pulse` (main).
- Go practices: context-aware funcs, no globals, dependency injection via constructors, errors wrapped with `%w`, table-driven tests, race-safe concurrency.
- Persistence: GORM + SQLite file `./pulse.db`. Migrations via AutoMigrate in `data/db.go`.
- API: REST resources (Poll, Option, Vote). SSE at `GET /polls/:id/results/stream`. Webhooks (`vote.created`, `poll.threshold_reached`, `poll.closed`) with `Pulse-Signature` HMAC-SHA256 over raw JSON body.
- OpenAPI: swaggo annotations on handlers; serve Swagger UI at `/swagger/index.html`.

## Commands (always run)
- Setup: `go mod tidy`
- Lint: `golangci-lint run`
- Tests: `go test ./... -race -count=1`
- Vet: `go vet ./...`
- Swagger: `swag init --parseDependency --parseInternal`
- Run: `go run ./cmd/pulse`

## Conventions
- Logging: structured (log/stdlib ok). Include request ID middleware.
- Validation: bind+validate request DTOs; never trust client totals—recount from DB.
- SSE: `text/event-stream`, no-cache, heartbeat keepalive, clean disconnect handling.
- Webhooks: retry w/ exponential backoff (max 5). Include timestamp + signature header `Pulse-Signature`. Never log secrets.
- Security: CORS allowlist via env; input validation; limit payload sizes.

## Env (configure, do not hardcode)
`PORT`, `DB_PATH`, `CORS_ORIGINS`, `WEBHOOK_MAX_RETRIES`.

## Checks (block merge if failing)
- Lint, vet, tests must pass.
- Swagger must regenerate with no diff except timestamp fields.

## Non-goals
- No WebSockets (SSE only). No global singletons. Keep handlers thin; business logic lives in `app/`.
