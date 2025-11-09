# structurelint Design Document

## Executive Summary

This document outlines the strategic design and implementation of **structurelint**, a next-generation linter created to enforce project structure, organization, and architectural integrity. The current market reveals a distinct gap: existing tools excel at naming conventions but fail to provide rules for project topology and metrics, such as directory depth, file counts, and subdirectory limits.

**structurelint** fills this gap by evolving from a simple filesystem linter into a comprehensive "Architectural Guardian." This evolution is structured in three phases:

- **Phase 0 (Core)**: Fulfill the primary request by implementing rules for filesystem metrics (e.g., max depth, file/subdir counts) and incorporate best-in-class competitor features for naming conventions, regex matching, and file existence. ✅ **IMPLEMENTED**
- **Phase 1 (Layers)**: Expand into static code analysis to parse import graphs and enforce architectural layer boundaries.
- **Phase 2 (Orphans)**: Leverage the import graph data to identify and report orphaned files and unused exports.

## Market Landscape and The "structurelint" Niche

The need for structural linting arises from the "chaos" that can emerge in large-scale projects. Without enforcement, repository structures degrade, increasing cognitive load and making maintenance difficult.

### Existing Tools

A review of the existing ecosystem reveals several tools, each with a narrow focus:

- **ls-lint**: The most mature competitor, lauded for its high speed (written in Go) and simple YAML configuration. However, its rules are almost exclusively focused on naming conventions (e.g., kebab-case).
- **folderslint**: Designed for front-end projects, validates directory paths against a predefined "allow-list" of permitted paths.
- **Project Structure Validator**: A VS Code extension that validates if files with certain extensions are in the correct destination folder.

### The Gap

None of these tools adequately address the **quantitative metrics** of a project's structure:

- Maximum directory depth
- Maximum subdirectories per directory
- Complex file requirements (e.g., "this directory must contain either an index.ts or a README.md")

**structurelint** captures this unoccupied niche by integrating naming conventions while focusing on enforcing the topology and metrics of the repository.

## Implementation Architecture

### Core Components

The implementation follows Go best practices with a modular architecture:

```
structurelint/
├── cmd/
│   └── structurelint/      # Main CLI entry point
│       └── main.go
├── internal/
│   ├── config/             # Configuration system
│   │   └── config.go       # YAML parsing and cascading
│   ├── walker/             # Filesystem traversal
│   │   └── walker.go       # Directory walking and metrics
│   ├── rules/              # Rule implementations
│   │   ├── rule.go         # Rule interface
│   │   ├── max_depth.go    # Max depth rule
│   │   ├── max_files.go    # Max files per directory
│   │   ├── max_subdirs.go  # Max subdirectories
│   │   ├── naming_convention.go
│   │   ├── regex_match.go
│   │   ├── file_existence.go
│   │   └── disallowed_patterns.go
│   └── linter/             # Linter orchestration
│       └── linter.go       # Rule execution engine
└── examples/               # Example configurations
    ├── react-project.yml
    ├── go-project.yml
    ├── python-project.yml
    └── monorepo.yml
```

### Component Details

#### 1. Configuration System (`internal/config`)

The configuration system is modeled after ESLint's cascading configuration:

- **File Discovery**: Searches for `.structurelint.yml` or `.structurelint.yaml` files
- **Cascading**: Configurations from parent directories are merged with child configs
- **Root Flag**: A `root: true` property stops the upward search (crucial for monorepos)
- **Overrides**: Allows applying different rules to specific file patterns

```go
type Config struct {
    Root      bool
    Extends   interface{}
    Rules     map[string]interface{}
    Overrides []Override
}
```

Key features:
- Converts relative paths to absolute paths for consistent resolution
- Merges multiple configurations with later configs overriding earlier ones
- Supports disabling rules by setting values to `0` or `false`

#### 2. Filesystem Walker (`internal/walker`)

The walker efficiently traverses the filesystem and collects metrics:

```go
type FileInfo struct {
    Path       string  // Relative path from root
    AbsPath    string  // Absolute path
    IsDir      bool
    Depth      int     // Nesting depth from root
    ParentPath string  // Path of parent directory
}

type DirInfo struct {
    Path        string
    FileCount   int     // Number of files in directory
    SubdirCount int     // Number of subdirectories
    Depth       int
}
```

