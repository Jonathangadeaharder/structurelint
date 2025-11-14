// @structurelint:ignore test-adjacency Graph export is tested through integration tests
package export

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/structurelint/structurelint/internal/graph"
)

// GraphExporter exports dependency graphs in various formats
type GraphExporter struct {
	graph *graph.ImportGraph
}

// NewGraphExporter creates a new graph exporter
func NewGraphExporter(g *graph.ImportGraph) *GraphExporter {
	return &GraphExporter{graph: g}
}

// ExportDOT exports the graph in Graphviz DOT format
func (e *GraphExporter) ExportDOT() string {
	var sb strings.Builder

	sb.WriteString("digraph dependencies {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box, style=rounded];\n\n")

	// Group nodes by layer
	if len(e.graph.Layers) > 0 {
		layerFiles := make(map[string][]string)
		for file, layer := range e.graph.FileLayers {
			if layer != nil {
				layerFiles[layer.Name] = append(layerFiles[layer.Name], file)
			}
		}

		// Create subgraphs for each layer
		for _, layer := range e.graph.Layers {
			if files, ok := layerFiles[layer.Name]; ok {
				sb.WriteString(fmt.Sprintf("  subgraph cluster_%s {\n", sanitizeName(layer.Name)))
				sb.WriteString(fmt.Sprintf("    label=\"%s\";\n", layer.Name))
				sb.WriteString("    style=filled;\n")
				sb.WriteString("    color=lightgrey;\n")

				for _, file := range files {
					nodeName := sanitizeName(file)
					displayName := filepath.Base(file)
					sb.WriteString(fmt.Sprintf("    \"%s\" [label=\"%s\"];\n", nodeName, displayName))
				}

				sb.WriteString("  }\n\n")
			}
		}
	}

	// Add nodes that aren't in any layer
	nodesInLayers := make(map[string]bool)
	for file := range e.graph.FileLayers {
		nodesInLayers[file] = true
	}

	for file := range e.graph.Dependencies {
		if !nodesInLayers[file] {
			nodeName := sanitizeName(file)
			displayName := filepath.Base(file)
			sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\"];\n", nodeName, displayName))
		}
	}
	sb.WriteString("\n")

	// Add edges
	edges := make(map[string]bool)
	for source, targets := range e.graph.Dependencies {
		sourceName := sanitizeName(source)
		for _, target := range targets {
			targetName := sanitizeName(target)
			edge := fmt.Sprintf("  \"%s\" -> \"%s\";\n", sourceName, targetName)
			if !edges[edge] {
				edges[edge] = true
				sb.WriteString(edge)
			}
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}

// ExportMermaid exports the graph in Mermaid format
func (e *GraphExporter) ExportMermaid() string {
	var sb strings.Builder

	sb.WriteString("graph LR\n")

	// Add nodes and edges
	edges := make(map[string]bool)
	allNodes := make(map[string]bool)

	// Collect all nodes
	for source := range e.graph.Dependencies {
		allNodes[source] = true
	}
	for _, targets := range e.graph.Dependencies {
		for _, target := range targets {
			allNodes[target] = true
		}
	}

	// Sort nodes for deterministic output
	var sortedNodes []string
	for node := range allNodes {
		sortedNodes = append(sortedNodes, node)
	}
	sort.Strings(sortedNodes)

	// Add node styling based on layers
	for _, node := range sortedNodes {
		nodeName := sanitizeName(node)
		displayName := filepath.Base(node)

		if layer, ok := e.graph.FileLayers[node]; ok && layer != nil {
			// Node in a layer - use layer name as class
			sb.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", nodeName, displayName))
			sb.WriteString(fmt.Sprintf("  class %s layer_%s\n", nodeName, sanitizeName(layer.Name)))
		} else {
			// Node not in a layer
			sb.WriteString(fmt.Sprintf("  %s[\"%s\"]\n", nodeName, displayName))
		}
	}

	sb.WriteString("\n")

	// Add edges
	for source, targets := range e.graph.Dependencies {
		sourceName := sanitizeName(source)
		for _, target := range targets {
			targetName := sanitizeName(target)
			edge := fmt.Sprintf("  %s --> %s\n", sourceName, targetName)
			if !edges[edge] {
				edges[edge] = true
				sb.WriteString(edge)
			}
		}
	}

	// Add layer styling
	if len(e.graph.Layers) > 0 {
		sb.WriteString("\n  %% Layer styles\n")
		colors := []string{"lightblue", "lightgreen", "lightyellow", "lightpink", "lightcyan"}
		for i, layer := range e.graph.Layers {
			color := colors[i%len(colors)]
			sb.WriteString(fmt.Sprintf("  classDef layer_%s fill:%s\n", sanitizeName(layer.Name), color))
		}
	}

	return sb.String()
}

// ExportJSON exports the graph in JSON format
func (e *GraphExporter) ExportJSON() string {
	var sb strings.Builder

	sb.WriteString("{\n")
	sb.WriteString("  \"nodes\": [\n")

	// Collect all unique nodes
	allNodes := make(map[string]bool)
	for source := range e.graph.Dependencies {
		allNodes[source] = true
	}
	for _, targets := range e.graph.Dependencies {
		for _, target := range targets {
			allNodes[target] = true
		}
	}

	// Sort nodes for deterministic output
	var sortedNodes []string
	for node := range allNodes {
		sortedNodes = append(sortedNodes, node)
	}
	sort.Strings(sortedNodes)

	// Output nodes
	for i, node := range sortedNodes {
		layer := ""
		if l, ok := e.graph.FileLayers[node]; ok && l != nil {
			layer = l.Name
		}

		sb.WriteString(fmt.Sprintf("    {\"id\": \"%s\", \"layer\": \"%s\"}", node, layer))
		if i < len(sortedNodes)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("  ],\n")
	sb.WriteString("  \"edges\": [\n")

	// Output edges
	type edge struct {
		from string
		to   string
	}
	var edges []edge
	edgeSet := make(map[string]bool)

	for source, targets := range e.graph.Dependencies {
		for _, target := range targets {
			edgeKey := source + "->" + target
			if !edgeSet[edgeKey] {
				edgeSet[edgeKey] = true
				edges = append(edges, edge{from: source, to: target})
			}
		}
	}

	// Sort edges for deterministic output
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].from != edges[j].from {
			return edges[i].from < edges[j].from
		}
		return edges[i].to < edges[j].to
	})

	for i, e := range edges {
		sb.WriteString(fmt.Sprintf("    {\"from\": \"%s\", \"to\": \"%s\"}", e.from, e.to))
		if i < len(edges)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("  ]\n")
	sb.WriteString("}\n")

	return sb.String()
}

// sanitizeName sanitizes a file path for use as a node name in graph formats
func sanitizeName(name string) string {
	// Replace special characters with underscores
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, " ", "_")
	return name
}
