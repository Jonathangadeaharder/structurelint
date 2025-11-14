# Plugin Examples

⬆️ **[Parent Directory](../README.md)**

## Overview

This directory contains example plugin implementations demonstrating how to extend structurelint with custom language parsers.

## Available Examples

### Rust Parser (`rust_parser.go`)

A complete parser plugin for Rust programming language that extracts:
- **Imports**: `use` statements (including nested imports and relative paths)
- **Exports**: Public functions, structs, enums, constants, and traits

**Supported patterns**:
```rust
use std::collections::HashMap;
use super::foo;
use crate::bar;
use foo::{bar, baz};

pub fn my_function() {}
pub struct MyStruct {}
pub enum MyEnum {}
pub const MY_CONST: i32 = 42;
pub trait MyTrait {}
```

### Ruby Parser (`ruby_parser.go`)

A complete parser plugin for Ruby programming language that extracts:
- **Imports**: `require` and `require_relative` statements
- **Exports**: Classes, modules, and public methods

**Supported patterns**:
```ruby
require 'foo'
require_relative '../bar'

class MyClass
end

module MyModule
end

def my_method
end
```

## Using These Examples

### As Templates

Copy and modify these examples to create parsers for other languages:

```bash
cp rust_parser.go your_language_parser.go
# Edit to implement parsing logic for your language
```

### As Reference

Study the implementation patterns:
1. **Regex-based parsing** for simple syntax extraction
2. **Line-by-line scanning** for performance
3. **Error handling** with descriptive messages
4. **Relative vs absolute imports** detection

### In Your Project

Import and register these example parsers:

```go
package main

import (
    "github.com/structurelint/structurelint/pkg/plugin"
    "github.com/structurelint/structurelint/pkg/plugin/examples"
)

func init() {
    // Register Rust parser
    plugin.RegisterParser(&examples.RustParser{})

    // Register Ruby parser
    plugin.RegisterParser(&examples.RubyParser{})
}
```

## Implementation Guidelines

### 1. Start Simple

Begin with basic pattern matching and expand:

```go
// Start with simple regex
importRegex := regexp.MustCompile(`import\s+(\w+)`)

// Add complexity incrementally
importRegex := regexp.MustCompile(`import\s+(?:{\s*([^}]+)\s*}\s+from\s+)?['"]([^'"]+)['"]`)
```

### 2. Handle Edge Cases

Consider language-specific edge cases:

```go
// Skip comments
if strings.HasPrefix(strings.TrimSpace(line), "//") {
    continue
}

// Skip multi-line comments
if inMultiLineComment {
    if strings.Contains(line, "*/") {
        inMultiLineComment = false
    }
    continue
}
```

### 3. Test Thoroughly

Create test files with various syntax patterns:

```go
func TestRustParser_ParseImports(t *testing.T) {
    parser := &RustParser{}

    testCases := []struct {
        name     string
        file     string
        expected int
    }{
        {"simple use", "testdata/simple.rs", 1},
        {"nested use", "testdata/nested.rs", 3},
        {"relative use", "testdata/relative.rs", 2},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            imports, err := parser.ParseImports(tc.file)
            if err != nil {
                t.Fatalf("ParseImports failed: %v", err)
            }

            if len(imports) != tc.expected {
                t.Errorf("Expected %d imports, got %d", tc.expected, len(imports))
            }
        })
    }
}
```

## Language Support Comparison

| Language | Import Parsing | Export Parsing | Complexity |
|----------|----------------|----------------|------------|
| Rust     | ✅ use statements | ✅ pub items | Medium |
| Ruby     | ✅ require | ✅ classes/modules | Low |
| Python   | Built-in | Built-in | - |
| TypeScript | Built-in | Built-in | - |
| Go       | Built-in | Built-in | - |

## Creating New Parsers

### Step 1: Define the Parser

```go
type MyLanguageParser struct{}

func (p *MyLanguageParser) Name() string {
    return "mylang-parser"
}

func (p *MyLanguageParser) Version() string {
    return "1.0.0"
}

func (p *MyLanguageParser) SupportedExtensions() []string {
    return []string{".mylang"}
}
```

### Step 2: Implement ParseImports

```go
func (p *MyLanguageParser) ParseImports(filePath string) ([]plugin.Import, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var imports []plugin.Import
    scanner := bufio.NewScanner(file)
    lineNum := 0

    // Define regex for your language's import syntax
    importRegex := regexp.MustCompile(`your_import_pattern_here`)

    for scanner.Scan() {
        lineNum++
        line := scanner.Text()

        if matches := importRegex.FindStringSubmatch(line); matches != nil {
            imports = append(imports, plugin.Import{
                ImportPath: matches[1], // Adjust based on your regex groups
                IsRelative: /* determine if relative */,
                Line:       lineNum,
            })
        }
    }

    return imports, scanner.Err()
}
```

### Step 3: Implement ParseExports

```go
func (p *MyLanguageParser) ParseExports(filePath string) ([]plugin.Export, error) {
    // Similar pattern to ParseImports
    // Extract exported functions, classes, variables, etc.
    return []plugin.Export{}, nil
}
```

### Step 4: Test and Document

Create comprehensive tests and document supported syntax patterns.

## Advanced Parsing Techniques

### AST-Based Parsing

For complex languages, consider using an AST parser:

```go
import "go/parser"
import "go/ast"

func (p *AdvancedParser) ParseImports(filePath string) ([]plugin.Import, error) {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filePath, nil, parser.ImportsOnly)
    if err != nil {
        return nil, err
    }

    // Walk the AST to extract imports
    var imports []plugin.Import
    for _, imp := range node.Imports {
        imports = append(imports, plugin.Import{
            ImportPath: imp.Path.Value,
            // ...
        })
    }

    return imports, nil
}
```

### Incremental Parsing

For large files, consider incremental parsing:

```go
func (p *IncrementalParser) ParseImports(filePath string) ([]plugin.Import, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    reader := bufio.NewReader(file)
    var imports []plugin.Import

    // Stop parsing after imports section
    for {
        line, err := reader.ReadString('\n')
        if err != nil {
            break
        }

        // Stop at first non-import line (language-specific)
        if !isImportLine(line) && len(imports) > 0 {
            break
        }

        // Parse import...
    }

    return imports, nil
}
```

## Contributing

Have you created a parser for a new language? Consider:

1. **Publishing as a separate package**: `structurelint-<language>-plugin`
2. **Submitting a PR**: Add to official examples
3. **Sharing in discussions**: Help others learn from your implementation

## See Also

- [Plugin Architecture Documentation](../README.md)
- [Parser Interface Specification](../plugin.go)
- [Built-in Parser Implementation](../../../internal/parser/README.md)
