package linter

import (
	"fmt"
	"strings"

	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/rules"
	"github.com/structurelint/structurelint/internal/walker"
)

// Linter is the main linter orchestrator
type Linter struct {
	config         *config.Config
	productionMode bool
}

// Violation is an alias for rules.Violation
type Violation = rules.Violation

// New creates a new Linter
func New() *Linter {
	return &Linter{
		productionMode: false,
	}
}

// WithProductionMode enables production mode (excludes test files from analysis)
func (l *Linter) WithProductionMode(enabled bool) *Linter {
	l.productionMode = enabled
	return l
}

// Lint runs the linter on the given path
func (l *Linter) Lint(path string) ([]Violation, error) {
	// Load configuration
	configs, err := config.FindConfigs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Merge configurations
	l.config = config.Merge(configs...)

	// Walk the filesystem
	w := walker.New(path).WithExclude(l.config.Exclude)
	if err := w.Walk(); err != nil {
		return nil, fmt.Errorf("failed to walk filesystem: %w", err)
	}

	files := w.GetFiles()
	dirs := w.GetDirs()

	// Filter files for production mode if enabled
	var filteredFiles []walker.FileInfo
	if l.productionMode {
		filteredFiles = l.filterProductionFiles(files)
	} else {
		filteredFiles = files
	}

	// Build import graph if layers are configured (Phase 1) or Phase 2 rules are enabled
	var importGraph *graph.ImportGraph
	needsGraph := len(l.config.Layers) > 0 ||
		l.isRuleEnabled("enforce-layer-boundaries") ||
		l.isRuleEnabled("disallow-orphaned-files") ||
		l.isRuleEnabled("disallow-unused-exports")

	if needsGraph {
		builder := graph.NewBuilder(path, l.config.Layers)
		var err error
		// Use filtered files for production mode
		importGraph, err = builder.Build(filteredFiles)
		if err != nil {
			return nil, fmt.Errorf("failed to build import graph: %w", err)
		}
	}

	// Create rules based on configuration
	rulesList := l.createRules(files, importGraph)

	// Execute all rules and collect violations
	var violations []Violation
	for _, rule := range rulesList {
		ruleViolations := rule.Check(files, dirs)
		violations = append(violations, ruleViolations...)
	}

	return violations, nil
}

// createRules instantiates rules based on the configuration using a registry pattern
func (l *Linter) createRules(files []walker.FileInfo, importGraph *graph.ImportGraph) []rules.Rule {
	var rulesList []rules.Rule

	// Simple rules that take a single integer "max" parameter
	simpleIntRules := map[string]func(int) rules.Rule{
		"max-depth":        func(max int) rules.Rule { return rules.NewMaxDepthRule(max) },
		"max-files-in-dir": func(max int) rules.Rule { return rules.NewMaxFilesRule(max) },
		"max-subdirs":      func(max int) rules.Rule { return rules.NewMaxSubdirsRule(max) },
	}

	for ruleName, factory := range simpleIntRules {
		if max, ok := l.getIntConfig(ruleName, "max"); ok {
			rulesList = append(rulesList, factory(max))
		}
	}

	// String map rules (naming-convention, file-existence, regex-match, file-hash)
	stringMapRules := map[string]func(map[string]string) rules.Rule{
		"naming-convention": func(patterns map[string]string) rules.Rule { return rules.NewNamingConventionRule(patterns) },
		"file-existence":    func(requirements map[string]string) rules.Rule { return rules.NewFileExistenceRule(requirements) },
		"regex-match":       func(patterns map[string]string) rules.Rule { return rules.NewRegexMatchRule(patterns) },
		"file-hash":         func(hashes map[string]string) rules.Rule { return rules.NewFileHashRule(hashes) },
	}

	for ruleName, factory := range stringMapRules {
		if config, ok := l.getStringMapConfig(ruleName); ok {
			rulesList = append(rulesList, factory(config))
		}
	}

	// String slice rules
	if patterns, ok := l.getStringSliceConfig("disallowed-patterns"); ok {
		rulesList = append(rulesList, rules.NewDisallowedPatternsRule(patterns))
	}

	// Complex rules that need custom handling
	l.addComplexRules(&rulesList, importGraph)

	return rulesList
}

