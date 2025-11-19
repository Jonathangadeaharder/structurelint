# Clean Architecture Example

## Overview

This configuration enforces Clean Architecture patterns for Go projects.

## Architecture Layers

```
┌─────────────────────────────────────┐
│         Interfaces/UI               │  (Controllers, CLI)
│  ┌───────────────────────────────┐  │
│  │      Infrastructure           │  │  (DB, APIs, External)
│  │  ┌─────────────────────────┐  │  │
│  │  │     Use Cases          │  │  │  (Business Rules)
│  │  │  ┌───────────────────┐ │  │  │
│  │  │  │     Domain        │ │  │  │  (Entities, Pure Logic)
│  │  │  └───────────────────┘ │  │  │
│  │  └─────────────────────────┘  │  │
│  └───────────────────────────────┘  │
└─────────────────────────────────────┘
```

## Dependency Rules

- **Domain**: No dependencies (pure business logic)
- **Use Cases**: Depends only on Domain
- **Infrastructure**: Depends on Domain and Use Cases
- **Interfaces**: Depends on all layers

## Usage

```bash
cp .structurelint.yml your-project/
cd your-project
structurelint
```

## Key Rules

- Domain purity (no DB, HTTP, external libs)
- Layer boundary enforcement
- Naming conventions for each layer
- Test location requirements
