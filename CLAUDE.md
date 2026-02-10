# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Zensor Server is a Go 1.23+ IoT backend service for managing tenants, devices, tasks, commands, and scheduled tasks. It follows clean architecture with event-driven processing via Kafka, PostgreSQL persistence, Redis caching, and MQTT device communication.

## Build & Run Commands

```bash
just init                              # Install all Go tools (golangci-lint, wire, arch-go, mockgen)
just build                             # Compile: go build -o server cmd/api/main.go
just run                               # Build + hot reload with entr (starts Docker deps unless ENV=local)
ENV=local just run                     # Run without Docker dependencies (uses in-memory PubSub)
```

## Testing

```bash
just unit                              # Run all unit tests (Ginkgo, randomized, race detection)
just unit internal/control_plane       # Run tests for a specific package
just tdd internal/control_plane        # Watch mode: re-runs tests on file changes
just functional permaculture "@tenant" # Functional tests for a module+tag (starts server automatically)
just functional maintenance            # Available modules: permaculture, maintenance
```

Unit tests use **Ginkgo v2** (BDD-style) with `gomega` matchers. Always use `_test` package suffix and only test exported methods. Mocks are generated with `mockgen` (`just mock`) and live in `test/unit/doubles/`.

Functional tests use **Godog** (Cucumber) with `.feature` files. Tags: `@tenant`, `@task`, `@device`, `@scheduled_task`, `@tenant_configuration`, `@pending`.

## Code Quality

```bash
just lint                              # golangci-lint (config: build/ci/golangci.yml)
just arch                              # Architecture validation with arch-go
just wire                              # Regenerate Wire DI code (cmd/api/wire/)
just mock                              # Regenerate all mockgen mocks
```

## Architecture

### Layer Structure

```
cmd/api/          → Entry point, Wire DI setup
internal/
  control_plane/  → HTTP controllers (httpapi/) + business logic (usecases/)
  data_plane/     → Event processing workers, DTOs
  infra/          → Infrastructure: HTTP server, Kafka, SQL, cache, PubSub
  shared_kernel/  → Cross-cutting domain utilities
  persistence/    → Repository implementations (GORM + PostgreSQL)
  maintenance/    → Maintenance module (conditionally enabled)
```

### Key Patterns

- **Dependency Injection**: Google Wire (compile-time). All wiring in `cmd/api/wire/`. Run `just wire` after changing providers.
- **Repository Pattern**: Interfaces defined in `control_plane/usecases/`, implementations in `persistence/`. Standard methods: `Create`, `GetByID`, `FindAll`, `Update`.
- **Workers**: All implement `async.Worker` interface (`Run(ctx, done)` + `Shutdown()`). Workers: CommandWorker, ScheduledTaskWorker, NotificationWorker, MetricWorker, PushNotificationWorker, ExecutionWorker.
- **Event-Driven**: Kafka topics with Avro schemas (in `schemas/`). In-memory PubSub for local dev (`ENV=local`).
- **HTTP Controllers**: Implement `Controller` interface with `AddRoutes(*http.ServeMux)`. User context extracted from headers: `X-User-ID`, `X-User-Name`, `X-User-Email`.
- **Multi-tenant**: All data scoped by Tenant. Entity hierarchy: Tenant → Devices → Tasks → Commands. ScheduledTasks are templates that generate Task instances.
- **Modules**: `permaculture` and `maintenance` modules are conditionally loaded via `config/server.yaml` (`modules.*.enabled`).

### Configuration

- YAML config: `config/server.yaml`
- Environment variable overrides: prefix `ZENSOR_SERVER_`, dots become underscores (e.g., `ZENSOR_SERVER_DATABASE_DSN`)
- `ENV=local` activates in-memory PubSub and replication service (no external deps needed)

### Infrastructure (Docker Compose)

`docker compose up -d --wait` starts: Kafka (19092), Schema Registry (8081), PostgreSQL (5432), Redis (6379), Prometheus, Grafana (3001), OpenTelemetry Collector (4317/4318).

## Code Style Rules

- Use `slog` global logger — never inject `*slog.Logger`
- Use `context.Context` as first parameter for blocking/I/O operations
- Use `any` instead of `interface{}`
- Prefer passing whole structs as method arguments instead of individual fields
- Comments only on exported items; no inline implementation comments — code should be self-documenting
- Error handling: always check errors, wrap with `fmt.Errorf("...: %w", err)`, never ignore with `_`
- All names in camelCase/PascalCase

## Testing Style

Ginkgo BDD structure:
```go
Context("MethodName", func() {
    var methodArg Type  // Declare at Context level

    When("scenario", func() {
        BeforeEach(func() {
            // Initialize methodArg and mocks
        })
        It("should do something", func() {
            gomega.Expect(result).To(gomega.Equal(expected))
        })
    })
})
```

## Git Workflow

- Never commit directly to main — use feature branches: `feature/<issue_id>`
- Commits follow conventional commit format (enforced by commitlint pre-commit hook)
- Production bug fixes: write a failing test first, then fix
