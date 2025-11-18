# Mixed Language Architecture: Visual Guide

## Current Architecture (Problematic)

```
┌─────────────────────────────────────────────────────────────┐
│ File Walker - Discovers all files in project               │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ Test Detection - Hardcoded patterns for ALL languages      │
│                                                             │
│ Patterns:                                                   │
│  ├─ _test      (Go)                                        │
│  ├─ .test      (TypeScript)                                │
│  ├─ .spec      (JavaScript/TypeScript)                     │
│  ├─ _spec      (Ruby)                                      │
│  ├─ Test       (Java)                                      │
│  └─ test_      (Python)  ← Detects Python tests!          │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ test-location Rule                                          │
│                                                             │
│ Checks ALL detected test files (Go, Python, etc.)          │
│ Applies Go-specific validation logic                       │
│ No file pattern scoping ✗                                  │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
              ┌──────┴──────┐
              │             │
        ✓ Go Tests    ✗ Python Tests
              │             │
              │             └─→ VIOLATION!
              │                 (Even if correctly placed)
              │
              └─────────→ Must manually exempt:
                          "tests/**/*.py"
```

**Problem**: Python tests detected → Go rules applied → False violations

---

## Proposed Architecture (Phase 1)

```
┌─────────────────────────────────────────────────────────────┐
│ File Walker - Discovers all files in project               │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ Test Detection - Hardcoded patterns (unchanged)             │
│                                                             │
│ Patterns: _test, .test, .spec, test_, Test, etc.           │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ test-location Rule (Enhanced)                               │
│                                                             │
│ NEW: File Pattern Filtering                                │
│  ├─ file-patterns: ["**/*_test.go"]                        │
│  └─ Only checks Go tests ✓                                 │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
              ┌──────┴──────┐
              │             │
        ✓ Go Tests    Python Tests
              │             │
              │             └─→ IGNORED (doesn't match pattern)
              │                 No violation ✓
              │
              └─────────→ Go-specific validation applied
                          Only to Go tests ✓
```

**Solution**: File patterns filter which tests are validated

---

## Code Flow Comparison

### Current (Problematic)

```go
func (r *TestLocationRule) Check(files []walker.FileInfo) []Violation {
    for _, file := range files {
        if !r.isTestFile(file.Path) {
            continue  // Skip non-test files
        }

        // ✗ PROBLEM: All detected tests checked here
        // Python, Go, TypeScript all validated with same logic

        if r.isExempted(file.Path) {
            continue  // Manual exemptions only
        }

        // Apply validation...
    }
}
```

### Proposed (Phase 1)

```go
func (r *TestLocationRule) Check(files []walker.FileInfo) []Violation {
    for _, file := range files {
        if !r.isTestFile(file.Path) {
            continue  // Skip non-test files
        }

        // ✓ NEW: Pattern-based filtering
        if !r.matchesFilePattern(file.Path) {
            continue  // Skip tests not matching patterns
        }

        if r.isExempted(file.Path) {
            continue  // Exemptions still work
        }

        // Apply validation... (only to matching tests)
    }
}
```

**Key Change**: Added pattern check after test detection

---

## Rule Consistency Comparison

### test-adjacency (Already Good)

```go
type TestAdjacencyRule struct {
    Pattern      string
    TestDir      string
    FilePatterns []string  // ✓ Has file patterns
    Exemptions   []string
}

// Config
test-adjacency:
  pattern: "adjacent"
  file-patterns: ["**/*.go"]  // ✓ Explicit scope
  exemptions: ["cmd/**/*.go"]
```

### test-location (Current - Inconsistent)

```go
type TestLocationRule struct {
    IntegrationTestDir string
    AllowAdjacent      bool
    // ✗ NO FilePatterns field
    Exemptions         []string
}

// Config
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  # ✗ No file-patterns option
  exemptions: ["tests/**/*.py"]  # Workaround
```

### test-location (Proposed - Consistent)

```go
type TestLocationRule struct {
    IntegrationTestDir string
    AllowAdjacent      bool
    FilePatterns       []string  // ✓ NEW - now consistent
    Exemptions         []string
}

// Config
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  // ✓ Explicit scope
  exemptions: ["testdata/**"]  # True exemptions only
```

**Result**: Both rules use same pattern-based approach

---

## Future Architecture (Phase 2)

```
┌─────────────────────────────────────────────────────────────┐
│ File Walker                                                 │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ Language Detection                                          │
│  ├─ **/*_test.go → Go                                      │
│  ├─ **/test_*.py → Python                                  │
│  └─ **/*.test.ts → TypeScript                              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│ Multi-Language test-location Rule                          │
│                                                             │
│ Per-language configurations:                               │
│  ├─ Go:                                                    │
│  │   ├─ file-patterns: ["**/*_test.go"]                   │
│  │   ├─ integration-test-dir: "tests"                     │
│  │   └─ allow-adjacent: true                              │
│  │                                                         │
│  ├─ Python:                                                │
│  │   ├─ file-patterns: ["**/test_*.py"]                   │
│  │   ├─ integration-test-dir: "tests"                     │
│  │   └─ allow-adjacent: false  (Python convention)        │
│  │                                                         │
│  └─ TypeScript:                                            │
│      ├─ file-patterns: ["**/*.test.ts"]                    │
│      ├─ integration-test-dir: "__tests__"                  │
│      └─ allow-adjacent: true                               │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
              ┌──────┴────────────────┐
              │      │                │
         Go Tests  Python Tests  TS Tests
              │      │                │
              │      │                │
              ▼      ▼                ▼
         Go Rules  Python Rules  TS Rules
              │      │                │
              └──────┴────────────────┘
                     │
                     ▼
              Language-appropriate validation
```

