# Command-Line Interface

## Overview

User-friendly CLI for the semantic code clone detection system.

## Commands

### ingest
Ingest and index code from a directory:
```bash
clone-detect ingest /path/to/code --index-path index.faiss
```

### search
Search for clones of a code snippet:
```bash
clone-detect search --query-file function.py --index-path index.faiss --threshold 0.85
```

### info
Display index statistics:
```bash
clone-detect info --index-path index.faiss
```

## Features

- Rich console output with progress bars
- Interactive feedback
- Configurable thresholds
- Multi-language support
