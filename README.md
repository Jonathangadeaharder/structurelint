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

**Phase 1 - Architectural Layer Enforcement:**
- **Import graph analysis** - Parse source files to build dependency graphs
- **Layer boundary validation** - Enforce architectural patterns (Clean Architecture, Hexagonal, Feature-Sliced Design, etc.)
- **Dependency rules** - Prevent violations like "domain importing from presentation"

**Phase 2 - Dead Code Detection:**
- **Orphaned file detection** - Find files never imported by other files
- **Unused export identification** - Locate dead exports that can be removed

**Phase 3 - Test Validation:** ✨ NEW
- **Test adjacency enforcement** - Ensure every source file has corresponding tests
- **Test location validation** - Prevent orphaned tests and enforce test directory structure
- **Multi-language support** - Python, Go, TypeScript, JavaScript, Java, Rust, Ruby, C/C++

**Phase 4 - File Content Templates:** ✨ NEW
- **Template system** - Define required file structures (READMEs, design docs, etc.)
- **Section validation** - Ensure documentation has required sections
- **Pattern enforcement** - Require or forbid specific content patterns

## Features

- **Fast**: Written in Go for blazing-fast performance, suitable for pre-commit hooks
- **Cascading Configuration**: ESLint-style `.structurelint.yml` files with inheritance
- **Flexible Rules**: From simple metrics to complex pattern matching
- **Architectural Enforcement**: Layer boundaries and import graph validation
- **Multi-Language Support**: TypeScript, JavaScript, Go, Python
- **Zero Dependencies**: Single binary, easy to install and distribute

## Installation

### Go Install (Recommended)

```bash
go install github.com/structurelint/structurelint/cmd/structurelint@latest
```

### Download Binary

