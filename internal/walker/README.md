# walker

⬆️ **[Parent Directory](../README.md)**

## Overview

The `walker` package handles filesystem traversal and collects information about files and directories.

## Key Features

- **Recursive Directory Walking**: Traverses entire directory trees
- **File Statistics**: Collects depth, file counts, subdirectory counts
- **Exclusion Support**: Respects exclude patterns from configuration
- **Automatic Filtering**: Skips `.git`, `node_modules`, `vendor` directories

## Main Types

- `Walker`: Main filesystem walker
- `FileInfo`: Information about a single file or directory
- `DirInfo`: Aggregated statistics for a directory

## Usage

```go
w := walker.New(path).WithExclude(excludePatterns)
err := w.Walk()
files := w.GetFiles()
dirs := w.GetDirs()
```

## Recent Refactoring

The `Walk` method was recently refactored to reduce cognitive complexity from 40 to under 30 by extracting logic into focused helper methods:
- `processPath`: Main path processing
- `shouldSkip`: Skip decision logic
- `isIgnoredDir`: Directory ignore checks
- `calculateDepth`: Depth calculation
- `normalizeParentPath`: Parent path normalization
- `updateDirectoryStats`: Statistics management
