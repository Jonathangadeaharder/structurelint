# Structurelint Presets

## Overview

Shareable configuration presets for common project types.

## Available Presets

### go-standard.yml

Sensible defaults for Go projects:
- Max depth: 5 levels
- Max files per directory: 15
- Snake_case file naming
- Adjacent test pattern
- Required files: README.md, go.mod, .gitignore

**Usage:**
```yaml
# .structurelint.yml
extends: ./examples/presets/go-standard.yml

# Override specific rules as needed
rules:
  max-depth: 6
```

### typescript-react.yml

Configuration for TypeScript/React projects with feature-sliced architecture:
- Max depth: 6 levels
- Max files per directory: 20
- PascalCase for components, camelCase for utilities
- Adjacent test pattern for .ts/.tsx files
- Layer boundaries: app → features → shared
- Required files: README.md, package.json, tsconfig.json

**Usage:**
```yaml
# .structurelint.yml
extends: ./examples/presets/typescript-react.yml

# Customize layers for your architecture
layers:
  - name: "app"
    path: "src/app/**"
    dependsOn: ["features", "shared"]
```

## Creating Custom Presets

1. Create a YAML file with your base configuration
2. Set `root: false` (unless you want to stop parent config search)
3. Define rules, layers, and other settings
4. Reference it using `extends` in your project config

```yaml
# my-team-preset.yml
root: false
rules:
  max-depth: 5
  max-files-in-dir: 20
```

## Future: NPM/PyPI Packages

In a future release, presets will be publishable as packages:

```yaml
# Coming soon!
extends: "@structurelint/preset-go-standard"
```
