# Test File Validation

structurelint provides powerful rules for enforcing test file organization and ensuring your codebase has proper test coverage.

## Rules

### 1. test-adjacency

Enforces that source files have corresponding test files, supporting two common patterns:

#### Pattern: Adjacent Tests

Tests live in the same directory as the source code they test.

```yaml
rules:
  test-adjacency:
    pattern: "adjacent"            # Tests next to source files
    file-patterns:
      - "**/*.go"                  # Check all Go files
      - "**/*.ts"                  # Check all TypeScript files
    exemptions:
      - "**/*_gen.go"              # Exempt generated files
      - "cmd/**"                    # Exempt entry point files
```

**Expected test file names by language:**
- Go: `file.go` → `file_test.go`
- TypeScript: `file.ts` → `file.test.ts`
- JavaScript: `file.js` → `file.spec.js`
- Python: `file.py` → `test_file.py`

#### Pattern: Separate Test Directory

Tests live in a dedicated test directory that mirrors the source structure.

```yaml
rules:
  test-adjacency:
    pattern: "separate"
    test-dir: "tests"              # Or "test", "__tests__", etc.
    file-patterns:
      - "src/**/*.ts"
    exemptions:
      - "src/generated/**"
```

**Example structure:**
```
src/
  user/
    service.ts
tests/
  user/
    service.test.ts
```

### 2. test-location

Validates that test files are in appropriate locations. Prevents test files from being misplaced.

```yaml
rules:
  test-location:
    integration-test-dir: "tests"  # Top-level integration test directory
    allow-adjacent: true           # Allow tests next to source
    exemptions:
      - "testdata/**"
```

**This rule catches:**
- Test files that aren't adjacent to their source code
- Test files that aren't in the integration test directory
- Orphaned test files with no corresponding source

**Example violations:**
```
src/
  user.ts
  utils/
    format_test.ts  # ❌ No format.ts file here

tests/
  random_test.ts    # ❌ Not adjacent to source, not matching structure
```

**Valid structures:**
```
src/
  user.ts
  user.test.ts      # ✅ Adjacent to source

tests/
  integration/
    api_test.ts     # ✅ In integration test directory
```

## Configuration Examples

### Go Project (Adjacent Tests)

```yaml
rules:
  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "**/*.go"
    exemptions:
      - "**/*_gen.go"
      - "vendor/**"
      - "cmd/**/*.go"

  test-location:
    integration-test-dir: "tests"
    allow-adjacent: true
```

### TypeScript Project (Separate Tests)

```yaml
rules:
  test-adjacency:
    pattern: "separate"
    test-dir: "__tests__"
    file-patterns:
      - "src/**/*.ts"
      - "src/**/*.tsx"
    exemptions:
      - "src/**/*.d.ts"
      - "src/types/**"

  test-location:
    integration-test-dir: "e2e"
    allow-adjacent: false
```

### Hybrid Approach

```yaml
rules:
  # Require unit tests adjacent to source
  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "src/**/*.ts"
    exemptions:
      - "src/index.ts"

  # Allow integration tests in dedicated directory
  test-location:
    integration-test-dir: "tests/integration"
    allow-adjacent: true
```

## Benefits

1. **Enforce Test Coverage**: Catch files without tests before they merge
2. **Consistent Organization**: Maintain uniform test file placement
3. **Prevent Orphaned Tests**: Detect test files with no corresponding source
4. **Language Agnostic**: Works with Go, TypeScript, JavaScript, Python, and more
5. **Flexible Patterns**: Support both adjacent and separate test directories
6. **Integration Tests**: Dedicated space for integration/e2e tests

## Common Use Cases

### Pre-commit Hook

Ensure new files have tests before commit:

```yaml
# .structurelint.yml
rules:
  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "src/**/*.ts"
```

### CI/CD Pipeline

Fail builds if test coverage structure is violated:

```bash
structurelint . || exit 1
```

### Monorepo

Different test patterns for different packages:

```yaml
# packages/frontend/.structurelint.yml
rules:
  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "**/*.tsx"

# packages/backend/.structurelint.yml
rules:
  test-adjacency:
    pattern: "separate"
    test-dir: "tests"
    file-patterns:
      - "**/*.go"
```
