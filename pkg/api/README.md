# api

⬆️ **[Parent Directory](../../README.md)**

## Overview

The `api` package provides a stable public API for using structurelint programmatically in Go applications. This allows you to integrate structurelint into your own tools, CI/CD pipelines, or testing frameworks.

## Features

- **Programmatic Linting**: Run structurelint from Go code
- **Configuration Builder**: Fluent API for building configurations
- **Architectural Rules**: ArchUnit-style fluent interface for defining architectural constraints
- **Backward Compatibility**: Stable public API that won't break between minor versions

## Installation

```bash
go get github.com/structurelint/structurelint/pkg/api
```

## Basic Usage

### Simple Linting

```go
package main

import (
    "fmt"
    "github.com/structurelint/structurelint/pkg/api"
)

func main() {
    // Create a new linter
    linter := api.NewLinter()

    // Run linting on current directory
    violations, err := linter.Lint(".")
    if err != nil {
        panic(err)
    }

    // Print violations
    for _, v := range violations {
        fmt.Printf("[%s] %s: %s\n", v.Rule, v.Path, v.Message)
    }
}
```

### Custom Configuration

```go
// Build configuration programmatically
cfg := api.NewConfig().
    EnableRule("no-empty-files", true).
    EnableRule("naming-convention", map[string]string{
        "**/*.go": "snake_case",
    }).
    AddExclude("vendor/**").
    AddExclude("node_modules/**").
    AddLayer("domain", "internal/domain/**").
    AddLayer("infrastructure", "internal/infrastructure/**")

// Create linter with config
linter := api.NewLinter().WithConfig(cfg)
violations, err := linter.Lint("./myproject")
```

### Production Mode

```go
// Enable production mode to exclude test files
linter := api.NewLinter().WithProductionMode(true)
violations, err := linter.Lint(".")

// This will only report issues in production code
for _, v := range violations {
    if v.Rule == "disallow-unused-exports" {
        fmt.Printf("Dead code in production: %s\n", v.Path)
    }
}
```

### Load Configuration from File

```go
// Load existing .structurelint.yml
cfg, err := api.LoadConfig(".")
if err != nil {
    panic(err)
}

linter := api.NewLinter().WithConfig(cfg)
violations, err := linter.Lint(".")
```

## Fluent Architectural Rules API

