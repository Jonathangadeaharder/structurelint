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
	config  *config.Config
	rootDir string // Root directory being linted
}

// Violation is an alias for rules.Violation
type Violation = rules.Violation

// New creates a new Linter
func New() *Linter {
	return &Linter{}
}

// Lint runs the linter on the given path
func (l *Linter) Lint(path string) ([]Violation, error) {
	// Store root directory for later use
	l.rootDir = path

	// Load configuration with auto-loaded .gitignore patterns
	configs, err := config.FindConfigsWithGitignore(path)
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
		l.isRuleEnabled("disallow-unused-exports") ||
		l.isRuleEnabled("property-enforcement")

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

	// String map rules (file-existence, regex-match)
	stringMapRules := map[string]func(map[string]string) rules.Rule{
		"file-existence":    func(requirements map[string]string) rules.Rule { return rules.NewFileExistenceRule(requirements) },
		"regex-match":       func(patterns map[string]string) rules.Rule { return rules.NewRegexMatchRule(patterns) },
	}

	for ruleName, factory := range stringMapRules {
		if config, ok := l.getStringMapConfig(ruleName); ok {
			rulesList = append(rulesList, factory(config))
		}
	}

	// Naming convention rule - with language-aware defaults (Priority 2 feature)
	if _, ok := l.getRuleConfig("naming-convention"); ok {
		userPatterns, _ := l.getStringMapConfig("naming-convention")

		// Check if auto-language-naming is enabled (default: true)
		autoLanguageNaming := true
		if l.config.AutoLanguageNaming != nil {
			autoLanguageNaming = *l.config.AutoLanguageNaming
		}

		if autoLanguageNaming {
			// Use language-aware naming with user overrides
			if rule, err := rules.NewLanguageAwareNamingConventionRule(l.rootDir, userPatterns); err == nil {
				rulesList = append(rulesList, rule)
			}
		} else {
			// Traditional behavior - only use user patterns
			if len(userPatterns) > 0 {
				rulesList = append(rulesList, rules.NewNamingConventionRule(userPatterns))
			}
		}
	}

	// String slice rules
	if patterns, ok := l.getStringSliceConfig("disallowed-patterns"); ok {
		rulesList = append(rulesList, rules.NewDisallowedPatternsRule(patterns))
	}

	// Uniqueness constraints rule (Priority 2 feature)
	if constraints, ok := l.getStringMapConfig("uniqueness-constraints"); ok {
		rulesList = append(rulesList, rules.NewUniquenessConstraintsRule(constraints))
	}

	// Complex rules that need custom handling
	l.addComplexRules(&rulesList, importGraph)

	return rulesList
}

// addComplexRules adds rules that require more complex configuration
func (l *Linter) addComplexRules(rulesList *[]rules.Rule, importGraph *graph.ImportGraph) {
	l.checkBreakingChanges()
	l.addComplexityRules(rulesList)
	l.addGraphDependentRules(rulesList, importGraph)
	l.addPathBasedLayerRule(rulesList)
	l.addTestValidationRules(rulesList)
	l.addContentRules(rulesList)
	l.addGitHubWorkflowsRule(rulesList)
	l.addLinterConfigRule(rulesList)
}

// checkBreakingChanges checks for deprecated rule configurations
func (l *Linter) checkBreakingChanges() {
	if _, ok := l.getRuleConfig("max-cyclomatic-complexity"); ok {
		panic("BREAKING CHANGE: 'max-cyclomatic-complexity' rule has been removed.\n" +
			"Use 'max-cognitive-complexity' instead - it's scientifically superior (r=0.54 vs cyclomatic's weak correlation).\n" +
			"See: https://github.com/structurelint/structurelint#phase-5-evidence-based-metrics")
	}
}

// addComplexityRules adds evidence-based complexity measurement rules
func (l *Linter) addComplexityRules(rulesList *[]rules.Rule) {
	l.addCognitiveComplexityRule(rulesList)
	l.addHalsteadEffortRule(rulesList)
}

// addCognitiveComplexityRule adds max cognitive complexity rule
func (l *Linter) addCognitiveComplexityRule(rulesList *[]rules.Rule) {
	cognitiveComplexity, ok := l.getRuleConfig("max-cognitive-complexity")
	if !ok {
		return
	}

	complexityMap, ok := cognitiveComplexity.(map[string]interface{})
	if !ok {
		return
	}

	max := l.getIntFromMap(complexityMap, "max")
	if max <= 0 {
		return
	}

	testMax := l.getIntFromMap(complexityMap, "test-max")
	filePatterns := l.getStringSliceFromMap(complexityMap, "file-patterns")

	rule := rules.NewMaxCognitiveComplexityRule(max, filePatterns)
	if testMax > 0 {
		rule = rule.WithTestMax(testMax)
	}
	*rulesList = append(*rulesList, rule)
}

