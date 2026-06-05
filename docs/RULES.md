# Structurelint Rules Reference

Complete reference for all available rules in structurelint.

## Table of Contents

- [Layer & Dependency Rules](#layer--dependency-rules)
- [Naming Convention Rules](#naming-convention-rules)
- [File Organization Rules](#file-organization-rules)
- [Code Quality Rules](#code-quality-rules)
- [Structural Integrity Rules](#structural-integrity-rules)
- [Testing Rules](#testing-rules)
- [Rule Types (Advanced)](#rule-types-advanced)

---

## ⚠️ Removed Rules

The following rules have been removed. Using them causes an error with migration advice:

| Removed Rule | Migration |
|---|---|
| `max-cyclomatic-complexity` | Use `max-cognitive-complexity` or a language-specific tool |
| `github-workflows` | Use actionlint, zizmor, or yamllint |
| `linter-config` | Use `file-existence` to require config files |
| `contract-framework` | Use `file-existence` to encode requirements |
| `api-spec` | Replace with `file-existence: { "api/openapi.yaml": "exists:1" }` |
| `spec-adr-enforcement` | Replace with `file-existence` + `naming-convention` |
| `file-content` | Use copier / cookiecutter for templating |
| `property-enforcement` | Use `disallow-import-cycles` + `path-based-layers` `forbiddenPaths` |

---

## Layer & Dependency Rules

### `enforce-layer-boundaries`

Enforces that layers only depend on allowed layers.

**Configuration**:

```yaml
layers:
  - name: domain
    path: "internal/domain/**"

  - name: application
    path: "internal/application/**"
    dependsOn: [domain]

rules:
  enforce-layer-boundaries: true
```

**Detects**:
- Imports from disallowed layers
- Circular dependencies between layers
- Violations of dependency hierarchy

**Example Violation**:

```go
// ❌ domain/user.go imports application layer
package domain
import "myapp/application/usecases" // VIOLATION
```

**Fix**:

```go
// ✓ domain defines interfaces, application implements
package domain
type UserRepository interface { ... }
```

---

### `disallowed-patterns`

Blocks specific file or directory patterns from existing in the project.

**Configuration**:

```yaml
rules:
  disallowed-patterns:
    - "src/utils/**"        # Disallow generic utils folder
    - "*.tmp"               # No temp files
    - ".DS_Store"           # No macOS metadata
    - "**/__pycache__"      # No Python cache dirs
```

**Use cases**:
- Prevent anti-pattern directories (generic `utils/`, `helpers/`)
- Block platform metadata files (`.DS_Store`, `Thumbs.db`)
- Disallow generated artifacts in source directories

---

### `path-based-layers`

Declarative layer validation without import graph parsing. Much faster than `enforce-layer-boundaries`.

**Configuration**:

```yaml
rules:
  path-based-layers:
    layers:
      - name: domain
        patterns: ["internal/domain/**"]
        canDependOn: []
      - name: application
        patterns: ["internal/application/**"]
        canDependOn: ["domain"]
      - name: infrastructure
        patterns: ["internal/infrastructure/**"]
        canDependOn: ["domain", "application"]
        forbiddenPaths: ["**/http/**", "**/sql/**"]
```

**Use cases**:
- Enforce architecture without parsing source files
- Works even when code doesn't compile
- 50x faster than import-graph-based validation

---

## Naming Convention Rules

### `naming-convention`

Enforces file naming patterns.

**Configuration**:

```yaml
rules:
  naming-convention:
    # React components must be PascalCase
    "src/components/*.tsx": "^[A-Z][a-zA-Z0-9]*\\.tsx$"

    # React hooks must start with 'use'
    "src/hooks/*.ts": "^use[A-Z][a-zA-Z0-9]*\\.ts$"

    # Go files must be snake_case
    "internal/**/*.go": "^[a-z][a-z0-9_]*\\.go$"
```

**Patterns**:
- **PascalCase**: `^[A-Z][a-zA-Z0-9]*$`
- **camelCase**: `^[a-z][a-zA-Z0-9]*$`
- **snake_case**: `^[a-z][a-z0-9_]*$`
- **kebab-case**: `^[a-z][a-z0-9-]*$`

**Auto-fix**: ✅ Available via `structurelint fix`

---

### `naming-convention` auto-language-naming

When `autoLanguageNaming: true` (default) in the config, language-specific conventions are automatically applied:

- **Go**: `snake_case.go`, `name_test.go`
- **TypeScript**: `PascalCase.tsx`, `camelCase.ts`
- **Python**: `snake_case.py`, `test_name.py`
- **Java**: `PascalCase.java`, `PascalCaseTest.java`

**Disable**:

```yaml
autoLanguageNaming: false
```

---

## File Organization Rules

### `max-depth`

Limits directory nesting depth.

**Configuration**:

```yaml
rules:
  max-depth:
    max: 4
```

**Why**: Deep nesting reduces readability and indicates poor organization.

**Violation Example**:

```
src/
  components/
    admin/
      users/
        profile/
          edit/
            form/  # 7 levels deep - VIOLATION
```

---

### `max-files-in-dir`

Limits number of files per directory.

**Configuration**:

```yaml
rules:
  max-files-in-dir:
    max: 20
    paths: ["src/components/*"]
```

**Why**: Too many files in one directory reduces navigability.

---

### `max-subdirs`

Limits number of subdirectories.

**Configuration**:

```yaml
rules:
  max-subdirs:
    max: 10
    paths: ["internal/*"]
```

---

### `file-existence`

Requires specific files to exist.

**Configuration**:

```yaml
rules:
  file-existence:
    "README.md": "Project must have README"
    ".gitignore": "Git ignore file required"
    "LICENSE": "License file required"
    "internal/domain/repository.go": "Repository interface required"
```

**Auto-fix**: ✅ Can generate files from templates

---

### `test-location`

Enforces test file placement.

**Configuration**:

```yaml
rules:
  test-location:
    integration-test-dir: "tests/integration"
    allow-adjacent: true
```

**Options**:
- `allow-adjacent: true` - Tests beside source files (Go style)
- `integration-test-dir` - Separate integration tests
- File patterns to identify test files

---

## Code Quality Rules

### `max-cognitive-complexity`

Limits cognitive complexity (scientifically superior to cyclomatic).

**Configuration**:

```yaml
rules:
  max-cognitive-complexity:
    max: 15
    test-max: 25
    file-patterns: ["**/*.go", "**/*.ts"]
```

**Why**: Cognitive complexity correlates better with bug density (r=0.54) than cyclomatic complexity.

**Measures**:
- Nested conditionals (+1 per level)
- Recursive calls
- Jumps (break, continue)
- Logical operators

---

### `max-halstead-effort`

Limits Halstead effort (data complexity).

**Configuration**:

```yaml
rules:
  max-halstead-effort:
    max: 1000000
    file-patterns: ["**/*.go"]
```

**Why**: High Halstead effort indicates difficult-to-understand code.

---

### `disallow-orphaned-files`

Finds files not imported anywhere.

**Configuration**:

```yaml
rules:
  disallow-orphaned-files: true
```

With options:

```yaml
rules:
  disallow-orphaned-files:
    entry-point-patterns:
      - "cmd/**"
      - "*_test.go"
```

**Detects**: Dead code, unused utilities

---

### `disallow-unused-exports`

Finds exported symbols never imported.

**Configuration**:

```yaml
rules:
  disallow-unused-exports: true
```

With options:

```yaml
rules:
  disallow-unused-exports:
    exclude-patterns:
      - "internal/domain/*.go"  # Exports for extensibility
    entry-point-patterns:
      - "src/index.ts"          # Entry points can have unused exports
```

---

## Structural Integrity Rules

### `uniqueness-constraints`

Prevents dual-implementation anti-patterns.

```yaml
rules:
  uniqueness-constraints:
    "index.ts|index.js": "exists:1"
```

### `case-conflicts`

Detects filename case collisions on case-insensitive filesystems.

```yaml
rules:
  case-conflicts: {}
```

### `disallow-empty-dirs`

Flags directories with no files.

```yaml
rules:
  disallow-empty-dirs: {}
```

### `disallow-symlinks`

Disallows symbolic links.

```yaml
rules:
  disallow-symlinks: {}
```

### `disallow-deep-relative-imports`

Flags imports with excessive `../` segments.

```yaml
rules:
  disallow-deep-relative-imports:
    max-parents: 3
```

### `disallow-import-cycles`

Detects circular dependencies.

```yaml
rules:
  disallow-import-cycles: true
```

---

## Testing Rules

---

## Rule Types (Advanced)

structurelint supports advanced rule composition for custom validation logic.

### `predicate-rule`

Create custom rules with predicates.

**Example**:

```yaml
rules:
  custom-domain-purity:
    type: predicate
    predicate:
      all:
        - in-layer: domain
        - not:
            depends-on: "*infrastructure*"
    message: "Domain must be pure"
```

**Predicates**:
- `in-layer(name)`
- `depends-on(pattern)`
- `has-import(pattern)`
- `file-matches(pattern)`
- `all(...)`, `any(...)`, `not(...)`

---

### `ast-query-rule`

Create rules with tree-sitter queries.

**Example** (Find global variables in Go):

```yaml
rules:
  no-global-vars:
    type: ast-query
    language: go
    query: |
      (var_declaration) @global
    message: "Global variables not allowed"
```

---

### `composite-rule`

Combine multiple rules with logic.

**Example**:

```yaml
rules:
  strict-domain:
    type: composite
    operator: all-of
    rules:
      - no-infrastructure-imports
      - no-global-state
      - pure-functions-only
```

**Operators**: `all-of`, `any-of`, `not`, `xor`

---

## Auto-fix Support

Rules with auto-fix capability:

| Rule | Auto-fix | Safe |
|------|----------|------|
| naming-convention | ✅ | ⚠️ |
| file-existence | ✅ | ✅ |
| test-location | ✅ | ⚠️ |

**Legend**:
- ✅ Safe - Can auto-apply
- ⚠️ Unsafe - Requires review

**Usage**:

```bash
# Preview fixes
structurelint fix --dry-run

# Apply safe fixes
structurelint fix --auto

# Interactive review
structurelint fix --interactive

# Fix only a specific rule
structurelint fix --rule naming-convention
```

---

## Rule Composition

Combine rules for powerful constraints:

```yaml
rules:
  # Clean Architecture enforcement
  clean-arch:
    type: composite
    operator: all-of
    rules:
      - enforce-layer-boundaries
      - domain-purity:
          type: predicate
          predicate:
            all:
              - in-layer: domain
              - not:
                  has-import: "database/sql"
      - use-case-isolation:
          type: predicate
          predicate:
            all:
              - in-layer: usecases
              - only-depends-on: [domain]
```

---

## Best Practices

### Start Simple

```yaml
rules:
  # Phase 1: Basic organization
  max-depth:
    max: 5

  naming-convention:
    "**/*.go": "^[a-z_]+\\.go$"
```

### Add Layers Gradually

```yaml
# Phase 2: Add layer boundaries
layers:
  - name: domain
    path: "internal/domain/**"
  - name: application
    path: "internal/application/**"
    dependsOn: [domain]
```

### Customize for Your Team

```yaml
# Phase 3: Team-specific rules
rules:
  max-cognitive-complexity:
    max: 10  # Stricter than default
    test-max: 20
```

---

## Troubleshooting

### "Rule not found"

Check rule name spelling:

```bash
structurelint --help  # List available rules
```

### "Too many violations"

Start with warnings:

```yaml
rules:
  enforce-layer-boundaries:
    enabled: true
    severity: warning  # Don't fail CI yet
```

### "False positives"

Add exemptions:

```yaml
rules:
  disallowed-patterns:
    "internal/domain/**":
      - "database/sql"
    exemptions:
      - "internal/domain/testing_utils.go"
```

---

## See Also

- [Example Configurations](../examples/)
- [Getting Started Guide](./GETTING_STARTED.md)
- [GitHub Actions Integration](./GITHUB_ACTION.md)
