# structurelint

**structurelint** is a next-generation linter designed to enforce project structure, organization, and architectural integrity. Unlike traditional linters that focus on code quality, structurelint ensures your project's filesystem topology remains clean, maintainable, and aligned with best practices.

## Why structurelint?

As projects grow, their directory structures often degrade into chaos:
- Deeply nested folder hierarchies that are hard to navigate
- Directories with hundreds of files lacking organization
- Inconsistent naming conventions across the codebase
- Missing critical files (like README.md or index files)

**structurelint** prevents this entropy by providing enforceable rules for:

**Phase 0 - Filesystem Linting:**
- **Directory depth limits** - Prevent unmanageable folder nesting
- **File count limits** - Keep directories focused and organized
- **Subdirectory limits** - Control complexity at each level
- **Naming conventions** - Enforce camelCase, kebab-case, PascalCase, etc.
- **File existence requirements** - Ensure critical files are present
- **Pattern restrictions** - Disallow problematic patterns

**Phase 1 - Architectural Layer Enforcement:** ✨ NEW
- **Import graph analysis** - Parse source files to build dependency graphs
- **Layer boundary validation** - Enforce architectural patterns (Clean Architecture, Hexagonal, Feature-Sliced Design, etc.)
- **Dependency rules** - Prevent violations like "domain importing from presentation"

## Features

- **Fast**: Written in Go for blazing-fast performance, suitable for pre-commit hooks
- **Cascading Configuration**: ESLint-style `.structurelint.yml` files with inheritance
- **Flexible Rules**: From simple metrics to complex pattern matching
- **Architectural Enforcement**: Layer boundaries and import graph validation
- **Multi-Language Support**: TypeScript, JavaScript, Go, Python
- **Zero Dependencies**: Single binary, easy to install and distribute

## Installation

```bash
# Download the binary (once released)
# For now, build from source:
go build -o structurelint ./cmd/structurelint

# Or install directly
go install github.com/structurelint/structurelint/cmd/structurelint@latest
```

## Quick Start

1. Create a `.structurelint.yml` file in your project root:

```yaml
root: true

rules:
  # Prevent deeply nested directories
  max-depth: { max: 7 }

  # Limit files per directory
  max-files-in-dir: { max: 20 }

  # Limit subdirectories per directory
  max-subdirs: { max: 10 }

  # Enforce naming conventions
  naming-convention:
    "*.ts": "camelCase"
    "src/components/**/": "PascalCase"
```

2. Run structurelint:

```bash
./structurelint .
```

## Configuration

### Configuration File

structurelint looks for `.structurelint.yml` or `.structurelint.yaml` files. Configuration cascades from parent directories, similar to ESLint.

### Root Configuration

Set `root: true` to stop the upward search for configuration files:

```yaml
root: true
rules:
  max-depth: { max: 5 }
```

### Cascading Configuration

You can have multiple configuration files in different directories:

```
project/
├── .structurelint.yml      # Root config
└── src/
    └── legacy/
        └── .structurelint.yml  # Override rules for legacy code
```

## Rules Reference

### Metric Rules

#### `max-depth`

Enforces a maximum directory nesting depth.

```yaml
rules:
  max-depth: { max: 7 }
```

**Example violation**: A file at `src/components/atoms/buttons/primary/variants/large/index.ts` with depth > 7.

#### `max-files-in-dir`

Limits the number of files in a single directory.

```yaml
rules:
  max-files-in-dir: { max: 20 }
```

**Example violation**: A directory containing 25 files when the limit is 20.

#### `max-subdirs`

Limits the number of subdirectories in a directory.

```yaml
rules:
  max-subdirs: { max: 10 }
```

**Example violation**: A directory with 15 subdirectories when the limit is 10.

### Naming Convention Rules

#### `naming-convention`

Enforces naming conventions for files and directories.

```yaml
rules:
  naming-convention:
    "*.ts": "camelCase"
    "*.js": "kebab-case"
    "src/components/**/": "PascalCase"
```

