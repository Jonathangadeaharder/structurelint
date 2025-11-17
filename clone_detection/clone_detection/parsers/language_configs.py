"""
Language configurations for Tree-sitter parsing.

This module defines the S-expression queries and configurations for extracting
functions and methods from various programming languages as specified in Table 1.1
of the blueprint.
"""

from dataclasses import dataclass
from typing import List, Optional


@dataclass
class LanguageConfig:
    """Configuration for a specific programming language."""

    name: str
    extensions: List[str]
    grammar_module: str
    function_query: str
    capture_name: str = "function.definition"


# Table 1.1: Tree-sitter Queries for Function Extraction
LANGUAGE_CONFIGS = {
    "python": LanguageConfig(
        name="python",
        extensions=[".py"],
        grammar_module="tree_sitter_python",
        function_query="""
            (function_definition
              name: (identifier) @function.name
              body: (block) @function.body) @function.definition
        """,
        capture_name="function.definition",
    ),
    "javascript": LanguageConfig(
        name="javascript",
        extensions=[".js", ".jsx", ".mjs"],
        grammar_module="tree_sitter_javascript",
        function_query="""
            [
                (function_declaration) @function.definition
                (arrow_function) @function.definition
                (method_definition) @function.definition
            ]
        """,
        capture_name="function.definition",
    ),
    "java": LanguageConfig(
        name="java",
        extensions=[".java"],
        grammar_module="tree_sitter_java",
        function_query="""
            (method_declaration) @function.definition
        """,
        capture_name="function.definition",
    ),
    "go": LanguageConfig(
        name="go",
        extensions=[".go"],
        grammar_module="tree_sitter_go",
        function_query="""
            (function_declaration) @function.definition
        """,
        capture_name="function.definition",
    ),
    "cpp": LanguageConfig(
        name="cpp",
        extensions=[".cpp", ".cc", ".cxx", ".hpp", ".h"],
        grammar_module="tree_sitter_cpp",
        function_query="""
            (function_definition) @function.definition
        """,
        capture_name="function.definition",
    ),
    "csharp": LanguageConfig(
        name="csharp",
        extensions=[".cs"],
        grammar_module="tree_sitter_c_sharp",
        function_query="""
            (method_declaration) @function.definition
        """,
        capture_name="function.definition",
    ),
}


def get_language_for_file(file_path: str) -> Optional[str]:
    """
    Determine the programming language based on file extension.

    Args:
        file_path: Path to the source file

    Returns:
        Language name if recognized, None otherwise
    """
    file_path_lower = file_path.lower()
    for lang_name, config in LANGUAGE_CONFIGS.items():
        for ext in config.extensions:
            if file_path_lower.endswith(ext):
                return lang_name
    return None
