# Product Context

## Why This Project Exists
IoT deployments require a robust backend to manage devices, process sensor data, and orchestrate commands. Zensor Server solves the problem of scalable, reliable device and event management, providing a unified platform for device registration, event ingestion, and command execution.

## Target Users
- IoT system administrators
- Developers integrating devices and sensors
- Product teams needing real-time device data
- End users (indirectly, via web interfaces)

## User Experience Goals
- Simple device onboarding and management
- Real-time visibility into device status and events
- Reliable and traceable command execution
- Easy integration with external systems (via APIs and event streams)

## How It Should Work
1. Devices are registered and associated with tenants.
2. Devices send events via MQTT, which are ingested and materialized.
3. Users interact with the system via HTTP APIs to manage devices, tasks, and evaluation rules.
4. Commands are queued, sequenced, and dispatched to devices based on rules and schedules.
5. System provides health, metrics, and observability endpoints for monitoring.

---

> _This file explains the "why" and "how" from a product perspective._ 