# examples

⬆️ **[Parent Directory](../README.md)**

## Overview

The `examples` directory contains example projects and configuration files demonstrating various structurelint use cases and patterns.

## Purpose

- **Documentation by Example**: Show real-world configurations
- **Testing**: Validate that structurelint works with different project structures
- **Learning**: Help users understand how to configure rules for their needs

## Available Configuration Examples

### Project Type Examples

| File | Description | Key Features |
|------|-------------|--------------|
| `basic.yml` | Minimal setup for any project | Basic metrics, naming conventions |
| `react-project.yml` | React/TypeScript configuration | Component organization, test adjacency |
| `go-project.yml` | Go project structure | Standard Go layout, adjacent tests |
| `python-project.yml` | Python project structure | Separate test directory |
| `monorepo.yml` | Monorepo configuration | Multiple packages, shared rules |

### Architecture Examples

| File | Description | Enforces |
|------|-------------|----------|
| `clean-architecture.yml` | Clean Architecture pattern | Domain, Application, Infrastructure layers |
| `hexagonal-architecture.yml` | Ports & Adapters pattern | Core, Ports, Adapters separation |
| `feature-sliced.yml` | Feature-Sliced Design | Shared, Entities, Features, Widgets, Pages |

### Advanced Feature Examples

| File | Description | Demonstrates |
|------|-------------|--------------|
| `complete-setup.yml` | All 5 phases enabled | Full structurelint capabilities |
| `dead-code-detection.yml` | Phase 2 features | Orphaned files, unused exports |
| `test-aaa-pattern.yml` | AAA pattern enforcement | Test quality and consistency |

## Using Examples

Copy an example configuration to your project:

```bash
# Copy and customize
cp examples/react-project.yml .structurelint.yml

# Or use --init for automatic configuration
structurelint --init
```

Test any example configuration:

```bash
# Dry run with an example config
structurelint . --config examples/go-project.yml
```
