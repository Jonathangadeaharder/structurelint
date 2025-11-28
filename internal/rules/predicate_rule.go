package rules

import (
	"fmt"

	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/rules/predicate"
	"github.com/structurelint/structurelint/internal/walker"
)

// PredicateRule is a rule based on predicate logic
type PredicateRule struct {
	name        string
	description string
	predicate   predicate.Predicate
	message     string
	graph       *graph.ImportGraph
}

// NewPredicateRule creates a new predicate-based rule
func NewPredicateRule(name, description string, pred predicate.Predicate, message string) *PredicateRule {
	return &PredicateRule{
		name:        name,
		description: description,
		predicate:   pred,
		message:     message,
	}
}

// WithGraph attaches an import graph to the rule
func (r *PredicateRule) WithGraph(g *graph.ImportGraph) *PredicateRule {
	r.graph = g
	return r
}

// Name returns the rule name
func (r *PredicateRule) Name() string {
	return r.name
}

// Check executes the rule and returns violations
func (r *PredicateRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Create context
	ctx := &predicate.Context{
		Graph:      r.graph,
		AllFiles:   files,
		AllDirs:    dirs,
		CustomData: make(map[string]interface{}),
	}
	_ = ctx

	// Check each file
	for _, file := range files {
		if r.predicate(file, ctx) {
			message := r.message
			if message == "" {
				message = fmt.Sprintf("%s: %s", r.description, file.Path)
			} else {
				message = fmt.Sprintf(message, file.Path)
			}

			violations = append(violations, Violation{
				Rule:    r.name,
				Path:    file.Path,
				Message: message,
			})
		}
	}

	return violations
}

// --- Predicate rule builders for common patterns ---

// DisallowFilesWhere creates a rule that disallows files matching a predicate
func DisallowFilesWhere(name, description string, pred predicate.Predicate) *PredicateRule {
	return NewPredicateRule(
		name,
		description,
		pred,
		"%s violates rule: "+description,
	)
}

// RequireFilesWhere creates a rule that requires at least one file matching a predicate
type RequireFileRule struct {
	name        string
	description string
	predicate   predicate.Predicate
	message     string
	graph       *graph.ImportGraph
}

// NewRequireFileRule creates a rule that requires at least one file matching a predicate
func NewRequireFileRule(name, description string, pred predicate.Predicate, message string) *RequireFileRule {
	return &RequireFileRule{
		name:        name,
		description: description,
		predicate:   pred,
		message:     message,
	}
}

// WithGraph attaches an import graph to the rule
func (r *RequireFileRule) WithGraph(g *graph.ImportGraph) *RequireFileRule {
	r.graph = g
	return r
}

// Name returns the rule name
func (r *RequireFileRule) Name() string {
	return r.name
}

// Check executes the rule and returns violations
func (r *RequireFileRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	// Create context
	ctx := &predicate.Context{
		Graph:      r.graph,
		AllFiles:   files,
		AllDirs:    dirs,
		CustomData: make(map[string]interface{}),
	}
	_ = ctx

	// Check if at least one file matches
	for _, file := range files {
		if r.predicate(file, ctx) {
			return nil // Found at least one match - rule satisfied
		}
	}

	// No files matched - create violation
	message := r.message
	if message == "" {
		message = fmt.Sprintf("Required file not found: %s", r.description)
	}

	return []Violation{{
		Rule:    r.name,
		Path:    ".",
		Message: message,
	}}
}

// --- Example predicate rule configurations ---

// Example: "Domain entities cannot depend on Infrastructure"
func ExampleDomainPurityRule(g *graph.ImportGraph) Rule {
	pred := predicate.All(
		predicate.InLayer("domain"),
		predicate.DependsOn("*infrastructure*"),
	)

	return NewPredicateRule(
		"domain-purity",
		"Domain layer must not depend on infrastructure",
		pred,
		"Domain file %s depends on infrastructure (violates clean architecture)",
	).WithGraph(g)
}

// Example: "Test files must be adjacent to source files"
func ExampleTestAdjacencyPredicateRule() Rule {
	pred := predicate.All(
		predicate.NameContains("_test"),
		predicate.Not(predicate.Custom(func(file walker.FileInfo, ctx *predicate.Context) bool {
			// Check if corresponding source file exists
			sourcePath := file.Path
			if len(sourcePath) > 8 && sourcePath[len(sourcePath)-8:] == "_test.go" {
				sourcePath = sourcePath[:len(sourcePath)-8] + ".go"
				// Check if source file exists in AllFiles
				for _, f := range ctx.AllFiles {
					if f.Path == sourcePath {
						return true
					}
				}
			}
			return false
		})),
	)

	return NewPredicateRule(
		"test-adjacency-predicate",
		"Test files must be adjacent to source files",
		pred,
		"Test file %s has no corresponding source file",
	)
}

// Example: "Large files must be in specific directories"
func ExampleLargeFileLocationRule() Rule {
	pred := predicate.All(
		predicate.SizeGreaterThan(10*1024), // 10KB
		predicate.Not(predicate.Any(
			predicate.PathContains("/vendor/"),
			predicate.PathContains("/generated/"),
			predicate.PathContains("/migrations/"),
		)),
	)

	return NewPredicateRule(
		"large-file-location",
		"Large files must be in designated directories",
		pred,
		"Large file %s is not in an allowed directory (vendor/, generated/, migrations/)",
	)
}

// Example: "Orphaned files are disallowed"
func ExampleNoOrphansRule(g *graph.ImportGraph) Rule {
	pred := predicate.All(
		predicate.IsFile(),
		predicate.Not(predicate.NameMatches("README*")),
		predicate.Not(predicate.NameMatches("LICENSE*")),
		predicate.Not(predicate.PathContains("/test/")),
		predicate.IsOrphaned(),
	)

	return NewPredicateRule(
		"no-orphans",
		"Files must be part of the dependency graph",
		pred,
		"Orphaned file %s has no imports or exports",
	).WithGraph(g)
}
