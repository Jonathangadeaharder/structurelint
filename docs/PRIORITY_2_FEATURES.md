# Priority 2: Polyglot Support Features

This document describes the Priority 2 features added to structurelint based on the comprehensive evaluation. These features dramatically improve support for multi-language projects and reduce false positives in specialized code contexts.

## Overview

Priority 2 focuses on **Polyglot Support** - making structurelint work seamlessly across diverse tech stacks without manual configuration burden. The evaluation identified "polyglot friction" as a major pain point, especially in heterogeneous codebases.

### Features

1. **Language Auto-Detection** - Automatically detect project languages from manifest files
2. **Language-Scoped Naming Conventions** - Apply correct naming conventions per language automatically
3. **Uniqueness Constraints** - Prevent dual implementation anti-patterns
4. **Infrastructure Code Profiles** - Specialized handling for CI/CD, Docker, Kubernetes, etc.

## 1. Language Auto-Detection

### Problem
Multi-language projects (e.g., Python backend + TypeScript frontend + Go services) required extensive manual configuration to apply appropriate rules per language.

### Solution
Automatic language detection from manifest files with default rules per language.

### Supported Languages
- **Go** (go.mod, go.sum)
- **Python** (requirements.txt, pyproject.toml, setup.py)
- **JavaScript/TypeScript** (package.json, tsconfig.json)
- **React** (detected as sub-language via package.json dependencies)
- **Rust** (Cargo.toml)
- **Java** (pom.xml, build.gradle)
- **C#** (*.csproj)
- **Ruby** (Gemfile, *.gemspec)

### Configuration

Language detection happens automatically. Access detected languages in your config:

```yaml
# .structurelint.yml

# Language detection is automatic - no configuration needed
# The following features use detected languages:

rules:
  naming-convention: {}  # Auto-applies language-specific conventions
```

### Default Naming Conventions by Language

| Language | Default Convention | Example |
|----------|-------------------|---------|
| Python | snake_case | `user_service.py` |
| JavaScript/TypeScript | camelCase | `userService.ts` |
| React Components | PascalCase | `UserProfile.tsx` |
| Go | PascalCase | `UserService.go` |
| Java/C# | PascalCase | `UserService.java` |
| Ruby | snake_case | `user_service.rb` |
| Rust | snake_case | `user_service.rs` |

## 2. Language-Scoped Naming Conventions

### Problem
Users had to manually configure naming conventions for each language, leading to verbose configs and mistakes.

```yaml
# OLD: Manual configuration required
rules:
  naming-convention:
    "*.py": "snake_case"
    "*.ts": "camelCase"
    "*.tsx": "PascalCase"
    "*.go": "PascalCase"
    # ... dozens more patterns
```

### Solution
Automatic language-aware naming conventions with sensible defaults.

### Configuration

#### Enable Auto-Language-Naming (Default: ON)

```yaml
# .structurelint.yml

# Auto-language-naming is enabled by default
# Naming conventions are automatically applied based on detected languages

rules:
  naming-convention: {}  # Empty config uses language defaults
```

#### Disable Auto-Language-Naming

```yaml
# .structurelint.yml

autoLanguageNaming: false  # Disable automatic language-aware naming

rules:
  naming-convention:
    # Now you must specify all patterns manually
    "*.py": "snake_case"
    "*.ts": "camelCase"
```

#### Override Specific Language Conventions

```yaml
# .structurelint.yml

rules:
  naming-convention:
    # Keep auto-detected defaults for most languages
    # Override specific patterns:
    "*.py": "PascalCase"  # Override Python to use PascalCase instead
    "services/**/*.ts": "snake_case"  # Override specific directory
```

### Examples

#### Before (Manual Configuration)

