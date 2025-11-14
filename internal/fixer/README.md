# fixer

⬆️ **[Parent Directory](../README.md)**

## Overview

The `fixer` package applies automated fixes to the codebase to resolve structurelint violations.

## Features

- **Automated remediation**: Automatically fixes violations that can be safely corrected
- **Dry-run mode**: Preview fixes without applying them
- **Multiple fix types**: Supports rename, delete, and modify operations
- **Verbose output**: Clear feedback about what's being fixed

## Supported Fix Types

- **Rename**: Rename files or directories (e.g., fix naming convention violations)
- **Delete**: Remove orphaned or unused files
- **Modify**: Update file content (e.g., remove unused exports)

## Usage

```go
import "github.com/structurelint/structurelint/internal/fixer"

// Create a new fixer (dryRun=false, verbose=true)
f := fixer.New(false, true)

// Apply fixes
err := f.Apply(fixes)
```

## CLI Usage

```bash
# Apply all fixes
structurelint --fix .

# Preview fixes without applying them
structurelint --dry-run .
```

## Supported Rules

Currently supports automated fixing for:
- `naming-convention`: Rename files to match the configured naming convention
- More rules coming soon...
