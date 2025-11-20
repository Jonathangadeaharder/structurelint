# Graph Export

## Overview

This package provides graph export to various visualization formats.

## Components

- **dot.go**: Graphviz DOT format export
- **mermaid.go**: Mermaid diagram export

## Supported Formats

### DOT (Graphviz)
- Standard graph visualization format
- Compatible with Graphviz tools
- Supports directed graphs, clusters, and styling

### Mermaid
- Markdown-compatible diagrams
- GitHub/GitLab native rendering
- Interactive in documentation

## Usage

### DOT Export
```go
exporter := export.NewDOTExporter()
dot := exporter.Export(graph)
os.WriteFile("graph.dot", []byte(dot), 0644)
```

### Mermaid Export
```go
exporter := export.NewMermaidExporter()
mermaid := exporter.Export(graph)
os.WriteFile("graph.mmd", []byte(mermaid), 0644)
```

## Visualization

DOT files can be rendered with:
```bash
dot -Tpng graph.dot -o graph.png
```
