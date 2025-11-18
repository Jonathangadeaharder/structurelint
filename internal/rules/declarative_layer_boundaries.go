// @structurelint:no-test Foundation for future declarative layer validation, tested via path_based_layers.go
package rules

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
)

// DeclarativeLayerBoundariesRule enforces layer boundaries using path-only validation
// This rule works without import graphs, making it suitable for:
// - Non-buildable code
// - Diverse architectures
// - Quick validation without parsing
type DeclarativeLayerBoundariesRule struct {
	Layers []DeclarativeLayer
}

// DeclarativeLayer represents a layer with regex-based path matching
type DeclarativeLayer struct {
	Name         string
	PathPattern  string          // Regex pattern for matching files in this layer
	pathRegex    *regexp.Regexp  // Compiled regex (internal)
	CanDependOn  []string        // Names of layers this layer can depend on
	CannotImport []string        // Regex patterns of paths this layer cannot import
}

// Name returns the rule name
func (r *DeclarativeLayerBoundariesRule) Name() string {
	return "declarative-layer-boundaries"
}

// Check validates layer boundaries based on file paths and simple heuristics
func (r *DeclarativeLayerBoundariesRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Compile regex patterns
	if err := r.compilePatterns(); err != nil {
		return []Violation{{
			Rule:    r.Name(),
			Path:    ".",
			Message: fmt.Sprintf("failed to compile layer patterns: %v", err),
		}}
	}

	// Build a map of files to their layers
	fileToLayer := make(map[string]*DeclarativeLayer)
	for _, file := range files {
		if file.IsDir {
			continue
		}
		for i := range r.Layers {
			layer := &r.Layers[i]
			if layer.pathRegex.MatchString(file.Path) {
				fileToLayer[file.Path] = layer
				break // First matching layer wins
			}
		}
	}

	// Check each file for potential violations
	for _, file := range files {
		if file.IsDir {
			continue
		}

		sourceLayer := fileToLayer[file.Path]
		if sourceLayer == nil {
			continue // File not in any layer
		}

		// Read file and extract potential imports (simple heuristic)
		imports := r.extractPotentialImports(file)

		for _, importPath := range imports {
			// Find which file this import might reference
			targetFile := r.resolveImportToFile(importPath, files)
			if targetFile == "" {
				continue // External or unresolved import
			}

			targetLayer := fileToLayer[targetFile]
			if targetLayer == nil {
				continue // Target not in any layer
			}

			// Check if this dependency is allowed
			if !r.canLayerDependOn(sourceLayer, targetLayer) {
				violations = append(violations, Violation{
					Rule: r.Name(),
					Path: file.Path,
					Message: fmt.Sprintf(
						"layer '%s' cannot depend on layer '%s' (references: %s)",
						sourceLayer.Name,
						targetLayer.Name,
						targetFile,
					),
				})
			}

			// Check CannotImport patterns
			for _, pattern := range sourceLayer.CannotImport {
				matched, err := regexp.MatchString(pattern, targetFile)
				if err == nil && matched {
					violations = append(violations, Violation{
						Rule: r.Name(),
						Path: file.Path,
						Message: fmt.Sprintf(
							"layer '%s' cannot import from '%s' (matches forbidden pattern: %s)",
							sourceLayer.Name,
							targetFile,
							pattern,
						),
					})
				}
			}
		}
	}

	return violations
}

// compilePatterns compiles all regex patterns for layers
func (r *DeclarativeLayerBoundariesRule) compilePatterns() error {
	for i := range r.Layers {
		layer := &r.Layers[i]
		regex, err := regexp.Compile(layer.PathPattern)
		if err != nil {
			return fmt.Errorf("layer '%s': %w", layer.Name, err)
		}
		layer.pathRegex = regex
	}
	return nil
}

// canLayerDependOn checks if sourceLayer can depend on targetLayer
func (r *DeclarativeLayerBoundariesRule) canLayerDependOn(source, target *DeclarativeLayer) bool {
	// Same layer can always reference itself
	if source.Name == target.Name {
		return true
	}

	// Check if target is in the allowed dependencies list
	for _, allowedDep := range source.CanDependOn {
		if allowedDep == target.Name {
			return true
		}
	}

	return false
}

// extractPotentialImports extracts import-like paths from a file
// This is a simple heuristic that looks for common import patterns
func (r *DeclarativeLayerBoundariesRule) extractPotentialImports(file walker.FileInfo) []string {
	// For now, this is a placeholder that would need to read files
	// In practice, we'd scan for:
	// - Go: import "path"
	// - Python: from x import y, import x
	// - TS/JS: import x from 'path', require('path')

	// For MVP, return empty to avoid file I/O in this first iteration
	// This will be enhanced in a follow-up
	return []string{}
}

// resolveImportToFile resolves an import path to an actual file
func (r *DeclarativeLayerBoundariesRule) resolveImportToFile(importPath string, files []walker.FileInfo) string {
	// Try exact match
	for _, file := range files {
		if file.Path == importPath {
			return file.Path
		}
	}

	// Try with extensions
	extensions := []string{".ts", ".tsx", ".js", ".jsx", ".go", ".py", ".rs", ".java", ".cs"}
	for _, ext := range extensions {
		testPath := importPath + ext
		for _, file := range files {
			if file.Path == testPath {
				return file.Path
			}
		}
	}

	// Try as directory with index
	indexFiles := []string{"index.ts", "index.tsx", "index.js", "index.jsx", "__init__.py", "mod.rs"}
	for _, indexFile := range indexFiles {
		testPath := filepath.Join(importPath, indexFile)
		for _, file := range files {
			if file.Path == testPath {
				return file.Path
			}
		}
	}

	// Try prefix matching (for package imports)
	for _, file := range files {
		if strings.HasPrefix(file.Path, importPath+"/") {
			return file.Path
		}
	}

	return ""
}

// NewDeclarativeLayerBoundariesRule creates a new DeclarativeLayerBoundariesRule
func NewDeclarativeLayerBoundariesRule(layers []DeclarativeLayer) *DeclarativeLayerBoundariesRule {
	return &DeclarativeLayerBoundariesRule{
		Layers: layers,
	}
}

// LayerFromConfig creates a DeclarativeLayer from configuration
func LayerFromConfig(name, pathPattern string, canDependOn []string) DeclarativeLayer {
	return DeclarativeLayer{
		Name:        name,
		PathPattern: pathPattern,
		CanDependOn: canDependOn,
	}
}