Features:
- Uses `filepath.WalkDir` for efficient traversal
- Calculates depth by counting path separators
- Aggregates directory statistics for metric rules
- Supports glob pattern matching including `**` wildcards

#### 3. Rules Engine (`internal/rules`)

Each rule implements a common interface:

```go
type Rule interface {
    Name() string
    Check(files []FileInfo, dirs map[string]*DirInfo) []Violation
}
```

This design allows for:
- Easy addition of new rules
- Parallel rule execution (future optimization)
- Consistent violation reporting

**Implemented Rules:**

1. **max-depth**: Enforces maximum directory nesting depth
2. **max-files-in-dir**: Limits files per directory
3. **max-subdirs**: Limits subdirectories per directory
4. **naming-convention**: Enforces camelCase, PascalCase, kebab-case, snake_case, etc.
5. **regex-match**: Validates filenames against regex patterns with substitution support
6. **file-existence**: Requires specific files to exist (supports ranges like `exists:1-10`)
7. **disallowed-patterns**: Blocks specific file or directory patterns

#### 4. Linter Orchestration (`internal/linter`)

The linter ties everything together:

1. Load and merge configurations from the filesystem
2. Create a walker and traverse the directory structure
3. Instantiate rules based on configuration
4. Execute all rules and collect violations
5. Report violations to the user

```go
type Linter struct {
    config *config.Config
}
```

The linter uses dynamic rule creation based on configuration, extracting and type-converting rule parameters from the YAML config.

## Configuration System Design

### Cascading Configuration

The cascading system mirrors ESLint's behavior:

```yaml
# project/.structurelint.yml (root)
root: true
rules:
  max-depth: { max: 7 }

# project/src/legacy/.structurelint.yml (child)
rules:
  max-depth: { max: 10 }  # Overrides parent
```

### Override System

Overrides provide surgical control over specific parts of the project:

```yaml
root: true
rules:
  max-depth: { max: 7 }

overrides:
  - files: ['src/components/**']
    rules:
      max-depth: { max: 10 }
      file-existence:
        "index.ts|index.tsx": "exists:1"
```

Overrides are processed in order, allowing for precise control.

### Configuration Merging Logic

The merging follows a strict precedence:

1. **Parent configs** are loaded first (lowest precedence)
2. **Child configs** override parent settings
3. **extends** configurations are loaded before local rules
4. **overrides** are applied last (highest precedence)

## Rule Specifications

### Metric Rules

#### max-depth

Prevents deeply nested directory structures that are hard to navigate.

```yaml
rules:
  max-depth: { max: 7 }
```

**Implementation**: Counts path separators in relative paths and compares against the configured maximum.

#### max-files-in-dir

Keeps directories focused and organized by limiting file count.

```yaml
rules:
  max-files-in-dir: { max: 20 }
```

**Implementation**: Aggregates file counts per directory during the walk phase, then validates against the limit.

#### max-subdirs

Controls complexity by limiting subdirectories.

```yaml
rules:
  max-subdirs: { max: 10 }
```

**Implementation**: Counts subdirectories for each directory and validates against the limit.

### Naming Convention Rules

#### naming-convention

Enforces consistent naming across the codebase.

```yaml
rules:
  naming-convention:
    "*.ts": "camelCase"
    "src/components/**/": "PascalCase"
```

**Supported conventions**:
- `camelCase`: firstWord (starts lowercase, no separators)
- `PascalCase`: FirstWord (starts uppercase, no separators)
- `kebab-case`: first-word (lowercase with hyphens)
- `snake_case`: first_word (lowercase with underscores)
- `lowercase`: allower
- `UPPERCASE`: ALLUPPER

**Implementation**: Pattern matching determines which files/directories to check, then applies the appropriate case validation function.

### Pattern Rules

#### regex-match

Provides powerful pattern validation with substitution.

```yaml
rules:
  regex-match:
    "src/components/*/*.tsx": "regex:${0}"  # Component name matches directory
    "*.js": "regex:![0-9]+"                 # Filename is not just numbers
```

**Special syntax**:
- `regex:pattern`: File must match
- `regex:!pattern`: File must NOT match (negation)
- `${0}`, `${1}`: Directory name substitution from wildcards