**Supported conventions**:
- `camelCase` - e.g., `myFile.ts`
- `PascalCase` - e.g., `MyComponent.tsx`
- `kebab-case` - e.g., `my-file.js`
- `snake_case` - e.g., `my_file.py`
- `lowercase` - e.g., `myfile.txt`
- `UPPERCASE` - e.g., `README.md`

### Pattern Rules

#### `regex-match`

Validates filenames against regex patterns.

```yaml
rules:
  regex-match:
    # Ensure component files match their directory name
    "src/components/*/*.tsx": "regex:${0}"
    # Disallow filenames that are just numbers
    "*.js": "regex:![0-9]+"
```

**Special syntax**:
- `regex:pattern` - File must match the regex
- `regex:!pattern` - File must NOT match the regex (negation)
- `${0}`, `${1}` - Substitutes directory names from wildcards

**Example**: `src/components/Button/Button.tsx` matches `${0}` (both are "Button")

#### `file-existence`

Requires specific files to exist in directories.

```yaml
rules:
  file-existence:
    # Every directory must have exactly one index file
    "index.ts|index.js": "exists:1"
    # Must have at least one test file
    "*.test.ts": "exists:1"
    # No subdirectories allowed
    ".dir": "exists:0"
    # Must have between 1 and 10 .md files
    "*.md": "exists:1-10"
```

**Syntax**:
- `exists:1` - Exactly 1 file must exist
- `exists:0` - No files of this type allowed
- `exists:1-10` - Between 1 and 10 files
- `.dir` - Special pattern for subdirectories

#### `disallowed-patterns`

Blocks specific file or directory patterns.

```yaml
rules:
  disallowed-patterns:
    - "src/utils/**"  # Disallow generic utils folder
    - "*.tmp"         # No temp files
    - ".DS_Store"     # No macOS metadata
```

## Advanced Configuration

### Overrides

Apply different rules to specific parts of your project:

```yaml
root: true

rules:
  max-depth: { max: 7 }
  max-files-in-dir: { max: 15 }

overrides:
  # Stricter rules for components
  - files: ['src/components/**']
    rules:
      max-depth: { max: 10 }
      file-existence:
        "index.ts|index.tsx": "exists:1"
      naming-convention:
        "**/": "PascalCase"

  # Relaxed rules for legacy code
  - files: ['src/legacy/**']
    rules:
      max-depth: 0        # Disable rule (0 = disabled)
      max-files-in-dir: 0
```

### Disabling Rules

Set a rule to `0` or `false` to disable it:

```yaml
rules:
  max-depth: 0           # Disabled
  naming-convention: false  # Also disabled
```

## Example Configurations

### React Project

```yaml
root: true

rules:
  max-depth: { max: 8 }
  max-files-in-dir: { max: 20 }
  max-subdirs: { max: 10 }

  naming-convention:
    "src/**/*.ts": "camelCase"
    "src/**/*.tsx": "PascalCase"
    "src/components/**/": "PascalCase"

  disallowed-patterns:
    - "src/components/atoms"      # Discourage atomic design
    - "src/components/molecules"
    - "src/utils/**"               # Prefer specific utility folders

overrides:
  - files: ['src/features/*']
    rules:
      file-existence:
        "index.ts|index.tsx": "exists:1"
      max-subdirs: { max: 5 }
```

### Go Project

```yaml
root: true

rules:
  max-depth: { max: 6 }
  max-files-in-dir: { max: 15 }

  naming-convention:
    "**/*.go": "snake_case"
    "cmd/**/": "snake_case"

overrides:
  - files: ['internal/**']
    rules:
      max-depth: { max: 5 }

  - files: ['pkg/*']
    rules:
      file-existence:
        "README.md": "exists:1"
```

### Python Project

```yaml
root: true

rules:
  max-depth: { max: 5 }
  max-files-in-dir: { max: 20 }

  naming-convention:
    "**/*.py": "snake_case"
    "**/": "snake_case"

  file-existence:
    "__init__.py": "exists:1"  # All packages need __init__.py

disallowed-patterns:
  - "**/__pycache__"
  - "**/*.pyc"
```

