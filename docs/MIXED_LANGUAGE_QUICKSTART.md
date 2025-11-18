# Mixed Language Support: Quick Start Guide

## TL;DR

**Current State**: Mixed-language projects (Go + Python) require manual exemptions
**Solution**: Add `file-patterns` to `test-location` rule (like `test-adjacency` already has)
**Timeline**: Phase 1 (file patterns) → Phase 2 (per-language config) → Phase 3 (auto-detection)

## Current Workaround (Temporary)

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  exemptions:
    - "testdata/**"
    - "internal/rules/rules_test.go"
    - "tests/**/*.py"  # Manual exemption for Python tests
```

**Problems**:
- Negative logic (hard to understand what IS checked)
- Must add exemption for each new language
- Inconsistent with test-adjacency approach

## Proposed Solution: Phase 1 (Recommended)

### What Changes

Add `file-patterns` field to `test-location` rule:

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  # NEW - explicitly check only Go tests
  exemptions:
    - "testdata/**"
    - "internal/rules/rules_test.go"
    # No more Python exemptions needed!
```

### Benefits

1. ✓ **Positive Logic**: Clearly states "check Go tests"
2. ✓ **Automatic**: Non-Go tests ignored automatically
3. ✓ **Consistent**: Same approach as `test-adjacency`
4. ✓ **Maintainable**: Add new languages without changing config
5. ✓ **Backward Compatible**: Existing configs still work

### Implementation Complexity

- **Code Changes**: ~50 lines
- **Risk**: Low (reuses existing pattern matching)
- **Testing**: Straightforward (pattern matching is well-tested)
- **Migration**: Automatic (old configs work as-is)

## What This Looks Like for Different Projects

### Go-Only Project

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]
```

### Go + Python Project (like structurelint)

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  # Python tests automatically ignored
```

### Go + TypeScript Project

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns:
    - "**/*_test.go"      # Go tests
    - "**/*.test.ts"      # TypeScript tests
    - "**/*.spec.ts"      # TypeScript spec tests
```

### Python-Only Project

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: false  # Python convention: centralized tests
  file-patterns:
    - "**/test_*.py"
    - "**/*_test.py"
```

## Comparison: Before vs After

### Before (Current - Exemption-Based)

```yaml
test-adjacency:
  pattern: "adjacent"
  file-patterns: ["**/*.go"]  # ✓ Has file patterns

test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  exemptions:  # ✗ No file patterns, only exemptions
    - "tests/**/*.py"
    - "tests/**/*.js"
    - "tests/**/*.ts"
```

**Problems**:
- Inconsistent approach between test-adjacency and test-location
- Negative logic (exemptions)
- Maintenance burden

### After (Phase 1 - Pattern-Based)

```yaml
test-adjacency:
  pattern: "adjacent"
  file-patterns: ["**/*.go"]  # ✓ Has file patterns

test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  # ✓ Now has file patterns too!
```

**Benefits**:
- Consistent approach
- Positive logic (file-patterns)
- Self-documenting

## Future Enhancements (Phase 2+)

### Phase 2: Per-Language Configuration

```yaml
test-location:
  languages:
    go:
      file-patterns: ["**/*_test.go"]
      integration-test-dir: "tests"
      allow-adjacent: true
    python:
      file-patterns: ["**/test_*.py"]
      integration-test-dir: "tests"
      allow-adjacent: false  # Python convention
```

### Phase 3: Auto-Detection

```yaml
test-location:
  auto-detect: true  # Automatically detect languages and apply conventions
  overrides:
    python:
      allow-adjacent: true  # Override detected convention
```

## How to Implement Phase 1

### Step 1: Update TestLocationRule Struct

```go
type TestLocationRule struct {
    IntegrationTestDir string
    AllowAdjacent      bool
    FilePatterns       []string  // NEW
    Exemptions         []string
}
```

### Step 2: Add Pattern Matching

```go
func (r *TestLocationRule) matchesFilePattern(path string) bool {
    if len(r.FilePatterns) == 0 {
        return true  // Backward compatible
    }
    for _, pattern := range r.FilePatterns {
        if matchesGlobPattern(path, pattern) {
            return true
        }
    }
    return false
}
```

### Step 3: Update Check Logic

```go
for _, file := range files {
    if !r.isTestFile(file.Path) {
        continue
    }
    if !r.matchesFilePattern(file.Path) {  // NEW
        continue
    }
    // ... existing logic
}
```

### Step 4: Update Config Loading

```go
filePatterns := l.getStringSliceFromMap(locMap, "file-patterns")
*rulesList = append(*rulesList, rules.NewTestLocationRule(
    integrationDir, allowAdjacent, filePatterns, exemptions,
))
```

### Step 5: Update .structurelint.yml

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  # NEW
  exemptions:
    - "testdata/**"
    - "internal/rules/rules_test.go"
```

## Testing the Change

```bash
# Before
go run ./cmd/structurelint .
# Error: tests/test_clone_detection_basic.py violation

# After (with file-patterns)
go run ./cmd/structurelint .
# ✓ All checks passed
```

## Migration Path

### Option 1: Keep Old Config (Works As-Is)

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  exemptions:
    - "tests/**/*.py"
```

**Result**: Still works, no breaking change

### Option 2: Migrate to New Config (Recommended)

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  # Add this
  exemptions:
    - "testdata/**"  # Remove language-specific exemptions
```

**Result**: Cleaner, more maintainable

## FAQ

### Q: Will this break existing projects?

**A**: No. If `file-patterns` is not specified (or empty), the rule behaves exactly as before.

### Q: Do I need to update my config immediately?

**A**: No. Existing configs with exemptions will continue to work.

### Q: What if I have multiple languages to check?

**A**: List all patterns:
```yaml
file-patterns:
  - "**/*_test.go"
  - "**/*.test.ts"
```

### Q: Can I still use exemptions?

**A**: Yes! Exemptions still work and apply after pattern matching.

### Q: What about test-adjacency?

**A**: test-adjacency already has file-patterns. This change makes test-location consistent.

## Recommended Actions

1. **Review** the full design doc: `docs/MIXED_LANGUAGE_SUPPORT.md`
2. **Prototype** Phase 1 changes in a feature branch
3. **Test** on structurelint itself (Go + Python)
4. **Document** migration guide
5. **Release** as minor version (backward compatible)

## Summary

| Aspect | Current | Phase 1 | Phase 2 | Phase 3 |
|--------|---------|---------|---------|---------|
| **Approach** | Exemptions | File Patterns | Per-Language | Auto-Detect |
| **Config Complexity** | Medium | Low | Medium | Low |
| **Maintainability** | Poor | Good | Excellent | Excellent |
| **Flexibility** | Low | Medium | High | High |
| **Risk** | - | Low | Medium | High |
| **Timeline** | - | 1 iteration | 2-3 iterations | 3-5 iterations |
| **Backward Compat** | - | ✓ Yes | Requires migration | Opt-in |

**Recommendation**: Implement Phase 1 now, plan for Phase 2/3 based on user feedback.
