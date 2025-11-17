"""
Semantic Code Clone Detection using GraphCodeBERT and FAISS.

This package provides a production-grade system for detecting semantic code clones
across multiple programming languages using deep learning and approximate nearest
neighbor search.
"""

__version__ = "0.1.0"
__author__ = "structurelint"

from clone_detection.parsers.tree_sitter_parser import CodeSnippet, TreeSitterParser
from clone_detection.embeddings.graphcodebert import GraphCodeBERTEmbedder
from clone_detection.indexing.faiss_index import FAISSIndexBuilder
from clone_detection.query.search import CloneSearcher, CloneMatch

__all__ = [
    "TreeSitterParser",
    "CodeSnippet",
    "GraphCodeBERTEmbedder",
    "FAISSIndexBuilder",
    "CloneSearcher",
    "CloneMatch",
]
