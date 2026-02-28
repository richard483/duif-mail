# Repository Guidelines

## Project Structure & Module Organization
- `cmd/api/main.go`: composition root and gRPC server bootstrap.
- `internal/domain/`: core entities and interfaces (no framework dependencies).
- `internal/usecase/`: business workflows implementing domain services.
- `internal/repository/`: data access adapters (current in-memory implementation).
- `internal/server/`: gRPC transport adapter.
- `internal/config/`: environment-driven configuration loading.
- `proto/`: `mail.proto` plus generated `mail.pb.go` and `mail_grpc.pb.go`.
- Root docs: `README.md` (setup/run) and `ARCHITECTURE.md` (layering rules).

## Build, Test, and Development Commands
- `go mod tidy`: sync and clean dependencies.
- `go run cmd/api/main.go`: run the service locally.
- `go build -o gomail.exe cmd/api/main.go`: build a local binary.
- `go test ./...`: run all tests across packages.
- `go test -cover ./...`: run tests with coverage.
- `protoc --proto_path=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/*.proto`: regenerate protobuf stubs after `.proto` changes.

## Coding Style & Naming Conventions
- Use `gofmt` formatting and idiomatic Go style before committing.
- Package names stay lowercase and concise (for example: `domain`, `usecase`).
- Exported identifiers use `PascalCase`; unexported use `camelCase`.
- Keep dependency flow inward only: `server/repository -> usecase/domain`, never reverse.
- Keep transport/protobuf conversion in `internal/server`, not in domain logic.

## Testing Guidelines
- Use Go's `testing` package; place tests as `*_test.go` beside source files.
- Name tests by behavior, for example `TestMailUseCase_SendEmail`.
- Prefer table-driven tests for validation and edge cases.
- Cover use case logic and gRPC handler request/response mapping for each feature change.

## Commit & Pull Request Guidelines
- Use Conventional Commits: `type(scope): short summary`.
- Allowed types: `feat`, `fix`, `refactor`, `test`, `docs`, `chore`.
- Example: `feat(mail): add template variables to send flow`.
- Keep subject lines imperative and under 72 characters.
- PRs must include: purpose, key changes, test evidence (`go test ./...`), and linked issue/task.
- For API/proto updates, include sample `grpcurl` request/response and list regenerated files.

## Security & Configuration Tips
- Do not commit secrets. Copy `.env.example` to `.env` for local development.
- Required mail credentials must be supplied via environment variables in non-dev environments.

## Roadmap Tracking
- Track implementation progress in `docs/ROADMAP.md`.
- Update status after each merged phase/task (`Planned`, `In Progress`, `Done`).
- Keep tasks practical and scoped; avoid adding non-essential complexity.
