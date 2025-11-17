# FAISS Indexing

## Overview

High-performance vector indexing using FAISS IndexIVFPQ for scalable similarity search.

## Index Configuration

- **Index Type**: IndexIVFPQ (Inverted File with Product Quantization)
- **Clusters (nlist)**: 4096
- **PQ Subquantizers (m)**: 64
- **Bits per code (nbits)**: 8
- **Search probes (nprobe)**: 16

## Key Features

- **L2 Normalization**: Enables cosine similarity search via L2 distance
- **Threshold-based Search**: Uses range_search for configurable similarity thresholds
- **Scalability**: Optimized for millions of code snippets

## Mathematical Foundation

The L2 normalization technique converts cosine similarity to L2 distance:
```
D_L2² = 2 - 2·cos_sim
```
