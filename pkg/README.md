# pkg

⬆️ **[Parent Directory](../README.md)**

## Overview

The `pkg` directory contains public-facing Go packages that provide stable APIs for using structurelint programmatically.

## Packages

- **[api](api/README.md)** - Stable public API for programmatic linting and architectural rules

## Purpose

Packages in `pkg/` follow Go conventions for public packages:
- Designed for external consumption by other Go programs
- Follow semantic versioning for stability guarantees
- Provide backward-compatible interfaces
- Well-documented with examples and godoc comments

## Usage

Import packages from the `pkg/` directory to integrate structurelint into your applications:

```go
import "github.com/structurelint/structurelint/pkg/api"
```

See individual package READMEs for detailed usage instructions.
