# Metrics Package

This package implements evidence-based software complexity metrics for Go code analysis.

## Overview

The metrics package provides analyzers for calculating various code complexity metrics that have been empirically validated to correlate with code comprehension and maintenance difficulty.

## Implemented Metrics

### Cognitive Complexity

**File:** `cognitive_complexity.go`

Implementation based on Schnappinger et al. (2020) research showing r=0.54 correlation with comprehension time.

**Key Features:**
- Penalizes nesting (human cognitive load increases with nesting)
- Ignores shorthand structures (they improve readability)
- Based on human assessment rather than mathematical models
- Superior to Cyclomatic Complexity for measuring understandability

**Usage:**
```go
analyzer := NewCognitiveComplexityAnalyzer()
metrics := analyzer.AnalyzeFunction(funcDecl)
```

### Halstead Complexity Measures

**File:** `halstead.go`

Implementation based on Halstead (1977) and Scalabrino et al. (2022) EEG study showing rs=0.901 correlation with cognitive load.

**Key Features:**
- Volume (V): Information content in bits
- Difficulty (D): How difficult to write/understand
- Effort (E): Mental effort required (D × V)
- >90% correlation with measured brain activity during code comprehension
- Captures "data complexity" (vocabulary, operators, operands)

**Usage:**
```go
analyzer := NewHalsteadAnalyzer()
metrics := analyzer.AnalyzeFunction(funcDecl)
```

## Common Types

### FunctionMetric
Represents metrics for a single function:
- `Name`: Function name
- `StartLine`, `EndLine`: Position in file
- `Value`: Primary metric value
- `Complexity`: Complexity score (for cognitive complexity)

### FileMetrics
Aggregated metrics for an entire file:
- `Functions`: Per-function metrics
- `FileLevel`: Aggregated statistics (total, average, max, count)

## Testing

All metrics implementations have comprehensive test coverage:
- `cognitive_complexity_test.go`: Tests various control flow patterns
- `halstead_test.go`: Tests operator/operand counting

Run tests with:
```bash
go test ./internal/metrics/...
```

## References

- Schnappinger et al. (2020): Cognitive complexity correlation study
- Halstead, M. (1977): "Elements of Software Science"
- Scalabrino et al. (2022): EEG study on Halstead metrics
