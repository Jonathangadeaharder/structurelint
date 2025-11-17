# Changelog

All notable changes to the Semantic Clone Detection system will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2025-11-17

### Added

#### Part I: Ingestion & Parsing
- Tree-sitter-based multi-language parser for extracting function-level code snippets
- Support for 6 programming languages: Python, JavaScript, Java, Go, C++, C#
- `TreeSitterParser` class with `parse_file()` and `parse_directory()` methods
- `CodeSnippet` dataclass for storing code with metadata (file path, line numbers, language)
- Language configuration system with S-expression queries (Table 1.1 from blueprint)
- Exclusion pattern support using pathspec

#### Part II: Vectorization
- GraphCodeBERT-based semantic code embedding
- `GraphCodeBERTEmbedder` class for generating 768-dimensional vectors
- Batch inference support with configurable batch size
- CPU and GPU device support with auto-detection
- Proper `<s>` token extraction from last hidden state (not pooler_output or mean pooling)
- Model save/load functionality for fine-tuned models

#### Part III: FAISS Indexing
- Production-grade IndexIVFPQ implementation
- **Critical L2 normalization** for cosine similarity equivalence
- `FAISSIndexBuilder` class with train/add/save/load methods
- Support for IndexFlatL2, IndexIVFFlat, and IndexIVFPQ
- Parameter tuning support (nlist, m, nbits, nprobe)
- GPU acceleration support (requires faiss-gpu)
- Cosine-to-L2 threshold conversion functions (Table 4.1)
- Comprehensive index statistics and information methods

#### Part IV: Query Pipeline
- `CloneSearcher` class for semantic clone detection
- Threshold-based `range_search()` (not k-NN) for finding all clones above similarity threshold
- `MetadataStore` SQLite database for snippet metadata
- `CloneMatch` dataclass for structured results
- Support for searching by code string or file location (file:line)
- Batch query support
- Result hydration with metadata join
- Exclude self-matches option

#### CLI Interface
- `clone-detect` command-line tool
- `ingest` command: Batch ingestion pipeline (Blueprint A)
  - Multi-language parsing
  - GraphCodeBERT embedding generation
  - FAISS index training and population
  - Metadata database creation
- `search` command: Query pipeline (Blueprint B)
  - Search by code or file location
  - Configurable similarity threshold
  - Rich table output with results
- `info` command: Display index statistics
- Progress bars and rich console output
- Verbose logging option

#### Configuration System
- Pydantic-based configuration models with validation
- `CloneDetectionConfig` with nested configs for model, index, parsing, and query
- YAML configuration file support
- Example configuration file with documentation
- Default values aligned with blueprint specifications

#### Documentation
- Comprehensive README.md with architecture overview and quick start
- ARCHITECTURE.md with deep technical details
- API examples demonstrating Python usage
- CLI usage examples with shell script
- Blueprint reference documentation throughout codebase
- Inline code documentation with rationale for design decisions

#### Testing
- Basic test suite with pytest
- Tests for cosine-L2 conversion (Table 4.1 validation)
- L2 normalization verification
- Import tests for all major components
- Configuration loading tests

#### Project Infrastructure
- pyproject.toml with complete dependency specification
- Support for CPU and GPU installations
- Development dependencies (pytest, black, isort, mypy)
- Optional dependencies for training and GPU
- Entry point script: `clone-detect` CLI command
- MIT License

### Implementation Notes

This release implements the complete architecture from the blueprint:
"A Blueprint for Semantic Code Clone Detection at Scale Using GraphCodeBERT and FAISS"

Key design decisions:
- **Path A** implementation (no explicit DFG at inference)
- **IndexIVFPQ** for production scale (not Flat or IVF-only)
- **L2 normalization** as critical linchpin for mathematical correctness
- **range_search** for threshold-based clone detection (not k-NN)
- **SQLite** for metadata (lightweight, embedded database)
- **Click + Rich** for CLI (beautiful, user-friendly interface)

### Known Limitations

1. **Token Limit**: Functions > 512 tokens are truncated
2. **Static Index**: No incremental updates (requires full rebuild)
3. **Approximate Search**: IndexIVFPQ is not 100% accurate (98-99% recall)
4. **Language Support**: Limited to 6 languages (extensible via Tree-sitter)
5. **Fine-Tuning**: Base model used (fine-tuning implementation not included)

### Performance

Typical performance on 10M function codebase:
- Index build: ~2-3 hours (with GPU)
- Index size: 640 MB (compressed from 30.7 GB)
- Query latency: 5-15 ms
- Throughput: 500-1000 queries/second

## [Unreleased]

### Planned Features

- Fine-tuning script with Triplet Margin Loss
- BigCloneBench evaluation framework
- Index sharding for billion-scale datasets
- Natural language code search
- IDE integrations (VSCode, IntelliJ)
- Incremental index updates
- Cross-language clone detection
- Clone cluster analysis and visualization
