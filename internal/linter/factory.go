package linter

import (
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	rulesci "github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
	rulesgraph "github.com/Jonathangadeaharder/structurelint/internal/rules/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/rules/quality"
	rulestest "github.com/Jonathangadeaharder/structurelint/internal/rules/test"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"

	// Blank imports to trigger sub-package init() registrations
	_ "github.com/Jonathangadeaharder/structurelint/internal/rules/structure"
)

type RuleFactory struct {
	rootDir     string
	config      *config.Config
	importGraph *graph.ImportGraph
	files       []walker.FileInfo
	dirs        map[string]*walker.DirInfo
}

func NewRuleFactory(rootDir string, cfg *config.Config, importGraph *graph.ImportGraph) *RuleFactory {
	return &RuleFactory{
		rootDir:     rootDir,
		config:      cfg,
		importGraph: importGraph,
	}
}

func (f *RuleFactory) CreateRules(files []walker.FileInfo, dirs map[string]*walker.DirInfo) ([]rules.Rule, error) {
	f.files = files
	f.dirs = dirs

	if err := f.checkBreakingChanges(); err != nil {
		return nil, err
	}

	var rulesList []rules.Rule
	rulesList = append(rulesList, f.createRegistryRules()...)
	rulesList = append(rulesList, f.createComplexityRules()...)
	rulesList = append(rulesList, f.createGraphDependentRules()...)
	rulesList = append(rulesList, f.createPathBasedLayerRules()...)
	rulesList = append(rulesList, f.createTestValidationRules()...)
	rulesList = append(rulesList, f.createContentRules()...)
	rulesList = append(rulesList, f.createCIRules()...)
	rulesList = append(rulesList, f.createLinterConfigRules()...)

	return rulesList, nil
}

func (f *RuleFactory) checkBreakingChanges() error {
	if _, ok := f.config.Rules["max-cyclomatic-complexity"]; ok {
		return fmt.Errorf("BREAKING CHANGE: 'max-cyclomatic-complexity' rule has been removed.\n" +
			"Use 'max-cognitive-complexity' instead - it's scientifically superior (r=0.54 vs cyclomatic's weak correlation).\n" +
			"See: https://github.com/Jonathangadeaharder/structurelint#phase-5-evidence-based-metrics")
	}
	return nil
}

func (f *RuleFactory) createRegistryRules() []rules.Rule {
	var rulesList []rules.Rule

	for ruleName, ruleConfig := range f.config.Rules {
		if !f.isRuleEnabled(ruleName) {
			continue
		}

		factory, ok := rules.GetFactory(ruleName)
		if !ok {
			continue
		}

		ctx := &rules.RuleContext{
			RootDir:     f.rootDir,
			ImportGraph: f.importGraph,
			Config:      f.normalizeConfig(ruleConfig),
		}

		rule, err := factory(ctx)
		if err == nil && rule != nil {
			rulesList = append(rulesList, rule)
		}
	}

	return rulesList
}

func (f *RuleFactory) normalizeConfig(config interface{}) map[string]interface{} {
	if m, ok := config.(map[string]interface{}); ok {
		return m
	}
	return map[string]interface{}{
		"": config,
	}
}

func (f *RuleFactory) createComplexityRules() []rules.Rule {
	var rulesList []rules.Rule

	if rule := f.createCognitiveComplexityRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	if rule := f.createHalsteadEffortRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}

	return rulesList
}

func (f *RuleFactory) createCognitiveComplexityRule() rules.Rule {
	config, ok := f.config.Rules["max-cognitive-complexity"]
	if !ok || !f.isRuleEnabled("max-cognitive-complexity") {
		return nil
	}
	complexityMap, ok := config.(map[string]interface{})
	if !ok {
		return nil
	}
	max := f.getIntFromMap(complexityMap, "max")
	if max <= 0 {
		return nil
	}
	testMax := f.getIntFromMap(complexityMap, "test-max")
	filePatterns := f.getStringSliceFromMap(complexityMap, "file-patterns")
	rule := quality.NewMaxCognitiveComplexityRule(max, filePatterns)
	if testMax > 0 {
		rule = rule.WithTestMax(testMax)
	}
	return rule
}

