# Architecture Documentation

## Overview

This document provides a deep dive into the architecture of the semantic code clone detection system, as implemented following the blueprint "A Blueprint for Semantic Code Clone Detection at Scale Using GraphCodeBERT and FAISS".

## System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    SEMANTIC CLONE DETECTION                      │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────┐    ┌──────────────┐    ┌──────────────┐       │
│  │   Part I    │───▶│   Part II    │───▶│   Part III   │       │
│  │  Ingestion  │    │ Vectorization│    │   Indexing   │       │
│  │ Tree-sitter │    │ GraphCodeBERT│    │    FAISS     │       │
│  └─────────────┘    └──────────────┘    └──────────────┘       │
│        │                   │                     │               │
│        ▼                   ▼                     ▼               │
│  Code Snippets      768-D Vectors      IndexIVFPQ               │
│                                                  │               │
│                                                  │               │
│  ┌─────────────────────────────────────────────┐│               │
│  │             Part IV: Query Pipeline          ││               │
│  │          range_search + Metadata             ││               │
│  └──────────────────────────────────────────────┘▼               │
│                                          Clone Matches           │
└─────────────────────────────────────────────────────────────────┘
```

## Part I: Ingestion & Parsing

### Implementation: `clone_detection/parsers/`

**Purpose**: Extract discrete, semantically-coherent units (functions/methods) from polyglot codebases.

**Key Components**:

1. **`language_configs.py`**: Defines S-expression queries for each language
   - Maps file extensions to languages
   - Specifies Tree-sitter query patterns
   - Supports: Python, JavaScript, Java, Go, C++, C#

2. **`tree_sitter_parser.py`**: Main parsing engine
   - `CodeSnippet` dataclass: Stores code + metadata
   - `TreeSitterParser`: Multi-language parser
   - Methods: `parse_file()`, `parse_directory()`

**Design Decisions**:
- **Granularity**: Function-level (not file-level or line-level)
  - Rationale: Matches semantic units, fits 512-token limit
- **Parser Choice**: Tree-sitter (not regex or language-specific ASTs)
  - Rationale: Polyglot, robust to syntax errors, fast
- **Metadata Storage**: Includes file path, line numbers, function name
  - Rationale: Required for result hydration in Part IV

### Data Flow

```
Source Files (.py, .js, .java, .go, ...)
    │
    ▼
Tree-sitter Parser
    │
    ▼
List[CodeSnippet(code, file_path, start_line, end_line, language, function_name)]
```

## Part II: Vectorization

### Implementation: `clone_detection/embeddings/`

**Purpose**: Transform code snippets into 768-dimensional semantic vectors.

**Key Components**:

1. **`graphcodebert.py`**: GraphCodeBERT embedding generator
   - `GraphCodeBERTEmbedder`: Main embedder class
   - Methods: `embed_batch()`, `embed_single()`
   - Model: `microsoft/graphcodebert-base`

**Design Decisions**:
- **Path A Implementation**: No explicit Data Flow Graph at inference
  - Rationale: Model pre-trained with DFG awareness; complex DFG extraction avoided
  - Trade-off: Simpler implementation, slightly lower theoretical accuracy
- **Embedding Extraction**: `<s>` token's last hidden state
  - Rationale: Pre-trained aggregate representation (not pooler_output or mean pooling)
- **Batch Inference**: Process multiple snippets simultaneously
  - Rationale: GPU efficiency, 10-100x speedup

### Data Flow

```
List[CodeSnippet]
    │
    ▼
RobertaTokenizer (padding, truncation, max_length=512)
    │
    ▼
GraphCodeBERT Forward Pass
    │
    ▼
Extract last_hidden_state[:, 0, :]  # <s> token
    │
    ▼
NumPy Array (N, 768)
```

### Fine-Tuning (Optional but Recommended)

The blueprint specifies a **Triplet Margin Loss** architecture for fine-tuning:

```
Triplet: (anchor, positive, negative)
    │
    ▼
GraphCodeBERT (shared weights)
    │
    ▼
(e_a, e_p, e_n)
    │
    ▼
