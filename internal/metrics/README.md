# Metrics Package

Evidence-based software quality metrics implementation.

## Overview

This package provides analyzers for calculating software quality metrics with scientific backing:

- **Cognitive Complexity**: Measures code understandability based on human cognitive load
- **Halstead Effort**: Measures mental effort required to comprehend code

## Components

### Cognitive Complexity Analyzer

Implementation based on:
- Schnappinger et al. (2020) Meta-Analysis: r=0.54 correlation with comprehension time
- Superior to Cyclomatic Complexity for measuring understandability

Key differences from Cyclomatic Complexity:
1. Penalizes nesting (human cognitive load increases with nesting)
2. Ignores shorthand structures (they improve readability)
3. Based on human assessment, not mathematical models

### Halstead Analyzer

Implementation based on:
- Halstead, M. (1977) "Elements of Software Science"
- Scalabrino et al. (2022) EEG Study: rs=0.901 correlation with cognitive load

Metrics:
- Volume (V): Information content in bits
- Difficulty (D): How difficult to write/understand
- Effort (E): Mental effort required (D × V)

## Usage

```go
import "github.com/structurelint/structurelint/internal/metrics"

// Cognitive Complexity
analyzer := metrics.NewCognitiveComplexityAnalyzer()
fileMetrics := analyzer.AnalyzeFile(astNode)

// Halstead Effort
halsteadAnalyzer := metrics.NewHalsteadAnalyzer()
halsteadMetrics := halsteadAnalyzer.AnalyzeFile(astNode)
```

## Scientific References

1. Schnappinger, M., Oertelt, A., Fietzke, A., & Pretschner, A. (2020). Human-Level Ordinal Comparison of Software Comprehensibility. In ESEM.
2. Scalabrino, S., Bavota, G., Russo, B., Oliveto, R., & Di Penta, M. (2022). Listening to the Developers: A Fine-Grained Analysis of Code Comprehension. IEEE TSE.
