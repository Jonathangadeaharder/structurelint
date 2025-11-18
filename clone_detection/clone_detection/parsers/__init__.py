"""
Part I: Ingestion & Parsing

Tree-sitter-based polyglot code parsing for extracting function-level code snippets.
"""

from clone_detection.parsers.tree_sitter_parser import CodeSnippet, TreeSitterParser
from clone_detection.parsers.language_configs import LANGUAGE_CONFIGS, LanguageConfig

__all__ = [
    "TreeSitterParser",
    "CodeSnippet",
    "LANGUAGE_CONFIGS",
    "LanguageConfig",
]
