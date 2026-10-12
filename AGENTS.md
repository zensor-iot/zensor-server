# AGENTS.md

Instrucciones para agentes de código (Claude Code, Codex, Cursor, etc.) que trabajan en este repositorio.

## UI de Zensor

**Cada vez que crees o modifiques UI** (cualquier archivo en `web/`, componentes, estilos, copy, íconos o assets), primero lee la guía de estilo:

- [`docs/design-system/STYLE.md`](docs/design-system/STYLE.md) — guía completa del sistema de diseño. Lee la **sección 1 (reglas duras)** antes de escribir código y valida el resultado con el **checklist de la sección 13** antes de entregar.

Reglas mínimas que no se negocian:

- Usa solo tokens semánticos de `docs/design-system/tokens.css` / clases del preset `docs/design-system/tailwind.preset.js`. Nada de hex sueltos ni colores por defecto de Tailwind.
- Fuente de verdad de los tokens: `docs/design-system/tokens.json` (no edites `tokens.css` a mano).
- Referencia visual: `docs/design-system/preview.html`. Logos en `docs/design-system/logos/`.
- Patrón principal: lista de tarjetas compactas + drawer derecho con el detalle.
- Copy en español de Chile, tuteo, sentence case, sin emojis ni datos inventados.

Implementación en `web/`:

- Tailwind v3 usa `docs/design-system/tailwind.preset.js` (reemplaza la escala por defecto) y `tokens.css` se importa en `web/src/main.jsx`. Vite resuelve `@design-system` → `docs/design-system`; los logos se sirven en `/ui/logos/`.
- Construye pantallas con el kit de `web/src/ui` (Button, Drawer, ConfirmDialog, StatusPill, Notice, EmptyState, FilterChips, Switch, Field/Input/Select, Table, SensorTile, HumidityGauge…) y los helpers de `web/src/lib/zensor.js` (formato es-CL, `humidityState`, tema). No agregues archivos `.css` por pantalla.
- `web/src/components/AppShell.jsx` es el único shell (sidebar con selector de cuenta). Las páginas empiezan con `<Page>` + `<PageHeader>`.
- Referencia del patrón lista + drawer: `web/src/components/devices/`.

## Project Overview

Zensor Server is a Go 1.23+ IoT backend service for managing tenants, devices, tasks, commands, and scheduled tasks. It follows clean architecture with PostgreSQL persistence (direct reads and writes), Redis caching, and MQTT device communication.

## Build & Run Commands

```bash
just init                              # Install all Go tools (golangci-lint, wire, arch-go, mockgen)
just build                             # Compile: go build -o server cmd/api/main.go
just run                               # Build + hot reload with air (starts Docker deps unless ENV=local)
ENV=local just run                     # Run without Docker dependencies (in-memory SQLite ORM, no-op MQTT client)
```

`just build`/`just run`/`just dev`/`just unit`/`just tdd` all first run `just _web-build` (pnpm install + vite build of `./web` into the go:embed'd `internal/infra/httpserver/web/dist`) — pnpm must be installed for these to work.

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
  infra/          → Infrastructure: HTTP server, MQTT, SQL, cache, async (in-process broker)
  shared_kernel/  → Cross-cutting domain utilities
  persistence/    → Repository implementations (GORM + PostgreSQL)
  maintenance/    → Maintenance module (conditionally enabled)
web/              → React/Vite frontend (SPA), built by pnpm and embedded into the Go binary
internal/infra/httpserver/web/ → go:embed package that serves web/'s build output (dist/)
```

### Key Patterns

- **Dependency Injection**: Google Wire (compile-time). All wiring in `cmd/api/wire/`. Run `just wire` after changing providers.
- **Repository Pattern**: Interfaces defined in `control_plane/usecases/`, implementations in `persistence/`. Standard methods: `Create`, `GetByID`, `FindAll`, `Update`.
- **Workers**: All implement `async.Worker` interface (`Run(ctx, done)` + `Shutdown()`). Workers: CommandWorker, ScheduledTaskWorker, NotificationWorker, MetricWorker, PushNotificationWorker, ExecutionWorker, LoraIntegrationWorker. They coordinate via an in-process `async.InternalBroker` and periodic ticker polling of Postgres — there is no external message broker.
- **HTTP Controllers**: Implement `Controller` interface with `AddRoutes(*http.ServeMux)`. User context extracted from headers: `X-User-ID`, `X-User-Name`, `X-User-Email`.
- **Multi-tenant**: All data scoped by Tenant. Entity hierarchy: Tenant → Devices → Tasks → Commands. ScheduledTasks are templates that generate Task instances.
- **Modules**: `permaculture` and `maintenance` modules are conditionally loaded via `config/server.yaml` (`modules.*.enabled`).

### Configuration

- YAML config: `config/server.yaml`
- Environment variable overrides: prefix `ZENSOR_SERVER_`, dots become underscores (e.g., `ZENSOR_SERVER_DATABASE_DSN`)
- `ENV=local` switches to an in-memory SQLite ORM and a no-op MQTT client (no external deps needed)

### Infrastructure (Docker Compose)

`docker compose up -d --wait` starts: PostgreSQL (5432), Redis (6379), Prometheus, Grafana (3001), OpenTelemetry Collector (4317/4318).

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
