# Predicate System

## Overview

This package provides a predicate-based rule system for custom constraints.

## Components

- **predicate.go**: Predicate evaluation engine

## Supported Predicates

- `in-layer(name)`: Check if file is in a specific layer
- `depends-on(pattern)`: Check if file depends on pattern
- `has-import(pattern)`: Check for specific imports
- `file-matches(pattern)`: Check file path matches pattern
- `all(...)`: All predicates must match
- `any(...)`: Any predicate must match
- `not(...)`: Negate predicate

## Usage

Predicates allow expressing complex architectural constraints:

```yaml
rules:
  custom-rule:
    type: predicate
    predicate:
      all:
        - in-layer: domain
        - not:
            depends-on: "*infrastructure*"
    message: "Domain must be pure"
```

## Implementation

Predicates are evaluated recursively with short-circuit logic for performance.
