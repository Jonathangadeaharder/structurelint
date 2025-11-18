"""
Tree-sitter-based code parser for extracting function-level code snippets.

This module implements Part I of the blueprint: Ingestion & Parsing.
It uses Tree-sitter to parse source code files and extract discrete,
semantically-coherent units (functions and methods) across multiple languages.
"""

import importlib
import logging
from dataclasses import dataclass
from pathlib import Path
from typing import Dict, List, Optional, Set

from tree_sitter import Language, Parser, Tree

from clone_detection.parsers.language_configs import (
    LANGUAGE_CONFIGS,
    LanguageConfig,
    get_language_for_file,
)

logger = logging.getLogger(__name__)


@dataclass
class CodeSnippet:
    """
    Represents a parsed code snippet with metadata.

    Attributes:
        code: The raw source code of the function
        file_path: Path to the source file
        start_line: Starting line number (1-indexed)
        end_line: Ending line number (1-indexed)
        language: Programming language
        function_name: Optional name of the function
    """

    code: str
    file_path: str
    start_line: int
    end_line: int
    language: str
    function_name: Optional[str] = None

    def __repr__(self) -> str:
        return (
            f"CodeSnippet(file={self.file_path}, "
            f"lines={self.start_line}-{self.end_line}, "
            f"lang={self.language}, "
            f"name={self.function_name})"
        )

    def to_dict(self) -> Dict:
        """Convert to dictionary for serialization."""
        return {
            "code": self.code,
            "file_path": self.file_path,
            "start_line": self.start_line,
            "end_line": self.end_line,
            "language": self.language,
            "function_name": self.function_name,
        }