func (f *RuleFactory) createHalsteadEffortRule() rules.Rule {
	config, ok := f.config.Rules["max-halstead-effort"]
	if !ok || !f.isRuleEnabled("max-halstead-effort") {
		return nil
	}
	effortMap, ok := config.(map[string]interface{})
	if !ok {
		return nil
	}
	max := f.getFloatFromMap(effortMap, "max")
	if max <= 0 {
		return nil
	}
	filePatterns := f.getStringSliceFromMap(effortMap, "file-patterns")
	return quality.NewMaxHalsteadEffortRule(max, filePatterns)
}

func (f *RuleFactory) createGraphDependentRules() []rules.Rule {
	if f.importGraph == nil {
		return nil
	}

	var rulesList []rules.Rule
	if rule := f.createLayerBoundariesRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	if rule := f.createOrphanedFilesRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	if rule := f.createUnusedExportsRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	if rule := f.createPropertyEnforcementRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	return rulesList
}

func (f *RuleFactory) createLayerBoundariesRule() rules.Rule {
	if _, ok := f.config.Rules["enforce-layer-boundaries"]; !ok {
		return nil
	}
	if !f.isRuleEnabled("enforce-layer-boundaries") || len(f.config.Layers) == 0 {
		return nil
	}
	return rulesgraph.NewLayerBoundariesRule(f.importGraph)
}

func (f *RuleFactory) createOrphanedFilesRule() rules.Rule {
	config, ok := f.config.Rules["disallow-orphaned-files"]
	if !ok || !f.isRuleEnabled("disallow-orphaned-files") {
		return nil
	}
	rule := rulesgraph.NewOrphanedFilesRule(f.importGraph, f.config.Entrypoints)
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return rule
	}
	patterns, ok := configMap["entry-point-patterns"].([]interface{})
	if !ok {
		return rule
	}
	entryPointPatterns := f.extractStringSlice(patterns)
	if len(entryPointPatterns) > 0 {
		rule = rule.WithEntryPointPatterns(entryPointPatterns)
	}
	return rule
}

func (f *RuleFactory) createUnusedExportsRule() rules.Rule {
	if _, ok := f.config.Rules["disallow-unused-exports"]; !ok {
		return nil
	}
	if !f.isRuleEnabled("disallow-unused-exports") {
		return nil
	}
	return rulesgraph.NewUnusedExportsRule(f.importGraph)
}

func (f *RuleFactory) createPropertyEnforcementRule() rules.Rule {
	config, ok := f.config.Rules["property-enforcement"]
	if !ok || !f.isRuleEnabled("property-enforcement") {
		return nil
	}
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil
	}
	enforcementConfig := rulesgraph.PropertyEnforcementConfig{
		MaxDependenciesPerFile: f.getIntFromMap(configMap, "max_dependencies_per_file"),
		MaxDependencyDepth:     f.getIntFromMap(configMap, "max_dependency_depth"),
		DetectCycles:           f.getBoolFromMap(configMap, "detect_cycles"),
		ForbiddenPatterns:      f.getStringSliceFromMap(configMap, "forbidden_patterns"),
	}
	return rulesgraph.NewPropertyEnforcementRule(f.importGraph, enforcementConfig)
}

func (f *RuleFactory) createPathBasedLayerRules() []rules.Rule {
	ruleConfig, ok := f.config.Rules["path-based-layers"]
	if !ok || !f.isRuleEnabled("path-based-layers") {
		return nil
	}

	configMap, ok := ruleConfig.(map[string]interface{})
	if !ok {
		return nil
	}

	layersConfig, ok := configMap["layers"].([]interface{})
	if !ok {
		return nil
	}

	pathLayers := f.parsePathLayers(layersConfig)
	if len(pathLayers) == 0 {
		return nil
	}

	return []rules.Rule{rulesgraph.NewPathBasedLayerRule(pathLayers)}
}

func (f *RuleFactory) parsePathLayers(layersConfig []interface{}) []rulesgraph.PathLayer {
	var pathLayers []rulesgraph.PathLayer

	for _, layerInterface := range layersConfig {
		layerMap, ok := layerInterface.(map[string]interface{})
		if !ok {
			continue
		}

		layer := rulesgraph.PathLayer{
			Name:           f.getStringFromMap(layerMap, "name"),
			Patterns:       f.getStringSliceFromMap(layerMap, "patterns"),
			CanDependOn:    f.getStringSliceFromMap(layerMap, "canDependOn"),
			ForbiddenPaths: f.getStringSliceFromMap(layerMap, "forbiddenPaths"),
		}
		pathLayers = append(pathLayers, layer)
	}

	return pathLayers
}

