package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
)

// PathBasedLayerRule enforces layer boundaries using only directory structure
// This is simpler than import-graph analysis and works even when code doesn't compile
type PathBasedLayerRule struct {
	Layers []PathLayer
}

// PathLayer defines a layer by its directory pattern
type PathLayer struct {
	Name            string   // Layer name (e.g., "presentation", "business", "data")
	Patterns        []string // Glob/regex patterns matching files in this layer
	CanDependOn     []string // Names of layers this can depend on
	ForbiddenPaths  []string // Path patterns that files in this layer cannot reference
	compiledRegexes []*regexp.Regexp
}

// Name returns the rule name
func (r *PathBasedLayerRule) Name() string {
	return "path-based-layers"
}

// Check validates that files don't reference forbidden layers based on path structure
func (r *PathBasedLayerRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Compile patterns
	if err := r.compileAllPatterns(); err != nil {
		return []Violation{{
			Rule:    r.Name(),
			Path:    ".",
			Message: fmt.Sprintf("pattern compilation error: %v", err),
		}}
	}

	// Map files to layers
	fileToLayer := r.mapFilesToLayers(files)

	// For each layer, check forbidden path violations
	for _, file := range files {
		if file.IsDir {
			continue
		}

		layer := fileToLayer[file.Path]
		if layer == nil {
			continue
		}

		// Check if file path violates layer separation
		// E.g., "presentation/controllers/user_controller.py" shouldn't be in "data/repositories/"
		violations = append(violations, r.checkFileLocation(file, layer, fileToLayer)...)
	}

	return violations
}

// compileAllPatterns compiles regex patterns for all layers
func (r *PathBasedLayerRule) compileAllPatterns() error {
	for i := range r.Layers {
		layer := &r.Layers[i]
		layer.compiledRegexes = make([]*regexp.Regexp, 0, len(layer.Patterns))

		for _, pattern := range layer.Patterns {
			// Convert glob-like patterns to regex
			regexPattern := globToRegex(pattern)
			regex, err := regexp.Compile(regexPattern)
			if err != nil {
				return fmt.Errorf("layer '%s' pattern '%s': %w", layer.Name, pattern, err)
			}
			layer.compiledRegexes = append(layer.compiledRegexes, regex)
		}
	}
	return nil
}

// mapFilesToLayers creates a mapping of file paths to their layers
func (r *PathBasedLayerRule) mapFilesToLayers(files []walker.FileInfo) map[string]*PathLayer {
	fileToLayer := make(map[string]*PathLayer)

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Find first matching layer
		for i := range r.Layers {
			layer := &r.Layers[i]
			if layer.matchesFile(file.Path) {
				fileToLayer[file.Path] = layer
				break
			}
		}
	}

	return fileToLayer
}

// matchesFile checks if a file path matches this layer's patterns
func (layer *PathLayer) matchesFile(path string) bool {
	for _, regex := range layer.compiledRegexes {
		if regex.MatchString(path) {
			return true
		}
	}
	return false
}

// checkFileLocation verifies that files in a layer don't violate separation rules
func (r *PathBasedLayerRule) checkFileLocation(
	file walker.FileInfo,
	layer *PathLayer,
	fileToLayer map[string]*PathLayer,
) []Violation {
	var violations []Violation

	// Check forbidden paths
	for _, forbiddenPattern := range layer.ForbiddenPaths {
		regex, err := regexp.Compile(globToRegex(forbiddenPattern))
		if err != nil {
			continue
		}

		// Check if this file's path contains the forbidden pattern
		// This catches cases like "presentation/data/repository.py" (presentation shouldn't have data paths)
		if regex.MatchString(file.Path) {
			violations = append(violations, Violation{
				Rule: r.Name(),
				Path: file.Path,
				Message: fmt.Sprintf(
					"layer '%s' file contains forbidden path pattern '%s'",
					layer.Name,
					forbiddenPattern,
				),
			})
		}
	}

	return violations
}

// globToRegex converts a simple glob pattern to regex
// Supports: *, **, ?, [abc]
func globToRegex(pattern string) string {
	// Escape regex special characters except glob wildcards
	pattern = regexp.QuoteMeta(pattern)

	// Restore and convert glob wildcards
	// Handle ** carefully - it should match zero or more path segments
	pattern = strings.ReplaceAll(pattern, `\*\*`, "DOUBLESTAR")
	pattern = strings.ReplaceAll(pattern, `\*`, "[^/]*")    // * -> [^/]* (any characters except /)
	pattern = strings.ReplaceAll(pattern, `\?`, ".")        // ? -> . (any single character)

	// Replace DOUBLESTAR with proper regex
	// Use (?:.*/)?  for zero or more path segments
	pattern = strings.ReplaceAll(pattern, "DOUBLESTAR/", "(?:.*/)?")  // **/ at start or middle
	pattern = strings.ReplaceAll(pattern, "/DOUBLESTAR", "(?:/.*)?")  // /** at end or middle
	pattern = strings.ReplaceAll(pattern, "DOUBLESTAR", ".*")         // ** by itself

	// Anchor the pattern
	return "^" + pattern + "$"
}

// NewPathBasedLayerRule creates a new PathBasedLayerRule
func NewPathBasedLayerRule(layers []PathLayer) *PathBasedLayerRule {
	return &PathBasedLayerRule{
		Layers: layers,
	}
}
