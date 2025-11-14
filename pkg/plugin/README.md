# plugin

⬆️ **[Parent Directory](../../README.md)**

## Overview

The `plugin` package provides an extensible plugin architecture for structurelint, allowing developers to add support for custom languages, parsers, and linting rules without modifying the core codebase.

## Features

- **Custom Language Parsers**: Add support for any programming language
- **Custom Linting Rules**: Define project-specific or language-specific rules
- **Plugin Registry**: Centralized registration and discovery of plugins
- **Type-Safe Interfaces**: Well-defined contracts for plugin implementations
- **Thread-Safe**: Safe for concurrent access

## Plugin Types

### 1. Parser Plugins

Parser plugins enable structurelint to understand new programming languages by extracting import/export information.

```go
type Parser interface {
    Name() string
    Version() string
    SupportedExtensions() []string
    ParseImports(filePath string) ([]Import, error)
    ParseExports(filePath string) ([]Export, error)
}
```

### 2. Rule Plugins

Rule plugins add custom linting rules that can validate project-specific requirements.

```go
type RulePlugin interface {
    Name() string
    Version() string
    Check(files []FileInfo) ([]Violation, error)
}
```

## Quick Start

### Creating a Parser Plugin

Here's a complete example of a Rust parser plugin:

```go
package main

import (
    "bufio"
    "os"
    "regexp"
    "strings"

    "github.com/structurelint/structurelint/pkg/plugin"
)

type RustParser struct{}

func (p *RustParser) Name() string {
    return "rust-parser"
}

func (p *RustParser) Version() string {
    return "1.0.0"
}

func (p *RustParser) SupportedExtensions() []string {
    return []string{".rs"}
}

func (p *RustParser) ParseImports(filePath string) ([]plugin.Import, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var imports []plugin.Import
    scanner := bufio.NewScanner(file)
    lineNum := 0

    useRegex := regexp.MustCompile(`^\s*use\s+([^;]+);`)

    for scanner.Scan() {
        lineNum++
        line := scanner.Text()

        if matches := useRegex.FindStringSubmatch(line); matches != nil {
            importPath := strings.TrimSpace(matches[1])

            isRelative := strings.HasPrefix(importPath, "super::") ||
                         strings.HasPrefix(importPath, "self::")

            imports = append(imports, plugin.Import{
                ImportPath: importPath,
                IsRelative: isRelative,
                Line:       lineNum,
            })
        }
    }

    return imports, scanner.Err()
}

func (p *RustParser) ParseExports(filePath string) ([]plugin.Export, error) {
    // Implementation similar to ParseImports
    // Extract public functions, structs, traits, etc.
    return []plugin.Export{}, nil
}

// Register the plugin
func init() {
    plugin.RegisterParser(&RustParser{})
}
```

### Creating a Rule Plugin

Example of a custom rule that enforces specific naming patterns:

```go
package main

import (
    "fmt"
    "path/filepath"
    "strings"

    "github.com/structurelint/structurelint/pkg/plugin"
)

type DatabaseMigrationRule struct{}

func (r *DatabaseMigrationRule) Name() string {
    return "database-migration-naming"
}

func (r *DatabaseMigrationRule) Version() string {
    return "1.0.0"
}

func (r *DatabaseMigrationRule) Check(files []plugin.FileInfo) ([]plugin.Violation, error) {
    var violations []plugin.Violation

    for _, file := range files {
        // Check if file is in migrations directory
        if !strings.Contains(file.Path, "migrations/") {
            continue
        }

        // Enforce naming: YYYYMMDD_description.sql
        basename := filepath.Base(file.Path)
        if !regexp.MustCompile(`^\d{8}_\w+\.sql$`).MatchString(basename) {
            violations = append(violations, plugin.Violation{
                Rule:    r.Name(),
                Path:    file.Path,
                Message: fmt.Sprintf("migration file must follow format YYYYMMDD_description.sql, got: %s", basename),
            })
        }
    }

    return violations, nil
}

// Register the plugin
func init() {
    plugin.RegisterRule(&DatabaseMigrationRule{})
}
```

## Using Plugins

### Method 1: Import in Custom CLI

Create a custom version of structurelint that imports your plugins:

```go
// cmd/my-structurelint/main.go
package main

import (
    _ "github.com/yourorg/structurelint-rust-plugin"  // Import for side effects
    "github.com/structurelint/structurelint/cmd/structurelint"
)

func main() {
    structurelint.Execute()
}
```

