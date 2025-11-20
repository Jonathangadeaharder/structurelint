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
	violationPaths := make(map[string]bool)
	var violations []Violation

	// Check forbidden paths
	violations = r.checkForbiddenPaths(file, layer, violationPaths)

	// Check CanDependOn constraints
	if len(layer.CanDependOn) > 0 {
		canDependOnViolations := r.checkCanDependOnConstraints(file, layer, violationPaths)
		violations = append(violations, canDependOnViolations...)
	}

	return violations
}

// checkForbiddenPaths checks if file path contains explicitly forbidden patterns
func (r *PathBasedLayerRule) checkForbiddenPaths(
	file walker.FileInfo,
	layer *PathLayer,
	violationPaths map[string]bool,
) []Violation {
	var violations []Violation

	for _, forbiddenPattern := range layer.ForbiddenPaths {
		if violation := r.checkSingleForbiddenPattern(file, layer, forbiddenPattern); violation != nil {
			violations = append(violations, *violation)
			violationPaths[forbiddenPattern] = true
		}
	}

	return violations
}

// checkSingleForbiddenPattern checks a single forbidden pattern
func (r *PathBasedLayerRule) checkSingleForbiddenPattern(
	file walker.FileInfo,
	layer *PathLayer,
	forbiddenPattern string,
) *Violation {
	regex, err := regexp.Compile(globToRegex(forbiddenPattern))
	if err != nil {
		return nil
	}

	if !regex.MatchString(file.Path) {
		return nil
	}

	return &Violation{
		Rule: r.Name(),
		Path: file.Path,
		Message: fmt.Sprintf(
			"layer '%s' file contains forbidden path pattern '%s'",
			layer.Name,
			forbiddenPattern,
		),
	}
}

// checkCanDependOnConstraints checks if file violates CanDependOn constraints
func (r *PathBasedLayerRule) checkCanDependOnConstraints(
	file walker.FileInfo,
	layer *PathLayer,
	violationPaths map[string]bool,
) []Violation {
	var violations []Violation
	forbiddenLayers := r.getForbiddenLayersFromCanDependOn(layer)

	for _, forbiddenLayerName := range forbiddenLayers {
		layerViolations := r.checkForbiddenLayer(file, layer, forbiddenLayerName, violationPaths)
		violations = append(violations, layerViolations...)
	}

	return violations
}

// checkForbiddenLayer checks if file contains paths from a forbidden layer
func (r *PathBasedLayerRule) checkForbiddenLayer(
	file walker.FileInfo,
	layer *PathLayer,
	forbiddenLayerName string,
	violationPaths map[string]bool,
) []Violation {
	forbiddenLayer := r.findLayerByName(forbiddenLayerName)
	if forbiddenLayer == nil {
		return nil
	}

	var violations []Violation
	for _, pattern := range forbiddenLayer.Patterns {
		if violation := r.checkForbiddenLayerPattern(file, layer, forbiddenLayerName, pattern, violationPaths); violation != nil {
			violations = append(violations, *violation)
		}
	}

	return violations
}

// findLayerByName finds a layer by its name
func (r *PathBasedLayerRule) findLayerByName(name string) *PathLayer {
	for i := range r.Layers {
		if r.Layers[i].Name == name {
			return &r.Layers[i]
		}
	}
	return nil
}

// checkForbiddenLayerPattern checks a single pattern from a forbidden layer
func (r *PathBasedLayerRule) checkForbiddenLayerPattern(
	file walker.FileInfo,
	layer *PathLayer,
	forbiddenLayerName string,
	pattern string,
	violationPaths map[string]bool,
) *Violation {
	layerSegment := r.extractLayerSegment(pattern)
	if layerSegment == "" || !strings.Contains(file.Path, layerSegment) {
		return nil
	}

	if r.isDuplicateViolation(layerSegment, violationPaths) {
		return nil
	}

	return &Violation{
		Rule: r.Name(),
		Path: file.Path,
		Message: fmt.Sprintf(
			"layer '%s' file contains path from layer '%s' which is not in its allowed dependencies (canDependOn: %v)",
			layer.Name,
			forbiddenLayerName,
			layer.CanDependOn,
		),
	}
}

// isDuplicateViolation checks if a violation has already been reported
func (r *PathBasedLayerRule) isDuplicateViolation(layerSegment string, violationPaths map[string]bool) bool {
	trimmedSegment := strings.Trim(layerSegment, "/")
	for forbiddenPattern := range violationPaths {
		if strings.Contains(forbiddenPattern, trimmedSegment) {
			return true
		}
	}
	return false
}

// getForbiddenLayersFromCanDependOn returns layer names that are implicitly forbidden
// based on the CanDependOn configuration
func (r *PathBasedLayerRule) getForbiddenLayersFromCanDependOn(layer *PathLayer) []string {
	var forbidden []string
	allowedMap := make(map[string]bool)
	allowedMap[layer.Name] = true // Layer can reference itself
	
	for _, allowed := range layer.CanDependOn {
		allowedMap[allowed] = true
	}
	
	for i := range r.Layers {
		otherLayer := &r.Layers[i]
		if !allowedMap[otherLayer.Name] {
			forbidden = append(forbidden, otherLayer.Name)
		}
	}
	
	return forbidden
}

// extractLayerSegment extracts the key path segment from a layer pattern
// e.g., "src/data/**" -> "/data/", "src/business/**" -> "/business/"
func (r *PathBasedLayerRule) extractLayerSegment(pattern string) string {
	// Remove glob wildcards
	cleaned := strings.ReplaceAll(pattern, "**", "")
	cleaned = strings.ReplaceAll(cleaned, "*", "")
	cleaned = strings.Trim(cleaned, "/")
	
	// Extract the last meaningful segment before wildcards
	// For "src/data/**", we want "data"
	parts := strings.Split(cleaned, "/")
	if len(parts) >= 2 {
		// Return as "/segment/" to match in paths
		return "/" + parts[len(parts)-1] + "/"
	}
	
	return ""
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
