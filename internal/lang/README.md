# Language Detection

⬆️ **[Parent Directory](../README.md)**

## Overview

The `lang` package provides language detection capabilities for structurelint's polyglot support. It identifies programming languages used in a project by analyzing manifest files (package.json, go.mod, Cargo.toml, etc.) and provides language-specific defaults.

## Components

### detector.go
- Language detection from manifest files
- Support for 9 programming languages
- Sub-language detection (e.g., React within TypeScript)
- Default naming convention mapping per language

### Language Support

**Supported Languages:**
- Go
- Python
- TypeScript
- JavaScript
- React (JSX/TSX)
- Rust
- Java
- C#
- Ruby

## Usage

```go
detector := lang.NewDetector(rootDir)
languages, err := detector.Detect()
for _, langInfo := range languages {
    convention := langInfo.Language.DefaultNamingConvention()
    // Apply language-specific rules
}
```

## Related

- **Priority 2 Features**: Part of polyglot support implementation
- **Naming Convention Rule**: Uses language detection for auto-configuration
- **Infrastructure Profile**: Language-aware exemptions
