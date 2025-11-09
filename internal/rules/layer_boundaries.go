package rules

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/walker"
)

// LayerBoundariesRule enforces architectural layer boundaries
type LayerBoundariesRule struct {
	Graph *graph.ImportGraph
}

// Name returns the rule name
func (r *LayerBoundariesRule) Name() string {
	return "enforce-layer-boundaries"
}

// Check validates that imports respect layer boundaries
func (r *LayerBoundariesRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	if r.Graph == nil {
		return []Violation{}
	}

	var violations []Violation

	// Check each file's imports
	for sourceFile, dependencies := range r.Graph.Dependencies {
		sourceLayer := r.Graph.GetLayerForFile(sourceFile)

		// Skip files not in any layer
		if sourceLayer == nil {
			continue
		}

		// Check each dependency
		for _, dep := range dependencies {
			// Resolve the dependency to a file in the project
			targetFile := r.resolveDependencyToFile(dep, files)
			if targetFile == "" {
				// External dependency or unresolved, skip
				continue
			}

			targetLayer := r.Graph.GetLayerForFile(targetFile)

			// Check if this dependency is allowed
			if !r.Graph.CanLayerDependOn(sourceLayer, targetLayer) {
				targetLayerName := "unknown"
				if targetLayer != nil {
					targetLayerName = targetLayer.Name
				}

				violations = append(violations, Violation{
					Rule: r.Name(),
					Path: sourceFile,
					Message: fmt.Sprintf(
						"layer '%s' cannot import from layer '%s' (imported: %s)",
						sourceLayer.Name,
						targetLayerName,
						targetFile,
					),
				})
			}
		}
	}

	return violations
}

// resolveDependencyToFile attempts to resolve an import path to an actual file in the project
func (r *LayerBoundariesRule) resolveDependencyToFile(dep string, files []walker.FileInfo) string {
	// Try exact match
	for _, file := range files {
		if file.Path == dep {
			return file.Path
		}
	}

	// Try with common extensions
	extensions := []string{".ts", ".tsx", ".js", ".jsx", ".go", ".py"}
	for _, ext := range extensions {
		testPath := dep + ext
		for _, file := range files {
			if file.Path == testPath {
				return file.Path
			}
		}
	}

	// Try as directory with index file
	indexFiles := []string{"index.ts", "index.tsx", "index.js", "index.jsx"}
	for _, indexFile := range indexFiles {
		testPath := filepath.Join(dep, indexFile)
		for _, file := range files {
			if file.Path == testPath {
				return file.Path
			}
		}
	}

	// Try matching any file in the dependency path (for Go packages)
	for _, file := range files {
		if strings.HasPrefix(file.Path, dep) {
			return file.Path
		}
	}

	return ""
}

// NewLayerBoundariesRule creates a new LayerBoundariesRule
func NewLayerBoundariesRule(importGraph *graph.ImportGraph) *LayerBoundariesRule {
	return &LayerBoundariesRule{
		Graph: importGraph,
	}
}
