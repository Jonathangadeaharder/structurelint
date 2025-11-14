// Package plugin provides a plugin architecture for extending structurelint
// with custom language parsers and rules.
//
// @structurelint:no-test Plugin interface with integration tests
package plugin

import (
	"fmt"
	"path/filepath"
	"sync"
)

// Parser is the interface that custom language parsers must implement
type Parser interface {
	// Name returns the unique name of the parser plugin
	Name() string

	// Version returns the semantic version of the plugin
	Version() string

	// SupportedExtensions returns file extensions this parser handles (e.g., [".rs", ".rb"])
	SupportedExtensions() []string

	// ParseImports extracts import statements from a file
	// Returns a list of Import structs representing dependencies
	ParseImports(filePath string) ([]Import, error)

	// ParseExports extracts exported symbols from a file
	// Returns a list of Export structs representing public APIs
	ParseExports(filePath string) ([]Export, error)
}

// Import represents an import/dependency statement
type Import struct {
	ImportPath string // The imported module/file path
	IsRelative bool   // Whether this is a relative import (./foo vs absolute)
	Line       int    // Line number where import appears
	Symbol     string // Specific symbol imported (optional, e.g., "func" from import)
}

// Export represents an exported symbol
type Export struct {
	Name string // Name of the exported symbol
	Kind string // Kind of export: "function", "class", "variable", "type", etc.
	Line int    // Line number where export is defined
}

// RulePlugin is the interface for custom linting rules
type RulePlugin interface {
	// Name returns the unique name of the rule
	Name() string

	// Version returns the semantic version of the plugin
	Version() string

	// Check validates files and returns violations
	Check(files []FileInfo) ([]Violation, error)
}

// FileInfo provides information about files to plugins
type FileInfo struct {
	Path    string // Relative path from root
	AbsPath string // Absolute path
	IsDir   bool   // Whether this is a directory
}

// Violation represents a rule violation found by a plugin
type Violation struct {
	Rule    string // Name of the rule that was violated
	Path    string // Path to the file with the violation
	Message string // Human-readable violation message
	Line    int    // Line number (0 if not applicable)
	Column  int    // Column number (0 if not applicable)
}

// Registry manages registered plugins
type Registry struct {
	parsers map[string]Parser     // Map from extension to parser
	rules   map[string]RulePlugin // Map from rule name to plugin
	mu      sync.RWMutex
}

var globalRegistry = &Registry{
	parsers: make(map[string]Parser),
	rules:   make(map[string]RulePlugin),
}

// GetRegistry returns the global plugin registry
func GetRegistry() *Registry {
	return globalRegistry
}

// RegisterParser registers a custom language parser
func (r *Registry) RegisterParser(parser Parser) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Register all supported extensions
	for _, ext := range parser.SupportedExtensions() {
		if existing, exists := r.parsers[ext]; exists {
			return fmt.Errorf("parser already registered for extension %s: %s", ext, existing.Name())
		}
		r.parsers[ext] = parser
	}

	return nil
}

// RegisterRule registers a custom linting rule
func (r *Registry) RegisterRule(rule RulePlugin) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, exists := r.rules[rule.Name()]; exists {
		return fmt.Errorf("rule already registered: %s (version %s)", existing.Name(), existing.Version())
	}

	r.rules[rule.Name()] = rule
	return nil
}

// GetParser returns the parser for a given file extension
func (r *Registry) GetParser(ext string) (Parser, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	parser, exists := r.parsers[ext]
	return parser, exists
}

// GetParserForFile returns the parser for a given file path
func (r *Registry) GetParserForFile(filePath string) (Parser, bool) {
	ext := filepath.Ext(filePath)
	return r.GetParser(ext)
}

// GetRule returns a custom rule by name
func (r *Registry) GetRule(name string) (RulePlugin, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rule, exists := r.rules[name]
	return rule, exists
}

// ListParsers returns all registered parsers
func (r *Registry) ListParsers() []ParserInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build unique list (one per parser, not per extension)
	seen := make(map[string]bool)
	var parsers []ParserInfo

	for _, parser := range r.parsers {
		if !seen[parser.Name()] {
			seen[parser.Name()] = true
			parsers = append(parsers, ParserInfo{
				Name:       parser.Name(),
				Version:    parser.Version(),
				Extensions: parser.SupportedExtensions(),
			})
		}
	}

	return parsers
}

// ListRules returns all registered custom rules
func (r *Registry) ListRules() []RuleInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var rules []RuleInfo
	for _, rule := range r.rules {
		rules = append(rules, RuleInfo{
			Name:    rule.Name(),
			Version: rule.Version(),
		})
	}

	return rules
}

// ParserInfo provides information about a registered parser
type ParserInfo struct {
	Name       string
	Version    string
	Extensions []string
}

// RuleInfo provides information about a registered rule
type RuleInfo struct {
	Name    string
	Version string
}

// Helper functions for plugin developers

// RegisterParser is a convenience function to register a parser with the global registry
func RegisterParser(parser Parser) error {
	return globalRegistry.RegisterParser(parser)
}

// RegisterRule is a convenience function to register a rule with the global registry
func RegisterRule(rule RulePlugin) error {
	return globalRegistry.RegisterRule(rule)
}
