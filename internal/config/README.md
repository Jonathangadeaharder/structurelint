# config

⬆️ **[Parent Directory](../README.md)**

## Overview

The `config` package handles loading, parsing, and merging `.structurelint.yml` configuration files with cascading/inheritance semantics similar to ESLint.

## Key Features

- **Cascading Configuration**: Searches up the directory tree for config files
- **YAML Parsing**: Robust YAML configuration parsing with validation
- **Config Merging**: Intelligent merging of multiple config files
- **Rule Configuration**: Type-safe access to rule settings

## Main Types

- `Config`: Main configuration structure containing rules, layers, and entrypoints
- Configuration files are merged from root to leaf, with deeper configs overriding parent settings

## Usage

```go
configs, err := config.FindConfigs(path)
merged := config.Merge(configs...)
```
