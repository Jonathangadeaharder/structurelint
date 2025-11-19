# Scaffolding Generator

## Overview

This package provides code scaffolding and template generation for structurelint.

## Components

- **generator.go**: Core template engine for generating code from templates
- **templates.go**: Built-in templates for Go, TypeScript, and Python

## Features

- Multi-language support (Go, TypeScript, Python)
- Variable substitution with automatic case conversion (PascalCase, camelCase, snake_case, kebab-case)
- Package detection from go.mod, package.json, etc.
- 10 built-in templates for common patterns

## Built-in Templates

### Go
- Service, Repository, Handler, Model

### TypeScript
- Service, Controller, Model

### Python
- Service, Repository, Model

## Usage

```go
gen := scaffold.NewGenerator()
vars := scaffold.Variables{
    Name: "UserService",
    Package: "github.com/example/app",
}
err := gen.Generate("service", "go", vars, ".")
```
