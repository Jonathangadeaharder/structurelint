# Domain-Driven Design Example

## Overview

This configuration enforces DDD patterns with bounded contexts.

## Bounded Contexts

```
┌────────────────────┐   ┌────────────────────┐   ┌────────────────────┐
│   User Context     │   │   Order Context    │   │  Payment Context   │
│                    │   │                    │   │                    │
│  ┌──────────────┐  │   │  ┌──────────────┐  │   │  ┌──────────────┐  │
│  │  Domain      │  │   │  │  Domain      │  │   │  │  Domain      │  │
│  │  - Aggregates│  │   │  │  - Aggregates│  │   │  │  - Aggregates│  │
│  │  - Entities  │  │   │  │  - Entities  │  │   │  │  - Entities  │  │
│  │  - VOs       │  │   │  │  - VOs       │  │   │  │  - VOs       │  │
│  └──────────────┘  │   │  └──────────────┘  │   │  └──────────────┘  │
│  ┌──────────────┐  │   │  ┌──────────────┐  │   │  ┌──────────────┐  │
│  │ Application  │  │   │  │ Application  │  │   │  │ Application  │  │
│  │  - Commands  │  │   │  │  - Commands  │  │   │  │  - Commands  │  │
│  │  - Queries   │  │   │  │  - Queries   │  │   │  │  - Queries   │  │
│  └──────────────┘  │   │  └──────────────┘  │   │  └──────────────┘  │
└────────────────────┘   └────────────────────┘   └────────────────────┘
         │                        │                        │
         └────────────────────────┴────────────────────────┘
                                  │
                          ┌───────▼────────┐
                          │ Shared Kernel  │
                          └────────────────┘
```

## Key DDD Patterns

- **Bounded Contexts**: Isolated domains with clear boundaries
- **Aggregates**: Consistency boundaries for entities
- **Value Objects**: Immutable domain concepts
- **Domain Events**: Capture state changes
- **Repositories**: Persistence abstraction in domain

## Structure

Each bounded context follows:
```
internal/{context}/
  domain/
    aggregates/
    entities/
    valueobjects/
    events/
    repository.go (interfaces)
  application/
    commands/
    queries/
  infrastructure/
```

## Usage

```bash
cp .structurelint.yml your-project/
cd your-project
structurelint
```
