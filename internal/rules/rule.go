// Package rules defines the linting rule interface and implementations.
//
// @structurelint:no-test Interface definitions and types only, tested through rule implementations
package rules

import (
	"github.com/structurelint/structurelint/internal/walker"
)

// Violation represents a rule violation
type Violation struct {
	Rule    string
	Path    string
	Message string
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
