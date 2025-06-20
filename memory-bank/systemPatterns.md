# System Patterns

## Architecture Overview
Zensor Server follows a layered, event-driven architecture inspired by the C4 model. It consists of:
- Web Application (API server)
- Device Controller, Service, and Command Publisher (control plane)
- Command Worker and Lora Integration Worker (data plane)
- Message Broker (Kafka/Redpanda)
- Database (Materialize/PostgreSQL)

## Key Technical Decisions
- Use of Kafka/Redpanda for event and command streaming
- Materialize as a query persistence layer for real-time views
- MQTT for device event ingestion
- Go modules for dependency management
- OpenTelemetry for observability
- Multi-tenancy support at the data and API layer

## Design Patterns in Use
- Dependency injection (Google Wire)
- Repository pattern for persistence
- Command pattern for device actions
- Worker pattern for background processing
- Event-driven communication via pub/sub
- DTOs for message passing

## Component Relationships
- Device Controller queues commands via Device Service
- Device Service dispatches commands through Command Publisher
- Command Publisher publishes to Kafka topics
- Command Worker processes pending commands and updates device state
- Lora Integration Worker ingests device events from MQTT and updates state
- All components interact via well-defined interfaces and event streams

---

> _This file captures the technical structure and reasoning behind it._ 