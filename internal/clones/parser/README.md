# parser

⬆️ **[Parent Directory](../README.md)**

## Overview

The `parser` package handles AST-based normalization of source code into canonical token streams.

## Components

### normalizer.go
Converts source code into normalized token streams using Go's `go/ast` package.

#### Key Functions

**`NewNormalizer() *Normalizer`**
- Creates a new normalizer instance
- Initializes token.FileSet for AST parsing

**`NormalizeFile(filePath string) (*types.FileTokens, error)`**
- Parses a Go source file into an AST
- Extracts and normalizes tokens
- Returns normalized token stream

**`extractTokens(file *ast.File, src []byte) []types.Token`**
- Traverses AST using visitor pattern
- Normalizes identifiers → `_ID_`
- Normalizes literals → `_LIT_`
- Preserves keywords and operators

#### Normalization Rules

| Input | Output |
|-------|--------|
| Identifiers (myVar, calculateSum) | `_ID_` |
| Literals ("hello", 42, 3.14) | `_LIT_` |
| Keywords (if, for, return) | as-is |
| Operators (+, ==, :=) | as-is |
| Punctuation ((, ), {) | as-is |

## Example

```go
// Original code
func calculateSum(a int, b int) int {
    result := a + b
    return result
}

// Normalized token stream
func _ID_ ( _ID_ _ID_ , _ID_ _ID_ ) _ID_ { _ID_ := _ID_ + _ID_ return _ID_ }
```

## Performance

- **Throughput**: ~30MB/s
- **Bottleneck**: CPU-bound (AST traversal)
- **Scalability**: Parallelizable per file
