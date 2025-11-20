# Tree-sitter Parser

## Overview

This package provides tree-sitter based parsing for multiple languages.

## Components

- **parser.go**: Multi-language AST parser using tree-sitter
- **imports.go**: Import extraction from source files
- **exports.go**: Export detection for public symbols
- **metrics.go**: Code metrics calculation from AST

## Supported Languages

- Go
- TypeScript/JavaScript
- Python
- Java
- Rust

## Features

- Fast, incremental parsing
- Language-agnostic API
- Import/export analysis
- Cognitive complexity calculation
- Halstead metrics

## Usage

```go
parser := treesitter.NewParser("go")
ast, err := parser.Parse(source)
imports := parser.ExtractImports(source)
exports := parser.ExtractExports(source)
```