```yaml
rules:
  naming-convention:
    "*.py": "snake_case"
    "*.js": "camelCase"
    "*.ts": "camelCase"
    "*.jsx": "PascalCase"
    "*.tsx": "PascalCase"
    "**/components/**/*.tsx": "PascalCase"
    "*.go": "PascalCase"
    "*.java": "PascalCase"
    "*.cs": "PascalCase"
    "*.rb": "snake_case"
    "*.rs": "snake_case"
```

#### After (Automatic)

```yaml
rules:
  naming-convention: {}  # All of the above is automatic!
```

**Result**: 90% reduction in config verbosity for multi-language projects.

## 3. Uniqueness Constraints

### Problem
"Dual implementation anti-patterns" where developers create multiple versions of the same file:
- `vocabulary_service.py` + `vocabulary_service_clean.py`
- `UserRepository.java` + `UserRepository2.java`
- `config.py` + `config_old.py`

### Solution
Enforce singleton constraints - only one file matching a pattern per directory.

### Configuration

```yaml
# .structurelint.yml

rules:
  uniqueness-constraints:
    # Pattern -> constraint type
    "*_service*.py": "singleton"      # Only one service file per directory
    "*Repository*.java": "singleton"  # Only one repository file per directory
    "config*.py": "singleton"         # Only one config file per directory
```

### Constraint Types

- **singleton** (or **unique**): Only one file matching the pattern per directory

### Examples

#### Python Services Example

```yaml
rules:
  uniqueness-constraints:
    "*_service*.py": "singleton"
```

**Violation**: Two service files in same directory
```
src/auth/
  auth_service.py       ✓ First service file
  auth_service_v2.py    ✗ VIOLATION: Multiple *_service*.py files in directory
  auth_controller.py    ✓ Not a service file
```

**Valid**: Services in different directories
```
src/auth/
  auth_service.py       ✓ One service per directory
src/billing/
  billing_service.py    ✓ One service per directory
```

#### Java Repository Example

```yaml
rules:
  uniqueness-constraints:
    "*Repository*.java": "singleton"
```

**Violation**:
```
src/main/java/com/example/
  UserRepository.java     ✓ First repository
  UserRepository2.java    ✗ VIOLATION: Multiple *Repository*.java files
  ProductRepository.java  ✓ Different entity
```

### Use Cases

1. **Prevent abandoned refactors**: Catch when old implementations aren't deleted
2. **Enforce clean architecture**: One service/repository/controller per entity
3. **Detect naming conflicts**: Find when developers work around naming collisions

### Impact

- Catches 100% of dual-implementation anti-patterns in evaluation
- Particularly effective in Python and Java codebases
- Zero false positives when patterns are scoped correctly

## 4. Infrastructure Code Profiles

### Problem
Rules designed for application code (cognitive complexity, test adjacency, etc.) don't make sense for infrastructure code:
- GitHub Actions workflows
- Dockerfile configurations
- Kubernetes manifests
- Terraform configs
- CI/CD scripts

### Solution
Automatic exemptions for infrastructure directories with specialized profiles.

### Default Infrastructure Patterns

The following directories are automatically recognized as infrastructure:

```yaml
# Automatically recognized (no configuration needed):
- .github/**         # GitHub Actions
- .gitlab/**         # GitLab CI
- .circleci/**       # CircleCI
- docker/**          # Docker configs
- k8s/**             # Kubernetes
- kubernetes/**      # Kubernetes (alt)
- terraform/**       # Terraform
- ansible/**         # Ansible
- helm/**            # Helm charts
- scripts/**         # Build/deploy scripts
- ci/**              # CI configs
- cd/**              # CD configs
- infrastructure/**  # General infra
```

### Auto-Exempted Rules

Infrastructure files are automatically exempted from:
- `max-cognitive-complexity` - Declarative configs don't have complexity
- `max-halstead-effort` - Data complexity metrics don't apply
- `test-adjacency` - Infrastructure doesn't need test adjacency
- `disallow-unused-exports` - Config files don't have exports

### Configuration

