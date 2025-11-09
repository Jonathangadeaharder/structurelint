# Test Fixtures

This directory contains test fixtures for integration testing structurelint.

## Directory Structure

```
testdata/
├── fixtures/
│   ├── good-project/       # Project that passes all checks
│   ├── bad-project/        # Project with Phase 0 violations
│   └── layer-violations/   # Project with Phase 1 violations
```

## Fixtures

### good-project

A well-structured project that passes all structurelint checks:
- Respects max depth (3 levels)
- Respects max files per directory (10)
- Uses kebab-case naming convention
- Has correct layer dependencies (services → domain)
- No orphaned files

**Expected result:** ✅ No violations found

### bad-project

A project with Phase 0 violations:
- **Max depth violation**: `src/deeply/nested/directories/` exceeds max depth of 2
- **Max files violation**: `src/` directory contains 5 files (max: 3)

**Expected result:** ❌ 4 violations found

### layer-violations

A project with Phase 1 architectural violations:
- **Layer boundary violation**: `domain` layer imports from `application` layer
  - File: `src/domain/user.ts`
  - Violation: Domain should not depend on application layer

The correct architecture should be:
```
presentation → application → domain
```

But the code has:
```
domain → application (❌ WRONG DIRECTION)
```

**Expected result:** ❌ 1 violation found

## Running Tests

Test all fixtures:

```bash
# Good project (should pass)
./structurelint testdata/fixtures/good-project

# Bad project (should fail with 4 violations)
./structurelint testdata/fixtures/bad-project

# Layer violations (should fail with 1 violation)
./structurelint testdata/fixtures/layer-violations
```

## Creating New Fixtures

To create a new test fixture:

1. Create a directory under `testdata/fixtures/`
2. Add a `.structurelint.yml` configuration file
3. Add source files that either pass or violate the rules
4. Document the expected behavior in this README
5. Add integration tests if needed