func (f *RuleFactory) createTestValidationRules() []rules.Rule {
	var rulesList []rules.Rule

	if testAdj, ok := f.config.Rules["test-adjacency"]; ok {
		if f.isRuleEnabled("test-adjacency") {
			if adjMap, ok := testAdj.(map[string]interface{}); ok {
				pattern := f.getStringFromMap(adjMap, "pattern")
				testDir := f.getStringFromMap(adjMap, "test-dir")
				filePatterns := f.getStringSliceFromMap(adjMap, "file-patterns")
				exemptions := f.getStringSliceFromMap(adjMap, "exemptions")

				if pattern != "" && len(filePatterns) > 0 {
					rulesList = append(rulesList, rulestest.NewTestAdjacencyRule(pattern, testDir, filePatterns, exemptions))
				}
			}
		}
	}

	if testLoc, ok := f.config.Rules["test-location"]; ok {
		if f.isRuleEnabled("test-location") {
			if locMap, ok := testLoc.(map[string]interface{}); ok {
				integrationDir := f.getStringFromMap(locMap, "integration-test-dir")
				allowAdjacent := f.getBoolFromMap(locMap, "allow-adjacent")
				filePatterns := f.getStringSliceFromMap(locMap, "file-patterns")
				exemptions := f.getStringSliceFromMap(locMap, "exemptions")

				rulesList = append(rulesList, rulestest.NewTestLocationRule(integrationDir, allowAdjacent, filePatterns, exemptions))
			}
		}
	}

	return rulesList
}

func (f *RuleFactory) createContentRules() []rules.Rule {
	fileContent, ok := f.config.Rules["file-content"]
	if !ok || !f.isRuleEnabled("file-content") {
		return nil
	}

	contentMap, ok := fileContent.(map[string]interface{})
	if !ok {
		return nil
	}

	templates := f.getStringMapFromMap(contentMap, "templates")
	templateDir := f.getStringFromMap(contentMap, "template-dir")

	if len(templates) == 0 || templateDir == "" {
		return nil
	}

	rootPath := f.rootDir
	if rootPath == "" {
		rootPath = "."
	}

	return []rules.Rule{rules.NewFileContentRule(templates, templateDir, rootPath)}
}

func (f *RuleFactory) createCIRules() []rules.Rule {
	var rulesList []rules.Rule

	if workflowConfig, ok := f.config.Rules["github-workflows"]; ok {
		if f.isRuleEnabled("github-workflows") {
			if configMap, ok := workflowConfig.(map[string]interface{}); ok {
				rule := rulesci.NewGitHubWorkflowsRule(rulesci.GitHubWorkflowsRule{
					RequireTests:           f.getBoolFromMap(configMap, "require-tests"),
					RequireSecurity:        f.getBoolFromMap(configMap, "require-security"),
					RequireQuality:         f.getBoolFromMap(configMap, "require-quality"),
					RequireLogCommits:      f.getBoolFromMap(configMap, "require-log-commits"),
					RequireRepomixArtifact: f.getBoolFromMap(configMap, "require-repomix-artifact"),
					RequiredJobs:           f.getStringSliceFromMap(configMap, "required-jobs"),
					RequiredTriggers:       f.getStringSliceFromMap(configMap, "required-triggers"),
					AllowMissing:           f.getStringSliceFromMap(configMap, "allow-missing"),
				})
				rulesList = append(rulesList, rule)
			}
		}
	}

	return rulesList
}

func (f *RuleFactory) createLinterConfigRules() []rules.Rule {
	var rulesList []rules.Rule
	if rule := f.createLinterConfigRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	if rule := f.createAPISpecRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	if rule := f.createContractFrameworkRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	if rule := f.createSpecADRRule(); rule != nil {
		rulesList = append(rulesList, rule)
	}
	return rulesList
}

func (f *RuleFactory) createLinterConfigRule() rules.Rule {
	config, ok := f.config.Rules["linter-config"]
	if !ok || !f.isRuleEnabled("linter-config") {
		return nil
	}
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil
	}
	return rulesci.NewLinterConfigRule(rulesci.LinterConfigRule{
		RequirePython:     f.getBoolFromMap(configMap, "require-python"),
		RequireTypeScript: f.getBoolFromMap(configMap, "require-typescript"),
		RequireGo:         f.getBoolFromMap(configMap, "require-go"),
		RequireHTML:       f.getBoolFromMap(configMap, "require-html"),
		RequireCSS:        f.getBoolFromMap(configMap, "require-css"),
		RequireSQL:        f.getBoolFromMap(configMap, "require-sql"),
		RequireRust:       f.getBoolFromMap(configMap, "require-rust"),
		CustomLinters:     f.getStringSliceFromMap(configMap, "custom-linters"),
	})
}