Loss = max(0, D(e_a, e_p) - D(e_a, e_n) + margin)
```

- **Dataset**: BigCloneBench or SemanticCloneBench
- **Objective**: Push similar code together, dissimilar code apart
- **Result**: Fine-tuned model checkpoint (use instead of base model)

## Part III: FAISS Indexing

### Implementation: `clone_detection/indexing/`

**Purpose**: Build scalable, compressed vector index for fast similarity search.

**Key Components**:

1. **`faiss_index.py`**: FAISS index builder
   - `IndexType` enum: FLAT, IVF_FLAT, IVF_PQ
   - `FAISSIndexBuilder`: Index construction and management
   - Functions: `cosine_to_l2_threshold()`, `l2_to_cosine_similarity()`

**Critical Design Decision: L2 Normalization**

This is the **most important non-obvious step** in the entire system.

**Problem**: GraphCodeBERT is trained for cosine similarity, but FAISS IndexIVFPQ uses L2 distance.

**Solution**: L2-normalize ALL vectors before indexing/searching.

**Mathematical Justification**:

For unit-length vectors (||u|| = 1, ||v|| = 1):

```
D_L2²(u, v) = ||u - v||²
            = (u - v) · (u - v)
            = u·u + v·v - 2·u·v
            = 1 + 1 - 2·u·v
            = 2 - 2·cos_sim(u, v)

Therefore:
D_L2(u, v) = √(2 - 2·cos_sim)
cos_sim = 1 - (D_L2² / 2)
```

**Implementation**:
```python
faiss.normalize_L2(vectors)  # CRITICAL! Do this for ALL vectors
```

### Index Structure: IndexIVFPQ

**IVF (Inverted File)**:
- Partitions space into `nlist` Voronoi cells using k-means
- Query searches only `nprobe` nearest cells
- Speed improvement: O(N) → O(N/nlist × nprobe)

**PQ (Product Quantization)**:
- Splits 768-dim vector into `m` sub-vectors (e.g., 64 sub-vectors of 12 dims)
- Each sub-vector compressed to 8-bit centroid ID
- Memory: 768×4 bytes → m×1 bytes (98% compression)

**Parameter Tuning (Table 3.2)**:

| Parameter | Purpose | Guidelines | Impact |
|-----------|---------|------------|--------|
| `nlist` | IVF clusters | 4√N to 16√N | Training time, search speed |
| `m` | PQ sub-vectors | Must divide 768 (e.g., 64) | Memory, accuracy |
| `nbits` | PQ bits | 8 (standard) | Centroids per sub-space |
| `nprobe` | Query cells | 16 (balanced), 32 (accurate) | Speed vs. accuracy |

### Data Flow

```
Embeddings (N, 768)
    │
    ▼
L2 Normalization (CRITICAL!)
    │
    ▼
Training Phase: k-means clustering
    │  - IVF: nlist centroids
    │  - PQ: m × 2^nbits sub-centroids
    │
    ▼
Add Phase: Assign vectors to cells, compress with PQ
    │
    ▼
IndexIVFPQ (compressed, fast)
```

## Part IV: Query Pipeline

### Implementation: `clone_detection/query/`

**Purpose**: Find clones using threshold-based range search with metadata retrieval.

**Key Components**:

1. **`metadata.py`**: SQLite metadata storage
   - `MetadataStore`: Maps vector IDs → code snippet metadata
   - Schema: (id, code, file_path, start_line, end_line, language, function_name)

2. **`search.py`**: Clone searcher
   - `CloneMatch` dataclass: Result structure
   - `CloneSearcher`: High-level search interface
   - Methods: `find_clones()`, `find_clones_by_location()`

**Design Decisions**:
- **range_search() vs. search()**: Use range search (not k-NN)
  - Rationale: A function may have 0 or 1000 clones; k is arbitrary
- **Threshold Conversion**: Cosine → L2 using formula
  - Implementation: Table 4.1 mappings
- **Result Hydration**: Join FAISS IDs with metadata database
  - Rationale: FAISS only stores vectors, not human-readable locations

### Query Data Flow

```
Query Code (string)
    │
    ▼
GraphCodeBERT Embedding
    │
    ▼
L2 Normalization (CRITICAL!)
    │
    ▼
Cosine Threshold → L2 Threshold Conversion
    │  Example: 0.95 → 0.316
    │
    ▼
FAISS range_search(query, threshold)
    │  Returns: (lims, distances, IDs)
    │
    ▼
L2 Distances → Cosine Similarities
    │
    ▼
Metadata Store Lookup (IDs → file paths, lines, code)
    │
    ▼