// addHalsteadEffortRule adds max Halstead effort rule
func (l *Linter) addHalsteadEffortRule(rulesList *[]rules.Rule) {
	halsteadEffort, ok := l.getRuleConfig("max-halstead-effort")
	if !ok {
		return
	}

	effortMap, ok := halsteadEffort.(map[string]interface{})
	if !ok {
		return
	}

	max := l.getFloatFromMap(effortMap, "max")
	if max <= 0 {
		return
	}

	filePatterns := l.getStringSliceFromMap(effortMap, "file-patterns")
	*rulesList = append(*rulesList, rules.NewMaxHalsteadEffortRule(max, filePatterns))
}

// addGraphDependentRules adds rules that require import graph analysis
func (l *Linter) addGraphDependentRules(rulesList *[]rules.Rule, importGraph *graph.ImportGraph) {
	if importGraph == nil {
		return
	}

	l.addLayerBoundariesRule(rulesList, importGraph)
	l.addOrphanedFilesRule(rulesList, importGraph)
	l.addUnusedExportsRule(rulesList, importGraph)
	l.addPropertyEnforcementRule(rulesList, importGraph)
}

// addLayerBoundariesRule adds layer boundaries enforcement rule
func (l *Linter) addLayerBoundariesRule(rulesList *[]rules.Rule, importGraph *graph.ImportGraph) {
	if _, ok := l.getRuleConfig("enforce-layer-boundaries"); ok {
		if len(l.config.Layers) > 0 {
			*rulesList = append(*rulesList, rules.NewLayerBoundariesRule(importGraph))
		}
	}
}

// addOrphanedFilesRule adds orphaned files detection rule
func (l *Linter) addOrphanedFilesRule(rulesList *[]rules.Rule, importGraph *graph.ImportGraph) {
	ruleConfig, ok := l.getRuleConfig("disallow-orphaned-files")
	if !ok {
		return
	}

	rule := rules.NewOrphanedFilesRule(importGraph, l.config.Entrypoints)

	// Check for entry-point-patterns in the rule config
	if configMap, ok := ruleConfig.(map[string]interface{}); ok {
		if patterns, ok := configMap["entry-point-patterns"].([]interface{}); ok {
			entryPointPatterns := l.extractStringSlice(patterns)
			if len(entryPointPatterns) > 0 {
				rule = rule.WithEntryPointPatterns(entryPointPatterns)
			}
		}
	}

	*rulesList = append(*rulesList, rule)
}