Build and use your custom version:

```bash
go build -o my-structurelint cmd/my-structurelint/main.go
./my-structurelint .
```

### Method 2: Programmatic Registration

For library usage, register plugins programmatically:

```go
package main

import (
    "fmt"

    "github.com/structurelint/structurelint/pkg/api"
    "github.com/structurelint/structurelint/pkg/plugin"
    "github.com/yourorg/rust-parser"
)

func main() {
    // Register custom parser
    plugin.RegisterParser(&rustparser.RustParser{})

    // Use structurelint with custom parser
    linter := api.NewLinter()
    violations, err := linter.Lint("./my-rust-project")

    if err != nil {
        panic(err)
    }

    for _, v := range violations {
        fmt.Printf("[%s] %s: %s\n", v.Rule, v.Path, v.Message)
    }
}
```

## Plugin Discovery

### Listing Registered Plugins

```go
package main

import (
    "fmt"

    "github.com/structurelint/structurelint/pkg/plugin"
)

func main() {
    registry := plugin.GetRegistry()

    // List all parsers
    fmt.Println("Registered Parsers:")
    for _, parser := range registry.ListParsers() {
        fmt.Printf("  - %s v%s (extensions: %v)\n",
            parser.Name, parser.Version, parser.Extensions)
    }

    // List all custom rules
    fmt.Println("\nRegistered Rules:")
    for _, rule := range registry.ListRules() {
        fmt.Printf("  - %s v%s\n", rule.Name, rule.Version)
    }
}
```

## API Reference

### Parser Interface

#### `Name() string`
Returns a unique identifier for the parser (e.g., "rust-parser").

#### `Version() string`
Returns the semantic version of the plugin (e.g., "1.0.0").

#### `SupportedExtensions() []string`
Returns file extensions this parser handles (e.g., `[".rs", ".toml"]`).

#### `ParseImports(filePath string) ([]Import, error)`
Extracts import/dependency statements from a file.

**Returns**: List of `Import` structs:
```go
type Import struct {
    ImportPath string // The imported module/file path
    IsRelative bool   // Whether this is a relative import
    Line       int    // Line number where import appears
    Symbol     string // Specific symbol imported (optional)
}
```

#### `ParseExports(filePath string) ([]Export, error)`
Extracts exported symbols from a file.

**Returns**: List of `Export` structs:
```go
type Export struct {
    Name string // Name of the exported symbol
    Kind string // Kind: "function", "class", "variable", "type", etc.
    Line int    // Line number where export is defined
}
```

### RulePlugin Interface

#### `Name() string`
Returns a unique identifier for the rule (e.g., "database-migration-naming").

#### `Version() string`
Returns the semantic version of the plugin.

#### `Check(files []FileInfo) ([]Violation, error)`
Validates files and returns violations.

**Parameters**:
- `files`: List of file information to check

**Returns**: List of `Violation` structs:
```go
type Violation struct {
    Rule    string // Name of the rule that was violated
    Path    string // Path to the file with the violation
    Message string // Human-readable violation message
    Line    int    // Line number (0 if not applicable)
    Column  int    // Column number (0 if not applicable)
}
```

### Registry Methods

#### `RegisterParser(parser Parser) error`
Registers a custom language parser. Returns error if a parser is already registered for any of the extensions.

#### `RegisterRule(rule RulePlugin) error`
Registers a custom linting rule. Returns error if a rule with the same name is already registered.

#### `GetParser(ext string) (Parser, bool)`
Retrieves a parser for a given file extension. Returns `(nil, false)` if not found.

#### `GetParserForFile(filePath string) (Parser, bool)`
Retrieves a parser for a given file path by extracting the extension.

#### `GetRule(name string) (RulePlugin, bool)`
Retrieves a custom rule by name. Returns `(nil, false)` if not found.

#### `ListParsers() []ParserInfo`
Returns information about all registered parsers.

#### `ListRules() []RuleInfo`
Returns information about all registered custom rules.

## Example Plugins

### Ruby Parser

```go
type RubyParser struct{}

func (p *RubyParser) SupportedExtensions() []string {
    return []string{".rb"}
}

func (p *RubyParser) ParseImports(filePath string) ([]plugin.Import, error) {
    // Parse: require 'foo'
    // Parse: require_relative 'foo'
    // Parse: require File.expand_path('foo')
    // ...
}
```

