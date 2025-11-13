# Output Formatters

## Overview

This package provides output formatters for structurelint violations.

## Available Formatters

- **TextFormatter**: Human-readable text output (default)
- **JSONFormatter**: Machine-readable JSON format for CI/CD integration
- **JUnitFormatter**: JUnit XML format for Jenkins, GitHub Actions, etc.

## Usage

```go
formatter, err := output.GetFormatter("json")
if err != nil {
    return err
}

formatted, err := formatter.Format(violations)
if err != nil {
    return err
}

fmt.Print(formatted)
```

## Supported Formats

- `text` - Plain text output with file:message format
- `json` - JSON with version, timestamp, and structured violations
- `junit` - JUnit XML grouped by rule name
