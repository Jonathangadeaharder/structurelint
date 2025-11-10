# linter

⬆️ **[Parent Directory](../README.md)**

## Overview

The `linter` package is the main orchestrator that coordinates all linting operations. It manages the workflow from configuration loading through rule execution.

## Key Responsibilities

1. **Configuration Management**: Loads and merges configuration files
2. **Filesystem Walking**: Coordinates filesystem traversal
3. **Import Graph Building**: Triggers graph construction when needed
4. **Rule Instantiation**: Creates rule instances based on configuration
5. **Violation Collection**: Aggregates violations from all rules

## Main Types

- `Linter`: Main orchestrator struct
- `Violation`: Represents a single rule violation (alias from rules package)

## Workflow

```
Load Config → Walk Filesystem → Build Import Graph (if needed) → Create Rules → Execute Rules → Return Violations
```

## Recent Refactoring

The `createRules` function was recently refactored to extract type assertion logic into helper methods (`getIntConfig`, `getStringMapConfig`, `getStringSliceConfig`), reducing cognitive complexity and eliminating linter suppressions.