// addComplexRules adds rules that require more complex configuration
func (l *Linter) addComplexRules(rulesList *[]rules.Rule, importGraph *graph.ImportGraph) {
	// Max cyclomatic complexity rule
	if complexity, ok := l.getRuleConfig("max-cyclomatic-complexity"); ok {
		if complexityMap, ok := complexity.(map[string]interface{}); ok {
			max := l.getIntFromMap(complexityMap, "max")
			filePatterns := l.getStringSliceFromMap(complexityMap, "file-patterns")
			if max > 0 {
				*rulesList = append(*rulesList, rules.NewMaxCyclomaticComplexityRule(max, filePatterns))
			}
		}
	}

	// Graph-dependent rules
	if importGraph != nil {
		if _, ok := l.getRuleConfig("enforce-layer-boundaries"); ok {
			if len(l.config.Layers) > 0 {
				*rulesList = append(*rulesList, rules.NewLayerBoundariesRule(importGraph))
			}
		}

		if _, ok := l.getRuleConfig("disallow-orphaned-files"); ok {
			*rulesList = append(*rulesList, rules.NewOrphanedFilesRule(importGraph, l.config.Entrypoints))
		}

		if _, ok := l.getRuleConfig("disallow-unused-exports"); ok {
			*rulesList = append(*rulesList, rules.NewUnusedExportsRule(importGraph))
		}
	}

	// Test validation rules (Phase 3)
	l.addTestValidationRules(rulesList)

	// Content rules (Phase 4)
	l.addContentRules(rulesList)
}

// addTestValidationRules adds Phase 3 test validation rules
func (l *Linter) addTestValidationRules(rulesList *[]rules.Rule) {
	// Test adjacency rule
	if testAdj, ok := l.getRuleConfig("test-adjacency"); ok {
		if adjMap, ok := testAdj.(map[string]interface{}); ok {
			pattern := l.getStringFromMap(adjMap, "pattern")
			testDir := l.getStringFromMap(adjMap, "test-dir")
			filePatterns := l.getStringSliceFromMap(adjMap, "file-patterns")
			exemptions := l.getStringSliceFromMap(adjMap, "exemptions")

			if pattern != "" && len(filePatterns) > 0 {
				*rulesList = append(*rulesList, rules.NewTestAdjacencyRule(pattern, testDir, filePatterns, exemptions))
			}
		}
	}

	// Test location rule
	if testLoc, ok := l.getRuleConfig("test-location"); ok {
		if locMap, ok := testLoc.(map[string]interface{}); ok {
			integrationDir := l.getStringFromMap(locMap, "integration-test-dir")
			allowAdjacent := l.getBoolFromMap(locMap, "allow-adjacent")
			exemptions := l.getStringSliceFromMap(locMap, "exemptions")

			*rulesList = append(*rulesList, rules.NewTestLocationRule(integrationDir, allowAdjacent, exemptions))
		}
	}
}

// addContentRules adds Phase 4 file content rules
func (l *Linter) addContentRules(rulesList *[]rules.Rule) {
	// File content rule
	if fileContent, ok := l.getRuleConfig("file-content"); ok {
		if contentMap, ok := fileContent.(map[string]interface{}); ok {
			templates := l.getStringMapFromMap(contentMap, "templates")
			templateDir := l.getStringFromMap(contentMap, "template-dir")

			if len(templates) > 0 && templateDir != "" {
				// Get root path from linter (need to pass it through)
				rootPath := "." // Default to current directory
				*rulesList = append(*rulesList, rules.NewFileContentRule(templates, templateDir, rootPath))
			}
		}
	}
}

// getIntConfig extracts an integer value from a rule's configuration map
func (l *Linter) getIntConfig(ruleName, key string) (int, bool) {
	config, ok := l.getRuleConfig(ruleName)
	if !ok {
		return 0, false
	}

	configMap, ok := config.(map[string]interface{})
	if !ok {
		return 0, false
	}

	value, ok := configMap[key].(int)
	return value, ok
}

