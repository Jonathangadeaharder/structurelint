// Package export provides graph visualization and export functionality
package export

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/graph"
)

// DOTExporter exports dependency graphs in GraphViz DOT format
type DOTExporter struct {
	graph   *graph.ImportGraph
	options DOTOptions
}

// DOTOptions configures the DOT export
type DOTOptions struct {
	// Title of the graph
	Title string

	// ShowLayers colors nodes by their layer
	ShowLayers bool

	// HighlightViolations marks illegal dependencies in red
	HighlightViolations bool

	// FilterLayer only shows files in this layer (empty = all)
	FilterLayer string

	// MaxDepth limits dependency depth (0 = unlimited)
	MaxDepth int

	// ShowCycles highlights circular dependencies
	ShowCycles bool

	// SimplifyPaths shortens file paths for readability
	SimplifyPaths bool
}

// NewDOTExporter creates a new DOT exporter
func NewDOTExporter(g *graph.ImportGraph, options DOTOptions) *DOTExporter {
	if options.Title == "" {
		options.Title = "Dependency Graph"
	}
	return &DOTExporter{
		graph:   g,
		options: options,
	}
}

// Export writes the graph in DOT format to the writer
func (e *DOTExporter) Export(w io.Writer) error {
	// Write header
	if _, err := fmt.Fprintf(w, "digraph \"%s\" {\n", e.options.Title); err != nil {
		return fmt.Errorf("failed to write DOT header: %w", err)
	}
	if _, err := fmt.Fprintf(w, "  rankdir=LR;\n"); err != nil {
		return fmt.Errorf("failed to write DOT rankdir: %w", err)
	}
	if _, err := fmt.Fprintf(w, "  node [shape=box, style=rounded];\n"); err != nil {
		return fmt.Errorf("failed to write DOT node style: %w", err)
	}
	if _, err := fmt.Fprintf(w, "  edge [arrowhead=vee];\n\n"); err != nil {
		return fmt.Errorf("failed to write DOT edge style: %w", err)
	}

	// Get nodes to display
	nodes := e.getFilteredNodes()
	if len(nodes) == 0 {
		if _, err := fmt.Fprintf(w, "  // No nodes to display\n"); err != nil {
			return fmt.Errorf("failed to write DOT comment: %w", err)
		}
		if _, err := fmt.Fprintf(w, "}\n"); err != nil {
			return fmt.Errorf("failed to write DOT closing: %w", err)
		}
		return nil
	}

	// Detect cycles if needed
	cycles := make(map[string]map[string]bool)
	if e.options.ShowCycles {
		cycles = e.detectAllCycles()
	}

	// Define nodes with colors
	nodeIDs := make(map[string]string)
	for i, node := range nodes {
		nodeID := fmt.Sprintf("n%d", i)
		nodeIDs[node] = nodeID

		label := node
		if e.options.SimplifyPaths {
			label = e.simplifyPath(node)
		}

		// Determine node color based on layer
		color := e.getNodeColor(node)
		fillColor := e.getNodeFillColor(node)

		if _, err := fmt.Fprintf(w, "  %s [label=\"%s\", color=\"%s\", fillcolor=\"%s\", style=\"rounded,filled\"];\n",
			nodeID, label, color, fillColor); err != nil {
			return fmt.Errorf("failed to write DOT node: %w", err)
		}
	}

	if _, err := fmt.Fprintf(w, "\n"); err != nil {
		return fmt.Errorf("failed to write DOT newline: %w", err)
	}

	// Add edges
	for _, fromNode := range nodes {
		fromID := nodeIDs[fromNode]
		deps := e.graph.GetDependencies(fromNode)

		for _, toNode := range deps {
			// Only show edge if target node is in our filtered set
			toID, exists := nodeIDs[toNode]
			if !exists {
				continue
			}

			// Check if this is a cycle
			isCycle := cycles[fromNode] != nil && cycles[fromNode][toNode]

			// Check if this is a violation
			isViolation := e.isViolation(fromNode, toNode)

			// Determine edge style
			edgeColor := "black"
			edgeStyle := "solid"
			edgeWidth := "1.0"

			if isCycle && e.options.ShowCycles {
				edgeColor = "orange"
				edgeWidth = "2.0"
				edgeStyle = "bold"
			}

			if isViolation && e.options.HighlightViolations {
				edgeColor = "red"
				edgeWidth = "2.0"
				edgeStyle = "bold"
			}

			if _, err := fmt.Fprintf(w, "  %s -> %s [color=\"%s\", style=\"%s\", penwidth=%s];\n",
				fromID, toID, edgeColor, edgeStyle, edgeWidth); err != nil {
				return fmt.Errorf("failed to write DOT edge: %w", err)
			}
		}
	}

	// Add legend if showing layers
	if e.options.ShowLayers {
		e.writeLegend(w)
	}

	if _, err := fmt.Fprintf(w, "}\n"); err != nil {
		return fmt.Errorf("failed to write DOT closing brace: %w", err)
	}
	return nil
}