Inspired by [ArchUnit](https://www.archunit.org/), the fluent API allows you to define architectural constraints in a readable, declarative way.

### Basic Architectural Rule

```go
// Define rule: Domain layer should not depend on Infrastructure
rule := api.NewArchRule().
    That(api.Layers().Matching("domain")).
    ShouldNot().
    DependOn(api.Layers().Matching("infrastructure"))

// Check violations
violations := rule.Check(files)
```

### File-Based Rules

```go
// All Go files should follow snake_case
rule := api.NewArchRule().
    That(api.Files().Matching("**/*.go")).
    Should().
    HaveNamingConvention("snake_case")
```

### Empty File Constraints

```go
// Test files should not be empty
rule := api.NewArchRule().
    That(api.Files().Matching("**/*_test.go")).
    ShouldNot().
    BeEmpty()
```

### Complex Architectural Tests

```go
package myproject_test

import (
    "testing"
    "github.com/structurelint/structurelint/pkg/api"
)

func TestArchitecture_LayerDependencies(t *testing.T) {
    // Load project
    linter := api.NewLinter()

    // Define architectural rules
    rules := []*api.ArchRule{
        // Rule 1: Domain should not depend on Infrastructure
        api.NewArchRule().
            That(api.Layers().Matching("domain")).
            ShouldNot().
            DependOn(api.Layers().Matching("infrastructure")),

        // Rule 2: Presentation should not depend on Data
        api.NewArchRule().
            That(api.Layers().Matching("presentation")).
            ShouldNot().
            DependOn(api.Layers().Matching("data")),

        // Rule 3: All use cases should be in application layer
        api.NewArchRule().
            That(api.Files().Matching("**/usecase/**")).
            Should().
            HaveNamingConvention("snake_case"),
    }

    // Validate all rules
    for _, rule := range rules {
        violations := rule.Check(files)
        if len(violations) > 0 {
            t.Errorf("Architectural rule violated: %v", violations)
        }
    }
}
```

## API Reference

### Core Types

#### `Linter`

Main entry point for programmatic linting.

```go
type Linter struct { ... }

func NewLinter() *Linter
func (l *Linter) WithConfig(cfg *Config) *Linter
func (l *Linter) WithProductionMode(enabled bool) *Linter
func (l *Linter) Lint(path string) ([]Violation, error)
```

#### `Violation`

Represents a linting rule violation.

```go
type Violation struct {
    Rule    string // Name of the rule that was violated
    Path    string // Path to the file with the violation
    Message string // Human-readable violation message
}
```

#### `Config`

Configuration for the linter.

```go
type Config struct { ... }

func NewConfig() *Config
func LoadConfig(path string) (*Config, error)
func (c *Config) EnableRule(name string, ruleConfig interface{}) *Config
func (c *Config) AddExclude(pattern string) *Config
func (c *Config) AddLayer(name string, pattern string) *Config
```

### Fluent API Types

#### `ArchRule`

Fluent builder for architectural rules.

```go
type ArchRule struct { ... }

func NewArchRule() *ArchRule
func (a *ArchRule) That(selector LayerSelector) *ArchRuleBuilder
func (a *ArchRule) WithGraph(g *graph.ImportGraph) *ArchRule
func (a *ArchRule) WithLayers(layers []config.Layer) *ArchRule
func (a *ArchRule) Check(files []walker.FileInfo) []Violation
```

#### `LayerSelector`

Selects which layers or files a rule applies to.

```go
func Layers() *LayerSelectorBuilder
func Files() *FileSelectorBuilder

func (b *LayerSelectorBuilder) Matching(pattern string) LayerSelector
func (b *FileSelectorBuilder) Matching(pattern string) LayerSelector
```

#### `ConstraintBuilder`

Builds constraints for architectural rules.

```go
func (b *ArchRuleBuilder) Should() *ConstraintBuilder
func (b *ArchRuleBuilder) ShouldNot() *ConstraintBuilder

func (cb *ConstraintBuilder) DependOn(target LayerSelector) *ArchRule
func (cb *ConstraintBuilder) BeEmpty() *ArchRule
func (cb *ConstraintBuilder) HaveNamingConvention(convention string) *ArchRule
```

### Utility Functions

```go
// Get information about all available rules
func AvailableRules() []RuleInfo

// Check if a rule supports automated fixing
func IsFixable(ruleName string) bool

// Generate fixes for violations (coming soon)
func GenerateFixes(path string, dryRun bool) ([]Fix, error)
```

## Rule Information

```go
type RuleInfo struct {
    Name        string
    Description string
    Fixable     bool
}
```

### Available Rules

| Rule | Description | Fixable |
|------|-------------|---------|
| `no-empty-files` | Disallow empty files | No |
| `disallowed-patterns` | Disallow specific path patterns | No |
| `required-files` | Require specific files to exist | No |
| `naming-convention` | Enforce file/directory naming conventions | Yes |
| `test-adjacency` | Require tests adjacent to implementation | No |
| `disallow-unused-exports` | Disallow unused exports | Yes |
| `file-hash` | Validate file contents via SHA256 hash | No |
| `granular-dependencies` | Fine-grained module dependency validation | No |

## Integration Examples

### In CI/CD Pipeline

```go
package main

import (
    "fmt"
    "os"
    "github.com/structurelint/structurelint/pkg/api"
)

func main() {
    linter := api.NewLinter()
    violations, err := linter.Lint(".")

    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }

    if len(violations) > 0 {
        for _, v := range violations {
            fmt.Printf("[%s] %s: %s\n", v.Rule, v.Path, v.Message)
        }
        os.Exit(1)
    }

    fmt.Println("✓ All architectural rules passed")
}
```

### In Unit Tests

```go
package myproject_test

import (
    "testing"
    "github.com/structurelint/structurelint/pkg/api"
)

func TestProjectStructure(t *testing.T) {
    linter := api.NewLinter()
    violations, err := linter.Lint(".")

    if err != nil {
        t.Fatalf("Linting failed: %v", err)
    }

    if len(violations) > 0 {
        t.Errorf("Found %d structural violations:", len(violations))
        for _, v := range violations {
            t.Logf("  [%s] %s: %s", v.Rule, v.Path, v.Message)
        }
    }
}
```

### Custom Tool Integration

```go
package main

import (
    "fmt"
    "github.com/structurelint/structurelint/pkg/api"
)

func analyzeProject(projectPath string) error {
    // Custom configuration for your tool
    cfg := api.NewConfig().
        EnableRule("no-empty-files", true).
        AddLayer("core", "src/core/**").
        AddLayer("ui", "src/ui/**")

    linter := api.NewLinter().WithConfig(cfg)
    violations, err := linter.Lint(projectPath)

    if err != nil {
        return err
    }

    // Custom violation handling
    for _, v := range violations {
        fmt.Printf("⚠️  %s in %s\n", v.Message, v.Path)
    }

    return nil
}
```

## Stability Guarantee

The `pkg/api` package follows semantic versioning:
- **Major version**: Breaking API changes
- **Minor version**: New features, backward compatible
- **Patch version**: Bug fixes only

You can safely upgrade between minor versions without code changes.

## See Also

- [Main README](../../README.md) - Project overview
- [CLI Documentation](../../cmd/structurelint/README.md) - Command-line interface
- [Configuration Guide](../../docs/configuration.md) - YAML configuration reference
- [Rules Documentation](../../docs/rules.md) - Available linting rules