// getStringMapConfig extracts a string-to-string map from a rule's configuration
func (l *Linter) getStringMapConfig(ruleName string) (map[string]string, bool) {
	config, ok := l.getRuleConfig(ruleName)
	if !ok {
		return nil, false
	}

	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil, false
	}

	result := make(map[string]string)
	for k, v := range configMap {
		if strVal, ok := v.(string); ok {
			result[k] = strVal
		}
	}

	if len(result) == 0 {
		return nil, false
	}

	return result, true
}

// getStringSliceConfig extracts a slice of strings from a rule's configuration
func (l *Linter) getStringSliceConfig(ruleName string) ([]string, bool) {
	config, ok := l.getRuleConfig(ruleName)
	if !ok {
		return nil, false
	}

	configSlice, ok := config.([]interface{})
	if !ok {
		return nil, false
	}

	result := make([]string, 0, len(configSlice))
	for _, v := range configSlice {
		if strVal, ok := v.(string); ok {
			result = append(result, strVal)
		}
	}

	if len(result) == 0 {
		return nil, false
	}

	return result, true
}

// getRuleConfig safely extracts a rule configuration
func (l *Linter) getRuleConfig(ruleName string) (interface{}, bool) {
	if l.config == nil || l.config.Rules == nil {
		return nil, false
	}

	value, exists := l.config.Rules[ruleName]
	if !exists {
		return nil, false
	}

	// Check if rule is disabled (value is 0 or false)
	switch v := value.(type) {
	case int:
		if v == 0 {
			return nil, false
		}
	case bool:
		if !v {
			return nil, false
		}
	}

	return value, true
}

// isRuleEnabled checks if a rule is enabled in the configuration
func (l *Linter) isRuleEnabled(ruleName string) bool {
	_, enabled := l.getRuleConfig(ruleName)
	return enabled
}

// getStringFromMap extracts a string value from a map
func (l *Linter) getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

// getBoolFromMap extracts a boolean value from a map
func (l *Linter) getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

// getIntFromMap extracts an integer value from a map
func (l *Linter) getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	// Also handle float64 (common from YAML parsing)
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

// getStringSliceFromMap extracts a string slice from a map
func (l *Linter) getStringSliceFromMap(m map[string]interface{}, key string) []string {
	if val, ok := m[key].([]interface{}); ok {
		result := make([]string, 0, len(val))
		for _, v := range val {
			if strVal, ok := v.(string); ok {
				result = append(result, strVal)
			}
		}
		return result
	}
	return nil
}

// getStringMapFromMap extracts a string map from a map
func (l *Linter) getStringMapFromMap(m map[string]interface{}, key string) map[string]string {
	if val, ok := m[key].(map[string]interface{}); ok {
		result := make(map[string]string)
		for k, v := range val {
			if strVal, ok := v.(string); ok {
				result[k] = strVal
			}
		}
		return result
	}
	return nil
}

// filterProductionFiles filters out test files from the file list
func (l *Linter) filterProductionFiles(files []walker.FileInfo) []walker.FileInfo {
	var production []walker.FileInfo

	for _, file := range files {
		if !l.isTestFile(file.Path) {
			production = append(production, file)
		}
	}

	return production
}

// isTestFile checks if a file is a test file based on common patterns
func (l *Linter) isTestFile(path string) bool {
	// Common test file patterns
	testPatterns := []string{
		"_test.go",       // Go tests
		"_test.ts",       // TypeScript tests
		"_test.js",       // JavaScript tests
		".test.ts",       // TypeScript tests (alternative)
		".test.js",       // JavaScript tests (alternative)
		".spec.ts",       // TypeScript specs
		".spec.js",       // JavaScript specs
		"test_",          // Python tests
		"__tests__/",     // Jest tests directory
		"/test/",         // Generic test directory
		"/tests/",        // Generic tests directory
		"spec/",          // RSpec or similar
	}

	for _, pattern := range testPatterns {
		if strings.Contains(path, pattern) {
			return true
		}
	}

	return false
}
