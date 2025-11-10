package linter

import (
	"fmt"

	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/rules"
	"github.com/structurelint/structurelint/internal/walker"
)

// Linter is the main linter orchestrator
type Linter struct {
	config *config.Config
}

// Violation is an alias for rules.Violation
type Violation = rules.Violation

// New creates a new Linter
func New() *Linter {
	return &Linter{}
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

	// Build import graph if layers are configured (Phase 1) or Phase 2 rules are enabled
	var importGraph *graph.ImportGraph
	needsGraph := len(l.config.Layers) > 0 ||
		l.isRuleEnabled("enforce-layer-boundaries") ||
		l.isRuleEnabled("disallow-orphaned-files") ||
		l.isRuleEnabled("disallow-unused-exports")

	if needsGraph {
		builder := graph.NewBuilder(path, l.config.Layers)
		var err error
		importGraph, err = builder.Build(files)
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

// createRules instantiates rules based on the configuration
func (l *Linter) createRules(files []walker.FileInfo, importGraph *graph.ImportGraph) []rules.Rule {
	var rulesList []rules.Rule

	// Max depth rule
	if max, ok := l.getIntConfig("max-depth", "max"); ok {
		rulesList = append(rulesList, rules.NewMaxDepthRule(max))
	}

	// Max files rule
	if max, ok := l.getIntConfig("max-files-in-dir", "max"); ok {
		rulesList = append(rulesList, rules.NewMaxFilesRule(max))
	}

	// Max subdirs rule
	if max, ok := l.getIntConfig("max-subdirs", "max"); ok {
		rulesList = append(rulesList, rules.NewMaxSubdirsRule(max))
	}

	// Naming convention rule
	if patterns, ok := l.getStringMapConfig("naming-convention"); ok {
		rulesList = append(rulesList, rules.NewNamingConventionRule(patterns))
	}

	// Disallowed patterns rule
	if patterns, ok := l.getStringSliceConfig("disallowed-patterns"); ok {
		rulesList = append(rulesList, rules.NewDisallowedPatternsRule(patterns))
	}

	// File existence rule
	if requirements, ok := l.getStringMapConfig("file-existence"); ok {
		rulesList = append(rulesList, rules.NewFileExistenceRule(requirements))
	}

	// Regex match rule
	if patterns, ok := l.getStringMapConfig("regex-match"); ok {
		rulesList = append(rulesList, rules.NewRegexMatchRule(patterns))
	}

	// Layer boundaries rule (Phase 1)
	if _, ok := l.getRuleConfig("enforce-layer-boundaries"); ok {
		if importGraph != nil && len(l.config.Layers) > 0 {
			rulesList = append(rulesList, rules.NewLayerBoundariesRule(importGraph))
		}
	}

	// Orphaned files rule (Phase 2)
	if _, ok := l.getRuleConfig("disallow-orphaned-files"); ok {
		if importGraph != nil {
			rulesList = append(rulesList, rules.NewOrphanedFilesRule(importGraph, l.config.Entrypoints))
		}
	}

	// Unused exports rule (Phase 2)
	if _, ok := l.getRuleConfig("disallow-unused-exports"); ok {
		if importGraph != nil {
			rulesList = append(rulesList, rules.NewUnusedExportsRule(importGraph))
		}
	}

	// Test adjacency rule
	if testAdj, ok := l.getRuleConfig("test-adjacency"); ok {
		if adjMap, ok := testAdj.(map[string]interface{}); ok {
			pattern := l.getStringFromMap(adjMap, "pattern")
			testDir := l.getStringFromMap(adjMap, "test-dir")
			filePatterns := l.getStringSliceFromMap(adjMap, "file-patterns")
			exemptions := l.getStringSliceFromMap(adjMap, "exemptions")

			if pattern != "" && len(filePatterns) > 0 {
				rulesList = append(rulesList, rules.NewTestAdjacencyRule(pattern, testDir, filePatterns, exemptions))
			}
		}
	}

	// Test location rule
	if testLoc, ok := l.getRuleConfig("test-location"); ok {
		if locMap, ok := testLoc.(map[string]interface{}); ok {
			integrationDir := l.getStringFromMap(locMap, "integration-test-dir")
			allowAdjacent := l.getBoolFromMap(locMap, "allow-adjacent")
			exemptions := l.getStringSliceFromMap(locMap, "exemptions")

			rulesList = append(rulesList, rules.NewTestLocationRule(integrationDir, allowAdjacent, exemptions))
		}
	}

	// File content rule
	if fileContent, ok := l.getRuleConfig("file-content"); ok {
		if contentMap, ok := fileContent.(map[string]interface{}); ok {
			templates := l.getStringMapFromMap(contentMap, "templates")
			templateDir := l.getStringFromMap(contentMap, "template-dir")

			if len(templates) > 0 && templateDir != "" {
				// Get root path from linter (need to pass it through)
				rootPath := "." // Default to current directory
				rulesList = append(rulesList, rules.NewFileContentRule(templates, templateDir, rootPath))
			}
		}
	}

	return rulesList
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
