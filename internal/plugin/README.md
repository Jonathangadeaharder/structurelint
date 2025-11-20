# Plugin System

## Overview

This package provides extensibility through plugins for custom rules and analysis.

## Components

- **semantic_clone.go**: Semantic clone detection plugin

## Plugin Architecture

The plugin system allows extending structurelint with custom:
- Rule implementations
- Analyzers
- Fixers
- Formatters

## Semantic Clone Detection

Detects code clones using:
- Abstract Syntax Tree (AST) comparison
- Semantic similarity analysis
- Type-aware matching

## Usage

Plugins can be registered at runtime to extend structurelint's capabilities without modifying core code.

```go
plugin := plugin.NewSemanticCloneDetector()
violations := plugin.Analyze(files)
```
