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
//
//nolint:gocognit,gocyclo // Factory function with many rule types - complexity is acceptable
func (l *Linter) createRules(files []walker.FileInfo, importGraph *graph.ImportGraph) []rules.Rule {
	var rulesList []rules.Rule

	// Max depth rule
	if maxDepthConfig, ok := l.getRuleConfig("max-depth"); ok {
		if maxDepth, ok := maxDepthConfig.(map[string]interface{}); ok {
			if max, ok := maxDepth["max"].(int); ok {
				rulesList = append(rulesList, rules.NewMaxDepthRule(max))
			}
		}
	}

	// Max files rule
	if maxFilesConfig, ok := l.getRuleConfig("max-files-in-dir"); ok {
		if maxFiles, ok := maxFilesConfig.(map[string]interface{}); ok {
			if max, ok := maxFiles["max"].(int); ok {
				rulesList = append(rulesList, rules.NewMaxFilesRule(max))
			}
		}
	}

	// Max subdirs rule
	if maxSubdirsConfig, ok := l.getRuleConfig("max-subdirs"); ok {
		if maxSubdirs, ok := maxSubdirsConfig.(map[string]interface{}); ok {
			if max, ok := maxSubdirs["max"].(int); ok {
				rulesList = append(rulesList, rules.NewMaxSubdirsRule(max))
			}
		}
	}

	// Naming convention rule
	if namingConfig, ok := l.getRuleConfig("naming-convention"); ok {
		if patterns, ok := namingConfig.(map[string]interface{}); ok {
			stringPatterns := make(map[string]string)
			for k, v := range patterns {
				if strVal, ok := v.(string); ok {
					stringPatterns[k] = strVal
				}
			}
			if len(stringPatterns) > 0 {
				rulesList = append(rulesList, rules.NewNamingConventionRule(stringPatterns))
			}
		}
	}

	// Disallowed patterns rule
	if disallowedConfig, ok := l.getRuleConfig("disallowed-patterns"); ok {
		if patterns, ok := disallowedConfig.([]interface{}); ok {
			stringPatterns := make([]string, 0, len(patterns))
			for _, p := range patterns {
				if strVal, ok := p.(string); ok {
					stringPatterns = append(stringPatterns, strVal)
				}
			}
			if len(stringPatterns) > 0 {
				rulesList = append(rulesList, rules.NewDisallowedPatternsRule(stringPatterns))
			}
		}
	}

	// File existence rule
	if existenceConfig, ok := l.getRuleConfig("file-existence"); ok {
		if requirements, ok := existenceConfig.(map[string]interface{}); ok {
			stringRequirements := make(map[string]string)
			for k, v := range requirements {
				if strVal, ok := v.(string); ok {
					stringRequirements[k] = strVal
				}
			}
			if len(stringRequirements) > 0 {
				rulesList = append(rulesList, rules.NewFileExistenceRule(stringRequirements))
			}
		}
	}

	// Regex match rule
	if regexConfig, ok := l.getRuleConfig("regex-match"); ok {
		if patterns, ok := regexConfig.(map[string]interface{}); ok {
			stringPatterns := make(map[string]string)
			for k, v := range patterns {
				if strVal, ok := v.(string); ok {
					stringPatterns[k] = strVal
				}
			}
			if len(stringPatterns) > 0 {
				rulesList = append(rulesList, rules.NewRegexMatchRule(stringPatterns))
			}
		}
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

	return rulesList
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