// getFilteredNodes returns nodes to display based on filter options
func (e *DOTExporter) getFilteredNodes() []string {
	var nodes []string

	// Get all files from graph
	allFiles := e.graph.AllFiles
	if len(allFiles) == 0 {
		// Fallback: collect from dependencies map
		for file := range e.graph.Dependencies {
			allFiles = append(allFiles, file)
		}
	}

	// Apply layer filter
	for _, file := range allFiles {
		if e.options.FilterLayer != "" {
			layer := e.graph.GetLayerForFile(file)
			if layer == nil || layer.Name != e.options.FilterLayer {
				continue
			}
		}
		nodes = append(nodes, file)
	}

	// Apply depth filter if specified
	if e.options.MaxDepth > 0 {
		nodes = e.filterByDepth(nodes, e.options.MaxDepth)
	}

	return nodes
}

// filterByDepth limits nodes by dependency depth
func (e *DOTExporter) filterByDepth(nodes []string, maxDepth int) []string {
	// Start with nodes that have no incoming dependencies (roots)
	roots := make([]string, 0)
	for _, node := range nodes {
		if e.graph.IncomingRefs[node] == 0 {
			roots = append(roots, node)
		}
	}

	// If no roots found, use first node
	if len(roots) == 0 && len(nodes) > 0 {
		roots = []string{nodes[0]}
	}

	// BFS to find nodes within depth
	visited := make(map[string]bool)
	depthMap := make(map[string]int)
	queue := make([]string, len(roots))
	copy(queue, roots)

	for _, root := range roots {
		depthMap[root] = 0
		visited[root] = true
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		currentDepth := depthMap[current]
		if currentDepth >= maxDepth {
			continue
		}

		deps := e.graph.GetDependencies(current)
		for _, dep := range deps {
			if !visited[dep] {
				visited[dep] = true
				depthMap[dep] = currentDepth + 1
				queue = append(queue, dep)
			}
		}
	}

	// Return only visited nodes
	filtered := make([]string, 0, len(visited))
	for _, node := range nodes {
		if visited[node] {
			filtered = append(filtered, node)
		}
	}

	return filtered
}

// getNodeColor returns the border color for a node based on its layer
func (e *DOTExporter) getNodeColor(node string) string {
	if !e.options.ShowLayers {
		return "black"
	}

	layer := e.graph.GetLayerForFile(node)
	if layer == nil {
		return "gray"
	}

	// Color scheme for different layers
	colors := map[string]string{
		"domain":         "#2E7D32", // Dark green
		"application":    "#1565C0", // Dark blue
		"infrastructure": "#C62828", // Dark red
		"presentation":   "#F57C00", // Dark orange
		"api":            "#6A1B9A", // Dark purple
		"cmd":            "#424242", // Dark gray
		"internal":       "#00695C", // Dark teal
	}

	if color, ok := colors[layer.Name]; ok {
		return color
	}

	return "black"
}

