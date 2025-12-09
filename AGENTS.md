# Repository Guidelines

## Project Structure & Module Organization
- `services/authService/` and `services/cloudRepositoryService/`: independent Go services (Echo) with `cmd/` entrypoints and `features/` split into `handler`, `usecase`, `repository`, and `model`.
- `shared/`: common config, logger, middleware, errors, AWS helpers, and utilities consumed by all services.
- `migrations/`: SQL migrations (see `scripts/migrate.sh`), plus service-specific schemas in `services/*/DB_SCHEMA.md`.
- `docs/` and `services/*/docs/`: Swagger artifacts; regenerate per service when endpoints change.
- Tests live alongside code (`*_test.go`) with Auth E2E suites in `services/authService/tests/e2e`.

## Build, Test, and Development Commands
- Auth Service: `cd services/authService`; `make run` (local), `make build`, `make test` or `make test-e2e`, `make test-coverage`, `make fmt`, `make lint`, `make swagger-init`. `make docker-up`/`docker-down` use the root compose for MySQL/S3 mocks.
- Cloud Repository Service: `cd services/cloudRepositoryService`; `make run`, `make build`, `make test`, `make dev` (air hot reload), `make docker-build`, `make docker-run`.
- Migrations: from repo root `./scripts/migrate.sh up|down|create <name>` (requires `migrate` CLI and DB env vars).
- Docker: `docker-compose.yml` for local stack; `docker-compose.prod.yml` for production-like runs.

## Coding Style & Naming Conventions
- Go 1.24; run `gofmt` (`make fmt`) before commits. Tabs + max 120 cols; keep handlers lean by delegating to usecases.
- Lint with `golangci-lint run` (wired via `make lint` in Auth). Prefer `context.Context` plumbing and `fmt.Errorf("%w", err)` wrapping.
- Package names lowercase, file names snake_case; Echo route handlers end with `Handler`, repositories `Repository`, and use cases `UseCase`.
- Update Swagger specs after API changes (`make swagger-init` in Auth; mirror pattern for other services).

## Testing Guidelines
- Default to table-driven Go tests; name with `TestXxx`. Use `make test` for unit/integration, `make test-e2e` for Auth E2E, and check coverage with `make test-coverage`.
- Place new Auth E2E cases under `services/authService/tests/e2e` per API domain; cover success, validation, and edge cases.
- Keep tests isolated: clean DB/fixtures in `TestMain` or per-suite helpers; avoid networked AWS by mocking S3/SES.

## Commit & Pull Request Guidelines
- Follow conventional commits seen in history (`feat:`, `fix:`, `docs:`); include service scope when clear (e.g., `fix(auth): refresh token expiry`).
- PRs should note impacted service(s), key changes, test commands run, and config/env additions. Attach Swagger diffs or screenshots for API-facing updates when helpful.
- Keep changes atomic; avoid committing generated artifacts except Swagger outputs when intentionally refreshed.

## Security & Configuration Tips
- Do not commit secrets; use per-service `.env` files (see examples in `README.md`). Rotate AWS keys locally and prefer IAM roles in deployment.
- Shared config lives in `shared/config`; keep defaults safe and guard optional features with env flags.
- Before pushing, ensure DB migrations are applied locally and documented in `migrations/` with matching up/down scripts.