## Integration

### Pre-commit Hook

Add to your `.git/hooks/pre-commit`:

```bash
#!/bin/sh
./structurelint . || exit 1
```

### GitHub Actions

```yaml
name: Lint Project Structure

on: [push, pull_request]

jobs:
  structurelint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run structurelint
        run: |
          go install github.com/structurelint/structurelint/cmd/structurelint@latest
          structurelint .
```

## Roadmap

### Phase 0 (Current) - Core Filesystem Linting
- ✅ Metric rules (max-depth, max-files, max-subdirs)
- ✅ Naming conventions
- ✅ File existence validation
- ✅ Pattern matching and disallowing

### Phase 1 - Architectural Layer Enforcement
- Import graph analysis
- Layer boundary enforcement
- Dependency rules

### Phase 2 - Dead Code Detection
- Orphaned file detection
- Unused export identification
- Compiler plugin system for non-standard files

## Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Acknowledgments

Inspired by:
- [ls-lint](https://ls-lint.org/) - Fast filesystem linter
- [ESLint](https://eslint.org/) - Configuration system design
- [Knip](https://github.com/webpro/knip) - Dead code detection
## Phase 1: Layer Boundary Enforcement

### Overview

Phase 1 adds powerful architectural validation by analyzing import/dependency graphs and enforcing layer boundaries. This allows you to define and enforce architectural patterns like Clean Architecture, Hexagonal Architecture, or Feature-Sliced Design.

### Configuration

Define layers in your `.structurelint.yml`:

```yaml
root: true

# Define architectural layers
layers:
  - name: 'domain'
    path: 'src/domain/**'
    dependsOn: []  # Domain has no dependencies

  - name: 'application'
    path: 'src/application/**'
    dependsOn: ['domain']  # Can depend on domain

  - name: 'presentation'
    path: 'src/presentation/**'
    dependsOn: ['application', 'domain']  # Can depend on application and domain

rules:
  # Enable layer boundary enforcement
  enforce-layer-boundaries: true
```

### How It Works

1. **Import Parsing**: structurelint parses source files (TypeScript, JavaScript, Go, Python) to extract import statements
2. **Graph Building**: Creates a dependency graph showing which files import which
3. **Layer Assignment**: Assigns each file to a layer based on the `path` patterns
4. **Boundary Validation**: Checks that imports respect the `dependsOn` rules

### Example: Preventing Layer Violations

Given this configuration:

```yaml
layers:
  - name: 'domain'
    path: 'src/domain/**'
    dependsOn: []

  - name: 'presentation'
    path: 'src/presentation/**'
    dependsOn: ['domain']
```

This violation will be detected:

```typescript
// src/domain/user.ts
import { UserComponent } from '../presentation/userComponent'  // ❌ VIOLATION!
// Domain cannot import from presentation (dependsOn: [])
```

Output:
```
src/domain/user.ts: layer 'domain' cannot import from layer 'presentation' (imported: src/presentation/userComponent.ts)
```

### Example Architectures

#### Clean Architecture

```yaml
layers:
  - name: 'domain'
    path: 'src/domain/**'
    dependsOn: []

  - name: 'application'
    path: 'src/application/**'
    dependsOn: ['domain']

  - name: 'infrastructure'
    path: 'src/infrastructure/**'
    dependsOn: ['domain', 'application']

  - name: 'presentation'
    path: 'src/presentation/**'
    dependsOn: ['application', 'domain']

rules:
  enforce-layer-boundaries: true
```

#### Hexagonal (Ports & Adapters)

```yaml
layers:
  - name: 'core'
    path: 'src/core/**'
    dependsOn: []

  - name: 'ports'
    path: 'src/ports/**'
    dependsOn: ['core']

  - name: 'adapters-in'
    path: 'src/adapters/in/**'
    dependsOn: ['ports', 'core']

  - name: 'adapters-out'
    path: 'src/adapters/out/**'
    dependsOn: ['ports', 'core']

rules:
  enforce-layer-boundaries: true
```

#### Feature-Sliced Design

```yaml
layers:
  - name: 'shared'
    path: 'src/shared/**'
    dependsOn: []

  - name: 'entities'
    path: 'src/entities/**'
    dependsOn: ['shared']

  - name: 'features'
    path: 'src/features/**'
    dependsOn: ['shared', 'entities']

  - name: 'widgets'
    path: 'src/widgets/**'
    dependsOn: ['shared', 'entities', 'features']

  - name: 'pages'
    path: 'src/pages/**'
    dependsOn: ['shared', 'entities', 'features', 'widgets']

rules:
  enforce-layer-boundaries: true
```

### Wildcard Dependencies

Use `'*'` to allow a layer to depend on all others (useful for app/config layers):

```yaml
layers:
  - name: 'app'
    path: 'src/app/**'
    dependsOn: ['*']  # Can import from any layer
```

### Supported Languages

- **TypeScript/JavaScript**: `.ts`, `.tsx`, `.js`, `.jsx`, `.mjs`
- **Go**: `.go`
- **Python**: `.py`

### Complete Example

See `examples/clean-architecture.yml`, `examples/hexagonal-architecture.yml`, or `examples/feature-sliced.yml` for full working examples.
## Phase 2: Dead Code Detection

### Overview

Phase 2 adds dead code detection by identifying orphaned files and unused exports. This helps eliminate project bloat and keeps your codebase clean.

### Features

**1. Orphaned File Detection**
- Identifies files that are never imported by any other file
- Respects configured entrypoints
- Automatically excludes configuration files, test files, and documentation

**2. Unused Export Detection**
- Finds exported symbols that are never imported elsewhere in the project
- Helps identify dead code that can be safely removed
- Works across TypeScript, JavaScript, Go, and Python

### Configuration

```yaml
root: true

# Define entry points (files that don't need to be imported)
entrypoints:
  - "src/index.ts"
  - "src/main.go"
  - "**/*test*"      # All test files
  - "**/__tests__/**" # Test directories

rules:
  # Enable Phase 2 dead code detection
  disallow-orphaned-files: true
  disallow-unused-exports: true
```

### Example Violations

**Orphaned File:**
```
src/unused-util.ts: file is orphaned (not imported by any other file)
```

**Unused Exports:**
```
src/helpers.ts: exports 'formatDate', 'parseNumber' but is never imported
```

### Automatic Exclusions

The orphaned files rule automatically excludes:
- **Configuration files**: `.structurelint.yml`, `package.json`, `tsconfig.json`, etc.
- **Documentation**: `*.md`, `*.txt` files
- **Test files**: Files matching `*test*`, `*spec*`
- **Common entrypoints**: `main.*`, `index.*`, `app.*`, `__init__.py`

### Combining with Overrides

You can disable dead code detection for specific paths:

```yaml
rules:
  disallow-orphaned-files: true
  disallow-unused-exports: true

overrides:
  # Don't check test files
  - files: ['**/*test*', '**/__tests__/**']
    rules:
      disallow-orphaned-files: 0
      disallow-unused-exports: 0

  # Entrypoints can have unused exports (they're for external use)
  - files: ['src/index.ts', 'src/main.ts']
    rules:
      disallow-unused-exports: 0
```

### Complete Example

See `examples/dead-code-detection.yml` and `examples/complete-setup.yml` for full working examples combining all three phases.

### How It Works

1. **Import Graph**: Phase 2 builds on the import graph from Phase 1
2. **Reference Counting**: Tracks how many times each file is imported
3. **Export Parsing**: Extracts all export statements from source files
4. **Cross-Reference**: Compares exports against import statements

### Benefits

- **Reduce Bundle Size**: Remove unused code that bloats your application
- **Improve Maintainability**: Clean codebase is easier to understand
- **Prevent Accumulation**: Catch dead code before it becomes technical debt
- **CI/CD Integration**: Enforce cleanliness in your build pipeline
