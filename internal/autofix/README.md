# Auto-Fix Framework

## Overview

This package provides the auto-fix framework for structurelint, allowing automatic correction of violations.

## Components

- **engine.go**: Core auto-fix engine with action-based architecture
- **file_location_fixer.go**: Fixer for file location violations with import rewriting

## Key Interfaces

### Fix
Represents a fix for a violation with actions, confidence level, and safety flag.

### Action
Interface for individual fix actions (Apply, Describe, Revert).

### Fixer
Interface for generating fixes for specific violation types.

## Usage

The auto-fix framework is used by the `structurelint fix` command to automatically correct violations.

```go
engine := autofix.NewEngine()
engine.RegisterFixer(&autofix.FileLocationFixer{})
fixes, err := engine.GenerateFixes(violations, files)
```
