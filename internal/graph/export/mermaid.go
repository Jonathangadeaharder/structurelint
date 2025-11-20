package export

import (
	"fmt"
	"io"

	"github.com/structurelint/structurelint/internal/graph"
)

// MermaidExporter exports dependency graphs in Mermaid format
type MermaidExporter struct {
	graph   *graph.ImportGraph
	options MermaidOptions
}

// MermaidOptions configures the Mermaid export
type MermaidOptions struct {
	// Title of the graph
	Title string

	// ShowLayers colors nodes by their layer
	ShowLayers bool

	// HighlightViolations marks illegal dependencies with red edges
	HighlightViolations bool

	// FilterLayer only shows files in this layer (empty = all)
	FilterLayer string

	// MaxDepth limits dependency depth (0 = unlimited)
	MaxDepth int

	// ShowCycles highlights circular dependencies
	ShowCycles bool

	// SimplifyPaths shortens file paths for readability
	SimplifyPaths bool

	// Direction of the graph (LR, RL, TB, BT)
	Direction string
}

// NewMermaidExporter creates a new Mermaid exporter
func NewMermaidExporter(g *graph.ImportGraph, options MermaidOptions) *MermaidExporter {
	if options.Title == "" {
		options.Title = "Dependency Graph"
	}
	if options.Direction == "" {
		options.Direction = "LR"
	}
	return &MermaidExporter{
		graph:   g,
		options: options,
	}
}

// Export writes the graph in Mermaid format to the writer
func (e *MermaidExporter) Export(w io.Writer) error {
	// Write header
	if _, err := fmt.Fprintf(w, "graph %s\n", e.options.Direction); err != nil {
		return fmt.Errorf("failed to write Mermaid header: %w", err)
	}

	// Get nodes to display (reuse DOT exporter logic)
	dotExporter := NewDOTExporter(e.graph, DOTOptions{
		FilterLayer:   e.options.FilterLayer,
		MaxDepth:      e.options.MaxDepth,
		SimplifyPaths: e.options.SimplifyPaths,
	})
	nodes := dotExporter.getFilteredNodes()

	if len(nodes) == 0 {
		if _, err := fmt.Fprintf(w, "  %% No nodes to display\n"); err != nil {
			return fmt.Errorf("failed to write Mermaid comment: %w", err)
		}
		return nil
	}

	// Detect cycles if needed
	cycles := make(map[string]map[string]bool)
	if e.options.ShowCycles {
		cycles = dotExporter.detectAllCycles()
	}

	// Create node ID map
	nodeIDs := make(map[string]string)
	for i, node := range nodes {
		nodeIDs[node] = fmt.Sprintf("n%d", i)
	}

	// Add edges (Mermaid defines nodes implicitly through edges)
	for _, fromNode := range nodes {
		fromID := nodeIDs[fromNode]
		fromLabel := fromNode
		if e.options.SimplifyPaths {
			fromLabel = dotExporter.simplifyPath(fromNode)
		}

		deps := e.graph.GetDependencies(fromNode)
		hasEdges := false

		for _, toNode := range deps {
			// Only show edge if target node is in our filtered set
			toID, exists := nodeIDs[toNode]
			if !exists {
				continue
			}

			hasEdges = true
			toLabel := toNode
			if e.options.SimplifyPaths {
				toLabel = dotExporter.simplifyPath(toNode)
			}

			// Check if this is a cycle or violation
			isCycle := cycles[fromNode] != nil && cycles[fromNode][toNode]
			isViolation := dotExporter.isViolation(fromNode, toNode)

			// Determine edge style
			var edgeStyle string
			if isCycle && e.options.ShowCycles {
				edgeStyle = fmt.Sprintf("  %s[\"%s\"] -.->|cycle| %s[\"%s\"]\n",
					fromID, fromLabel, toID, toLabel)
			} else if isViolation && e.options.HighlightViolations {
				edgeStyle = fmt.Sprintf("  %s[\"%s\"] -.->|violation| %s[\"%s\"]\n",
					fromID, fromLabel, toID, toLabel)
			} else {
				edgeStyle = fmt.Sprintf("  %s[\"%s\"] --> %s[\"%s\"]\n",
					fromID, fromLabel, toID, toLabel)
			}

			if _, err := fmt.Fprintf(w, "%s", edgeStyle); err != nil {
				return fmt.Errorf("failed to write Mermaid edge: %w", err)
			}
		}

		// If node has no edges, define it explicitly
		if !hasEdges {
			if _, err := fmt.Fprintf(w, "  %s[\"%s\"]\n", fromID, fromLabel); err != nil {
				return fmt.Errorf("failed to write Mermaid node: %w", err)
			}
		}
	}

	// Add styling if showing layers
	if e.options.ShowLayers {
		e.writeStyles(w, nodeIDs, nodes)
	}

	return nil
}