Download pre-built binaries from the [releases page](https://github.com/structurelint/structurelint/releases):

```bash
# Linux (amd64)
curl -L https://github.com/structurelint/structurelint/releases/latest/download/structurelint-linux-amd64 -o structurelint
chmod +x structurelint
sudo mv structurelint /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/structurelint/structurelint/releases/latest/download/structurelint-darwin-arm64 -o structurelint
chmod +x structurelint
sudo mv structurelint /usr/local/bin/
```

### Build from Source

```bash
git clone https://github.com/structurelint/structurelint.git
cd structurelint
go build -o structurelint ./cmd/structurelint
```

## Quick Start

### Option 1: Automatic Configuration (Recommended)

Let structurelint analyze your project and generate configuration automatically:

```bash
# Analyze your project and create .structurelint.yml
structurelint --init

# Review and customize the generated config
# Then run the linter
structurelint .
```

The `--init` command automatically detects:
- Programming languages (Python, Go, TypeScript, Java, etc.)
- Test patterns (adjacent tests vs separate test directories)
- Project structure metrics
- Documentation style

### Option 2: Manual Configuration

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
structurelint .
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

## Phase 5: Evidence-Based Software Quality Metrics ✨ NEW

### Overview

Phase 5 adds **scientifically-validated software quality metrics** based on systematic literature reviews, meta-analyses, and neuroscience research. This framework moves beyond traditional metrics like Cyclomatic Complexity to provide better predictors of code quality, maintainability, and defect-proneness.

### Why Evidence-Based Metrics?

Traditional metrics have significant limitations:

**Cyclomatic Complexity (CC) - The Problem**:
- ❌ Weak predictor of maintainability (mathematical model is "unsatisfactory")
- ❌ Deviates from human perception (EEG studies show poor correlation with cognitive load)
- ❌ Often outperformed by simple Lines of Code (LOC) in defect prediction
- ❌ Treats `switch` with 20 cases (easy to read) same as 20 nested `if` statements (hard to read)

**Evidence-Based Alternatives**:
- ✅ **Cognitive Complexity**: r=0.54 correlation with comprehension time (meta-analysis)
- ✅ **Halstead Effort**: rs=0.901 correlation with measured brain activity (EEG study)
- ✅ Combined metrics provide complete picture of code complexity

### Implemented Metrics

#### 1. Cognitive Complexity (CoC)

**Evidence Level**: Meta-analysis of 14 studies
**Correlation**: r=0.54 with comprehension time, r=-0.29 with subjective difficulty

**Why Superior to Cyclomatic Complexity**:
- Penalizes nesting (matches exponential increase in human cognitive load)
- Ignores shorthand operators that improve readability (`&&`, `||`, `?:`)
- Based on human assessment, not mathematical models

**Calculation Rules**:
```
1. Base complexity = 0 (not 1 like CC)
2. +1 for each flow break: if, for, while, catch, switch, goto
3. +1 additional for each level of nesting
4. No penalty for shorthand operators in sequence
```

**Example**:
```go
// Cyclomatic Complexity = 4
// Cognitive Complexity = 7
func processItems(items []Item) {
    for _, item := range items {        // +1 (for) = 1
        if item.IsActive {              // +2 (+1 for if, +1 for nesting) = 3
            if item.HasPermission {     // +3 (+1 for if, +2 for nesting) = 6
                process(item)
            }
        }
    }
}

// Note: A switch with 20 cases would have CC=20, CoC=21 (1 for switch + 20 for cases)
// but nested ifs are much harder to understand due to nesting penalties
```

#### 2. Halstead Metrics

**Evidence Level**: Neuroscience (EEG study)
**Correlation**: rs=0.901 with measured cognitive load

**Why Critical**:
- Captures **data complexity** (vocabulary, operators, operands)
- **Complements** Cognitive Complexity (which captures control-flow)
- Highest correlation with actual brain activity during code comprehension

**Metrics Calculated**:
```
n1 = distinct operators (if, +, =, func, etc.)
n2 = distinct operands (variables, constants)
N1 = total operators
N2 = total operands

Volume (V) = N × log₂(n)              // Information content in bits
Difficulty (D) = (n1/2) × (N2/n2)     // How hard to write/understand
Effort (E) = D × V                     // Mental effort required ⭐ PRIMARY METRIC
```

### Configuration

#### Replace Cyclomatic Complexity with Evidence-Based Metrics

```yaml
root: true

rules:
  # DEPRECATED: Traditional Cyclomatic Complexity
  max-cyclomatic-complexity: 0  # Disable (deprecated)

  # RECOMMENDED: Evidence-Based Metrics
  max-cognitive-complexity:
    max: 15
    file-patterns:
      - "**/*.go"
      - "**/*.ts"
      - "**/*.py"

  max-halstead-effort:
    max: 100000
    file-patterns:
      - "**/*.go"
```

### Thresholds and Interpretation

#### Cognitive Complexity Thresholds
- **0-5**: Simple, easy to understand ✅
- **6-10**: Moderate complexity, acceptable ⚠️
- **11-15**: High complexity, consider refactoring 🔶
- **16-25**: Very high complexity, should refactor 🔴
- **26+**: Extremely complex, high maintenance risk 🚨

#### Halstead Effort Thresholds
- **0-10,000**: Low effort ✅
- **10,000-50,000**: Moderate effort ⚠️
- **50,000-100,000**: High effort 🔶
- **100,000+**: Very high effort, high cognitive load 🚨

### Example Configurations

#### Evidence-Based Go Project
```yaml
root: true

rules:
  # Replace CC with Cognitive Complexity
  max-cyclomatic-complexity: 0  # Disabled
  max-cognitive-complexity:
    max: 15
    file-patterns: ["**/*.go"]

  # Add Halstead for data complexity
  max-halstead-effort:
    max: 100000
    file-patterns: ["**/*.go"]
```

See complete examples:
- `examples/evidence-based-go.yml`
- `examples/evidence-based-typescript.yml`

### Metric Comparison Table

| Metric | Evidence Level | Use Case | Correlation | Status |
|--------|---------------|----------|-------------|--------|
| **Cognitive Complexity** | Meta-analysis | Understandability | r=0.54 with time | ✅ Recommended |
| **Halstead Effort** | EEG Study | Cognitive Load | rs=0.901 with brain | ✅ Recommended |
| Cyclomatic Complexity | Outdated | Testing Paths | Often < LOC | ⚠️ Deprecated |
| Lines of Code | Strong | Size Baseline | Strong predictor | ✅ Use as control |

### Scientific Evidence

**Cognitive Complexity**:
- Schnappinger et al. (2020). "Meta-Analysis of Cognitive Complexity"
- Finding: r=0.54 correlation with comprehension time across 14 studies
- Conclusion: "First validated code-based metric reflecting code understandability"

**Halstead Effort**:
- Scalabrino et al. (2022). "EEG Study on Code Complexity Metrics"
- Finding: rs=0.901 correlation with measured cognitive load (brain activity)
- Conclusion: CC-based metrics "deviate considerably," Halstead captures data complexity

**Why Both Metrics Are Needed**:
- **Low CoC + High Halstead**: Data-flow nightmare (complex state, many variables)
- **High CoC + Low Halstead**: Control-flow nightmare (deep nesting, conditionals)
- **Both Required**: Complete picture of cognitive complexity

### Future Enhancements (Phase 6+)

**CK Suite (Object-Oriented Metrics)**:
- ✅ Evidence Level: Multiple SLRs, 2023 large-scale study
- 🔮 CBO (Coupling Between Objects): Strong defect predictor
- 🔮 RFC (Response For a Class): Interaction complexity
- 🔮 LCOM5 (Lack of Cohesion): "Among highest-performing metrics" (2023)

**Process Metrics** (Strongest Predictors):
- ✅ Evidence Level: SLR - "Overall effectively better than static code attributes"
- 🔮 Code Churn: Lines added/deleted/modified
- 🔮 Revision Count: Number of commits
- 🔮 Bug Fix Count: Historical defect-proneness
- 🔮 Developer Count: Ownership diffusion

**Statistical Framework**:
- 🔮 Multivariate logistic regression
- 🔮 LOC confounding variable control
- 🔮 Project-specific feature selection
- 🔮 Defect probability prediction

## Roadmap

### Phase 0 - Core Filesystem Linting ✅ COMPLETE
- ✅ Metric rules (max-depth, max-files, max-subdirs)
- ✅ Naming conventions
- ✅ File existence validation
- ✅ Pattern matching and disallowing

### Phase 1 - Architectural Layer Enforcement ✅ COMPLETE
- ✅ Import graph analysis
- ✅ Layer boundary enforcement
- ✅ Dependency rules

### Phase 2 - Dead Code Detection ✅ COMPLETE
- ✅ Orphaned file detection
- ✅ Unused export identification
- ✅ Entrypoint configuration

### Phase 3 - Test Validation ✅ COMPLETE
- ✅ Test adjacency enforcement (adjacent and separate patterns)
- ✅ Test location validation
- ✅ Multi-language support (Python, Go, TypeScript, Java, Rust, Ruby, C/C++)
- ✅ Language-specific test naming conventions

### Phase 4 - File Content Templates ✅ COMPLETE
- ✅ Template system for file structure validation
- ✅ Section validation (required sections)
- ✅ Pattern enforcement (required/forbidden patterns)
- ✅ Content structure validation (must-start-with, must-end-with)

### Phase 5 - Evidence-Based Quality Metrics ✅ COMPLETE
- ✅ Cognitive Complexity (replaces Cyclomatic Complexity)
- ✅ Halstead Metrics (Volume, Difficulty, Effort)
- ✅ Scientific evidence documentation
- ✅ Example configurations

### Phase 6 - Automatic Configuration ✅ COMPLETE
- ✅ `--init` command for automatic configuration generation
- ✅ Language detection (8+ languages)
- ✅ Test pattern recognition
- ✅ Smart defaults based on project structure
- ✅ Project metrics analysis

### Future Enhancements
- 🔮 CK Suite metrics (CBO, RFC, LCOM5) for OO languages
- 🔮 Process metrics from Git history (churn, revisions, bug fixes)
- 🔮 Statistical framework with LOC control
- 🔮 Multivariate defect prediction models
- 🔮 Monorepo support with per-package configurations
- 🔮 Framework-specific detection (pytest, Jest, JUnit)
- 🔮 Integration test directory detection
- 🔮 Compiler plugin system for non-standard files
- 🔮 Advanced dead code detection with call graph analysis

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

## Phase 3: Test Validation

### Overview

Phase 3 ensures comprehensive test coverage by validating that every source file has corresponding tests and that test files are properly organized.

### Features

**1. Test Adjacency Enforcement**
- Validates that source files have corresponding test files
- Supports both "adjacent" and "separate" test patterns
- Language-specific test file naming (e.g., `_test.go`, `.test.ts`, `test_*.py`)

**2. Test Location Validation**
- Prevents orphaned tests (tests without corresponding source files)
- Enforces proper test directory structure
- Supports integration test directories

**3. Self-Documenting Exemptions with `@structurelint:no-test`** ✨ NEW
- Declare test exemptions directly in source code with `// @structurelint:no-test <reason>`
- Self-documenting: reason for no tests is visible in the code
- Consistency validation: warns if file claims "no test needed" but has a test file
- Reduces need for long exemption lists in configuration

### Configuration

#### Adjacent Test Pattern

For projects where tests live next to source files (Go, TypeScript):

```yaml
rules:
  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "**/*.go"
      - "**/*.ts"
    exemptions:
      - "cmd/**/*.go"      # Entry points don't need tests
      - "**/*_gen.go"      # Generated files
      - "**/*.d.ts"        # Type definitions
```

#### Separate Test Pattern

For projects with dedicated test directories (Python, Java):

```yaml
rules:
  test-adjacency:
    pattern: "separate"
    test-dir: "tests"
    file-patterns:
      - "**/*.py"
    exemptions:
      - "**/__init__.py"  # Package initializers
      - "setup.py"         # Setup scripts
```

#### Test Location Validation

```yaml
rules:
  test-location:
    integration-test-dir: "tests"    # Directory for integration tests
    allow-adjacent: true              # Allow unit tests next to source
    exemptions:
      - "testdata/**"                 # Test fixtures
```

#### Using @structurelint:no-test Directive

Instead of adding files to exemption lists, declare exemptions directly in source code:

**Example - Interface definition (Go):**
```go
// Package rules defines the linting rule interface.
//
// @structurelint:no-test Interface definitions only, tested through implementations
package rules

type Rule interface {
    Name() string
    Check(files []FileInfo) []Violation
}
```

**Example - Simple utility (TypeScript):**
```typescript
// Re-export module for convenient imports
//
// @structurelint:no-test Simple re-export, tested via consuming code
package utils

export * from './helpers';
export * from './validators';
```

**Benefits over configuration exemptions:**
- ✅ Self-documenting: reason visible in code
- ✅ Consistency validation: warns if directive conflicts with test file existence
- ✅ Code review friendly: reviewers see justification
- ✅ Cleaner config: no long exemption lists

**See [NO_TEST_DIRECTIVE.md](docs/NO_TEST_DIRECTIVE.md) for complete documentation.**

### Example Violations

**Missing Test File (Adjacent Pattern):**
```
src/calculator.ts: missing test file (expected: src/calculator.test.ts)
```

**Orphaned Test File:**
```
tests/old-feature.test.ts: test file has no corresponding source file
```

**Test in Wrong Location:**
```
src/utils/helper.test.ts: test file should be in 'tests/' directory (separate pattern)
```

### Language-Specific Support

| Language | Adjacent Pattern | Separate Pattern | Test Naming |
|----------|-----------------|------------------|-------------|
| Go | ✅ Default | ✅ Supported | `*_test.go` |
| Python | ✅ Supported | ✅ Default | `test_*.py`, `*_test.py` |
| TypeScript/JS | ✅ Default | ✅ Supported | `*.test.ts`, `*.spec.js` |
| Java | ❌ | ✅ Default | `*Test.java`, `*IT.java` |
| Rust | ✅ Default | ✅ Supported | `*_test.rs` |
| Ruby | ❌ | ✅ Default | `*_spec.rb` |
| C/C++ | ✅ Supported | ✅ Supported | `test_*.cpp`, `*_test.cpp` |

### Complete Examples

#### Go Project with Adjacent Tests

```yaml
rules:
  test-adjacency:
    pattern: "adjacent"
    file-patterns:
      - "**/*.go"
    exemptions:
      - "cmd/**/*.go"
      - "**/*_gen.go"
      - "vendor/**"

  test-location:
    integration-test-dir: "tests"
    allow-adjacent: true
```

#### Python Project with Separate Tests

```yaml
rules:
  test-adjacency:
    pattern: "separate"
    test-dir: "tests"
    file-patterns:
      - "**/*.py"
    exemptions:
      - "**/__init__.py"
      - "**/conftest.py"
      - "setup.py"

  test-location:
    integration-test-dir: "tests"
    allow-adjacent: false
```

### Using --init for Test Configuration

The `--init` command automatically detects your test patterns:

```bash
$ structurelint --init

Analyzing project structure...

🔍 Project Analysis Summary
===========================

Languages Detected:
  [✓] python (42 files)
      Test pattern: separate
      Test directory: tests/

# Automatically generates appropriate test-adjacency config
```

See [docs/TEST_VALIDATION.md](docs/TEST_VALIDATION.md) for complete documentation.

## Phase 4: File Content Templates

### Overview

Phase 4 enables validation of file contents using templates. This ensures documentation and configuration files follow consistent structures.

### Features

**1. Section Validation**
- Require specific sections in markdown files (e.g., "## Overview", "## Installation")
- Enforce consistent documentation structure

**2. Pattern Matching**
- Require specific patterns (e.g., must start with heading)
- Forbid unwanted patterns (e.g., no TODO comments in production)

**3. Content Structure**
- Validate file must start/end with specific content
- Ensure proper formatting

**4. Test Pattern Enforcement**
- Enforce Arrange-Act-Assert (AAA) pattern in test files
- Improve test readability and consistency across teams
- Support for Go, TypeScript/JavaScript, Python tests

### Configuration

Define templates in `.structurelint/templates/`:

**.structurelint/templates/readme.yml:**
```yaml
# Template for README.md files
required-sections:
  - "# "              # Must have a main heading
  - "## Overview"     # Must have Overview section

required-patterns:
  - "^#\\s+\\w+"      # Must start with heading

must-start-with: "# " # Must start with main heading
```

**Reference templates in .structurelint.yml:**
```yaml
rules:
  file-content:
    templates:
      "**/README.md": "readme"
      "docs/design/*.md": "design-doc"
      "CONTRIBUTING.md": "contributing"
```

### Example Templates

#### README Template

```yaml
required-sections:
  - "# "
  - "## Overview"
  - "⬆️ **[Parent Directory]"  # Building lobby pattern

required-patterns:
  - "^#\\s+\\w+"                # Starts with heading
```

#### Design Document Template

```yaml
required-sections:
  - "# "
  - "## Problem Statement"
  - "## Proposed Solution"
  - "## Alternatives Considered"

forbidden-patterns:
  - "TODO"     # No TODOs in final design docs
  - "FIXME"
```

#### Contributing Guide Template

```yaml
required-sections:
  - "# Contributing"
  - "## Code of Conduct"
  - "## How to Contribute"
  - "## Development Setup"

must-end-with: "## License"
```

#### Test File Templates (AAA Pattern)

Enforce the Arrange-Act-Assert pattern for better test readability:

```yaml
rules:
  file-content:
    templates:
      # Go tests
      "**/*_test.go": "test-go"

      # TypeScript/JavaScript tests
      "**/*.test.ts": "test-typescript"
      "**/*.spec.js": "test-typescript"

      # Python tests
      "**/test_*.py": "test-python"
```

**Example compliant Go test:**
```go
func TestCalculator_Add_ReturnsSum(t *testing.T) {
    // Arrange
    calc := NewCalculator()
    a, b := 2, 3

    // Act
    result := calc.Add(a, b)

    // Assert
    assert.Equal(t, 5, result)
}
```

Available test templates:

**AAA Pattern** (structure only):
- `test-go.yml` - Lenient AAA enforcement for Go
- `test-typescript.yml` - Lenient AAA enforcement for TypeScript/JavaScript
- `test-python.yml` - Lenient AAA enforcement for Python
- `test-strict-aaa.yml` - Strict AAA enforcement (all languages)

**Given-When-Then** (naming + structure):
- `test-gwt-go.yml` - GWT naming + AAA for Go
- `test-gwt-typescript.yml` - GWT naming + AAA for TypeScript/JavaScript
- `test-gwt-python.yml` - GWT naming + AAA for Python
- `test-gwt-strict.yml` - Ultra-strict GWT + AAA (all languages)

See [docs/TEST_AAA_PATTERN.md](docs/TEST_AAA_PATTERN.md) and [docs/TEST_GWT_NAMING.md](docs/TEST_GWT_NAMING.md) for complete guides.

### Example Violations

**Missing Required Section:**
```
docs/api.md: missing required section "## Installation" (template: readme)
```

**Forbidden Pattern:**
```
docs/design/auth.md: contains forbidden pattern "TODO" (template: design-doc)
```

**Invalid Structure:**
```
README.md: must start with "# " (template: readme)
```

### Building Lobby Pattern

Enforce that every directory has a README serving as a navigation guide:

```yaml
rules:
  # Every directory must have exactly one README
  file-existence:
    "README.md": "exists:1"

  # READMEs must follow template
  file-content:
    templates:
      "**/README.md": "readme"
```

**.structurelint/templates/readme.yml:**
```yaml
required-sections:
  - "# "                              # Directory name
  - "⬆️ **[Parent Directory]"         # Link to parent
  - "## Overview"                     # Description

required-patterns:
  - "⬆️ \\*\\*\\[Parent Directory\\]\\(.*README\\.md\\)\\*\\*"  # Parent link
```

This creates a "building lobby" where each directory's README guides you through the codebase.

### Complete Example

```yaml
root: true

rules:
  # Require READMEs everywhere
  file-existence:
    "README.md": "exists:1"

  # Validate README content
  file-content:
    templates:
      "**/README.md": "readme"
      "docs/**/*.md": "documentation"
      "docs/design/*.md": "design-doc"

exclude:
  - node_modules/**
  - .git/**
```

See [docs/FILE_CONTENT_TEMPLATES.md](docs/FILE_CONTENT_TEMPLATES.md) for complete documentation.

## Documentation & Resources

### Getting Started

- 📖 **[Getting Started Guide](docs/GETTING_STARTED.md)** - Comprehensive tutorial from installation to advanced usage
- 🎯 **[Quick Start](#quick-start)** - Get up and running in 5 minutes

### Integration

- 🔗 **[Pre-commit Hooks](docs/PRE_COMMIT.md)** - Integrate with pre-commit framework
- 🤖 **[GitHub Actions](docs/GITHUB_ACTION.md)** - CI/CD integration examples
- ⚙️ **[Configuration Reference](#configuration)** - Full configuration documentation

### Contributing

- 🤝 **[Contributing Guide](CONTRIBUTING.md)** - How to contribute to structurelint
- 🐛 **[Issue Tracker](https://github.com/structurelint/structurelint/issues)** - Report bugs or request features
- 📋 **[Pull Request Template](.github/pull_request_template.md)** - Submit changes

### Project Information

- 📊 **[Test Coverage](MUTATION_TESTING.md)** - Mutation testing results (75.76% efficacy)
- 📈 **[Complexity Metrics](COMPLEXITY.md)** - Cyclomatic and cognitive complexity
- 📜 **[License](LICENSE)** - MIT License

### Examples

- 🎨 **[Integration Test Fixtures](testdata/fixtures/)** - Real examples:
  - `good-project/` - Clean structure (0 violations)
  - `bad-project/` - Phase 0 violations
  - `layer-violations/` - Architectural violations

## License

MIT License - see [LICENSE](LICENSE) for details

## Support

- 💬 Ask questions in [GitHub Issues](https://github.com/structurelint/structurelint/issues)
- 📖 Read the [documentation](docs/)
- 🤝 Contribute via [pull requests](https://github.com/structurelint/structurelint/pulls)

---

**Made with ❤️ for better codebases**
