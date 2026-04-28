# Plan: Sub-package internal/rules/

## Current State
- `internal/rules/` has 65 files in a flat package
- Core types: `rule.go` (Rule, Violation, AutoFix), `registry.go` (RuleContext, RuleFactory, Register, GetFactory)
- ~30 rule implementation files
- External consumers: `linter/`, `autofix/`, `output/`, `plugin/`

## Target Structure
```
internal/rules/
├── rule.go              # Rule interface, Violation, AutoFix (stays)
├── registry.go          # RuleContext, RuleFactory, Register, GetFactory (stays)
├── structure/           # Filesystem rules
│   ├── max_depth.go
│   ├── max_files.go
│   ├── max_subdirs.go
│   ├── naming_convention.go
│   ├── file_existence.go
│   ├── regex_match.go
│   └── disallowed_patterns.go
├── graph/               # Architecture rules
│   ├── layer_boundaries.go
│   ├── orphaned_files.go
│   ├── unused_exports.go
│   ├── path_based_layers.go
│   └── property_enforcement.go
├── quality/             # Metrics rules
│   ├── max_cognitive_complexity.go
│   └── max_halstead_effort.go
├── content/             # File content rules
│   └── file_content.go
├── ci/                  # CI/CD rules
│   ├── github_workflows.go
│   ├── linter_config.go
│   ├── openapi_asyncapi.go
│   ├── contract_framework.go
│   └── spec_adr.go
└── predicate/           # (already exists)
```

## Import Chain (No Circular Dependencies)
```
rules (base types) ← rules/structure (imports rules for Rule, Violation)
rules (base types) ← rules/graph (imports rules for Rule, Violation)
rules (base types) ← linter (imports rules for Rule, Violation)
rules/structure ← linter/factory (imports for constructors)
```

## Execution Steps

### Step 1: Create sub-package directories
- `internal/rules/structure/`
- `internal/rules/graph/`
- `internal/rules/quality/`
- `internal/rules/content/`
- `internal/rules/ci/`

### Step 2: Move rule files (one sub-package at a time)
For each file:
1. Copy file to sub-package
2. Change `package rules` to `package <subpackage>`
3. Add `import "github.com/Jonathangadeaharder/structurelint/internal/rules"`
4. Prefix all base types with `rules.` (Rule → rules.Rule, Violation → rules.Violation, etc.)
5. Delete original file

### Step 3: Update external consumers
- `linter/factory.go`: Add imports for sub-packages, update constructor calls
- `linter/linter.go`: No changes needed (uses rules.Rule interface)
- `autofix/`, `output/`, `plugin/`: No changes needed (use rules.Violation)

### Step 4: Move registrations to sub-packages
Each sub-package gets `init.go` that registers its rules via `rules.Register()`

### Step 5: Update tests
Move test files to sub-packages, update imports

### Step 6: Verify
- `go build ./...`
- `go test ./...`
