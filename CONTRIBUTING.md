# Contributing to structurelint

Thank you for your interest in contributing to structurelint! This document provides guidelines and instructions for contributing.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Testing](#testing)
- [Code Quality](#code-quality)
- [Submitting Changes](#submitting-changes)
- [Project Structure](#project-structure)

## Code of Conduct

Be respectful, constructive, and collaborative. We aim to build a welcoming community for all contributors.

## Getting Started

1. **Fork the repository** on GitHub
2. **Clone your fork** locally:
   ```bash
   git clone https://github.com/YOUR_USERNAME/structurelint.git
   cd structurelint
   ```
3. **Add upstream remote**:
   ```bash
   git remote add upstream https://github.com/structurelint/structurelint.git
   ```

## Development Setup

### Prerequisites

- Go 1.21 or later
- Git

### Install Dependencies

```bash
go mod download
```

### Build the Binary

```bash
go build -o structurelint ./cmd/structurelint
```

### Run Tests

```bash
go test ./...
```

## Making Changes

### Create a Branch

Always create a new branch for your changes:

```bash
git checkout -b feature/your-feature-name
```

Branch naming conventions:
- `feature/` - New features
- `fix/` - Bug fixes
- `docs/` - Documentation changes
- `refactor/` - Code refactoring
- `test/` - Test additions or improvements

### Coding Standards

1. **Follow Go conventions**:
   - Use `gofmt` for formatting
   - Follow [Effective Go](https://golang.org/doc/effective_go.html)
   - Run `go vet` before committing

2. **Complexity limits**:
   - Cyclomatic complexity: ≤ 20 per function
   - Cognitive complexity: ≤ 85 per function
   - See `COMPLEXITY.md` for details

3. **Write clear commit messages**:
   ```
   Add regex pattern matching for naming conventions

   - Implement wildcard substitution in patterns
   - Add tests for boundary conditions
   - Update documentation
   ```

### Adding New Rules

To add a new linting rule:

1. Create a new file in `internal/rules/`:
   ```go
   // internal/rules/your_rule.go
   package rules

   type YourRule struct {
       // Config fields
   }

   func NewYourRule(config map[string]interface{}) *YourRule {
       // Initialize from config
   }

   func (r *YourRule) Name() string {
       return "your-rule"
   }

   func (r *YourRule) Check(files []walker.FileInfo, importGraph *graph.ImportGraph) []Violation {
       // Implementation
   }
   ```

2. Register the rule in `internal/linter/linter.go`:
   ```go
   case "your-rule":
       rules = append(rules, NewYourRule(ruleCfg))
   ```

3. Add comprehensive tests in `internal/rules/your_rule_test.go`

4. Update README.md with rule documentation

## Testing

### Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run specific package tests
go test ./internal/rules/...
```

### Test Coverage Requirements

- **Minimum coverage**: 70%
- **Target coverage**: 80%+
- Current coverage: 79.2%

### Mutation Testing

We use mutation testing to ensure test quality:

```bash
# Install gremlins
go install github.com/go-gremlins/gremlins/cmd/gremlins@latest

# Run mutation testing
gremlins unleash --workers=2 ./internal/rules
```

- **Target efficacy**: 75%+
- Current efficacy: 75.76%

### Writing Tests

1. **Test file naming**: `*_test.go`
2. **Test function naming**: `TestFunctionName_Scenario`
3. **Test edge cases**:
   - Empty inputs
   - Boundary conditions
   - Error cases
   - Large inputs

Example:
```go
func TestYourRule_EmptyInput(t *testing.T) {
    rule := NewYourRule(nil)
    violations := rule.Check([]walker.FileInfo{}, nil)

    if len(violations) != 0 {
        t.Errorf("Expected 0 violations, got %d", len(violations))
    }
}
```

## Code Quality

### Before Committing

Run these checks locally:

```bash
# Format code
gofmt -w .

# Run linter
golangci-lint run

# Run tests
go test ./...

# Check complexity (optional)
gocyclo -over 20 .
gocognit -over 85 .
```

### Continuous Integration

All pull requests must pass:
- ✅ All tests
- ✅ golangci-lint
- ✅ Complexity checks
- ✅ Self-lint (structurelint on its own codebase)
- ✅ Multi-platform builds

## Submitting Changes

### Pull Request Process

1. **Update your branch** with latest upstream:
   ```bash
   git fetch upstream
   git rebase upstream/main
   ```

2. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

3. **Create a Pull Request** on GitHub with:
   - Clear title describing the change
   - Description of what changed and why
   - Reference to related issues (if any)
   - Screenshots/examples (if applicable)

### Pull Request Template

```markdown
## Summary
Brief description of changes

## Changes
- Item 1
- Item 2

## Testing
- [ ] All tests pass
- [ ] Added tests for new functionality
- [ ] Ran mutation testing (if applicable)
- [ ] Manual testing completed

## Related Issues
Closes #123
```

### Review Process

- All PRs require at least one approval
- Address review comments promptly
- Keep PRs focused and reasonably sized
- Be patient and respectful during review

## Project Structure

```
structurelint/
├── cmd/
│   └── structurelint/      # CLI entry point
├── internal/
│   ├── config/             # Configuration parsing
│   ├── graph/              # Import graph building
│   ├── linter/             # Linter orchestration
│   ├── parser/             # File parsing (imports, exports)
│   ├── rules/              # All linting rules
│   └── walker/             # Filesystem walking
├── testdata/
│   └── fixtures/           # Integration test projects
├── .github/
│   └── workflows/          # CI/CD pipelines
├── .structurelint.yml      # Self-linting config
├── .golangci.yml           # Linter config
└── docs/                   # Documentation
```

## Need Help?

- 📖 Read the [README.md](README.md)
- 🐛 Report bugs via [GitHub Issues](https://github.com/structurelint/structurelint/issues)
- 💬 Ask questions in issues or pull requests
- 📧 Contact maintainers (see README)

## License

By contributing, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing to structurelint! 🎉