// getNodeFillColor returns the fill color for a node based on its layer
func (e *DOTExporter) getNodeFillColor(node string) string {
	if !e.options.ShowLayers {
		return "#FFFFFF"
	}

	layer := e.graph.GetLayerForFile(node)
	if layer == nil {
		return "#F5F5F5"
	}

	// Light fill colors for different layers
	colors := map[string]string{
		"domain":         "#C8E6C9", // Light green
		"application":    "#BBDEFB", // Light blue
		"infrastructure": "#FFCDD2", // Light red
		"presentation":   "#FFE0B2", // Light orange
		"api":            "#E1BEE7", // Light purple
		"cmd":            "#E0E0E0", // Light gray
		"internal":       "#B2DFDB", // Light teal
	}

	if color, ok := colors[layer.Name]; ok {
		return color
	}

	return "#FFFFFF"
}

// isViolation checks if a dependency violates layer rules
func (e *DOTExporter) isViolation(from, to string) bool {
	fromLayer := e.graph.GetLayerForFile(from)
	toLayer := e.graph.GetLayerForFile(to)

	return !e.graph.CanLayerDependOn(fromLayer, toLayer)
}

// detectAllCycles finds all circular dependencies in the graph
func (e *DOTExporter) detectAllCycles() map[string]map[string]bool {
	cycles := make(map[string]map[string]bool)

	// Use DFS to detect cycles
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	path := make([]string, 0)

	var dfs func(node string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true
		path = append(path, node)

		for _, dep := range e.graph.GetDependencies(node) {
			if !visited[dep] {
				if dfs(dep) {
					return true
				}
			} else if recStack[dep] {
				// Found a cycle - mark all edges in the cycle
				cycleStart := -1
				for i, n := range path {
					if n == dep {
						cycleStart = i
						break
					}
				}
				if cycleStart >= 0 {
					for i := cycleStart; i < len(path); i++ {
						from := path[i]
						to := ""
						if i+1 < len(path) {
							to = path[i+1]
						} else {
							to = dep
						}
						if cycles[from] == nil {
							cycles[from] = make(map[string]bool)
						}
						cycles[from][to] = true
					}
				}
				return true
			}
		}

		path = path[:len(path)-1]
		recStack[node] = false
		return false
	}

	// Check all nodes
	for _, node := range e.graph.AllFiles {
		if !visited[node] {
			dfs(node)
		}
	}

	return cycles
}

// simplifyPath shortens a file path for display
func (e *DOTExporter) simplifyPath(path string) string {
	// Remove common prefixes
	path = strings.TrimPrefix(path, "./")

	// Use only filename if in same directory
	if !strings.Contains(path, "/") {
		return path
	}

	// Show only last 2 path components
	parts := strings.Split(path, "/")
	if len(parts) > 2 {
		return filepath.Join(parts[len(parts)-2:]...)
	}

	return path
}

// writeLegend adds a legend showing layer colors
func (e *DOTExporter) writeLegend(w io.Writer) {
	_, _ = fmt.Fprintf(w, "\n  // Legend\n")
	_, _ = fmt.Fprintf(w, "  subgraph cluster_legend {\n")
	_, _ = fmt.Fprintf(w, "    label=\"Layers\";\n")
	_, _ = fmt.Fprintf(w, "    style=filled;\n")
	_, _ = fmt.Fprintf(w, "    fillcolor=\"#F0F0F0\";\n")

	// Collect unique layers
	layerMap := make(map[string]*config.Layer)
	for _, layer := range e.graph.Layers {
		layerMap[layer.Name] = &layer
	}

	i := 0
	for _, layer := range e.graph.Layers {
		// Temporarily set layer for color lookup
		tempFile := "temp"
		e.graph.FileLayers[tempFile] = &layer
		color := e.getNodeColor(tempFile)
		fillColor := e.getNodeFillColor(tempFile)
		delete(e.graph.FileLayers, tempFile)

		_, _ = fmt.Fprintf(w, "    legend%d [label=\"%s\", color=\"%s\", fillcolor=\"%s\", style=\"rounded,filled\"];\n",
			i, layer.Name, color, fillColor)
		i++
	}

	_, _ = fmt.Fprintf(w, "  }\n")
}
