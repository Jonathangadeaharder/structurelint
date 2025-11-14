# export

⬆️ **[Parent Directory](../README.md)**

## Overview

The `export` package handles exporting dependency graphs in various visualization formats for analysis and documentation purposes.

## Supported Formats

- **DOT**: Graphviz DOT format for rendering with graphviz tools
- **Mermaid**: Mermaid diagram format for embedding in markdown
- **JSON**: Machine-readable JSON format with nodes and edges

## Usage

```go
exporter := export.NewGraphExporter(importGraph)

// Export as DOT format (for Graphviz)
dotOutput := exporter.ExportDOT()

// Export as Mermaid format (for GitHub/GitLab markdown)
mermaidOutput := exporter.ExportMermaid()

// Export as JSON
jsonOutput := exporter.ExportJSON()
```

## CLI Usage

```bash
# Export graph in DOT format and render with Graphviz
structurelint --export-graph dot . | dot -Tpng -o graph.png

# Export graph in Mermaid format
structurelint --export-graph mermaid . > ARCHITECTURE.md

# Export graph in JSON format
structurelint --export-graph json . > graph.json
```

## Features

- **Layer-aware**: Visualizations group files by architectural layers
- **Deterministic output**: Consistent ordering for version control
- **Multiple formats**: Choose the best format for your use case
