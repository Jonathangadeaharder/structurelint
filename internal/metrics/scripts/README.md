# Metrics Scripts

This directory contains language-specific scripts for calculating code complexity metrics in Python, JavaScript, and TypeScript files.

## Overview

These scripts are invoked by the Go `MultiLanguageAnalyzer` to compute:
- **Cognitive Complexity**: Measures code understandability based on control flow and nesting
- **Halstead Metrics**: Measures cognitive load based on operators and operands

## Files

- `python_metrics.py` - Python AST-based metrics calculator
- `js_metrics.js` - JavaScript/TypeScript metrics calculator using Babel parser
- `package.json` - Node.js dependencies for the JS metrics script

## Installation

Before using JavaScript/TypeScript metrics, install the required Node.js dependencies:

```bash
cd internal/metrics/scripts
npm install
```

This will install `@babel/parser` locally in this directory.

## Usage

These scripts are called automatically by structurelint. They can also be run directly:

```bash
# Python cognitive complexity
python3 python_metrics.py cognitive-complexity /path/to/file.py

# Python Halstead metrics
python3 python_metrics.py halstead /path/to/file.py

# JavaScript cognitive complexity
node js_metrics.js cognitive-complexity /path/to/file.js

# JavaScript Halstead metrics
node js_metrics.js halstead /path/to/file.js
```

## Output Format

All scripts output JSON with the following structure:

```json
{
  "functions": [
    {
      "name": "function_name",
      "start_line": 10,
      "end_line": 25,
      "complexity": 5,
      "value": 123.45
    }
  ],
  "file_level": {
    "total": 15.0,
    "average": 7.5,
    "max": 10.0,
    "function_count": 2.0
  }
}
```

On error, the output includes an "error" key with the error message.

## Requirements

- **Python**: Python 3.x with standard library (no external dependencies)
- **Node.js**: Node.js 14+ with `@babel/parser` (installed via npm)
