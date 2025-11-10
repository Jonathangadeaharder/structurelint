# cmd

⬆️ **[Parent Directory](../README.md)**

## Overview

The `cmd` directory contains command-line entry points for structurelint following Go's standard project layout.

## Subdirectories

| Directory | Purpose |
|-----------|---------|
| [`structurelint/`](structurelint/README.md) | Main CLI application entry point |

## Design

Following the [Standard Go Project Layout](https://github.com/golang-standards/project-layout), each subdirectory represents a separate executable that can be built.

## Building

```bash
go build ./cmd/structurelint
```

This produces the `structurelint` binary that users invoke.
