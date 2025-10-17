# Project Brief

## Project Name
Zensor Server

## Project Association
This repository is associated with the **zensor-iot** organization's **Main Prioritization** project (Project #2). All development tasks, features, and work items must be tracked and managed through this GitHub project to ensure proper prioritization and visibility.

## Purpose
Zensor Server is the core backend of the Zensor Ecosystem, responsible for orchestrating device management, event processing, and command sequencing for IoT devices. It provides APIs and integrations for device registration, event ingestion, command dispatch, and real-time monitoring.

## Scope
Included:
- Device registration and management
- Event ingestion from MQTT and Kafka
- Command sequencing and dispatch to devices
- Evaluation rules for device behavior
- Task scheduling and execution
- Multi-tenant support
- HTTP API for device, event, and task management
- Integration with Materialize (query persistence layer)
- Observability (metrics, health checks)

Excluded:
- Frontend web application (handled separately)
- Direct device firmware
- External analytics (beyond event streaming)

## Stakeholders
- IoT device operators and administrators
- Product and engineering teams
- End users (via web application)
- System integrators

## Success Criteria
- Reliable device and event management at scale
- Accurate and timely command dispatch
- Robust multi-tenant isolation
- High system availability and observability
- Extensible for new device types and integrations

## Timeline
- MVP: Device registration, event ingestion, command dispatch, and basic API endpoints (Complete)
- Next: Device resource API, sensor resource API, OpenAPI documentation, configuration improvements
- Ongoing: Feature expansion, bug fixes, and performance improvements

---

> _This document is the foundation for all other Memory Bank files. Update as the project evolves._ 