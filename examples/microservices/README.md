# Microservices Architecture Example

## Overview

This configuration enforces microservices independence and best practices.

## Service Architecture

```
┌─────────────────┐   ┌─────────────────┐   ┌─────────────────┐
│  User Service   │   │  Order Service  │   │ Payment Service │
│                 │   │                 │   │                 │
│  ┌───────────┐  │   │  ┌───────────┐  │   │  ┌───────────┐  │
│  │ OpenAPI   │  │   │  │ OpenAPI   │  │   │  │ OpenAPI   │  │
│  │ Contract  │  │   │  │ Contract  │  │   │  │ Contract  │  │
│  └───────────┘  │   │  └───────────┘  │   │  └───────────┘  │
│  ┌───────────┐  │   │  ┌───────────┐  │   │  ┌───────────┐  │
│  │   API     │  │   │  │   API     │  │   │  │   API     │  │
│  │ Handlers  │  │   │  │ Handlers  │  │   │  │ Handlers  │  │
│  └───────────┘  │   │  └───────────┘  │   │  └───────────┘  │
│  ┌───────────┐  │   │  ┌───────────┐  │   │  ┌───────────┐  │
│  │ Business  │  │   │  │ Business  │  │   │  │ Business  │  │
│  │  Logic    │  │   │  │  Logic    │  │   │  │  Logic    │  │
│  └───────────┘  │   │  └───────────┘  │   │  └───────────┘  │
│  ┌───────────┐  │   │  ┌───────────┐  │   │  ┌───────────┐  │
│  │ Database  │  │   │  │ Database  │  │   │  │ Database  │  │
│  └───────────┘  │   │  └───────────┘  │   │  └───────────┘  │
└─────────────────┘   └─────────────────┘   └─────────────────┘
         │                     │                     │
         └─────────────────────┴─────────────────────┘
                               │
                    ┌──────────▼──────────┐
                    │  Shared Libraries   │
                    │    (pkg/)          │
                    └────────────────────┘
```

## Key Principles

- **Service Independence**: No direct service-to-service imports
- **API Contracts**: OpenAPI specifications required
- **Database per Service**: No shared databases
- **Shared Libraries**: Common utilities in pkg/

## Enforcement

- Services cannot import from other services
- Each service requires OpenAPI contract
- Each service has Dockerfile
- Standardized structure per service

## Usage

```bash
cp .structurelint.yml your-project/
cd your-project
structurelint
```
