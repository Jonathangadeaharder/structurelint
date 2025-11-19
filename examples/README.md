# Structurelint Examples

## Overview

This directory contains example configurations for common architectural patterns. Use these as templates for your own projects.

## Available Examples

### 1. [Clean Architecture](./clean-architecture/)

**Best for**: Backend services with clear business logic separation

**Key features**:
- Domain layer with no dependencies
- Use cases depend only on domain
- Infrastructure at the outer layer
- Strict layer boundaries

**Use when**: Building backend services with complex business logic

### 2. [Hexagonal Architecture](./hexagonal-architecture/)

**Best for**: Services with multiple input/output channels

**Key features**:
- Application core is isolated
- Ports define interfaces  
- Adapters implement ports
- No adapter-to-adapter communication

**Use when**: Multiple input sources (REST, gRPC, CLI)

### 3. [Domain-Driven Design](./ddd/)

**Best for**: Complex domains with multiple bounded contexts

**Key features**:
- Bounded contexts are isolated
- Aggregates own their entities
- Domain events
- Ubiquitous language

**Use when**: Complex business domain, event-driven architecture

### 4. [Microservices](./microservices/)

**Best for**: Distributed systems with independent services

**Key features**:
- Service independence
- API contracts (OpenAPI)
- No direct database sharing

**Use when**: Multiple teams, different release cycles

### 5. [Frontend Monorepo](./monorepo-frontend/)

**Best for**: Multiple frontend apps sharing components

**Key features**:
- Shared design system
- Independent apps
- React/TypeScript

**Use when**: Multiple frontend applications, shared components

## Quick Start

```bash
# Copy example to your project
cp examples/clean-architecture/.structurelint.yml your-project/

# Customize for your structure
vim your-project/.structurelint.yml

# Run linter
cd your-project && structurelint
```
