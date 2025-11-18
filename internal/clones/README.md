# clones

⬆️ **[Parent Directory](../README.md)**

## Overview

The `clones` package implements state-of-the-art code clone detection for structurelint. It identifies duplicated code (clones) across your codebase using syntactic analysis based on AST normalization and Rabin-Karp rolling hash.

## Architecture

```
clones/
├── types/          # Core data structures (Token, Shingle, Clone)
├── parser/         # AST normalization
├── syntactic/      # Hashing, indexing, expansion
└── detector/       # Pipeline orchestration and reporting
```

## Packages

### types/
Core data structures used throughout the clone detection system:
- `Token`: Normalized token representation
- `Shingle`: K-gram window with hash
- `Clone`: Detected clone with locations
- `Location`: Position in source file

### parser/
AST-based normalization:
- Parses Go source files using `go/ast`
- Normalizes identifiers to `_ID_`
- Normalizes literals to `_LIT_`
- Outputs canonical token stream

### syntactic/
Syntactic clone detection algorithms:
- `hasher.go`: Rabin-Karp rolling hash for k-gram shingling
- `index.go`: Inverted index (hash → locations)
- `expander.go`: Greedy bidirectional match expansion

### detector/
Pipeline orchestration:
- `detector.go`: Main detection workflow
- `reporter.go`: Output formatting (console, JSON, SARIF)

## Usage

See [Clone Detection Documentation](../../docs/CLONE_DETECTION.md) for usage guide.

## Implementation Status

**Current**: Strategy A - Syntactic Detection (Type-1, Type-2, Type-3 clones)
**Future**: Strategy B - Semantic Detection (Type-4 clones via GraphCodeBERT)

## References

- [Clone Detection User Guide](../../docs/CLONE_DETECTION.md)
- [Clone Detection Architecture](../../docs/CLONE_DETECTION_ARCHITECTURE.md)