### PHP Parser

```go
type PHPParser struct{}

func (p *PHPParser) SupportedExtensions() []string {
    return []string{".php"}
}

func (p *PHPParser) ParseImports(filePath string) ([]plugin.Import, error) {
    // Parse: use Foo\Bar;
    // Parse: require 'foo.php';
    // Parse: include 'bar.php';
    // ...
}
```

### API Version Rule

```go
type APIVersionRule struct{}

func (r *APIVersionRule) Check(files []plugin.FileInfo) ([]plugin.Violation, error) {
    // Enforce that API files are versioned: api/v1/*, api/v2/*
    // Reject unversioned API files: api/*.go
}
```

## Best Practices

### 1. Semantic Versioning
Use semantic versioning for your plugins:
- **Major**: Breaking changes to the plugin interface
- **Minor**: New features, backward compatible
- **Patch**: Bug fixes only

### 2. Error Handling
Always handle errors gracefully and return descriptive error messages:

```go
func (p *MyParser) ParseImports(filePath string) ([]plugin.Import, error) {
    file, err := os.Open(filePath)
    if err != nil {
        return nil, fmt.Errorf("failed to open file: %w", err)
    }
    defer file.Close()

    // Parse logic...
}
```

### 3. Testing
Write comprehensive tests for your plugins:

```go
func TestRustParser_ParseImports(t *testing.T) {
    parser := &RustParser{}

    imports, err := parser.ParseImports("testdata/sample.rs")
    if err != nil {
        t.Fatalf("ParseImports failed: %v", err)
    }

    if len(imports) != 3 {
        t.Errorf("Expected 3 imports, got %d", len(imports))
    }
}
```

### 4. Documentation
Document your plugin's behavior, supported language features, and limitations:

```go
// RustParser is a plugin that parses Rust source files.
//
// Supported features:
// - Basic use statements: use foo::bar;
// - Nested imports: use foo::{bar, baz};
// - Relative imports: use super::foo;
//
// Limitations:
// - Does not parse inline modules (mod { ... })
// - Does not handle conditional compilation (#[cfg])
type RustParser struct{}
```

## Performance Considerations

- **Lazy Parsing**: Only parse files when needed
- **Caching**: Cache parse results for large codebases
- **Concurrency**: Ensure your plugin is thread-safe (use mutex if needed)

```go
type CachedParser struct {
    cache map[string][]plugin.Import
    mu    sync.RWMutex
}

func (p *CachedParser) ParseImports(filePath string) ([]plugin.Import, error) {
    p.mu.RLock()
    if cached, ok := p.cache[filePath]; ok {
        p.mu.RUnlock()
        return cached, nil
    }
    p.mu.RUnlock()

    // Parse and cache...
}
```

## Community Plugins

### Publishing Your Plugin

1. Create a GitHub repository: `structurelint-<language>-plugin`
2. Add comprehensive README with usage examples
3. Tag releases with semantic versions
4. Submit to the structurelint plugin registry (coming soon)

### Finding Plugins

- [structurelint-rust-plugin](https://github.com/example/structurelint-rust-plugin) - Rust language support
- [structurelint-ruby-plugin](https://github.com/example/structurelint-ruby-plugin) - Ruby language support
- [structurelint-php-plugin](https://github.com/example/structurelint-php-plugin) - PHP language support

*(Note: These are example links - replace with actual community plugins)*

## Troubleshooting

### Plugin Not Recognized

**Problem**: File extensions not being parsed by your plugin.

**Solution**: Ensure the plugin is registered before linting:
```go
func init() {
    plugin.RegisterParser(&MyParser{})
}
```

### Parser Conflicts

**Problem**: `parser already registered for extension .rs`

**Solution**: Only register one parser per extension. Check if another plugin already handles that extension:

```go
registry := plugin.GetRegistry()
if parser, exists := registry.GetParser(".rs"); exists {
    fmt.Printf("Extension .rs already handled by: %s\n", parser.Name())
}
```

### Rule Not Running

**Problem**: Custom rule doesn't appear in violations.

**Solution**: Ensure the rule is registered and configure it in `.structurelint.yml`:

```yaml
rules:
  database-migration-naming: true
```

## See Also

- [Public API Documentation](../api/README.md) - Using structurelint programmatically
- [Parser Documentation](../../internal/parser/README.md) - Built-in parser implementation
- [Plugin Examples](./examples/) - Complete plugin examples
