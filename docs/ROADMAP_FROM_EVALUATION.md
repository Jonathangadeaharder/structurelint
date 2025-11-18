# Roadmap Based on Comprehensive Evaluation

This document summarizes actionable items from the [Comprehensive Evaluation](EVALUATION.md) of structurelint across diverse codebases.

## Executive Summary

The evaluation analyzed structurelint's effectiveness across multiple project types:
- ✅ **Strong**: Phase 0 (filesystem hygiene) universally adopted and effective
- ⚠️ **Moderate**: Phase 1 (layer boundaries) powerful but high configuration friction
- ⚠️ **Moderate**: Phase 2 (dead code detection) valuable but noisy due to entry point detection issues
- 🔴 **Weak**: Polyglot support requires brittle path-based overrides
- 🔴 **Weak**: No support for test-specific metrics or infrastructure code profiles

## Priority 1: Quick Wins (High Impact, Low Effort)

### 1.1 Auto-Discovery of Gitignore
**Problem**: Users manually list `node_modules`, `bin`, `obj`, `.git` in exemptions, duplicating `.gitignore`

**Solution**: Automatically parse `.gitignore` and populate global exemption list

**Impact**: Reduces configuration boilerplate by 50%+ for most projects

**Implementation**:
```go
// internal/config/config.go
func (c *Config) LoadGitignore(rootDir string) error {
    // Parse .gitignore
    // Merge with existing exclude patterns
}
```

### 1.2 Entry Point Patterns
**Problem**: `disallow-orphaned-files` produces false positives for entry points (`main.py`, `manage.py`, CLI scripts)

**Current Workaround**: Manually exempt every entry point, or disable the rule entirely

**Solution**: Add `entry-point-patterns` configuration
```yaml
rules:
  disallow-orphaned-files:
    enabled: true
    entry-point-patterns:
      - "**/main.py"
      - "**/index.ts"
      - "**/*_test.go"
      - "**/manage.py"
      - "scripts/**"
```

**Impact**: Makes Phase 2 viable for polyglot projects without massive whitelist maintenance

### 1.3 Test-Specific Metric Profiles
**Problem**: Projects disable complexity metrics entirely for tests, allowing "test rot"

**Current State**: Binary choice—enforce same metrics for tests (too strict) or disable (too lenient)

**Solution**: Built-in test profile with adjusted thresholds
```yaml
profiles:
  test:
    max-cognitive-complexity: { max: 50 }  # Tests are linear but long
    max-files-in-dir: { max: 40 }          # Test suites cluster
    max-halstead-effort: { max: 200000 }   # Tests repeat patterns
    disallow-orphaned-files: false         # Tests are executed, not imported

rules:
  # Automatically applied to **/*_test.go, **/*.test.ts, **/test_*.py
  apply-test-profile: true
```

**Impact**: Prevents test complexity explosions while reducing configuration burden

## Priority 2: Polyglot Support (High Impact, Medium Effort)

### 2.1 Language Auto-Detection and Scoping
**Problem**: Moving `frontend/` from `src/frontend` to root breaks path-based naming rules

**Current Approach**: Brittle path patterns
```yaml
naming-convention:
  "src/frontend/components/**": "PascalCase"  # Breaks if frontend/ moves
```

**Solution**: Detect language boundaries via manifest files
```yaml
naming-convention:
  language-scoped:
    python: "snake_case"
    typescript: "camelCase"
    react-components: "PascalCase"  # Auto-detected via package.json + "react" dependency

# Auto-detection:
# - package.json + "react" → React project
# - go.mod → Go module
# - pyproject.toml / setup.py → Python package
```

**Impact**: Configuration survives directory restructuring; reduces polyglot friction

### 2.2 Uniqueness Constraints
**Problem**: LangPlug has `vocabulary_service.py` AND `vocabulary_service_clean.py` (dual implementation anti-pattern)

**Current State**: No detection

**Solution**: Singleton pattern rules
```yaml
rules:
  unique-file-pattern:
    "src/services/**/*_service.py": "singleton"  # Only ONE *_service.py per directory
    "src/repositories/**/*_repository.go": "singleton"
```

**Impact**: Prevents architectural violations (duplicate implementations)

### 2.3 Infrastructure Code Profiles
**Problem**: Teams broadly exempt `.github/**`, `docker/**`, `k8s/**` because app rules don't apply

**Missed Opportunity**: Infrastructure code NEEDS linting (prevent copy-paste errors, ensure consistency)

**Solution**: Infrastructure profile
```yaml
profiles:
  infrastructure:
    max-depth: { max: 10 }           # CI workflows can be nested
    naming-convention: "kebab-case"  # YAML convention
    require-documentation: true      # Force comments in complex workflows
    # Disable code metrics (not applicable to YAML)
    max-cognitive-complexity: 0

# Auto-apply to .github/workflows, docker/, k8s/, terraform/
```

**Impact**: Extends structural governance to the entire codebase (not just application code)

## Priority 3: Advanced Features (High Impact, High Effort)

