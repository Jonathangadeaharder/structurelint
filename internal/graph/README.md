# graph

⬆️ **[Parent Directory](../README.md)**

## Overview

The `graph` package builds import dependency graphs from source code and validates architectural layer boundaries.

## Key Features

- **Import Graph Construction**: Parses source files to build complete dependency graphs
- **Layer Validation**: Enforces architectural boundaries (Clean Architecture, Hexagonal, etc.)
- **Multi-Language Support**: Works with TypeScript, JavaScript, Go, and Python
- **Dependency Analysis**: Tracks both outgoing and incoming references

## Main Types

- `ImportGraph`: Complete dependency graph with file-to-file import relationships
- `Builder`: Constructs import graphs from parsed source files
- Layer configuration and validation logic

## Usage

```go
builder := graph.NewBuilder(path, layers)
importGraph, err := builder.Build(files)
```

## Architecture

This package is central to Phase 1 functionality (architectural layer enforcement) and Phase 2 (dead code detection).