class TreeSitterParser:
    """
    Multi-language code parser using Tree-sitter.

    This parser implements the extraction strategy described in Section 1.3
    of the blueprint, using S-expression queries to find function definitions
    across multiple programming languages.

    Example:
        >>> parser = TreeSitterParser(languages=["python", "java"])
        >>> snippets = parser.parse_file("example.py")
        >>> print(f"Found {len(snippets)} functions")
    """

    def __init__(self, languages: Optional[List[str]] = None):
        """
        Initialize the parser with specified languages.

        Args:
            languages: List of language names to support. If None, all available
                      languages are enabled.
        """
        self.enabled_languages = languages or list(LANGUAGE_CONFIGS.keys())
        self.parsers: Dict[str, Parser] = {}
        self.queries: Dict[str, object] = {}

        self._initialize_parsers()

    def _initialize_parsers(self) -> None:
        """
        Load Tree-sitter grammars and compile queries for each enabled language.

        This implements the initialization described in Section 1.3.
        """
        for lang_name in self.enabled_languages:
            if lang_name not in LANGUAGE_CONFIGS:
                logger.warning(f"Unknown language: {lang_name}, skipping")
                continue

            config = LANGUAGE_CONFIGS[lang_name]

            try:
                # Dynamically import the language grammar module
                # e.g., "tree_sitter_python" -> tree_sitter_python.language()
                grammar_module = importlib.import_module(config.grammar_module)
                language_func = grammar_module.language
                language = Language(language_func())

                # Create parser
                parser = Parser(language)
                self.parsers[lang_name] = parser

                # Compile the S-expression query
                query = language.query(config.function_query)
                self.queries[lang_name] = query

                logger.info(f"Initialized parser for {lang_name}")

            except ImportError as e:
                logger.error(
                    f"Failed to import grammar for {lang_name}: {e}. "
                    f"Install with: pip install {config.grammar_module.replace('_', '-')}"
                )
            except Exception as e:
                logger.error(f"Failed to initialize parser for {lang_name}: {e}")

    def parse_file(self, file_path: str) -> List[CodeSnippet]:
        """
        Parse a single source file and extract all function snippets.

        Args:
            file_path: Path to the source code file

        Returns:
            List of extracted code snippets

        Example:
            >>> parser = TreeSitterParser(languages=["python"])
            >>> snippets = parser.parse_file("my_module.py")
        """
        file_path = str(Path(file_path).resolve())

        # Determine language from file extension
        lang_name = get_language_for_file(file_path)
        if lang_name is None:
            logger.debug(f"Unsupported file type: {file_path}")
            return []

        if lang_name not in self.parsers:
            logger.debug(f"Parser not initialized for {lang_name}: {file_path}")
            return []

        # Read file content
        try:
            with open(file_path, "rb") as f:
                source_bytes = f.read()
        except Exception as e:
            logger.error(f"Failed to read file {file_path}: {e}")
            return []

        return self._parse_source(source_bytes, file_path, lang_name)

    def _parse_source(
        self, source_bytes: bytes, file_path: str, lang_name: str
    ) -> List[CodeSnippet]:
        """
        Parse source code bytes and extract function snippets.

        This implements the extraction logic from Section 1.3:
        1. Parse the file into an AST
        2. Run the S-expression query
        3. Extract function metadata and code

        Args:
            source_bytes: Raw source code as bytes
            file_path: Path to the source file
            lang_name: Programming language name

        Returns:
            List of extracted code snippets
        """
        parser = self.parsers[lang_name]
        query = self.queries[lang_name]
        config = LANGUAGE_CONFIGS[lang_name]

        # Parse the source code
        tree: Tree = parser.parse(source_bytes)

        # Run the query to find all function definitions
        captures = query.captures(tree.root_node)

        snippets = []
        for node, capture_name in captures:
            if capture_name == config.capture_name:
                # Extract the raw text of the function
                function_code = node.text.decode("utf8")

                # Extract metadata
                # Note: start_point and end_point are 0-indexed (row, column)
                # We add 1 to convert to 1-indexed line numbers
                start_line = node.start_point[0] + 1
                end_line = node.end_point[0] + 1

                # Try to extract function name (if available in captures)
                function_name = self._extract_function_name(node, source_bytes)

                snippet = CodeSnippet(
                    code=function_code,
                    file_path=file_path,
                    start_line=start_line,
                    end_line=end_line,
                    language=lang_name,
                    function_name=function_name,
                )
                snippets.append(snippet)

        logger.debug(f"Extracted {len(snippets)} functions from {file_path}")
        return snippets

    def _extract_function_name(self, node, source_bytes: bytes) -> Optional[str]:
        """
        Extract the function name from a function definition node.

        Args:
            node: Tree-sitter node representing the function
            source_bytes: Source code bytes

        Returns:
            Function name if found, None otherwise
        """
        # Look for a child node of type "identifier" or "name"
        for child in node.children:
            if child.type in ["identifier", "name"]:
                return child.text.decode("utf8")

            # Recursive search in case the name is nested
            for subchild in child.children:
                if subchild.type in ["identifier", "name"]:
                    return subchild.text.decode("utf8")

        return None

    def parse_directory(
        self,
        directory: str,
        exclude_patterns: Optional[List[str]] = None,
        max_files: Optional[int] = None,
    ) -> List[CodeSnippet]:
        """
        Recursively parse all source files in a directory.

        This implements the "CodebaseWalker" component from Blueprint A
        (Batch Ingestion Pipeline).

        Args:
            directory: Root directory to scan
            exclude_patterns: List of glob patterns to exclude (e.g., "*.test.py")
            max_files: Maximum number of files to process (for testing)

        Returns:
            List of all extracted code snippets

        Example:
            >>> parser = TreeSitterParser(languages=["python"])
            >>> snippets = parser.parse_directory(
            ...     "/path/to/codebase",
            ...     exclude_patterns=["**/test_*.py", "**/__pycache__/**"]
            ... )
        """
        from pathspec import PathSpec
        from pathspec.patterns import GitWildMatchPattern

        # Build exclusion pathspec
        exclude_spec = None
        if exclude_patterns:
            exclude_spec = PathSpec.from_lines(GitWildMatchPattern, exclude_patterns)

        # Collect supported file extensions
        supported_extensions: Set[str] = set()
        for lang_name in self.enabled_languages:
            if lang_name in LANGUAGE_CONFIGS:
                supported_extensions.update(LANGUAGE_CONFIGS[lang_name].extensions)

        # Walk directory and collect files
        all_snippets = []
        file_count = 0

        root_path = Path(directory).resolve()
        for file_path in root_path.rglob("*"):
            # Skip directories
            if not file_path.is_file():
                continue

            # Check if file has a supported extension
            if not any(str(file_path).endswith(ext) for ext in supported_extensions):
                continue

            # Check exclusion patterns
            relative_path = file_path.relative_to(root_path)
            if exclude_spec and exclude_spec.match_file(str(relative_path)):
                logger.debug(f"Excluded: {relative_path}")
                continue

            # Parse file
            snippets = self.parse_file(str(file_path))
            all_snippets.extend(snippets)

            file_count += 1
            if max_files and file_count >= max_files:
                logger.info(f"Reached max_files limit: {max_files}")
                break

        logger.info(
            f"Parsed {file_count} files, extracted {len(all_snippets)} function snippets"
        )
        return all_snippets

    def get_supported_languages(self) -> List[str]:
        """Get list of languages with initialized parsers."""
        return list(self.parsers.keys())
