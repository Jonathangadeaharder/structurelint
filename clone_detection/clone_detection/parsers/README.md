# Code Parsers

## Overview

Multi-language code parsing module using Tree-sitter for extracting function-level code snippets.

## Supported Languages

- Python
- JavaScript
- Java
- Go
- C++
- C#

## Components

- **language_configs.py**: Language-specific S-expression queries
- **tree_sitter_parser.py**: TreeSitterParser class for code extraction

## Features

- Function-level code extraction
- Metadata tracking (file path, line numbers, function names)
- Batch directory parsing
