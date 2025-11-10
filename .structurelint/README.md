# .structurelint

⬆️ **[Parent Directory](../README.md)**

## Overview

The `.structurelint/` directory contains project-specific structurelint configuration and templates.

## Contents

| Directory | Purpose |
|-----------|---------|
| [`templates/`](templates/README.md) | File content templates for validation |

## Purpose

This directory houses:
- **Templates**: Define required structure for documentation files
- **Local Configuration**: Project-specific linter settings
- **Custom Rules**: Project-specific rule definitions (future)

## Usage

Templates in `templates/` are referenced by the `file-content` rule in `.structurelint.yml`:

```yaml
rules:
  file-content:
    template-dir: ".structurelint/templates"
    templates:
      "**/README.md": "readme"
      "CONTRIBUTING.md": "contributing"
```

See [File Content Templates](../docs/FILE_CONTENT_TEMPLATES.md) for full documentation.
