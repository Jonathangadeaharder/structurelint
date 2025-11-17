# Mixed Language Project Support: Design Plan

## Current State Analysis

### Problem Statement

Structurelint currently has limited support for mixed-language projects (e.g., Go + Python). The `test-location` rule applies uniform logic to all detected test files across languages, leading to false positives.

**Example Issue:**
- Python test file: `tests/test_clone_detection_basic.py`
- Detected as test file via `test_` pattern (line 108 in location_validation.go)
- Go-specific validation applied (expecting Go conventions)
- Result: Violation despite being correctly located in integration test directory
- Workaround: Manual exemption `tests/**/*.py` in config

### Root Causes

1. **Language-Agnostic Test Detection**: `isTestFile()` detects tests across multiple languages but doesn't track which language
2. **No File Pattern Scoping**: Unlike `test-adjacency` (which has `FilePatterns`), `test-location` checks ALL detected test files
3. **Hardcoded Patterns**: Test patterns are hardcoded in Go code rather than configurable per-language
4. **Exemption-Based Workarounds**: Current solution relies on manual exemptions rather than proper language awareness

### Current Architecture

```go
// test-adjacency (GOOD - has file pattern scoping)
type TestAdjacencyRule struct {
    Pattern      string
    TestDir      string
    FilePatterns []string  // ✓ Scopes which files to check
    Exemptions   []string
}

// test-location (NEEDS IMPROVEMENT - no file pattern scoping)
type TestLocationRule struct {
    IntegrationTestDir string
    AllowAdjacent      bool
    Exemptions         []string  // ✗ Only manual exemptions
}
```

## Proposed Architecture

### Option 1: Add File Pattern Scoping (Recommended - Quick Win)

**Goal**: Make `test-location` consistent with `test-adjacency` by adding file pattern filtering.

**Changes**:
```go
type TestLocationRule struct {
    IntegrationTestDir string
    AllowAdjacent      bool
    FilePatterns       []string  // NEW: Only check these patterns
    Exemptions         []string
}
```

**Config**:
```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns:  # NEW: Explicitly scope to Go files
    - "**/*_test.go"
  exemptions:
    - "testdata/**"
    - "internal/rules/rules_test.go"
```

**Pros**:
- ✓ Small, focused change
- ✓ Consistent with test-adjacency
- ✓ Backward compatible (file-patterns optional, defaults to all)
- ✓ Solves immediate problem cleanly

**Cons**:
- Still requires manual configuration per language
- Doesn't auto-detect language conventions

### Option 2: Per-Language Configuration (Future Enhancement)

**Goal**: Support multiple language-specific configurations in a single rule.

**Config**:
```yaml
test-location:
  languages:
    go:
      file-patterns: ["**/*_test.go"]
      integration-test-dir: "tests"
      allow-adjacent: true
    python:
      file-patterns: ["**/test_*.py", "**/*_test.py"]
      integration-test-dir: "tests"
      allow-adjacent: false  # Python convention: centralized tests
    typescript:
      file-patterns: ["**/*.test.ts", "**/*.spec.ts"]
      integration-test-dir: "__tests__"
      allow-adjacent: true
```

**Pros**:
- ✓ Explicit, self-documenting configuration
- ✓ Different rules per language
- ✓ Supports language-specific conventions

**Cons**:
- Requires more complex config parsing
- Backward incompatible without migration logic
- Verbose configuration

### Option 3: Language Auto-Detection (Advanced)

**Goal**: Automatically detect primary language and exempt others.

**Changes**:
```go
type TestLocationRule struct {
    IntegrationTestDir string
    AllowAdjacent      bool
    PrimaryLanguage    string     // NEW: "go", "python", etc.
    AutoExemptOthers   bool       // NEW: Auto-exempt non-primary languages
    FilePatterns       []string
    Exemptions         []string
}
```

**Pros**:
- ✓ Minimal configuration
- ✓ Works automatically for most projects

**Cons**:
- Complex heuristics needed
- May not work for truly polyglot projects
- Less explicit/predictable

## Recommended Implementation Plan

### Phase 1: File Pattern Scoping (Immediate - Low Risk)

**Timeline**: 1 iteration

**Tasks**:
1. Add `FilePatterns []string` field to `TestLocationRule`
2. Add `matchesFilePattern()` method (reuse from test-adjacency)
3. Update `Check()` to skip files not matching patterns
4. Update config loading in `linter.go`
5. Update `.structurelint.yml` with explicit Go patterns
6. Add tests for pattern matching
7. Update documentation

