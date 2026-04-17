# Architecture Gaps — zensor-server

> Expert review against Architecture and Observability. Reviewers: Alex the Architect, Kira the Observability Engineer. Expert selection: `1,3`.

---

## Alex — Architecture

### GAP-A1: Runtime configuration is hardcoded in bootstrap paths (🟠 High)

**What the code does:** Runtime-critical values are hardcoded in bootstrap instead of being consistently loaded from config, for example HTTP listen address and telemetry service metadata.

```go
// internal/infra/httpserver/server.go
Addr: ":3000",

// cmd/api/main.go
semconv.ServiceNameKey.String("zensor-server")
```

**Why it matters:** Hardcoded deployment values increase environment drift risk, make staging/production parity harder, and create hidden coupling between code and infrastructure.

**What it should be:**
```go
addr := fmt.Sprintf(":%d", appConfig.Server.Port)
server := &http.Server{Addr: addr, Handler: handler}

tp := trace.NewTracerProvider(
    trace.WithResource(resource.NewWithAttributes(
        semconv.SchemaURL,
        semconv.ServiceNameKey.String(appConfig.OTel.ServiceName),
    )),
)
```

---

### GAP-A2: Panic-based control flow in non-recoverable runtime paths (🟠 High)

**What the code does:** Multiple runtime paths crash the entire process using `panic` for operational failures (server lifecycle, worker setup, handler registration).

```go
// internal/infra/httpserver/server.go
if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    panic(err)
}

// cmd/api/main.go
if err := replicationService.RegisterHandler(deviceHandler); err != nil {
    slog.Error("failed to register device handler", slog.Any("error", err))
    panic(err)
}
```

**Why it matters:** Process-wide crashes turn localized failures into full outages and reduce resilience under partial dependency failures.

**What it should be:**
```go
if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
    return fmt.Errorf("starting HTTP server: %w", err)
}

if err := replicationService.RegisterHandler(deviceHandler); err != nil {
    return fmt.Errorf("registering device handler: %w", err)
}
```

---

### GAP-A3: Composition root is overloaded with orchestration concerns (🟡 Medium)

**What the code does:** `cmd/api/main.go` mixes wiring, module feature flags, worker lifecycle, replication registration, MQTT setup, and OTel bootstrap in one large function.

```go
// cmd/api/main.go (single main flow handling all concerns)
controllers := []httpserver.Controller{...}
go httpServer.Run()
replicationService := initializeReplicationService()
...
go handleWireInjector(wire.InitializeLoraIntegrationWorker(...)).(async.Worker).Run(appCtx, wg.Done)
...
signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
```

**Why it matters:** A large composition root raises regression risk, complicates testability of startup behavior, and makes feature additions harder to isolate.

**What it should be:**
```go
func run(ctx context.Context, cfg config.AppConfig) error {
    app, err := bootstrapApp(cfg)
    if err != nil {
        return fmt.Errorf("bootstrapping app: %w", err)
    }
    return app.Run(ctx)
}
```

---

## Kira — Observability

### GAP-K1: PII is written into trace attributes (🔴 Critical)

**What the code does:** Request middleware sends `user.name` and `user.email` directly into tracing attributes.

```go
// internal/infra/httpserver/server.go
if userName != "" {
    span.SetAttributes(attribute.String("user.name", userName))
}
if userEmail != "" {
    span.SetAttributes(attribute.String("user.email", userEmail))
}
```

**Why it matters:** PII in telemetry increases compliance and data-retention risk, and can leak sensitive data to third-party observability backends.

**What it should be:**
```go
if userID != "" {
    span.SetAttributes(attribute.String("user.id", userID))
}
// Do not attach user.name or user.email to spans/logs/metrics.
```

---

### GAP-K2: Trace propagator configuration is not centralized in OTel bootstrap (🟠 High)

**What the code does:** B3 extraction/injection is done ad hoc in HTTP middleware, but there is no global `otel.SetTextMapPropagator(...)` configuration in startup.

```go
// internal/infra/httpserver/server.go
propagator := b3.New()
ctx := propagator.Extract(r.Context(), propagation.HeaderCarrier(r.Header))
...
propagator.Inject(ctx, propagation.HeaderCarrier(w.Header()))
```

**Why it matters:** Without centralized propagator setup, trace context behavior can drift between HTTP, async, and future outbound clients, causing broken trace continuity.

**What it should be:**
```go
// during OTel initialization
b3Propagator := b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader))
otel.SetTextMapPropagator(b3Propagator)

// middleware uses global propagator
propagator := otel.GetTextMapPropagator()
```

---

### GAP-K3: Non-structured stdout logging bypasses observability pipeline (🟡 Medium)

**What the code does:** In-memory pub/sub error paths emit logs with `fmt.Printf` and lose structured fields and trace correlation.

```go
// internal/infra/pubsub/memory.go
fmt.Printf("panic in message handler: %v\n", r)
fmt.Printf("error in message handler: %v\n", err)
```

**Why it matters:** Unstructured logs are harder to query, harder to correlate with traces, and inconsistent with the rest of the `slog`-based logging strategy.

**What it should be:**
```go
slog.Error("panic in message handler", slog.Any("panic", r), slog.String("topic", string(event.Topic)))
slog.Error("error in message handler", slog.Any("error", err), slog.String("topic", string(event.Topic)))
```

---

### GAP-K4: Profiling visibility is undefined for production diagnosis (🟢 Low)

**What the code does:** The HTTP server imports `net/http/pprof` but does not establish a dedicated internal profiling server strategy.

```go
// internal/infra/httpserver/server.go
import _ "net/http/pprof"
```

**Why it matters:** Incident-time CPU/memory diagnosis is slower when profiling availability and access boundaries are not explicit in runtime design.

**What it should be:**
```go
go func() {
    if err := http.ListenAndServe("127.0.0.1:6060", nil); err != nil {
        slog.Error("pprof server failed", slog.Any("error", err))
    }
}()
```

---

## Priority summary

| ID | Severity | Area | Gap |
|----|----------|------|-----|
| GAP-K1 | 🔴 Critical | Observability | PII is written into trace attributes (`user.name`, `user.email`) |
| GAP-A1 | 🟠 High | Architecture | Runtime configuration is hardcoded in bootstrap paths |
| GAP-A2 | 🟠 High | Architecture | Panic-based control flow in non-recoverable runtime paths |
| GAP-K2 | 🟠 High | Observability | Trace propagator configuration is not centralized in OTel bootstrap |
| GAP-A3 | 🟡 Medium | Architecture | Composition root is overloaded with orchestration concerns |
| GAP-K3 | 🟡 Medium | Observability | Non-structured stdout logging bypasses observability pipeline |
| GAP-K4 | 🟢 Low | Observability | Profiling visibility is undefined for production diagnosis |
