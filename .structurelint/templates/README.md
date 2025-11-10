# templates

⬆️ **[Parent Directory](../README.md)**

## Overview

This directory contains YAML template files that define content structure requirements for various file types.

## Available Templates

| Template | Purpose | Used For |
|----------|---------|----------|
| `readme.yml` | README structure | All README.md files |
| `contributing.yml` | Contributing guidelines | CONTRIBUTING.md |
| `design-doc.yml` | Design document structure | Technical design docs |

## Template Format

Templates are YAML files with the following structure:

```yaml
# Required section headers (exact match)
required-sections:
  - "## Overview"
  - "## Usage"

# Regex patterns that must be present
required-patterns:
  - "^#\\s+\\w+"

# Regex patterns that must NOT be present
forbidden-patterns:
  - "TODO"
  - "FIXME"

# Pattern file must start with
must-start-with: "^#\\s+"

# Pattern file must end with (optional)
must-end-with: "\\n$"
```

## Adding New Templates

1. Create a new `.yml` file in this directory
2. Define structure requirements using the format above
3. Reference it in `.structurelint.yml`:

```yaml
rules:
  file-content:
    template-dir: ".structurelint/templates"
    templates:
      "path/pattern/**/*.md": "your-template-name"
```

## Documentation

See [File Content Templates](../../docs/FILE_CONTENT_TEMPLATES.md) for complete documentation and examples.
