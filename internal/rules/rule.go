// Package rules defines the linting rule interface and implementations.
//
// @structurelint:no-test Interface definitions and types only, tested through rule implementations
package rules

import (
	"github.com/structurelint/structurelint/internal/parser"
	"github.com/structurelint/structurelint/internal/walker"
)

// Violation represents a rule violation
type Violation struct {
	Rule    string
	Path    string
	Message string
}

// Fix represents an automated fix for a violation
type Fix struct {
	FilePath string      // Path to the file to fix
	Action   string      // Type of action: "rename", "delete", "modify"
	OldValue string      // Old value (e.g., old filename, line to remove)
	NewValue string      // New value (e.g., new filename, replacement content)
	Metadata interface{} // Additional metadata for complex fixes
}

// Rule defines the interface for all linter rules
type Rule interface {
	// Name returns the name of the rule
	Name() string

	// Check validates the rule against the filesystem data
	Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation
}

// FixableRule defines the interface for rules that support automated fixing
type FixableRule interface {
	Rule

	// Fix generates fixes for violations
	Fix(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Fix
}

// RuleConfig is a marker interface for rule configurations
type RuleConfig interface {
	IsEnabled() bool
}

// ShouldIgnoreFile checks if a file has a directive to ignore a specific rule
// Returns (shouldIgnore, reason)
// Directives are parsed lazily and cached for performance
func ShouldIgnoreFile(file walker.FileInfo, ruleName string) (bool, string) {
	if file.IsDir {
		return false, ""
	}
	// Parse directives lazily - only when needed, with caching
	directives := parser.ParseDirectives(file.AbsPath)
	return parser.HasDirectiveForRule(directives, ruleName)
}

// FilterIgnoredFiles filters out files that have directives to ignore the specified rule
func FilterIgnoredFiles(files []walker.FileInfo, ruleName string) []walker.FileInfo {
	var filtered []walker.FileInfo
	for _, file := range files {
		if shouldIgnore, _ := ShouldIgnoreFile(file, ruleName); !shouldIgnore {
			filtered = append(filtered, file)
		}
	}
	return filtered
}
