# GraphCodeBERT Embeddings

## Overview

Neural code embedding generation using Microsoft's GraphCodeBERT model.

## Model

- **Model ID**: microsoft/graphcodebert-base
- **Embedding Dimension**: 768
- **Extraction Method**: `<s>` token's last hidden state

## Features

- Batch inference support
- GPU/CPU auto-detection
- Semantic code representation
- Pre-trained on Data Flow Graphs

## Usage

```python
from clone_detection.embeddings.graphcodebert import GraphCodeBERTEmbedder

embedder = GraphCodeBERTEmbedder()
embeddings = embedder.encode_batch(code_snippets)
```