List[CloneMatch(file_path, line, similarity, code)]
```

### Cosine-L2 Threshold Table (Table 4.1)

| Cosine Similarity | L2 Distance |
|-------------------|-------------|
| 0.995 (very strict) | 0.100 |
| 0.990 | 0.141 |
| 0.980 (strict) | 0.200 |
| 0.950 (common) | 0.316 |
| 0.900 | 0.447 |
| 0.850 (loose) | 0.548 |

## Part V: Scaling & Production

### GPU Acceleration

**Embedding Generation**:
- Move model to CUDA: `model.to("cuda")`
- Batch inference: Process 32-128 snippets simultaneously
- Speedup: 10-100x vs. CPU

**FAISS Index**:
- Install `faiss-gpu` (via conda)
- Move index to GPU: `faiss.index_cpu_to_gpu()`
- Query latency: 5-15ms for 10M vectors

### Horizontal Sharding

For billion-scale indexes:
```
Index Shard 1 (100M vectors) ──┐
Index Shard 2 (100M vectors) ──┼─▶ Master Node ─▶ Merged Results
Index Shard 3 (100M vectors) ──┤
...                             │
Index Shard N (100M vectors) ──┘
```

Each shard runs `range_search()` in parallel; master merges results.

### Evaluation Framework

**Benchmarks**:
- BigCloneBench (BCB): Industry standard
- SemanticCloneBench (SCB): Type-4 focused

**Metrics**:
- Precision: TP / (TP + FP)
- Recall: TP / (TP + FN)
- F1-score: 2·(P·R) / (P + R)

**Process**:
1. Index benchmark "training" set
2. Query each function in "test" set
3. Compare results to ground truth
4. Sweep similarity threshold to generate P-R curve

## CLI Architecture

### Command Structure

```
clone-detect
├── ingest      # Blueprint A: Batch Ingestion Pipeline
├── search      # Blueprint B: Query Pipeline
└── info        # Index inspection
```

### Ingest Pipeline (Blueprint A)

```
Source Directory
    │
    ▼
TreeSitterParser.parse_directory()
    │
    ▼
GraphCodeBERTEmbedder.embed_batch()
    │
    ▼
FAISSIndexBuilder.build()
    │  - train(training_sample)
    │  - add(all_vectors)
    │
    ▼
Save: index.faiss + metadata.db
```

### Search Pipeline (Blueprint B)

```
Query (code or file:line)
    │
    ▼
Load: index.faiss + metadata.db + embedder
    │
    ▼
CloneSearcher.find_clones()
    │  - Embed query
    │  - range_search()
    │  - Hydrate results
    │
    ▼
Display: List[CloneMatch]
```

## Configuration System

### Pydantic Models

- `ModelConfig`: GraphCodeBERT settings
- `IndexConfig`: FAISS parameters
- `ParsingConfig`: Language and exclusions
- `QueryConfig`: Search defaults

### YAML Configuration

```yaml
model:
  name: microsoft/graphcodebert-base
  device: cuda
index:
  nlist: 4096
  nprobe: 16
parsing:
  languages: [python, java]
query:
  default_similarity: 0.95
```

## Error Handling and Edge Cases

1. **Empty codebase**: Return early with warning
2. **Unsupported languages**: Skip files, log warning
3. **Truncated functions**: GraphCodeBERT truncates to 512 tokens
4. **Exact self-matches**: Optionally filtered (similarity ≈ 1.0)
5. **Missing metadata**: Log warning, skip result

## Performance Characteristics

**10M Function Index (NVIDIA A100)**:

| Operation | Time | Throughput |
|-----------|------|------------|
| Parse (Python) | ~10 min | 16,667 func/min |
| Embed (batch=64) | ~2 hours | 1,389 func/min |
| Index training | ~15 min | - |
| Index population | ~10 min | - |
| Single query (nprobe=16) | 5-15 ms | 66-200 qps |

**Memory Footprint**:

| Component | Size (10M vectors) |
|-----------|--------------------|
| Raw embeddings | 30.7 GB |
| IndexFlatL2 | 30.7 GB |
| IndexIVFPQ (m=64) | 640 MB (98% reduction) |
| Metadata DB | ~5 GB |

## Security Considerations

1. **Code Injection**: User-provided code is NOT executed; only embedded
2. **Path Traversal**: File paths validated before database storage
3. **SQL Injection**: Parameterized queries used throughout
4. **Model Security**: Use official HuggingFace models or verified checkpoints

## Future Enhancements

1. **Incremental Indexing**: Support for adding new vectors without full rebuild
2. **Multi-Modal Search**: Natural language queries ("find sorting algorithms")
3. **Cluster Analysis**: Identify large clone families for refactoring
4. **Real-Time IDE Integration**: VSCode/IntelliJ plugins
5. **Cross-Language Clones**: Detect Python ↔ Java semantic equivalence

## References

1. Guo et al. (2020). "GraphCodeBERT: Pre-training Code Representations with Data Flow"
2. Johnson et al. (2019). "Billion-scale similarity search with GPUs" (FAISS paper)
3. Tree-sitter documentation: https://tree-sitter.github.io/
4. FAISS wiki: https://github.com/facebookresearch/faiss/wiki

---

For implementation details, see the source code in `clone_detection/` directory.