// writeStyles adds CSS styling for layer colors
func (e *MermaidExporter) writeStyles(w io.Writer, nodeIDs map[string]string, nodes []string) {
	_, _ = fmt.Fprintf(w, "\n  %% Layer styling\n")

	// Color scheme for different layers (Mermaid uses fill and stroke)
	layerStyles := map[string]string{
		"domain":         "fill:#C8E6C9,stroke:#2E7D32,stroke-width:2px",
		"application":    "fill:#BBDEFB,stroke:#1565C0,stroke-width:2px",
		"infrastructure": "fill:#FFCDD2,stroke:#C62828,stroke-width:2px",
		"presentation":   "fill:#FFE0B2,stroke:#F57C00,stroke-width:2px",
		"api":            "fill:#E1BEE7,stroke:#6A1B9A,stroke-width:2px",
		"cmd":            "fill:#E0E0E0,stroke:#424242,stroke-width:2px",
		"internal":       "fill:#B2DFDB,stroke:#00695C,stroke-width:2px",
	}

	// Group nodes by layer
	layerNodes := make(map[string][]string)
	for _, node := range nodes {
		layer := e.graph.GetLayerForFile(node)
		if layer != nil {
			layerNodes[layer.Name] = append(layerNodes[layer.Name], node)
		}
	}

	// Apply styles to each layer's nodes
	for layerName, layerNodeList := range layerNodes {
		style, ok := layerStyles[layerName]
		if !ok {
			continue
		}

		for _, node := range layerNodeList {
			nodeID := nodeIDs[node]
			_, _ = fmt.Fprintf(w, "  style %s %s\n", nodeID, style)
		}
	}

	// Style for violations (if enabled)
	if e.options.HighlightViolations {
		_, _ = fmt.Fprintf(w, "  linkStyle default stroke:#333,stroke-width:1px\n")
	}

	// Style for cycles (if enabled)
	if e.options.ShowCycles {
		_, _ = fmt.Fprintf(w, "  linkStyle default stroke:#333,stroke-width:1px\n")
	}
}

// ExportWithWrapper wraps the Mermaid graph in markdown code fences
func (e *MermaidExporter) ExportWithWrapper(w io.Writer) error {
	_, _ = fmt.Fprintf(w, "```mermaid\n")
	if err := e.Export(w); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "```\n")
	return nil
}

// ExportHTML generates an HTML file with embedded Mermaid
func (e *MermaidExporter) ExportHTML(w io.Writer) error {
	_, _ = fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>%s</title>
  <script src="https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"></script>
  <style>
    body {
      font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
      margin: 20px;
      background: #f5f5f5;
    }
    .container {
      max-width: 100%%;
      background: white;
      padding: 20px;
      border-radius: 8px;
      box-shadow: 0 2px 4px rgba(0,0,0,0.1);
    }
    h1 {
      color: #333;
      margin-top: 0;
    }
    .mermaid {
      text-align: center;
    }
  </style>
</head>
<body>
  <div class="container">
    <h1>%s</h1>
    <div class="mermaid">
`, e.options.Title, e.options.Title)

	if err := e.Export(w); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(w, `    </div>
  </div>
  <script>
    mermaid.initialize({ startOnLoad: true, theme: 'default' });
  </script>
</body>
</html>
`)

	return nil
}