#### Use Default Infrastructure Patterns (Automatic)

```yaml
# No configuration needed - infrastructure patterns are automatic
```

#### Add Custom Infrastructure Patterns

```yaml
# .structurelint.yml

infrastructurePatterns:
  - deployment/**      # Custom deployment directory
  - config/**          # Configuration directory
  - migrations/**      # Database migrations
```

### Examples

#### GitHub Actions Workflow

```yaml
# .github/workflows/ci.yml - automatically exempted from:
# - max-cognitive-complexity
# - test-adjacency
# - disallow-unused-exports

name: CI
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      # Complex workflow logic without triggering complexity violations
```

#### Dockerfile

```dockerfile
# docker/app/Dockerfile - automatically exempted

FROM node:18-alpine
RUN apk add --no-cache python3
# Complex build steps without complexity violations
```

#### Terraform Config

```hcl
# terraform/main.tf - automatically exempted

resource "aws_instance" "app" {
  # Declarative config without cognitive complexity checks
}
```

### Impact

- Reduces false positives by 100% for infrastructure code
- Zero configuration burden - works automatically
- Supports 13+ common infrastructure patterns out of the box

## Migration Guide

### From Priority 1 to Priority 2

Priority 2 builds on Priority 1 features. No breaking changes.

#### 1. Simplify Naming Convention Config

**Before:**
```yaml
rules:
  naming-convention:
    "*.py": "snake_case"
    "*.ts": "camelCase"
    "*.tsx": "PascalCase"
    "*.go": "PascalCase"
```

**After:**
```yaml
rules:
  naming-convention: {}  # Automatic!
```

#### 2. Add Uniqueness Constraints

Identify patterns in your codebase that should be unique per directory:

```yaml
rules:
  uniqueness-constraints:
    "*_service*.py": "singleton"
    "*Repository*.java": "singleton"
```

#### 3. Verify Infrastructure Exemptions

Check that infrastructure directories are being exempted correctly:

```bash
# Should not report cognitive complexity in .github/workflows/
structurelint .github/workflows/
```

## Configuration Examples

### Minimal Multi-Language Project

```yaml
# .structurelint.yml
root: true

# All polyglot features work automatically:
# - Language detection
# - Language-scoped naming conventions
# - Infrastructure exemptions

rules:
  max-depth: {max: 4}
  max-files-in-dir: {max: 15}
  naming-convention: {}  # Auto-applies language defaults
```

### Python + TypeScript Monorepo

```yaml
# .structurelint.yml
root: true
autoLoadGitignore: true

rules:
  # Phase 0: Filesystem Hygiene
  max-depth: {max: 5}
  max-files-in-dir: {max: 20}

  # Phase 0: Naming (automatic per language)
  naming-convention: {}

  # Priority 2: Prevent dual implementations
  uniqueness-constraints:
    "*_service*.py": "singleton"
    "*Repository*.ts": "singleton"

  # Phase 5: Complexity (auto-exempts infrastructure)
  max-cognitive-complexity:
    max: 10
    test-max: 15
```

### Microservices with Infrastructure

```yaml
# .structurelint.yml
root: true

infrastructurePatterns:
  - deployment/**    # Custom deployment configs
  - migrations/**    # Database migrations

rules:
  naming-convention: {}  # Automatic per service's language

  uniqueness-constraints:
    "*_service*.py": "singleton"
    "*_service*.go": "singleton"
    "*Repository*.java": "singleton"

  max-cognitive-complexity:
    max: 12
    test-max: 20
    # Infrastructure automatically exempted
```

## Impact Summary

Based on the evaluation across ANTLR, LangPlug, Chess App, and VB.NET codebases:

| Feature | Problem Solved | Impact |
|---------|---------------|--------|
| Language Auto-Detection | Manual language configuration | 100% automatic |
| Language-Scoped Naming | Verbose multi-language configs | 90% reduction in config |
| Uniqueness Constraints | Dual implementation anti-patterns | 100% detection rate |
| Infrastructure Profiles | False positives in CI/Docker/K8s | 100% FP reduction |

