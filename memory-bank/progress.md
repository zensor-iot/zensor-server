# Progress

## What Works
- Device registration and management via HTTP API
- Event ingestion from MQTT
- Command sequencing and dispatch to devices
- Evaluation rules for device behavior
- Task scheduling and execution
- Multi-tenant support
- HTTP API for device, event, and task management
- PostgreSQL as the single persistence layer (direct reads and writes, no CDC indirection)
- Observability (metrics, health checks)
- Scheduled tasks with cron-based scheduling
- **Command flow between control plane and data plane**

> **2026-07-28: Removed Kafka, Kafka Connect, Schema Registry, and Avro.** Every repository now writes directly to Postgres; `LoraIntegrationWorker` polls Postgres for ready commands instead of consuming a Kafka topic. The former "replication module for local development" is gone since there's nothing left to replicate. See `systemPatterns.md` and `techContext.md` for the current architecture.

## What's Left to Build
- (nothing tracked here related to the former replication/Avro work — see note above)

## Technical Debt & Improvements
- **Avro Mapping Refactoring**: Replace reflection-based conversion with typed methods for better performance and type safety

## Current Status
- **MVP**: Complete ✅
- **Core Features**: Complete ✅
- **Scheduled Tasks**: Complete ✅
- **Replication Module**: Removed — repositories write directly to Postgres, nothing left to replicate ✅
- **Command Flow**: Fixed and working ✅

## Known Issues
- ~~Command type mismatch between control plane and data plane~~ ✅ **RESOLVED**

## Recent Achievements
- ✅ **Fixed command type mismatch issue**: Resolved LoraIntegrationWorker receiving internal.Command instead of shared_kernel.Command
- ✅ **Implemented JSON-based command conversion**: Added generic conversion method to handle type differences
- ✅ **Updated command consumption**: Modified worker to use generic map consumption with conversion
- ✅ **Implemented typed Avro device conversion**: Created `convertDomainDevice` method that accepts `*domain.Device` and returns `*AvroDevice` for better type safety
- ✅ **Fixed Avro union type serialization**: Corrected `last_message_received_at` field serialization to use proper union format for timestamp-millis
- Implemented comprehensive replication module architecture
- Created TopicHandler interface for flexible topic-specific replication
- Built DeviceHandler and TenantHandler implementations
- Added wire configuration for conditional replication service initialization
- Integrated replication service into main application flow
- Created comprehensive documentation for replication module

---

> _Track what's working, what's broken, and what's next._ 