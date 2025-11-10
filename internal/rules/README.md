# rules

⬆️ **[Parent Directory](../README.md)**

## Overview

The `rules` package contains all rule implementations for structurelint. Each rule validates a specific aspect of project structure or architecture.

## Rule Interface

All rules implement the `Rule` interface:

```go
type Rule interface {
    Name() string
    Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation
}
```

## Available Rules

### Phase 0: Filesystem Rules
- **max-depth**: Limits directory nesting depth
- **max-files-in-dir**: Limits files per directory
- **max-subdirs**: Limits subdirectories per directory
- **naming-convention**: Enforces naming patterns (camelCase, kebab-case, etc.)
- **disallowed-patterns**: Blocks specific file patterns
- **file-existence**: Requires specific files to exist
- **regex-match**: Validates filenames match regex patterns

### Phase 1: Architectural Rules
- **enforce-layer-boundaries**: Validates architectural layer dependencies

### Phase 2: Dead Code Detection
- **disallow-orphaned-files**: Detects files not imported anywhere
- **disallow-unused-exports**: Finds exported symbols that are never imported

## Adding New Rules

1. Create a new file `your_rule.go`
2. Implement the `Rule` interface
3. Add factory function `NewYourRule(...) *YourRule`
4. Register in `linter/linter.go`'s `createRules` method
5. Add tests in `your_rule_test.go`
