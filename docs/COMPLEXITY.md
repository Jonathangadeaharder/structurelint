# Code Complexity Analysis

This document summarizes the cyclomatic and cognitive complexity metrics for the structurelint codebase.

## Complexity Linters Enabled

- **gocyclo**: Traditional cyclomatic complexity (McCabe)
- **gocognit**: Cognitive complexity (accounts for nesting weight)

## Current Complexity Metrics

### Cyclomatic Complexity (gocyclo)

Functions ranked by cyclomatic complexity:

1. `linter.createRules`: **37** (suppressed - factory function)
2. `graph.Build`: **16**
3. `file_existence.checkRequirement`: **15**
4. `walker.Walk`: **14**
5. `disallowed_patterns.matchesGlobPattern`: **13**
6. `config.FindConfigs`: **12**
7. `naming_convention.matchesPattern`: **12**
8. `unused_exports.Check`: **12**
9. `layer_boundaries.resolveDependencyToFile`: **11**

**Threshold**: Functions with complexity > 17 will be flagged

### Cognitive Complexity (gocognit)

Functions ranked by cognitive complexity:

1. `linter.createRules`: **80** (suppressed - factory function)

**Threshold**: Functions with complexity > 85 will be flagged

## Suppressed Functions

### `linter.createRules` (internal/linter/linter.go:77)
- **Cyclomatic**: 37
- **Cognitive**: 80
- **Reason**: Factory function that instantiates multiple rule types based on configuration. The complexity is inherent to its purpose as a conditional factory and is acceptable.
- **Suppression**: `//nolint:gocognit,gocyclo`

## Recommendations

Most functions in the codebase have reasonable complexity. The highest non-suppressed complexity is 16 (graph.Build), which is acceptable for a function that:
- Parses multiple file types
- Builds an import dependency graph
- Maps files to architectural layers

### Future Refactoring Candidates

If complexity becomes a concern, consider refactoring:

1. **graph.Build** (complexity 16): Could extract language-specific parsing logic
2. **file_existence.checkRequirement** (complexity 15): Could split requirement parsing and validation
3. **walker.Walk** (complexity 14): Could extract directory traversal logic

However, these functions are currently readable and maintainable despite their complexity scores.

## Continuous Monitoring

The `.golangci.yml` configuration ensures:
- New functions with complexity > 17 (cyclomatic) will be flagged
- New functions with complexity > 85 (cognitive) will be flagged
- Test files are excluded from complexity checks
- Existing complex functions are documented and justified

Run `golangci-lint run ./...` to check complexity metrics.