func (f *RuleFactory) createAPISpecRule() rules.Rule {
	config, ok := f.config.Rules["api-spec"]
	if !ok || !f.isRuleEnabled("api-spec") {
		return nil
	}
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil
	}
	return rulesci.NewOpenAPIAsyncAPIRule(rulesci.OpenAPIAsyncAPIRule{
		RequireOpenAPI:  f.getBoolFromMap(configMap, "require-openapi"),
		RequireAsyncAPI: f.getBoolFromMap(configMap, "require-asyncapi"),
		CustomSpecs:     f.getStringSliceFromMap(configMap, "custom-specs"),
	})
}

func (f *RuleFactory) createContractFrameworkRule() rules.Rule {
	config, ok := f.config.Rules["contract-framework"]
	if !ok || !f.isRuleEnabled("contract-framework") {
		return nil
	}
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil
	}
	return rulesci.NewContractFrameworkRule(rulesci.ContractFrameworkRule{
		RequirePython:     f.getBoolFromMap(configMap, "require-python"),
		RequireRust:       f.getBoolFromMap(configMap, "require-rust"),
		RequireTypeScript: f.getBoolFromMap(configMap, "require-typescript"),
		RequireGo:         f.getBoolFromMap(configMap, "require-go"),
		RequireJava:       f.getBoolFromMap(configMap, "require-java"),
		RequireCSharp:     f.getBoolFromMap(configMap, "require-csharp"),
		RequireCPlusPlus:  f.getBoolFromMap(configMap, "require-cplusplus"),
		CustomFrameworks:  f.getStringSliceFromMap(configMap, "custom-frameworks"),
	})
}

func (f *RuleFactory) createSpecADRRule() rules.Rule {
	config, ok := f.config.Rules["spec-adr-enforcement"]
	if !ok || !f.isRuleEnabled("spec-adr-enforcement") {
		return nil
	}
	configMap, ok := config.(map[string]interface{})
	if !ok {
		return nil
	}
	return rulesci.NewSpecADRRule(rulesci.SpecADRRule{
		RequireSpecFolder:   f.getBoolFromMap(configMap, "require-spec-folder"),
		RequireADRFolder:    f.getBoolFromMap(configMap, "require-adr-folder"),
		EnforceSpecTemplate: f.getBoolFromMap(configMap, "enforce-spec-template"),
		EnforceADRTemplate:  f.getBoolFromMap(configMap, "enforce-adr-template"),
		SpecFolderPaths:     f.getStringSliceFromMap(configMap, "spec-folder-paths"),
		ADRFolderPaths:      f.getStringSliceFromMap(configMap, "adr-folder-paths"),
		SpecFilePatterns:    f.getStringSliceFromMap(configMap, "spec-file-patterns"),
		ADRFilePatterns:     f.getStringSliceFromMap(configMap, "adr-file-patterns"),
	})
}

func (f *RuleFactory) isRuleEnabled(ruleName string) bool {
	if f.config == nil || f.config.Rules == nil {
		return false
	}

	value, exists := f.config.Rules[ruleName]
	if !exists {
		return false
	}

	switch v := value.(type) {
	case int:
		return v != 0
	case bool:
		return v
	}

	return true
}

func (f *RuleFactory) extractStringSlice(patterns []interface{}) []string {
	var result []string
	for _, p := range patterns {
		if str, ok := p.(string); ok {
			result = append(result, str)
		}
	}
	return result
}

func (f *RuleFactory) getStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func (f *RuleFactory) getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key].(bool); ok {
		return val
	}
	return false
}

func (f *RuleFactory) getIntFromMap(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	if val, ok := m[key].(float64); ok {
		return int(val)
	}
	return 0
}

func (f *RuleFactory) getFloatFromMap(m map[string]interface{}, key string) float64 {
	if val, ok := m[key].(float64); ok {
		return val
	}
	if val, ok := m[key].(int); ok {
		return float64(val)
	}
	return 0
}

func (f *RuleFactory) getStringSliceFromMap(m map[string]interface{}, key string) []string {
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

func (f *RuleFactory) getStringMapFromMap(m map[string]interface{}, key string) map[string]string {
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
