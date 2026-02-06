# Project Roadmap

This tracker is the single source of truth for implementation progress.

## Status Key
- `Planned`: not started
- `In Progress`: actively being implemented
- `Done`: implemented and verified

## Phase 1: Core Cleanup + Test Baseline
Status: `Done`
- [x] Standardize gRPC error handling with status codes.
- [x] Improve request validation (required fields, email format, CC parsing).
- [x] Persist rendered template body consistently.
- [x] Add tests for `internal/usecase` and `internal/server`.
- [x] Reach at least 80% coverage for critical package(s).
Notes:
- `internal/usecase` coverage: 91.0%
- `internal/server` coverage: 88.6%

## Phase 2: PostgreSQL Persistence + DB Templates
Status: `Done`
- [x] Implement PostgreSQL repository (replace in-memory for runtime usage).
- [x] Add DB schema for `emails` and `email_templates`.
- [x] Add minimal send log table if needed for troubleshooting.
- [x] Move template source from code to DB.
- [x] Provide SQL DDL + seed data examples (including templates).
Notes:
- SQL files added: `db/001_init.sql`, `db/002_seed.sql`
- Runtime storage selectable via `STORAGE_DRIVER` (`memory` or `postgres`)

## Phase 3: Dockerization (Jenkins-Ready)
Status: `Done`
- [x] Add multi-stage `Dockerfile` with Alpine runtime.
- [x] Run app as non-root user.
- [x] Document container build/run commands in `README.md`.
Notes:
- Docker assets added: `Dockerfile`, `.dockerignore`

## Phase 4: Production-Sensible Defaults
Status: `Done`
- [x] Disable gRPC reflection outside development.
- [x] Remove sensitive startup logs.
- [x] Add simple readiness/health approach for runtime checks.
- [x] Keep SMTP provider configuration single and clear.
Notes:
- Registered gRPC health service and set serving status in `cmd/api/main.go`

## Phase 5: Documentation Cleanup
Status: `Done`
- [x] Update `README.md` for local/dev/prod-like flow.
- [x] Document DB setup and migration steps.
- [x] Document template lifecycle and update flow.
Notes:
- Added PostgreSQL `psql` setup commands.
- Added template UUID reference and `grpcurl` example for `template_id`.
- Added template update SQL example and health check command.

## Operating Notes
- Keep changes simple and readable; avoid unnecessary abstractions.
- Complete one phase at a time and update this file immediately after verification.

## Post-Phase Updates
Status: `Done`
- [x] Added `ListTemplates` gRPC endpoint (UUID-based templates).
- [x] Confirmed `GetAllEmails` returns all statuses (`sent`, `failed`, `pending`).
- [x] Added send-log persistence for every send attempt (success and failure).
