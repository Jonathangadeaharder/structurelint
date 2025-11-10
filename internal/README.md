# internal

⬆️ **[Parent Directory](../README.md)**

## Overview

The `internal` directory contains all internal implementation packages for structurelint. These packages are not intended for external import and implement the core functionality of the linter.

## Package Architecture

This directory follows Go's internal package convention - code here is only importable by packages within the structurelint project.

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| [`config/`](config/README.md) | Configuration loading and merging |
| [`graph/`](graph/README.md) | Import graph analysis and layer validation |
| [`linter/`](linter/README.md) | Main linter orchestration |
| [`parser/`](parser/README.md) | Source code parsing for imports/exports |
| [`rules/`](rules/README.md) | Rule implementations |
| [`walker/`](walker/README.md) | Filesystem traversal and analysis |

## Design Principles

- **Separation of Concerns**: Each package has a single, well-defined responsibility
- **Testability**: All packages are designed with testability in mind
- **No Circular Dependencies**: Package dependencies flow in one direction (walker → parser → graph → rules → linter)