**Implementation**: Extracts directory names from wildcard positions, substitutes into regex, then validates filename (without extension).

#### file-existence

Ensures required files are present.

```yaml
rules:
  file-existence:
    "index.ts|index.js": "exists:1"    # Exactly one index file
    "*.test.ts": "exists:1"            # At least one test
    ".dir": "exists:0"                 # No subdirectories
    "*.md": "exists:1-10"              # 1-10 markdown files
```

**Syntax**:
- `exists:N`: Exactly N files
- `exists:N-M`: Between N and M files
- `.dir`: Special pattern for subdirectories
- `pattern1|pattern2`: OR logic for multiple patterns

**Implementation**: Groups files by directory, applies pattern matching (including OR logic), counts matches, validates against requirements.

#### disallowed-patterns

Blocks problematic patterns.

```yaml
rules:
  disallowed-patterns:
    - "src/utils/**"
    - "**/__pycache__"
    - ".DS_Store"
```

**Implementation**: Uses glob matching (including `**` for recursive wildcards) against all file paths.

## Performance Considerations

### Why Go?

Go was chosen for several critical reasons:

1. **Concurrency**: Go's goroutines make it trivial to parallelize filesystem walking and rule execution
2. **Performance**: Compiled binary with fast startup time, suitable for pre-commit hooks
3. **Distribution**: Single static binary, easy to distribute across platforms
4. **Ecosystem**: Excellent standard library for filesystem operations
5. **Proven**: ls-lint demonstrates Go's effectiveness for filesystem linting

### Optimization Opportunities (Future)

- **Parallel rule execution**: Run independent rules concurrently
- **Incremental checking**: Only check modified directories (git integration)
- **Caching**: Cache directory metrics between runs
- **Ignore patterns**: Skip `.git`, `node_modules`, etc.

## Roadmap

### Phase 1: Architectural Layer Enforcement ✅ IMPLEMENTED

Layer boundary enforcement is now fully implemented! structurelint can parse source files, build import graphs, and validate architectural boundaries.

**Configuration**:

```yaml
layers:
  - name: 'domain'
    path: 'src/domain/**'
    dependsOn: []
  - name: 'application'
    path: 'src/application/**'
    dependsOn: ['domain']
  - name: 'presentation'
    path: 'src/presentation/**'
    dependsOn: ['application', 'domain']

rules:
  enforce-layer-boundaries: true
```

**Implementation**:

The Phase 1 implementation includes:

1. **Multi-Language Parser** (`internal/parser/`):
   - TypeScript/JavaScript: Parses `import` and `require()` statements
   - Go: Parses `import` blocks and single imports
   - Python: Parses `import` and `from...import` statements
   - Resolves relative imports to project paths

2. **Import Graph Builder** (`internal/graph/`):
   - Builds dependency map from file paths to imported paths
   - Assigns files to layers based on path patterns
   - Validates layer dependency rules

3. **Layer Boundaries Rule** (`internal/rules/layer_boundaries.go`):
   - Checks each file's imports against layer rules
   - Resolves import paths to actual files in the project
   - Reports clear violation messages with layer names

4. **Example Configurations**:
   - `examples/clean-architecture.yml`: Clean Architecture pattern
   - `examples/hexagonal-architecture.yml`: Ports & Adapters pattern
   - `examples/feature-sliced.yml`: Feature-Sliced Design pattern

**Example Violation Detection**:

```
src/domain/product.ts: layer 'domain' cannot import from layer 'presentation' (imported: src/presentation/userComponent.ts)
```

### Phase 2: Dead Code Detection ✅ IMPLEMENTED

Dead code detection is now fully implemented! structurelint can identify orphaned files and unused exports to help eliminate project bloat.

**Configuration**:

```yaml
entrypoints:
  - "src/index.ts"
  - "**/*test*"

rules:
  disallow-orphaned-files: true
  disallow-unused-exports: true
```

**Implementation**:

The Phase 2 implementation includes:

1. **Enhanced Import Graph** (`internal/graph/`):
   - Tracks all files in the project
   - Builds incoming reference count for each file
   - Extracts and stores exports from all files
   - Supports export parsing across TypeScript, JavaScript, Go, Python

