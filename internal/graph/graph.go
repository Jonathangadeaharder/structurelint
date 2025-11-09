package graph

import (
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/parser"
	"github.com/structurelint/structurelint/internal/walker"
)

// ImportGraph represents the import dependency graph of the project
type ImportGraph struct {
	// Map from file path to list of files it imports
	Dependencies map[string][]string

	// Map from file path to its layer
	FileLayers map[string]*config.Layer

	// All imports extracted from the project
	AllImports []parser.Import

	// Layer definitions
	Layers []config.Layer
}

// Builder builds an import graph from the filesystem
type Builder struct {
	rootPath string
	parser   *parser.Parser
	layers   []config.Layer
}

// NewBuilder creates a new graph builder
func NewBuilder(rootPath string, layers []config.Layer) *Builder {
	return &Builder{
		rootPath: rootPath,
		parser:   parser.New(rootPath),
		layers:   layers,
	}
}

// Build constructs the import graph from the given files
func (b *Builder) Build(files []walker.FileInfo) (*ImportGraph, error) {
	graph := &ImportGraph{
		Dependencies: make(map[string][]string),
		FileLayers:   make(map[string]*config.Layer),
		AllImports:   []parser.Import{},
		Layers:       b.layers,
	}

	// First pass: collect all imports
	for _, file := range files {
		if file.IsDir {
			continue
		}

		imports, err := b.parser.ParseFile(file.AbsPath)
		if err != nil {
			// Skip files we can't parse
			continue
		}

		graph.AllImports = append(graph.AllImports, imports...)

		// Build dependency map
		for _, imp := range imports {
			// Resolve relative imports to absolute paths
			resolvedPath := imp.ImportPath
			if imp.IsRelative {
				resolvedPath = b.parser.ResolveImportPath(file.Path, imp.ImportPath)
			}

			graph.Dependencies[file.Path] = append(graph.Dependencies[file.Path], resolvedPath)
		}
	}

	// Second pass: assign files to layers
	for _, file := range files {
		if file.IsDir {
			continue
		}

		layer := b.findLayerForFile(file.Path)
		if layer != nil {
			graph.FileLayers[file.Path] = layer
		}
	}

	return graph, nil
}

// findLayerForFile determines which layer a file belongs to
func (b *Builder) findLayerForFile(filePath string) *config.Layer {
	for i := range b.layers {
		layer := &b.layers[i]
		if b.matchesLayerPath(filePath, layer.Path) {
			return layer
		}
	}
	return nil
}

// matchesLayerPath checks if a file path matches a layer's path pattern
func (b *Builder) matchesLayerPath(filePath, layerPath string) bool {
	// Handle glob patterns
	if strings.Contains(layerPath, "**") {
		parts := strings.Split(layerPath, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			// Check prefix
			if prefix != "" && !strings.HasPrefix(filePath, prefix) {
				return false
			}

			// Check suffix
			if suffix != "" {
				matched, _ := filepath.Match(suffix, filepath.Base(filePath))
				if !matched {
					// Also try matching the full remaining path
					if !strings.HasSuffix(filePath, suffix) && !strings.Contains(filePath, suffix) {
						return false
					}
				}
			}

			return true
		}
	}

	// Simple prefix match
	return strings.HasPrefix(filePath, layerPath)
}

// GetLayerForFile returns the layer for a given file
func (g *ImportGraph) GetLayerForFile(filePath string) *config.Layer {
	return g.FileLayers[filePath]
}

// GetDependencies returns all files that a given file imports
func (g *ImportGraph) GetDependencies(filePath string) []string {
	return g.Dependencies[filePath]
}

// CanLayerDependOn checks if layerA is allowed to depend on layerB
func (g *ImportGraph) CanLayerDependOn(layerA, layerB *config.Layer) bool {
	if layerA == nil || layerB == nil {
		// If either layer is nil, allow the dependency
		return true
	}

	// A layer can always depend on itself
	if layerA.Name == layerB.Name {
		return true
	}

	// Check if layerB is in layerA's dependsOn list
	for _, allowedDep := range layerA.DependsOn {
		if allowedDep == "*" {
			// Wildcard - can depend on any layer
			return true
		}
		if allowedDep == layerB.Name {
			return true
		}
	}

	return false
}

// FindLayerByName finds a layer by its name
func (g *ImportGraph) FindLayerByName(name string) *config.Layer {
	for i := range g.Layers {
		if g.Layers[i].Name == name {
			return &g.Layers[i]
		}
	}
	return nil
}