---

## Decision Tree: When to Use What

```
Is this a multi-language project?
│
├─ No (Go only)
│  └─ Use standard config with file-patterns: ["**/*_test.go"]
│
└─ Yes (Go + Python/etc.)
   │
   ├─ Want to validate both languages?
   │  └─ Phase 2: Per-language config (future)
   │
   └─ Want to validate only Go?
      └─ Phase 1: file-patterns: ["**/*_test.go"]
         (Other languages auto-ignored) ✓
```

---

## Implementation Effort Matrix

| Phase | Lines Changed | Files Modified | Risk | Complexity | Value |
|-------|--------------|----------------|------|------------|-------|
| **Phase 1** | ~50 | 3 | Low | Simple | High |
| Phase 2 | ~200 | 5 | Med | Medium | Medium |
| Phase 3 | ~500 | 10 | High | Complex | High |

**Recommendation**: Phase 1 gives best value/effort ratio

---

## Pattern Matching Logic

### Current (test-adjacency) - Reusable!

```go
// internal/rules/adjacency_validation.go:173-180
func (r *TestAdjacencyRule) matchesFilePattern(path string) bool {
    for _, pattern := range r.FilePatterns {
        if matchesGlobPattern(path, pattern) {
            return true
        }
    }
    return false
}
```

**Key Insight**: This logic already exists and works well!

### Proposed (test-location) - Copy Pattern

```go
// internal/rules/location_validation.go (NEW)
func (r *TestLocationRule) matchesFilePattern(path string) bool {
    // Backward compatible: empty patterns = match all
    if len(r.FilePatterns) == 0 {
        return true
    }

    // Check against each pattern
    for _, pattern := range r.FilePatterns {
        if matchesGlobPattern(path, pattern) {
            return true
        }
    }
    return false
}
```

**Reuse**: Can potentially extract to shared utility function

---

## Config Evolution Timeline

### Today (Exemption-Based)

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  exemptions:
    - "testdata/**"
    - "tests/**/*.py"    # Workaround
    - "tests/**/*.js"    # More workarounds
    - "tests/**/*.ts"    # Even more workarounds
```

### Phase 1 (Pattern-Based)

```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  # Positive logic
  exemptions:
    - "testdata/**"  # Only true exemptions
```

### Phase 2 (Per-Language)

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
      allow-adjacent: false
```

### Phase 3 (Auto-Detect)

```yaml
test-location:
  auto-detect: true
  # Convention-aware, minimal config needed
```

---

## Testing Strategy Visual

```
┌─────────────────────────────────────────────────────────────┐
│ Test Suite for Phase 1                                      │
└────────────────────┬────────────────────────────────────────┘
                     │
    ┌────────────────┼────────────────┐
    │                │                │
    ▼                ▼                ▼
┌─────────┐    ┌─────────┐    ┌──────────┐
│ Go Only │    │ Python  │    │ Mixed    │
│ Project │    │ Only    │    │ Go+Py    │
└────┬────┘    └────┬────┘    └────┬─────┘
     │              │              │
     │              │              │
     ▼              ▼              ▼
┌─────────────────────────────────────┐
│ Pattern Matching Tests              │
│  ├─ Empty patterns (backward compat)│
│  ├─ Single pattern                  │
│  ├─ Multiple patterns                │
│  └─ Glob patterns                   │
└─────────────────────────────────────┘
```

---

## Migration Safety

```
Old Config (v1)          New Config (v2)
───────────────          ───────────────

test-location:           test-location:
  integration-test-dir     integration-test-dir
  allow-adjacent           allow-adjacent
  exemptions               file-patterns ← NEW
                          exemptions

│                         │
└────── Both Work ────────┘

Backward Compatible: ✓
Migration Required: ✗
Breaking Change: ✗
```

**Migration Strategy**:
1. Ship Phase 1 with both approaches working
2. Document new approach in release notes
3. Deprecate exemption-based approach in next major version
4. Remove exemption support only if absolutely necessary (probably never)

---

## Summary

**Current Problem**:
- test-location checks ALL languages
- Requires manual exemptions
- Inconsistent with test-adjacency

**Phase 1 Solution**:
- Add file-patterns to test-location
- Reuse existing pattern matching
- Backward compatible
- ~50 lines of code

**Future Enhancements**:
- Phase 2: Per-language configs
- Phase 3: Auto-detection

**Recommendation**: Implement Phase 1 immediately for quick win