2. **Export Parser** (`internal/parser/exports.go`):
   - TypeScript/JavaScript: Parses `export`, `export default`, `export { }` statements
   - Go: Identifies exported symbols (uppercase identifiers)
   - Python: Extracts `__all__` definitions and top-level public definitions
   - Handles named and default exports

3. **Orphaned Files Rule** (`internal/rules/orphaned_files.go`):
   - Detects files with zero incoming references
   - Respects configured entrypoints
   - Automatically excludes config files, docs, and test files
   - Reports files that can be safely removed

4. **Unused Exports Rule** (`internal/rules/unused_exports.go`):
   - Identifies exported symbols never imported
   - Cross-references exports against import statements
   - Reports files exporting dead code
   - Helps reduce bundle size

5. **Example Configurations**:
   - `examples/dead-code-detection.yml`: Phase 2 focused setup
   - `examples/complete-setup.yml`: All three phases together

**Example Detection**:

```
src/unused-util.ts: file is orphaned (not imported by any other file)
src/helpers.ts: exports 'formatDate', 'parseNumber' but is never imported
```

## Testing Strategy

### Unit Tests

Each component should have comprehensive unit tests:

- **Config**: Test YAML parsing, cascading, merging logic
- **Walker**: Test filesystem traversal, depth calculation, pattern matching
- **Rules**: Test each rule's violation detection logic
- **Linter**: Test orchestration and rule execution

### Integration Tests

End-to-end tests with real directory structures:

- Create temporary test directories
- Apply configurations
- Validate expected violations are detected
- Ensure no false positives

### Performance Tests

Benchmark against large codebases:

- Measure time to lint repositories with 10K+ files
- Ensure pre-commit hook viability (< 1 second for typical projects)
- Profile and optimize hot paths

## Adoption Strategy

### "Best Practice" Presets

Future feature: `structurelint --init` command that generates configuration based on project type.

Example presets to create:
- `@structurelint/preset-go-standard`: golang-standards/project-layout
- `@structurelint/preset-python-src`: Python src layout
- `@structurelint/preset-react-feature`: Feature-based React architecture
- `@structurelint/preset-monorepo`: Monorepo best practices

### Integration Points

1. **Pre-commit hooks**: Lint on every commit
2. **CI/CD**: GitHub Actions, GitLab CI, etc.
3. **IDE extensions**: VSCode extension for real-time feedback
4. **Git hooks**: Automatic setup via `structurelint --setup-hooks`

## Conclusion

**structurelint** has successfully delivered all three phases:

### Phase 0 ✅ COMPLETE - Filesystem Linting
- ✅ All core filesystem metric rules (max-depth, max-files, max-subdirs)
- ✅ Comprehensive naming convention enforcement
- ✅ Powerful pattern matching and regex validation
- ✅ File existence requirements with flexible syntax
- ✅ ESLint-style cascading configuration
- ✅ Fast, efficient Go implementation
- ✅ Example configurations for multiple project types

### Phase 1 ✅ COMPLETE - Architectural Layer Enforcement
- ✅ Multi-language import parser (TypeScript, JavaScript, Go, Python)
- ✅ Import graph builder with dependency mapping
- ✅ Layer boundary enforcement rule
- ✅ Support for Clean Architecture, Hexagonal, Feature-Sliced Design
- ✅ Clear violation reporting with layer names
- ✅ Example architectural configurations

### Phase 2 ✅ COMPLETE - Dead Code Detection
- ✅ Orphaned file detection with smart exclusions
- ✅ Unused export identification
- ✅ Enhanced export parser for multiple languages
- ✅ Incoming reference tracking
- ✅ Configurable entrypoints
- ✅ Automatic exclusion of config/doc files

The tool is **production-ready** and feature-complete across all three phases. It represents a comprehensive solution for:
- **Filesystem Organization**: Enforce structural metrics and naming
- **Architectural Integrity**: Maintain layer boundaries and dependencies
- **Code Cleanliness**: Eliminate dead code and orphaned files

By providing a unified linting solution across these three dimensions, structurelint enables teams to:
- Maintain architectural sanity with enforced layer boundaries
- Prevent structural degradation at scale
- Eliminate technical debt from unused code
- Enforce conventions consistently across the codebase
- Support multiple architectural patterns across different languages
- Scale confidently with automated enforcement
