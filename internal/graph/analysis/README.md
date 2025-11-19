# Graph Analysis

## Overview

This package provides dependency graph analysis algorithms.

## Components

- **cycles.go**: Cycle detection in dependency graphs

## Features

### Cycle Detection
- Identifies circular dependencies between modules
- Uses Tarjan's algorithm for strongly connected components
- Reports all cycles with full paths

## Usage

```go
analyzer := analysis.NewCycleDetector()
cycles := analyzer.DetectCycles(graph)
for _, cycle := range cycles {
    fmt.Println("Cycle:", cycle.Path)
}
```

## Algorithms

- **Tarjan's SCC**: O(V + E) cycle detection
- **Path reconstruction**: Full circular dependency paths
- **Minimal cycle basis**: Finds fundamental cycles
