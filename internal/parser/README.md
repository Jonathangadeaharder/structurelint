# parser

⬆️ **[Parent Directory](../README.md)**

## Overview

The `parser` package handles source code parsing to extract import and export statements from various programming languages.

## Supported Languages

- **TypeScript/JavaScript**: ES6 imports, CommonJS require, exports
- **Go**: Import declarations, exported symbols
- **Python**: Import statements, `__all__` exports

## Key Features

- **Import Extraction**: Identifies all import/require statements
- **Export Detection**: Finds exported symbols and declarations
- **Path Resolution**: Resolves relative and absolute import paths
- **Multi-Language**: Unified interface for different language parsers

## Main Functions

- `ParseTypeScriptImports(content string) []string`
- `ParseGoImports(content string) []string`
- `ParsePythonImports(content string) []string`
- `ParseTypeScriptExports(content string) []string`
- `ParseGoExports(content string) []string`
- `ParsePythonExports(content string) []string`
- `ResolveImportPath(importPath, fromFile string) string`

## Implementation

Uses regex-based parsing for simplicity and performance. For production use cases requiring more accuracy, could be extended to use proper AST parsers.
