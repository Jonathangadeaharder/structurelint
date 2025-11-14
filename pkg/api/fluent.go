// @structurelint:no-test Fluent API with comprehensive example tests in example_test.go
package api

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/walker"
)

// ArchRule provides a fluent interface for defining architectural rules
// Inspired by ArchUnit (Java) - allows writing tests like:
//   ArchRule().That(Layers().Matching("domain")).Should().NotDependOn(Layers().Matching("infrastructure"))
type ArchRule struct {
	selector   LayerSelector
	constraint Constraint
	graph      *graph.ImportGraph
	layers     []config.Layer
}

// ArchRule creates a new architectural rule builder
func NewArchRule() *ArchRule {
	return &ArchRule{}
}

// That specifies which layers/files this rule applies to
func (a *ArchRule) That(selector LayerSelector) *ArchRuleBuilder {
	a.selector = selector
	return &ArchRuleBuilder{rule: a}
}

// ArchRuleBuilder is the intermediate builder after specifying "That"
type ArchRuleBuilder struct {
	rule *ArchRule
}

// Should begins the constraint specification
func (b *ArchRuleBuilder) Should() *ConstraintBuilder {
	return &ConstraintBuilder{rule: b.rule}
}

// ShouldNot begins the negated constraint specification
func (b *ArchRuleBuilder) ShouldNot() *ConstraintBuilder {
	cb := &ConstraintBuilder{rule: b.rule}
	cb.negated = true
	return cb
}

// ConstraintBuilder builds dependency constraints
type ConstraintBuilder struct {
	rule    *ArchRule
	negated bool
}

// DependOn specifies a dependency constraint
func (cb *ConstraintBuilder) DependOn(target LayerSelector) *ArchRule {
	cb.rule.constraint = &DependencyConstraint{
		target:  target,
		negated: cb.negated,
	}
	return cb.rule
}

// BeEmpty specifies an emptiness constraint
func (cb *ConstraintBuilder) BeEmpty() *ArchRule {
	cb.rule.constraint = &EmptyConstraint{
		negated: cb.negated,
	}
	return cb.rule
}

// HaveNamingConvention specifies a naming convention constraint
func (cb *ConstraintBuilder) HaveNamingConvention(convention string) *ArchRule {
	cb.rule.constraint = &NamingConstraint{
		convention: convention,
		negated:    cb.negated,
	}
	return cb.rule
}

// WithGraph sets the import graph for dependency analysis
func (a *ArchRule) WithGraph(g *graph.ImportGraph) *ArchRule {
	a.graph = g
	return a
}

// WithLayers sets the layer configuration
func (a *ArchRule) WithLayers(layers []config.Layer) *ArchRule {
	a.layers = layers
	return a
}

// Check validates the rule and returns violations
func (a *ArchRule) Check(files []walker.FileInfo) []Violation {
	if a.selector == nil || a.constraint == nil {
		return []Violation{}
	}

	// Get files matching the selector
	selectedFiles := a.selector.Select(files, a.layers)

	// Apply constraint
	return a.constraint.Validate(selectedFiles, a.graph, a.layers)
}

// LayerSelector selects which layers/files a rule applies to
type LayerSelector interface {
	Select(files []walker.FileInfo, layers []config.Layer) []walker.FileInfo
}

// Layers creates a layer selector builder
func Layers() *LayerSelectorBuilder {
	return &LayerSelectorBuilder{}
}

// Files creates a file pattern selector builder
func Files() *FileSelectorBuilder {
	return &FileSelectorBuilder{}
}

// LayerSelectorBuilder builds layer-based selectors
type LayerSelectorBuilder struct {
	patterns []string
}

// Matching adds layer name patterns to match
func (b *LayerSelectorBuilder) Matching(pattern string) LayerSelector {
	b.patterns = append(b.patterns, pattern)
	return b
}

// Select implements LayerSelector
func (b *LayerSelectorBuilder) Select(files []walker.FileInfo, layers []config.Layer) []walker.FileInfo {
	var result []walker.FileInfo

	// Find layers matching our patterns
	var matchedLayers []config.Layer
	for _, layer := range layers {
		for _, pattern := range b.patterns {
			if matchPattern(layer.Name, pattern) {
				matchedLayers = append(matchedLayers, layer)
				break
			}
		}
	}

	// Select files belonging to matched layers
	for _, file := range files {
		for _, layer := range matchedLayers {
			matched, _ := filepath.Match(layer.Path, file.Path)
			if matched {
				result = append(result, file)
				break
			}
		}
	}

	return result
}

// FileSelectorBuilder builds file pattern-based selectors
type FileSelectorBuilder struct {
	patterns []string
}

// Matching adds file patterns to match
func (b *FileSelectorBuilder) Matching(pattern string) LayerSelector {
	b.patterns = append(b.patterns, pattern)
	return b
}

