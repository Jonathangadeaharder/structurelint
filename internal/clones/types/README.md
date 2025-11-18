# types

⬆️ **[Parent Directory](../README.md)**

## Overview

The `types` package defines core data structures used throughout the clone detection system.

## Data Types

### Token
Represents a normalized token from source code:
- `Type`: TokenType (Keyword, Identifier, Literal, Operator, Punctuation)
- `Value`: Normalized value (`_ID_`, `_LIT_`, or raw keyword/operator)
- `Line`, `Column`: Source position
- `Position`: Position in token stream

### TokenType
Categories of tokens:
- `TokenKeyword`: Keywords (if, for, func, return)
- `TokenIdentifier`: Variables/functions (normalized to `_ID_`)
- `TokenLiteral`: Strings/numbers (normalized to `_LIT_`)
- `TokenOperator`: Operators (+, -, ==, :=)
- `TokenPunctuation`: Punctuation ((, ), {, }, ;)

### Shingle
K-gram window with rolling hash:
- `Hash`: Rabin-Karp hash value (uint64)
- `StartToken`, `EndToken`: Window bounds in token stream
- `FilePath`: Source file
- `StartLine`, `EndLine`: Line numbers

### Clone
Detected code clone:
- `Type`: CloneType (Type1, Type2, Type3, Type4)
- `Locations`: All clone locations (2+)
- `TokenCount`: Size in tokens
- `LineCount`: Approximate size in lines
- `Hash`: Hash value (for syntactic clones)
- `Similarity`: Similarity score (0.0-1.0)

### CloneType
Classification of clones:
- `Type1`: Exact copy-paste (whitespace/comments differ)
- `Type2`: Renamed identifiers/literals
- `Type3`: Modified statements (additions/deletions)
- `Type4`: Semantic equivalence (different implementation)

### Location
Position in source file:
- `FilePath`: Path to file
- `StartLine`, `EndLine`: Line range (1-indexed)
- `StartToken`, `EndToken`: Token range in normalized stream

### FileTokens
File's normalized token stream:
- `FilePath`: Source file path
- `Tokens`: Normalized token sequence