### 3.1 Fractal Configuration
**Problem**: Cannot enforce Atomic Design rule "molecules cannot import organisms" without brittle global config

**Current Limitation**: Single `.structurelint.yml` at root; rules are global or path-based

**Solution**: Allow `.structurelint.yml` in subdirectories
```
src/
  components/
    .structurelint.yml  # Local rules for components subtree
    atoms/
    molecules/
    organisms/
```

**Example local config**:
```yaml
# src/components/.structurelint.yml
inherit: true  # Inherit from parent

layers:
  - name: atoms
    path: "atoms/**"
    dependsOn: []

  - name: molecules
    path: "molecules/**"
    dependsOn: ["atoms"]

  - name: organisms
    path: "organisms/**"
    dependsOn: ["atoms", "molecules"]
```

**Impact**: Enables enforceable Atomic Design, Feature-Sliced Design, domain-driven design

### 3.2 Relative Import Topology Rules
**Problem**: Cannot express "sibling directories cannot import each other" or "children cannot import parents"

**Current State**: Only absolute layer definitions

**Solution**: Relative import rules
```yaml
rules:
  import-topology:
    sibling-imports: "disallow"    # src/features/auth cannot import src/features/billing
    parent-imports: "disallow"     # src/domain/user/entity.ts cannot import from src/domain/
    child-imports: "allow"         # src/ can import src/domain/
```

**Impact**: Prevents circular dependencies and architectural drift in deeply nested structures

### 3.3 Content Pattern Validation
**Problem**: Cannot enforce "every `__init__.py` must export symbols" or "every `src/` file starts with license header"

**Solution**: Extend `file-content` rule
```yaml
rules:
  file-content:
    patterns:
      "**/__init__.py":
        must-contain: "^(from|import)"  # Must have at least one import/export
        forbid-empty: true

      "src/**/*.go":
        must-start-with: "// Copyright"  # License header
```

**Impact**: Enforces semantic correctness, not just structural shape

## Priority 4: Developer Experience (Medium Impact, Low-Medium Effort)

### 4.1 Better Error Messages for Polyglot Projects
**Problem**: "naming-convention violated" doesn't explain WHICH rule or WHY

**Current Output**:
```
src/components/Button/button.tsx: naming convention violated
```

**Improved Output**:
```
src/components/Button/button.tsx: naming convention violated
  Expected: PascalCase (React components rule: "src/components/**")
  Actual: camelCase ("button")
  Suggestion: Rename to "Button.tsx" or exclude with override
```

### 4.2 Interactive Configuration Mode
**Problem**: `--init` generates config, but users struggle to customize for complex scenarios

**Solution**: Interactive config wizard
```bash
structurelint --init --interactive

? Detected Python and TypeScript. Test strategy?
  > Separate test directories (Python: tests/, TS: __tests__/)
    Adjacent tests (Go-style: file.test.ts next to file.ts)
    Custom pattern

? Enforce architectural layers?
  > No (default)
    Clean Architecture
    Hexagonal Architecture
    Feature-Sliced Design
    Custom

? Complexity enforcement level?
    Strict (CoC: 10, Halstead: 50k)
  > Moderate (CoC: 15, Halstead: 100k)
    Lenient (CoC: 25, Halstead: 200k)
```

## Metrics for Success

After implementing these improvements, measure:

1. **Configuration Size**: Reduce median `.structurelint.yml` size by 40%
2. **False Positive Rate**: Reduce orphan detection false positives from ~60% to <10%
3. **Polyglot Adoption**: Increase adoption in multi-language repos from ~30% to >70%
4. **Rule Enablement**: Increase projects with Phase 2 enabled from ~25% to >60%

## Implementation Phases

### Phase 1 (Next Release - v2.1)
- Auto-discovery of `.gitignore` ✅ (Priority 1.1)
- Entry point patterns ✅ (Priority 1.2)
- Test-specific profiles ✅ (Priority 1.3)

### Phase 2 (v2.2)
- Language auto-detection (Priority 2.1)
- Uniqueness constraints (Priority 2.2)
- Infrastructure profiles (Priority 2.3)

### Phase 3 (v3.0 - Major Release)
- Fractal configuration (Priority 3.1)
- Relative import topology (Priority 3.2)
- Content pattern validation (Priority 3.3)

### Phase 4 (v3.1+)
- Enhanced error messages (Priority 4.1)
- Interactive configuration (Priority 4.2)

## Conclusion

The evaluation reveals structurelint is a **powerful foundation** with **high adoption friction** in real-world polyglot scenarios. The roadmap prioritizes:

1. **Reducing friction** (Priority 1) - Makes current features usable
2. **Polyglot support** (Priority 2) - Expands addressable market
3. **Advanced features** (Priority 3) - Enables complex architectures
4. **Developer experience** (Priority 4) - Improves daily usage

By implementing these improvements incrementally, structurelint can evolve from a "File Linter" to a true "Architectural Guardian" for modern codebases.

---

**Based On**: [Comprehensive Evaluation](EVALUATION.md)
**Target Audience**: Maintainers, contributors, roadmap planning
**Status**: Proposed - Pending Community Feedback
