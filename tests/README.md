# Integration Tests

## Overview

Root-level integration tests for all components of structurelint.

## Test Organization

- **test_clone_detection_basic.py**: Integration tests for the semantic code clone detection system
- Additional integration tests for structurelint features

## Running Tests

### Go Tests
```bash
go test ./...
```

### Python Tests
```bash
cd clone_detection
pytest ../tests/test_clone_detection_basic.py -v
```

## Test Structure

All tests follow the Arrange-Act-Assert (AAA) pattern as enforced by structurelint's file content template rules.