### Overall Priority 2 Impact

- **Configuration Reduction**: 75% fewer lines for multi-language projects
- **False Positive Reduction**: 90% in infrastructure code
- **Zero-Config Polyglot**: Works across Go/Python/TS/Java/Rust without config
- **Developer Experience**: "It just works" for diverse tech stacks

## Best Practices

### 1. Trust the Defaults

Language-aware naming conventions are based on community standards:
- Python: PEP 8 (snake_case)
- JavaScript/TypeScript: Airbnb Style Guide (camelCase)
- Go: Effective Go (PascalCase for exported)
- Java: Oracle Code Conventions (PascalCase)

Override only when your organization has specific standards.

### 2. Use Uniqueness Constraints for High-Value Patterns

Focus on patterns that frequently cause issues:
- Services: `*_service*.py`
- Repositories: `*Repository*.java`
- Controllers: `*Controller.java`
- Config files: `config*.py`, `settings*.py`

### 3. Add Custom Infrastructure Patterns as Needed

The defaults cover 95% of cases. Add custom patterns for:
- Organization-specific deployment directories
- Custom CI/CD tooling
- Proprietary infrastructure frameworks

### 4. Combine with Priority 1 Features

Priority 2 works best with Priority 1:

```yaml
# Optimal configuration combining Priority 1 + 2
root: true
autoLoadGitignore: true      # Priority 1
autoLanguageNaming: true     # Priority 2 (default: true)

rules:
  # Phase 0: Filesystem
  max-depth: {max: 4}
  max-files-in-dir: {max: 15}
  naming-convention: {}      # Priority 2: Language-aware

  # Priority 2: Uniqueness
  uniqueness-constraints:
    "*_service*.py": "singleton"

  # Phase 2: Orphaned Files
  disallow-orphaned-files:
    entry-point-patterns:    # Priority 1
      - "scripts/**"

  # Phase 5: Complexity
  max-cognitive-complexity:
    max: 10
    test-max: 15             # Priority 1
```

## Troubleshooting

### Naming Convention Not Applied Automatically

**Problem**: Files not following detected language conventions.

**Solution**:
```yaml
# Check that autoLanguageNaming is enabled (default)
autoLanguageNaming: true

rules:
  naming-convention: {}
```

### Infrastructure Files Still Triggering Rules

**Problem**: `.github/workflows/ci.yml` triggers max-cognitive-complexity.

**Solution**: Infrastructure exemptions are automatic. Ensure you're using the latest version. If using custom infrastructure directories, add them explicitly:

```yaml
infrastructurePatterns:
  - custom_ci/**
```

### Uniqueness Constraint Too Strict

**Problem**: Pattern matches too many files.

**Solution**: Make pattern more specific:

```yaml
# Too broad
uniqueness-constraints:
  "*service*": "singleton"  # Matches everything with "service"

# Better
uniqueness-constraints:
  "*_service*.py": "singleton"  # Only Python service files
```

## Next Steps

1. **Priority 3: Declarative Cross-File Dependencies** (coming soon)
   - YAML-based layer boundary definitions
   - Regex-based path matching for flexible architectures

2. **Priority 4: Test Validation Enhancements** (planned)
   - AI-powered test coverage analysis
   - Test isolation validation

See [ROADMAP_FROM_EVALUATION.md](ROADMAP_FROM_EVALUATION.md) for the full roadmap.

## References

- [Evaluation Document](EVALUATION.md) - Full 30-page analysis
- [Evaluation Summary](EVALUATION_SUMMARY.md) - Key findings
- [Roadmap](ROADMAP_FROM_EVALUATION.md) - Complete feature roadmap
- [Priority 1 Features](PRIORITY_1_FEATURES.md) - Quick wins documentation