**Backward Compatibility**:
- If `file-patterns` is empty/nil → check all test files (current behavior)
- If `file-patterns` is specified → only check matching files

**Example Migration**:
```yaml
# Before (implicit - checks all test files)
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true

# After (explicit - only Go tests)
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  # NEW - scope to Go
```

### Phase 2: Per-Language Configuration (Future - Medium Risk)

**Timeline**: 2-3 iterations

**Prerequisites**: Phase 1 complete

**Tasks**:
1. Design multi-language config schema
2. Create migration path from Phase 1 config
3. Update struct to support language-specific settings
4. Implement language detection utilities
5. Add comprehensive tests for all supported languages
6. Update `structurelint init` to detect multiple languages
7. Update documentation with multi-language examples

**Backward Compatibility**:
- Support both old single-language and new multi-language config
- Auto-migrate on first run with warning

### Phase 3: Convention Detection (Future - High Risk)

**Timeline**: 3-5 iterations

**Prerequisites**: Phase 2 complete

**Tasks**:
1. Build language convention database (test patterns, locations, etc.)
2. Implement heuristic detection algorithms
3. Add confidence scoring
4. Create override mechanisms for edge cases
5. Extensive testing across diverse codebases
6. Add telemetry/feedback mechanism

**Backward Compatibility**:
- Opt-in via `auto-detect: true` flag
- Falls back to explicit config if detection fails

## Implementation Details: Phase 1

### Code Changes

**1. Update TestLocationRule struct** (`internal/rules/location_validation.go`):
```go
type TestLocationRule struct {
    IntegrationTestDir string
    AllowAdjacent      bool
    FilePatterns       []string  // NEW
    Exemptions         []string
}
```

**2. Add pattern matching** (`internal/rules/location_validation.go`):
```go
// matchesFilePattern checks if a file matches any of the configured patterns
func (r *TestLocationRule) matchesFilePattern(path string) bool {
    // If no patterns specified, match all files (backward compatible)
    if len(r.FilePatterns) == 0 {
        return true
    }

    for _, pattern := range r.FilePatterns {
        if matchesGlobPattern(path, pattern) {
            return true
        }
    }
    return false
}
```

**3. Update Check() method** (`internal/rules/location_validation.go:52`):
```go
// Check each test file
for _, file := range files {
    if file.IsDir {
        continue
    }

    if !r.isTestFile(file.Path) {
        continue
    }

    // NEW: Skip if file doesn't match configured patterns
    if !r.matchesFilePattern(file.Path) {
        continue
    }

    if r.isExempted(file.Path) {
        continue
    }

    // ... rest of existing logic
}
```

**4. Update config loading** (`internal/linter/linter.go:191-197`):
```go
if testLoc, ok := l.getRuleConfig("test-location"); ok {
    if locMap, ok := testLoc.(map[string]interface{}); ok {
        integrationDir := l.getStringFromMap(locMap, "integration-test-dir")
        allowAdjacent := l.getBoolFromMap(locMap, "allow-adjacent")
        filePatterns := l.getStringSliceFromMap(locMap, "file-patterns")  // NEW
        exemptions := l.getStringSliceFromMap(locMap, "exemptions")

        *rulesList = append(*rulesList, rules.NewTestLocationRule(
            integrationDir,
            allowAdjacent,
            filePatterns,  // NEW parameter
            exemptions,
        ))
    }
}
```

**5. Update constructor** (`internal/rules/location_validation.go:179-186`):
```go
func NewTestLocationRule(integrationTestDir string, allowAdjacent bool, filePatterns []string, exemptions []string) *TestLocationRule {
    return &TestLocationRule{
        IntegrationTestDir: integrationTestDir,
        AllowAdjacent:      allowAdjacent,
        FilePatterns:       filePatterns,  // NEW
        Exemptions:         exemptions,
    }
}
```

### Config Changes

**Update `.structurelint.yml`**:
```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns:  # NEW - explicitly scope to Go test files
    - "**/*_test.go"
  exemptions:
    - "testdata/**"
    - "internal/rules/rules_test.go"
    # Remove: - "tests/**/*.py"  # No longer needed!
```

### Testing Strategy

