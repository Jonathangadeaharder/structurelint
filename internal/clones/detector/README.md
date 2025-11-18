# detector

⬆️ **[Parent Directory](../README.md)**

## Overview

The `detector` package orchestrates the complete clone detection pipeline, from file discovery to result reporting.

## Components

### detector.go
Main pipeline orchestration and workflow management.

#### Detection Pipeline

1. **File Discovery**
   - Walk directory tree
   - Filter by extension (.go)
   - Apply exclude patterns

2. **Parallel Normalization**
   - Worker pool (configurable workers)
   - Parse and normalize each file
   - Build token cache

3. **Shingling & Indexing**
   - Generate k-gram shingles
   - Build inverted index
   - Track hash collisions

4. **Collision Detection**
   - Find all hash collisions
   - Optionally filter to cross-file only

5. **Match Expansion**
   - Verify seed matches
   - Expand greedily forward/backward
   - Create clone objects

6. **Filtering & Reporting**
   - Filter by minimum token/line count
   - Format output
   - Report statistics

#### Configuration

```go
type Config struct {
    MinTokens      int      // Minimum clone size in tokens
    MinLines       int      // Minimum line count
    KGramSize      int      // Shingle window size
    ExcludePattern []string // File patterns to exclude
    CrossFileOnly  bool     // Only cross-file clones
    NumWorkers     int      // Parallel workers
}
```

### reporter.go
Output formatting for multiple formats.

#### Supported Formats

**Console (Default)**
- Human-readable output
- Grouped clone pairs
- File paths and line numbers
- Similarity scores

**JSON**
- Machine-readable format
- CI/CD integration
- Structured data for analysis

**SARIF**
- Static Analysis Results Interchange Format
- IDE integration (VS Code, GitHub)
- Code scanning compatibility

#### Key Functions

**`Report(clones []*Clone) string`**
- Formats clones in specified format
- Returns formatted output string

**`Summary(clones []*Clone) string`**
- Quick summary statistics
- Clone type breakdown
- Total tokens/lines

## Usage Example

```go
// Create detector
config := detector.DefaultConfig()
config.MinTokens = 50
config.MinLines = 10

d := detector.NewDetector(config)

// Run detection
clones, err := d.DetectClones("./internal")

// Report results
reporter := detector.NewReporter("console")
output := reporter.Report(clones)
fmt.Print(output)
```

## Performance Characteristics

| Stage | Time (39 files, 15K LOC) |
|-------|--------------------------|
| File discovery | <10ms |
| Normalization | ~500ms |
| Shingling | ~200ms |
| Indexing | ~100ms |
| Collision detection | <50ms |
| Expansion | ~1s |
| **Total** | **~2s** |

**Throughput**: ~7.5K LOC/s
