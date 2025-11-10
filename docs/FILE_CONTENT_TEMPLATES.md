# File Content Templates

structurelint can enforce file content structure using templates. This ensures documentation, configuration files, and other important files follow your team's standards.

## Overview

The `file-content` rule validates file content against predefined templates stored in `.structurelint/templates/`.

## Configuration

```yaml
rules:
  file-content:
    template-dir: ".structurelint/templates"
    templates:
      "**/README.md": "readme"           # Use readme.yml template
      "CONTRIBUTING.md": "contributing"  # Use contributing.yml template
      "docs/*.md": "design-doc"          # Use design-doc.yml template
```

## Template Structure

Templates are YAML files that define content requirements:

```yaml
# .structurelint/templates/readme.yml

# Required section headers
required-sections:
  - "## Overview"
  - "## Installation"
  - "## Usage"

# Regex patterns that must be present
required-patterns:
  - "^#\\s+\\w+"           # Must start with heading
  - "\\[.*\\]\\(.*\\)"     # Must contain at least one link

# Regex patterns that must NOT be present
forbidden-patterns:
  - "TODO"
  - "FIXME"
  - "\\[TBD\\]"

# Content must start with this pattern
must-start-with: "^#\\s+"

# Content must end with this pattern (optional)
must-end-with: "\\n$"
```

## Template Options

### required-sections

List of exact strings that must appear in the file content:

```yaml
required-sections:
  - "## Overview"
  - "## Features"
  - "## Installation"
```

### required-patterns

Regex patterns that must match somewhere in the content:

```yaml
required-patterns:
  - "^#\\s+\\w+"              # Heading at start
  - "##\\s+Installation"       # Installation section
  - "```[a-z]*\\n[\\s\\S]*?```" # At least one code block
```

### forbidden-patterns

Regex patterns that must NOT match:

```yaml
forbidden-patterns:
  - "TODO.*implement"    # No unimplemented TODOs
  - "HACK"               # No hacks
  - "XXX"                # No XXX markers
```

### must-start-with

Regex pattern that the file must start with:

```yaml
must-start-with: "^#\\s+[A-Z]"  # Must start with # Title
```

### must-end-with

Regex pattern that the file must end with:

```yaml
must-end-with: "\\n$"  # Must end with newline
```

## Example Templates

### README Template

```yaml
# .structurelint/templates/readme.yml
required-sections:
  - "## Overview"
  - "## Installation"
  - "## Usage"
  - "## Contributing"

required-patterns:
  - "^#\\s+\\w+"

must-start-with: "^#\\s+"
```

### Contributing Guidelines

```yaml
# .structurelint/templates/contributing.yml
required-sections:
  - "## Getting Started"
  - "## Development Setup"
  - "## Pull Request Process"
  - "## Code Style"
  - "## Testing"

forbidden-patterns:
  - "TODO"
  - "TBD"
```

### Design Document

```yaml
# .structurelint/templates/design-doc.yml
required-sections:
  - "## Problem Statement"
  - "## Proposed Solution"
  - "## Alternatives Considered"
  - "## Implementation Plan"
  - "## Testing Strategy"

required-patterns:
  - "(?i)motivation"
  - "(?i)trade-?offs?"

forbidden-patterns:
  - "\\[TBD\\]"
  - "TODO.*write"
```

### API Documentation

```yaml
# .structurelint/templates/api-doc.yml
required-sections:
  - "## Endpoints"
  - "## Authentication"
  - "## Request Format"
  - "## Response Format"
  - "## Error Codes"

required-patterns:
  - "`GET|POST|PUT|DELETE|PATCH`"
  - "```json"
```

### Configuration File Documentation

```yaml
# .structurelint/templates/config-doc.yml
required-sections:
  - "## Configuration Options"
  - "## Examples"

required-patterns:
  - "```yaml"
  - "Default:"
```

## Use Cases

### Enforce Documentation Standards

```yaml
rules:
  file-content:
    template-dir: ".structurelint/templates"
    templates:
      "**/README.md": "readme"
      "**/API.md": "api-doc"
      "docs/*.md": "design-doc"
```

### Require Complete Design Docs

Ensure design documents are complete before merging:

```yaml
# .structurelint/templates/design-doc.yml
required-sections:
  - "## Problem"
  - "## Solution"
  - "## Alternatives"
  - "## Security Considerations"
  - "## Performance Impact"

forbidden-patterns:
  - "\\[TBD\\]"
  - "TODO"
```

### Documentation Quality Gates

```yaml
# .structurelint/templates/readme.yml
required-patterns:
  - "\\[.*\\]\\(.*\\)"           # Must have links
  - "```"                         # Must have code examples
  - "(?i)(install|setup|usage)"   # Must explain usage

forbidden-patterns:
  - "Lorem ipsum"
  - "Example text"
  - "\\[\\]\\(\\)"                # No empty links
```

## Benefits

1. **Consistent Documentation**: All docs follow the same structure
2. **Quality Gates**: Prevent incomplete documentation from merging
3. **Onboarding**: New contributors know what's expected
4. **Automation**: No manual doc reviews needed
5. **Customizable**: Templates adapt to your team's needs
6. **Language Agnostic**: Works with any text file format

## Integration

### Pre-commit Hook

```yaml
repos:
  - repo: local
    hooks:
      - id: structurelint
        name: structurelint
        entry: structurelint
        language: system
        pass_filenames: false
```

### CI/CD

```yaml
# .github/workflows/lint.yml
- name: Validate File Content
  run: structurelint .
```

### Git Hooks

```bash
#!/bin/sh
# .git/hooks/pre-commit
structurelint . || exit 1
```

## Tips

1. **Start Simple**: Begin with basic templates and expand
2. **Use Comments**: Document why rules exist in template files
3. **Test Templates**: Validate templates work on existing files first
4. **Version Control**: Keep templates in `.structurelint/templates/`
5. **Team Buy-in**: Discuss templates with team before enforcing
6. **Gradual Rollout**: Start with warnings, then enforce