// Select implements LayerSelector
func (b *FileSelectorBuilder) Select(files []walker.FileInfo, layers []config.Layer) []walker.FileInfo {
	var result []walker.FileInfo

	for _, file := range files {
		for _, pattern := range b.patterns {
			matched, _ := filepath.Match(pattern, file.Path)
			if matched || matchGlob(file.Path, pattern) {
				result = append(result, file)
				break
			}
		}
	}

	return result
}

// Constraint represents an architectural constraint
type Constraint interface {
	Validate(files []walker.FileInfo, graph *graph.ImportGraph, layers []config.Layer) []Violation
}

// DependencyConstraint validates dependency relationships
type DependencyConstraint struct {
	target  LayerSelector
	negated bool
}

// Validate implements Constraint
func (c *DependencyConstraint) Validate(files []walker.FileInfo, graph *graph.ImportGraph, layers []config.Layer) []Violation {
	if graph == nil {
		return []Violation{}
	}

	var violations []Violation

	// Get target files
	targetFiles := c.target.Select(graph.AllFilesInfo(), layers)
	targetPaths := make(map[string]bool)
	for _, f := range targetFiles {
		targetPaths[f.Path] = true
	}

	// Check each source file
	for _, file := range files {
		if deps, ok := graph.Dependencies[file.Path]; ok {
			for _, dep := range deps {
				isTargetDep := targetPaths[dep]

				// If negated (should NOT depend on), violation when dependency exists
				// If not negated (should depend on), we'd need to check all files have at least one dep
				if c.negated && isTargetDep {
					violations = append(violations, Violation{
						Rule:    "arch-rule-dependency",
						Path:    file.Path,
						Message: fmt.Sprintf("file should not depend on %s", dep),
					})
				}
			}
		}
	}

	return violations
}

// EmptyConstraint validates whether files are empty
type EmptyConstraint struct {
	negated bool
}

// Validate implements Constraint
func (c *EmptyConstraint) Validate(files []walker.FileInfo, graph *graph.ImportGraph, layers []config.Layer) []Violation {
	var violations []Violation

	for _, file := range files {
		// Check file size using os.Stat
		info, err := os.Stat(file.AbsPath)
		if err != nil {
			continue // Skip files we can't stat
		}

		isEmpty := info.Size() == 0

		if c.negated && isEmpty {
			violations = append(violations, Violation{
				Rule:    "arch-rule-empty",
				Path:    file.Path,
				Message: "file should not be empty",
			})
		} else if !c.negated && !isEmpty {
			violations = append(violations, Violation{
				Rule:    "arch-rule-empty",
				Path:    file.Path,
				Message: "file should be empty",
			})
		}
	}

	return violations
}

// NamingConstraint validates naming conventions
type NamingConstraint struct {
	convention string
	negated    bool
}

// Validate implements Constraint
func (c *NamingConstraint) Validate(files []walker.FileInfo, graph *graph.ImportGraph, layers []config.Layer) []Violation {
	var violations []Violation

	for _, file := range files {
		basename := filepath.Base(file.Path)
		matches := matchesConvention(basename, c.convention)

		if c.negated && matches {
			violations = append(violations, Violation{
				Rule:    "arch-rule-naming",
				Path:    file.Path,
				Message: fmt.Sprintf("file should not follow %s naming convention", c.convention),
			})
		} else if !c.negated && !matches {
			violations = append(violations, Violation{
				Rule:    "arch-rule-naming",
				Path:    file.Path,
				Message: fmt.Sprintf("file should follow %s naming convention", c.convention),
			})
		}
	}

	return violations
}

// Helper functions

func matchPattern(s, pattern string) bool {
	// Simple glob matching with * support
	if pattern == "*" {
		return true
	}
	if strings.Contains(pattern, "*") {
		parts := strings.Split(pattern, "*")
		if len(parts) == 2 {
			return strings.HasPrefix(s, parts[0]) && strings.HasSuffix(s, parts[1])
		}
	}
	return s == pattern
}

func matchGlob(path, pattern string) bool {
	// Simple ** glob support
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(path, prefix) {
				return false
			}

			if suffix != "" {
				matched, _ := filepath.Match(suffix, filepath.Base(path))
				return matched
			}

			return true
		}
	}

	matched, _ := filepath.Match(pattern, path)
	return matched
}

func matchesConvention(name, convention string) bool {
	switch convention {
	case "kebab-case":
		return strings.ToLower(name) == name && !strings.Contains(name, "_") && !strings.ContainsAny(name, "ABCDEFGHIJKLMNOPQRSTUVWXYZ")
	case "snake_case":
		return strings.ToLower(name) == name && !strings.Contains(name, "-")
	case "camelCase":
		return !strings.Contains(name, "-") && !strings.Contains(name, "_") && len(name) > 0 && name[0] >= 'a' && name[0] <= 'z'
	case "PascalCase":
		return !strings.Contains(name, "-") && !strings.Contains(name, "_") && len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z'
	default:
		return true
	}
}
