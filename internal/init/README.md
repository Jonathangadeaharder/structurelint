# init

⬆️ **[Parent Directory](../README.md)**

## Overview

The `init` package handles automatic project detection and configuration generation for structurelint's `--init` command.

## Key Features

- **Language Detection**: Automatically identifies programming languages used in the project
- **Test Pattern Detection**: Discovers existing test file organization patterns
- **Smart Defaults**: Generates appropriate configuration based on detected patterns
- **Multi-Language Support**: Python, Go, TypeScript, JavaScript, Java, Rust, Ruby, C/C++

## Main Components

### detector.go

Scans the project and detects:
- Programming languages and file counts
- Test file patterns (adjacent vs separate)
- Integration test directories
- Project structure metrics (depth, files per dir, subdirectories)
- Documentation completeness

### generator.go

Generates `.structurelint.yml` configuration with:
- Language-specific test validation rules
- Appropriate exclusions for each ecosystem
- Structure limits based on current project
- Commented architectural layer examples
- Sensible defaults for detected patterns

## Supported Language Patterns

| Language | Test Pattern | File Naming |
|----------|--------------|-------------|
| Go | Adjacent | `*_test.go` |
| Python | Separate/Adjacent | `test_*.py`, `*_test.py` |
| TypeScript/JS | Adjacent | `*.test.ts`, `*.spec.js` |
| Java | Separate | `*Test.java`, `*IT.java` |
| Rust | Inline/Separate | `*_test.rs`, `tests/` |
| Ruby | Separate | `*_spec.rb` in `spec/` |
| C/C++ | Separate | `test_*.cpp`, `*_test.cpp` |

## Usage

```bash
# Initialize configuration for current project
structurelint --init

# View analysis without creating config
structurelint --init --dry-run
```

## Detection Logic

1. **Walk filesystem** with standard exclusions (node_modules, vendor, etc.)
2. **Count source files** by extension
3. **Identify test files** by naming patterns
4. **Analyze test organization**:
   - If test file has adjacent source → "adjacent" pattern
   - If test file in `tests/`, `test/`, `__tests__/` → "separate" pattern
5. **Calculate structure metrics** from actual project
6. **Generate configuration** with detected settings
