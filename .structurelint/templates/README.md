# templates

⬆️ **[Parent Directory](../README.md)**

## Overview

This directory contains YAML template files that define content structure requirements for various file types.

## Available Templates

### Documentation Templates

| Template | Purpose | Used For |
|----------|---------|----------|
| `readme.yml` | README structure | All README.md files |
| `contributing.yml` | Contributing guidelines | CONTRIBUTING.md |
| `design-doc.yml` | Design document structure | Technical design docs |

### Test Templates (AAA Pattern)

| Template | Purpose | Used For |
|----------|---------|----------|
| `test-go.yml` | Go test AAA pattern | `*_test.go` files |
| `test-typescript.yml` | TypeScript/JS test AAA pattern | `*.test.ts`, `*.spec.js` files |
| `test-python.yml` | Python test AAA pattern | `test_*.py`, `*_test.py` files |
| `test-strict-aaa.yml` | Strict AAA enforcement | Any test files (multi-language) |

The test templates enforce the Arrange-Act-Assert (AAA) pattern for better test readability and consistency. See [Test AAA Pattern](../../docs/TEST_AAA_PATTERN.md) for details.

### Test Templates (Given-When-Then Naming)

| Template | Purpose | Used For |
|----------|---------|----------|
| `test-gwt-go.yml` | Go test with GWT naming + AAA | `*_test.go` files |
| `test-gwt-typescript.yml` | TypeScript/JS test with GWT naming + AAA | `*.test.ts`, `*.spec.js` files |
| `test-gwt-python.yml` | Python test with GWT naming + AAA | `test_*.py`, `*_test.py` files |
| `test-gwt-strict.yml` | Ultra-strict GWT + AAA enforcement | Any test files (multi-language) |

The GWT templates combine descriptive Given-When-Then naming with AAA structure for maximum test clarity. See [Given-When-Then Naming](../../docs/TEST_GWT_NAMING.md) for details.

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
