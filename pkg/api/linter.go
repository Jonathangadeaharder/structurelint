// Package api provides a stable public API for using structurelint programmatically.
// This package exposes internal functionality in a backward-compatible way.
//
// @structurelint:no-test Public API with comprehensive example tests in example_test.go
package api

import (
	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/linter"
	"github.com/structurelint/structurelint/internal/rules"
)

// Linter provides programmatic access to structurelint functionality
type Linter struct {
	internal *linter.Linter
	config   *Config
}

// NewLinter creates a new linter instance
func NewLinter() *Linter {
	return &Linter{
		internal: linter.New(),
		config:   nil,
	}
}

// WithConfig sets a custom configuration for the linter
func (l *Linter) WithConfig(cfg *Config) *Linter {
	l.config = cfg
	return l
}

// WithProductionMode enables production mode (excludes test files from analysis)
func (l *Linter) WithProductionMode(enabled bool) *Linter {
	l.internal = l.internal.WithProductionMode(enabled)
	return l
}

// Lint runs linting on the specified path and returns violations
func (l *Linter) Lint(path string) ([]Violation, error) {
	violations, err := l.internal.Lint(path)
	if err != nil {
		return nil, err
	}

	// Convert internal violations to public API violations
	result := make([]Violation, len(violations))
	for i, v := range violations {
		result[i] = Violation{
			Rule:    v.Rule,
			Path:    v.Path,
			Message: v.Message,
		}
	}

	return result, nil
}

// Violation represents a linting rule violation
type Violation struct {
	Rule    string // Name of the rule that was violated
	Path    string // Path to the file with the violation
	Message string // Human-readable violation message
}

// Config wraps the internal configuration
type Config struct {
	internal *config.Config
}

// LoadConfig loads configuration from a path
func LoadConfig(path string) (*Config, error) {
	configs, err := config.FindConfigs(path)
	if err != nil {
		return nil, err
	}

	merged := config.Merge(configs...)
	return &Config{internal: merged}, nil
}

// NewConfig creates a new empty configuration
func NewConfig() *Config {
	return &Config{
		internal: &config.Config{
			Rules:   make(map[string]interface{}),
			Exclude: []string{},
			Layers:  []config.Layer{},
		},
	}
}

// EnableRule enables a rule with the given configuration
func (c *Config) EnableRule(name string, ruleConfig interface{}) *Config {
	c.internal.Rules[name] = ruleConfig
	return c
}

// AddExclude adds a path pattern to exclude from linting
func (c *Config) AddExclude(pattern string) *Config {
	c.internal.Exclude = append(c.internal.Exclude, pattern)
	return c
}

// AddLayer adds a layer definition to the configuration
func (c *Config) AddLayer(name string, path string) *Config {
	c.internal.Layers = append(c.internal.Layers, config.Layer{
		Name: name,
		Path: path,
	})
	return c
}

// Fix generates and applies fixes for violations
type Fix struct {
	FilePath string
	Action   string
	OldValue string
	NewValue string
}

// GenerateFixes generates fixes for the given path
func GenerateFixes(path string, dryRun bool) ([]Fix, error) {
	// This would integrate with the fixer package
	// For now, return a placeholder
	// TODO: Implement full fix generation through public API
	return []Fix{}, nil
}

// RuleInfo provides information about available rules
type RuleInfo struct {
	Name        string
	Description string
	Fixable     bool
}

// AvailableRules returns information about all available rules
func AvailableRules() []RuleInfo {
	return []RuleInfo{
		{Name: "no-empty-files", Description: "Disallow empty files", Fixable: false},
		{Name: "disallowed-patterns", Description: "Disallow specific path patterns", Fixable: false},
		{Name: "required-files", Description: "Require specific files to exist", Fixable: false},
		{Name: "naming-convention", Description: "Enforce file/directory naming conventions", Fixable: true},
		{Name: "test-adjacency", Description: "Require tests adjacent to implementation", Fixable: false},
		{Name: "disallow-unused-exports", Description: "Disallow unused exports", Fixable: true},
		{Name: "file-hash", Description: "Validate file contents via SHA256 hash", Fixable: false},
		{Name: "granular-dependencies", Description: "Fine-grained module dependency validation", Fixable: false},
	}
}

// IsFixable returns whether a rule supports automated fixing
func IsFixable(ruleName string) bool {
	for _, info := range AvailableRules() {
		if info.Name == ruleName {
			return info.Fixable
		}
	}
	return false
}

// Convert internal rules.Fix to public api.Fix
func convertFix(internal rules.Fix) Fix {
	return Fix{
		FilePath: internal.FilePath,
		Action:   internal.Action,
		OldValue: internal.OldValue,
		NewValue: internal.NewValue,
	}
}
