# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make run           # go run cmd/server/main.go
make build          # build to bin/server
make test           # go test -v ./...
make docker-up      # docker-compose up -d --build
make docker-down    # docker-compose down -v
```

Run a single test:
```bash
go test -v ./internal/service/... -run TestHashPasswordAndCheck
```

The server needs a `.env` file (see `.env.example`) or system env vars. `PORT` defaults to `8080`, `ENV` defaults to `development` (set to `production` or `release` to enable Gin release mode). MySQL connection vars support both `DB_*` and Railway-style `MYSQL*` names (see `internal/config/config.go`).

## Architecture

Standard Gin + GORM layered service backend, wired together entirely in `cmd/server/main.go` (no DI framework — repos → services → handlers are constructed and passed in by hand). When adding a new resource, follow this same chain and register routes in `main.go`.

- `internal/domain` — GORM models (`models.go`) and request/response DTOs (`dto.go`). This is the single source of truth for the schema; GORM auto-migrates from these structs.
- `internal/repository` — one file per aggregate (`user_repository.go`, `request_repository.go`, etc.) plus `db.go`, which owns `InitDB`.
- `internal/service` — business logic, one file per domain area. Services depend on repositories (and sometimes other services, e.g. `monitorSvc` depends on `emailSvc`).
- `internal/handler` — Gin handlers, thin wrappers over services using `pkg/response` for all JSON output.
- `internal/middleware` — `JWTAuth` (populates `userID`/`userRole`/`userBranchID`/`approverAccessType`/`firstLogin` into Gin context) and `RequireRoles(...)` for RBAC, plus request logging and panic recovery.
- `pkg/` — cross-cutting utilities: `jwt`, `hash` (bcrypt), `email`, `logger` (file logger writing to `logs/app.log`, also generates unique error codes), `response` (standard `APIResponse` envelope: `success`/`message`/`data`/`errorCode`/`errors`).

### Request lifecycle & workflow model

Routes are grouped under `/api/v1` in `main.go`: an unauthenticated `auth/login`, then a `protected` group behind `JWTAuth`, with an `admin` subgroup behind `RequireRoles("ADMIN")`. Non-admin write routes use per-route `RequireRoles(...)` (e.g. `BRANCH_REQUESTER`, `APPROVER`, `AGENCY`, `MONITOR`).

The domain models a stationery request/fulfillment workflow through a fixed role sequence:
`Request` (branch requester creates, with `RequestItem`s) → `approve` (`APPROVER` creates `ApprovalItem`s) → `deliver` (`AGENCY` creates `Delivery`/`DeliveryItem`s) → `verify` (`BRANCH_REQUESTER` creates `VerificationItem`s). `SlaSettings` defines max days allowed per stage; `MonitorService`/`monitor` routes surface delayed orders. Each request also has a chat thread (`ChatMessage`, targeted by role) and every mutating action should be recorded via `AuditRepository`/`AuditLog`.

Roles are seeded fixed-ID rows (`ADMIN`=1, `BRANCH_REQUESTER`=2, `APPROVER`=3, `AGENCY`=4, `MONITOR`=5) — see `RequireRoles` calls in `main.go` for which roles can hit which routes, and `repository/db.go`'s `SeedInitialData` for the seed data (including the default admin login `admin@stationery.com` / `Admin@123`).

### Database bootstrapping

`InitDB` in `internal/repository/db.go` does more than open a connection: it creates the database if missing, drops a couple of legacy FK constraints if present, runs `AutoMigrate` over all domain models, then does manual `HasColumn`/`AddColumn` checks for columns added after the initial migration, and finally calls `SeedInitialData`. There's also a standalone `migrations/000001_init_schema.up/down.sql` pair and `seed/seed.sql` — the Go-side `InitDB` path is what actually runs on `make run`/server start, the SQL files are a secondary reference/migration-tool path.

### API docs

`docs/swagger.yaml` documents the REST API — update it when routes/DTOs change.
