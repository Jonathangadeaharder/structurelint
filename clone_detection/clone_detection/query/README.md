# Query and Search

## Overview

Clone detection search functionality with threshold-based retrieval and metadata management.

## Components

- **search.py**: CloneSearcher for finding similar code snippets
- **metadata.py**: SQLite-based metadata storage for vector-to-code mapping

## Features

- Threshold-based clone detection
- Cosine similarity scoring
- Metadata retrieval (file paths, line numbers, function names)
- Result ranking and filtering

## Usage

```python
from clone_detection.query.search import CloneSearcher

searcher = CloneSearcher(index_path, metadata_path)
matches = searcher.find_clones(query_code, threshold=0.85)
```
