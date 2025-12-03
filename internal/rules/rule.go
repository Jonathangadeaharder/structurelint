// Package rules defines the linting rule interface and implementations.
//
// @structurelint:no-test Interface definitions and types only, tested through rule implementations
package rules

import (
	"github.com/Jonathangadeaharder/structurelint/internal/parser"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// AutoFix contains information about an automatic fix for a violation
type AutoFix struct {
	FilePath string // Path where the fix should be applied (relative to project root)
	Content  string // Content to write to the file
}

// Violation represents a rule violation
type Violation struct {
	Rule        string
	Path        string
	Message     string
	Expected    string   // Optional: What was expected (e.g., "PascalCase")
	Actual      string   // Optional: What was found (e.g., "camelCase")
	Suggestions []string // Optional: Fix suggestions
	Context     string   // Optional: Rule context (e.g., "React components rule: src/components/**")
	AutoFix     *AutoFix // Optional: Automatic fix content
}

// FormatDetailed returns a detailed, human-friendly violation message
// with expected/actual values, context, and suggestions
func (v *Violation) FormatDetailed() string {
	msg := v.Path + ": " + v.Message

	// Add expected/actual if available
	if v.Expected != "" && v.Actual != "" {
		msg += "\n  Expected: " + v.Expected
		msg += "\n  Actual: " + v.Actual
	}

	// Add context if available
	if v.Context != "" {
		msg += "\n  Context: " + v.Context
	}

	// Add suggestions if available
	if len(v.Suggestions) > 0 {
		msg += "\n  Suggestions:"
		for _, suggestion := range v.Suggestions {
			msg += "\n    - " + suggestion
		}
	}

	return msg
}

// Rule defines the interface for all linter rules
type Rule interface {
	// Name returns the name of the rule
	Name() string

	// Check validates the rule against the filesystem data
	Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation
}

// RuleConfig is a marker interface for rule configurations
type RuleConfig interface {
	IsEnabled() bool
}

// ShouldIgnoreFile checks if a file has a directive to ignore a specific rule
// Returns (shouldIgnore, reason)
func ShouldIgnoreFile(file walker.FileInfo, ruleName string) (bool, string) {
	if file.IsDir {
		return false, ""
	}
	return parser.HasDirectiveForRule(file.Directives, ruleName)
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
