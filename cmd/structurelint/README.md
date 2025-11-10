# structurelint CLI

This directory contains the main entry point for the structurelint command-line tool.

## Overview

The structurelint CLI is a project structure and architecture linter that helps enforce consistency and best practices in your codebase.

## Usage

```bash
# Lint current directory
structurelint

# Lint specific path
structurelint ./src

# Initialize configuration for your project
structurelint --init

# Show version
structurelint --version

# Show help
structurelint --help
```

## Features

- **Project Analysis**: Automatically detects project structure and patterns
- **Configuration Generation**: Creates smart default configurations based on your project
- **Flexible Linting**: Enforces rules for file organization, naming conventions, and architecture

## Building

```bash
go build -o structurelint ./cmd/structurelint
```

## Documentation

For detailed documentation, see the main repository README and docs directory.
