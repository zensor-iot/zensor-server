# Tech Context

## Technologies Used
- Go (1.23.6)
- Kafka/Redpanda (event streaming)
- Materialize/PostgreSQL (database, query layer)
- MQTT (device event ingestion)
- Google Wire (dependency injection)
- Goka (stream processing)
- OpenTelemetry (metrics, tracing)
- GORM (ORM)
- Viper (configuration)
- Prometheus (metrics)

## Development Setup
- Clone the repository
- Install Go 1.23.6+
- Run `just setup` to start dependencies (Redpanda, Materialize, etc.)
- Run `plz run //server` to start the server
- Use `just tdd` or `just unit` for tests
- Configuration via `config/server.yaml` and environment variables

## Technical Constraints
- All configuration must be externalized (no hardcoding)
- Only stable, tagged Go module dependencies
- Multi-tenant data isolation
- Eventual consistency for device/event state
- Observability required for all services

## Dependencies
- github.com/eclipse/paho.mqtt.golang: MQTT client
- github.com/lovoo/goka: Stream processing
- github.com/jackc/pgx/v5: PostgreSQL driver
- github.com/gorilla/websocket: WebSocket support
- github.com/prometheus/client_golang: Metrics
- github.com/spf13/viper: Configuration
- go.opentelemetry.io/otel: Observability
- gorm.io/gorm: ORM

---

> _This file helps onboard new developers and clarify technical boundaries._ 