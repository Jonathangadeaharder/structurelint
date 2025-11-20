# Hexagonal Architecture Example

## Overview

This configuration enforces Hexagonal Architecture (Ports & Adapters) patterns.

## Architecture Overview

```
        ┌──────────────────────────┐
        │   Primary Adapters       │  (REST, gRPC, CLI)
        │    (Inbound)             │
        └───────────┬──────────────┘
                    │
        ┌───────────▼──────────────┐
        │       Ports              │  (Interfaces)
        │  ┌────────────────────┐  │
        │  │   Application Core │  │  (Business Logic)
        │  └────────────────────┘  │
        └───────────┬──────────────┘
                    │
        ┌───────────▼──────────────┐
        │  Secondary Adapters      │  (DB, External APIs)
        │    (Outbound)            │
        └──────────────────────────┘
```

## Key Concepts

- **Core**: Business logic isolated from external concerns
- **Ports**: Interfaces defining communication contracts
- **Primary Adapters**: Drive the application (HTTP handlers, CLI)
- **Secondary Adapters**: Driven by application (DB, APIs)

## Dependency Rules

- Core has no dependencies on adapters
- Ports define interfaces, not implementations
- Adapters implement ports
- No adapter-to-adapter communication

## Usage

```bash
cp .structurelint.yml your-project/
cd your-project
structurelint
```