**Unit Tests** (`internal/rules/location_validation_test.go`):
```go
func TestTestLocationRule_FilePatterns_Go(t *testing.T) {
    // Arrange
    rule := NewTestLocationRule("tests", true, []string{"**/*_test.go"}, nil)

    files := []walker.FileInfo{
        {Path: "internal/foo_test.go", IsDir: false},     // Go test
        {Path: "tests/test_bar.py", IsDir: false},        // Python test
    }

    // Act
    violations := rule.Check(files, nil)

    // Assert
    // Should only check foo_test.go, not test_bar.py
    assert.Equal(t, 1, len(violations))
    assert.Contains(t, violations[0].Path, "foo_test.go")
}

func TestTestLocationRule_FilePatterns_Empty_BackwardCompatible(t *testing.T) {
    // Arrange - no file patterns (backward compatible)
    rule := NewTestLocationRule("tests", true, []string{}, nil)

    // Should check all test files like before
}

func TestTestLocationRule_FilePatterns_MultipleLanguages(t *testing.T) {
    // Arrange - multiple patterns
    rule := NewTestLocationRule("tests", true, []string{
        "**/*_test.go",
        "**/*.test.ts",
    }, nil)

    // Should check both Go and TypeScript tests
}
```

### Documentation Updates

**1. Update README.md** - Add Phase 3 section:
```markdown
### Phase 3: Test Location Validation

The `test-location` rule ensures test files are in appropriate locations.

**Configuration**:
- `integration-test-dir`: Directory for integration tests
- `allow-adjacent`: Allow tests adjacent to source code
- `file-patterns`: Glob patterns to scope which files to check (optional)
- `exemptions`: Patterns to exempt from validation

**Example** (Go project):
```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  # Only check Go tests
```

**Example** (Mixed Go + Python):
```yaml
test-location:
  integration-test-dir: "tests"
  allow-adjacent: true
  file-patterns: ["**/*_test.go"]  # Scope to Go; Python tests auto-ignored
```
```

**2. Create migration guide** (`docs/MIXED_LANGUAGE_MIGRATION.md`):
- How to update from exemption-based to pattern-based config
- Examples for common language combinations
- Troubleshooting guide

## Benefits of This Approach

### Immediate (Phase 1)
1. **Cleaner Config**: Remove manual exemptions like `tests/**/*.py`
2. **Explicit Intent**: `file-patterns` clearly documents what the rule checks
3. **Consistency**: Aligns with `test-adjacency` pattern
4. **Maintainability**: Easier to understand and modify

### Future (Phase 2+)
1. **Multi-Language Projects**: Native support for polyglot codebases
2. **Convention-Aware**: Auto-detect and apply language-specific best practices
3. **Extensibility**: Easy to add new languages without code changes
4. **User Experience**: `structurelint init` can generate optimal configs

## Alternative Considered: Exclude Patterns

**Why Not Just Add More Exemptions?**

Current workaround:
```yaml
test-location:
  exemptions:
    - "tests/**/*.py"
    - "tests/**/*.js"
    - "tests/**/*.ts"
    - "**/*.test.ts"
    - "**/*.spec.ts"
```

**Problems**:
1. **Negative Logic**: Harder to understand what IS checked
2. **Maintenance Burden**: Must update for each new language
3. **Error Prone**: Easy to miss patterns
4. **No Intent**: Doesn't communicate "this rule is for Go"
5. **Inconsistent**: Different approach than test-adjacency

**File Patterns Approach** (Better):
```yaml
test-location:
  file-patterns: ["**/*_test.go"]  # Clear: checks Go tests only
```

## Success Metrics

### Phase 1
- [ ] Zero manual exemptions needed for language-specific tests
- [ ] Config clearly documents which languages are validated
- [ ] No regressions in existing Go-only projects
- [ ] Test coverage >90% for new pattern matching logic

### Phase 2
- [ ] Support at least 5 languages (Go, Python, TypeScript, Java, Rust)
- [ ] Migration from Phase 1 is automatic or single-command
- [ ] Documentation includes examples for all supported languages

### Phase 3
- [ ] `structurelint init` correctly detects multi-language projects
- [ ] <5% false positive rate across diverse codebases
- [ ] Positive user feedback on auto-detection accuracy

## Next Steps

1. **Review this plan** with stakeholders
2. **Prototype Phase 1** in feature branch
3. **Test on real mixed-language projects** (structurelint itself!)
4. **Document migration path** for existing users
5. **Release Phase 1** as minor version (backward compatible)
6. **Gather feedback** before proceeding to Phase 2

## References

- Current implementation: `internal/rules/location_validation.go`
- Similar pattern: `internal/rules/adjacency_validation.go` (has FilePatterns)
- Config loading: `internal/linter/linter.go:191-197`
- Issue tracking: This session's mixed-language challenges