// extractStringSlice converts []interface{} to []string
func (l *Linter) extractStringSlice(patterns []interface{}) []string {
	var result []string
	for _, p := range patterns {
		if str, ok := p.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

// addUnusedExportsRule adds unused exports detection rule
func (l *Linter) addUnusedExportsRule(rulesList *[]rules.Rule, importGraph *graph.ImportGraph) {
	if _, ok := l.getRuleConfig("disallow-unused-exports"); ok {
		*rulesList = append(*rulesList, rules.NewUnusedExportsRule(importGraph))
	}
}

// addPropertyEnforcementRule adds comprehensive dependency management rule
func (l *Linter) addPropertyEnforcementRule(rulesList *[]rules.Rule, importGraph *graph.ImportGraph) {
	propertyConfig, ok := l.getRuleConfig("property-enforcement")
	if !ok {
		return
	}

	configMap, ok := propertyConfig.(map[string]interface{})
	if !ok {
		return
	}

	config := rules.PropertyEnforcementConfig{
		MaxDependenciesPerFile: l.getIntFromMap(configMap, "max_dependencies_per_file"),
		MaxDependencyDepth:     l.getIntFromMap(configMap, "max_dependency_depth"),
		DetectCycles:           l.getBoolFromMap(configMap, "detect_cycles"),
		ForbiddenPatterns:      l.getStringSliceFromMap(configMap, "forbidden_patterns"),
	}
	*rulesList = append(*rulesList, rules.NewPropertyEnforcementRule(importGraph, config))
}

// addPathBasedLayerRule adds path-based layer validation rule
func (l *Linter) addPathBasedLayerRule(rulesList *[]rules.Rule) {
	ruleConfig, ok := l.getRuleConfig("path-based-layers")
	if !ok {
		return
	}

	configMap, ok := ruleConfig.(map[string]interface{})
	if !ok {
		return
	}

	layersConfig, ok := configMap["layers"].([]interface{})
	if !ok {
		return
	}

	pathLayers := l.parsePathLayers(layersConfig)
	if len(pathLayers) > 0 {
		*rulesList = append(*rulesList, rules.NewPathBasedLayerRule(pathLayers))
	}
}

// parsePathLayers parses path layer configurations
func (l *Linter) parsePathLayers(layersConfig []interface{}) []rules.PathLayer {
	var pathLayers []rules.PathLayer

	for _, layerInterface := range layersConfig {
		layerMap, ok := layerInterface.(map[string]interface{})
		if !ok {
			continue
		}

		layer := rules.PathLayer{
			Name:           l.getStringFromMap(layerMap, "name"),
			Patterns:       l.getStringSliceFromMap(layerMap, "patterns"),
			CanDependOn:    l.getStringSliceFromMap(layerMap, "canDependOn"),
			ForbiddenPaths: l.getStringSliceFromMap(layerMap, "forbiddenPaths"),
		}
		pathLayers = append(pathLayers, layer)
	}

	return pathLayers
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
			filePatterns := l.getStringSliceFromMap(locMap, "file-patterns")
			exemptions := l.getStringSliceFromMap(locMap, "exemptions")

			*rulesList = append(*rulesList, rules.NewTestLocationRule(integrationDir, allowAdjacent, filePatterns, exemptions))
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

// addGitHubWorkflowsRule adds the GitHub workflows rule
func (l *Linter) addGitHubWorkflowsRule(rulesList *[]rules.Rule) {
	if workflowConfig, ok := l.getRuleConfig("github-workflows"); ok {
		if configMap, ok := workflowConfig.(map[string]interface{}); ok {
			requireTests := l.getBoolFromMap(configMap, "require-tests")
			requireSecurity := l.getBoolFromMap(configMap, "require-security")
			requireQuality := l.getBoolFromMap(configMap, "require-quality")
			requireLogCommits := l.getBoolFromMap(configMap, "require-log-commits")
			requireRepomixArtifact := l.getBoolFromMap(configMap, "require-repomix-artifact")
			requiredJobs := l.getStringSliceFromMap(configMap, "required-jobs")
			requiredTriggers := l.getStringSliceFromMap(configMap, "required-triggers")
			allowMissing := l.getStringSliceFromMap(configMap, "allow-missing")

			rule := rules.NewGitHubWorkflowsRule(rules.GitHubWorkflowsRule{
				RequireTests:          requireTests,
				RequireSecurity:       requireSecurity,
				RequireQuality:        requireQuality,
				RequireLogCommits:     requireLogCommits,
				RequireRepomixArtifact: requireRepomixArtifact,
				RequiredJobs:          requiredJobs,
				RequiredTriggers:      requiredTriggers,
				AllowMissing:          allowMissing,
			})
			*rulesList = append(*rulesList, rule)
		}
	}
}

// addLinterConfigRule adds the linter configuration rule
func (l *Linter) addLinterConfigRule(rulesList *[]rules.Rule) {
	if linterConfig, ok := l.getRuleConfig("linter-config"); ok {
		if configMap, ok := linterConfig.(map[string]interface{}); ok {
			requirePython := l.getBoolFromMap(configMap, "require-python")
			requireTypeScript := l.getBoolFromMap(configMap, "require-typescript")
			requireGo := l.getBoolFromMap(configMap, "require-go")
			requireHTML := l.getBoolFromMap(configMap, "require-html")
			requireCSS := l.getBoolFromMap(configMap, "require-css")
			requireSQL := l.getBoolFromMap(configMap, "require-sql")
			requireRust := l.getBoolFromMap(configMap, "require-rust")
			customLinters := l.getStringSliceFromMap(configMap, "custom-linters")

			rule := rules.NewLinterConfigRule(rules.LinterConfigRule{
				RequirePython:     requirePython,
				RequireTypeScript: requireTypeScript,
				RequireGo:         requireGo,
				RequireHTML:       requireHTML,
				RequireCSS:        requireCSS,
				RequireSQL:        requireSQL,
				RequireRust:       requireRust,
				CustomLinters:     customLinters,
			})
			*rulesList = append(*rulesList, rule)
		}
	}

	// API specification rule
	l.addOpenAPIAsyncAPIRule(rulesList)

	// Contract framework rule
	l.addContractFrameworkRule(rulesList)

	// Specification and ADR enforcement rule
	l.addSpecADRRule(rulesList)
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

// getFloatFromMap extracts a float64 value from a map
func (l *Linter) getFloatFromMap(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	// Also handle int (convert to float64)
	if val, ok := m[key].(int); ok {
		return float64(val)
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

// addOpenAPIAsyncAPIRule adds the API specification rule
func (l *Linter) addOpenAPIAsyncAPIRule(rulesList *[]rules.Rule) {
	if apiSpecConfig, ok := l.getRuleConfig("api-spec"); ok {
		if configMap, ok := apiSpecConfig.(map[string]interface{}); ok {
			requireOpenAPI := l.getBoolFromMap(configMap, "require-openapi")
			requireAsyncAPI := l.getBoolFromMap(configMap, "require-asyncapi")
			customSpecs := l.getStringSliceFromMap(configMap, "custom-specs")

			rule := rules.NewOpenAPIAsyncAPIRule(rules.OpenAPIAsyncAPIRule{
				RequireOpenAPI:  requireOpenAPI,
				RequireAsyncAPI: requireAsyncAPI,
				CustomSpecs:     customSpecs,
			})
			*rulesList = append(*rulesList, rule)
		}
	}
}

// addContractFrameworkRule adds the contract framework rule
func (l *Linter) addContractFrameworkRule(rulesList *[]rules.Rule) {
	if contractConfig, ok := l.getRuleConfig("contract-framework"); ok {
		if configMap, ok := contractConfig.(map[string]interface{}); ok {
			requirePython := l.getBoolFromMap(configMap, "require-python")
			requireRust := l.getBoolFromMap(configMap, "require-rust")
			requireTypeScript := l.getBoolFromMap(configMap, "require-typescript")
			requireGo := l.getBoolFromMap(configMap, "require-go")
			requireJava := l.getBoolFromMap(configMap, "require-java")
			requireCSharp := l.getBoolFromMap(configMap, "require-csharp")
			requireCPlusPlus := l.getBoolFromMap(configMap, "require-cplusplus")
			customFrameworks := l.getStringSliceFromMap(configMap, "custom-frameworks")

			rule := rules.NewContractFrameworkRule(rules.ContractFrameworkRule{
				RequirePython:     requirePython,
				RequireRust:       requireRust,
				RequireTypeScript: requireTypeScript,
				RequireGo:         requireGo,
				RequireJava:       requireJava,
				RequireCSharp:     requireCSharp,
				RequireCPlusPlus:  requireCPlusPlus,
				CustomFrameworks:  customFrameworks,
			})
			*rulesList = append(*rulesList, rule)
		}
	}
}

// addSpecADRRule adds the specification and ADR enforcement rule
func (l *Linter) addSpecADRRule(rulesList *[]rules.Rule) {
	if specADRConfig, ok := l.getRuleConfig("spec-adr-enforcement"); ok {
		if configMap, ok := specADRConfig.(map[string]interface{}); ok {
			requireSpecFolder := l.getBoolFromMap(configMap, "require-spec-folder")
			requireADRFolder := l.getBoolFromMap(configMap, "require-adr-folder")
			enforceSpecTemplate := l.getBoolFromMap(configMap, "enforce-spec-template")
			enforceADRTemplate := l.getBoolFromMap(configMap, "enforce-adr-template")
			specFolderPaths := l.getStringSliceFromMap(configMap, "spec-folder-paths")
			adrFolderPaths := l.getStringSliceFromMap(configMap, "adr-folder-paths")
			specFilePatterns := l.getStringSliceFromMap(configMap, "spec-file-patterns")
			adrFilePatterns := l.getStringSliceFromMap(configMap, "adr-file-patterns")

			rule := rules.NewSpecADRRule(rules.SpecADRRule{
				RequireSpecFolder:   requireSpecFolder,
				RequireADRFolder:    requireADRFolder,
				EnforceSpecTemplate: enforceSpecTemplate,
				EnforceADRTemplate:  enforceADRTemplate,
				SpecFolderPaths:     specFolderPaths,
				ADRFolderPaths:      adrFolderPaths,
				SpecFilePatterns:    specFilePatterns,
				ADRFilePatterns:     adrFilePatterns,
			})
			*rulesList = append(*rulesList, rule)
		}
	}
}
