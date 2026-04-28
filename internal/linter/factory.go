package linter

import (
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

func (f *RuleFactory) CreateRules(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Rule {
	f.files = files
	f.dirs = dirs

	f.checkBreakingChanges()

	var rulesList []rules.Rule
	rulesList = append(rulesList, f.createRegistryRules()...)
	rulesList = append(rulesList, f.createComplexityRules()...)
	rulesList = append(rulesList, f.createGraphDependentRules()...)
	rulesList = append(rulesList, f.createPathBasedLayerRules()...)
	rulesList = append(rulesList, f.createTestValidationRules()...)
	rulesList = append(rulesList, f.createContentRules()...)
	rulesList = append(rulesList, f.createCIRules()...)
	rulesList = append(rulesList, f.createLinterConfigRules()...)

	return rulesList
}

func (f *RuleFactory) checkBreakingChanges() {
	if _, ok := f.config.Rules["max-cyclomatic-complexity"]; ok {
		panic("BREAKING CHANGE: 'max-cyclomatic-complexity' rule has been removed.\n" +
			"Use 'max-cognitive-complexity' instead - it's scientifically superior (r=0.54 vs cyclomatic's weak correlation).\n" +
			"See: https://github.com/Jonathangadeaharder/structurelint#phase-5-evidence-based-metrics")
	}
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

	if cognitiveComplexity, ok := f.config.Rules["max-cognitive-complexity"]; ok {
		if f.isRuleEnabled("max-cognitive-complexity") {
			if complexityMap, ok := cognitiveComplexity.(map[string]interface{}); ok {
				max := f.getIntFromMap(complexityMap, "max")
				if max > 0 {
					testMax := f.getIntFromMap(complexityMap, "test-max")
					filePatterns := f.getStringSliceFromMap(complexityMap, "file-patterns")

					rule := quality.NewMaxCognitiveComplexityRule(max, filePatterns)
					if testMax > 0 {
						rule = rule.WithTestMax(testMax)
					}
					rulesList = append(rulesList, rule)
				}
			}
		}
	}

	if halsteadEffort, ok := f.config.Rules["max-halstead-effort"]; ok {
		if f.isRuleEnabled("max-halstead-effort") {
			if effortMap, ok := halsteadEffort.(map[string]interface{}); ok {
				max := f.getFloatFromMap(effortMap, "max")
				if max > 0 {
					filePatterns := f.getStringSliceFromMap(effortMap, "file-patterns")
					rulesList = append(rulesList, quality.NewMaxHalsteadEffortRule(max, filePatterns))
				}
			}
		}
	}

	return rulesList
}

func (f *RuleFactory) createGraphDependentRules() []rules.Rule {
	if f.importGraph == nil {
		return nil
	}

	var rulesList []rules.Rule

	if _, ok := f.config.Rules["enforce-layer-boundaries"]; ok {
		if f.isRuleEnabled("enforce-layer-boundaries") && len(f.config.Layers) > 0 {
			rulesList = append(rulesList, rulesgraph.NewLayerBoundariesRule(f.importGraph))
		}
	}

	if orphanedConfig, ok := f.config.Rules["disallow-orphaned-files"]; ok {
		if f.isRuleEnabled("disallow-orphaned-files") {
			rule := rulesgraph.NewOrphanedFilesRule(f.importGraph, f.config.Entrypoints)

			if configMap, ok := orphanedConfig.(map[string]interface{}); ok {
				if patterns, ok := configMap["entry-point-patterns"].([]interface{}); ok {
					entryPointPatterns := f.extractStringSlice(patterns)
					if len(entryPointPatterns) > 0 {
						rule = rule.WithEntryPointPatterns(entryPointPatterns)
					}
				}
			}

			rulesList = append(rulesList, rule)
		}
	}

	if _, ok := f.config.Rules["disallow-unused-exports"]; ok {
		if f.isRuleEnabled("disallow-unused-exports") {
			rulesList = append(rulesList, rulesgraph.NewUnusedExportsRule(f.importGraph))
		}
	}

	if propertyConfig, ok := f.config.Rules["property-enforcement"]; ok {
		if f.isRuleEnabled("property-enforcement") {
			if configMap, ok := propertyConfig.(map[string]interface{}); ok {
				config := rulesgraph.PropertyEnforcementConfig{
					MaxDependenciesPerFile: f.getIntFromMap(configMap, "max_dependencies_per_file"),
					MaxDependencyDepth:     f.getIntFromMap(configMap, "max_dependency_depth"),
					DetectCycles:           f.getBoolFromMap(configMap, "detect_cycles"),
					ForbiddenPatterns:      f.getStringSliceFromMap(configMap, "forbidden_patterns"),
				}
				rulesList = append(rulesList, rulesgraph.NewPropertyEnforcementRule(f.importGraph, config))
			}
		}
	}

	return rulesList
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

	if linterConfig, ok := f.config.Rules["linter-config"]; ok {
		if f.isRuleEnabled("linter-config") {
			if configMap, ok := linterConfig.(map[string]interface{}); ok {
				rule := rulesci.NewLinterConfigRule(rulesci.LinterConfigRule{
					RequirePython:     f.getBoolFromMap(configMap, "require-python"),
					RequireTypeScript: f.getBoolFromMap(configMap, "require-typescript"),
					RequireGo:         f.getBoolFromMap(configMap, "require-go"),
					RequireHTML:       f.getBoolFromMap(configMap, "require-html"),
					RequireCSS:        f.getBoolFromMap(configMap, "require-css"),
					RequireSQL:        f.getBoolFromMap(configMap, "require-sql"),
					RequireRust:       f.getBoolFromMap(configMap, "require-rust"),
					CustomLinters:     f.getStringSliceFromMap(configMap, "custom-linters"),
				})
				rulesList = append(rulesList, rule)
			}
		}
	}

	if apiSpecConfig, ok := f.config.Rules["api-spec"]; ok {
		if f.isRuleEnabled("api-spec") {
			if configMap, ok := apiSpecConfig.(map[string]interface{}); ok {
				rule := rulesci.NewOpenAPIAsyncAPIRule(rulesci.OpenAPIAsyncAPIRule{
					RequireOpenAPI:  f.getBoolFromMap(configMap, "require-openapi"),
					RequireAsyncAPI: f.getBoolFromMap(configMap, "require-asyncapi"),
					CustomSpecs:     f.getStringSliceFromMap(configMap, "custom-specs"),
				})
				rulesList = append(rulesList, rule)
			}
		}
	}

	if contractConfig, ok := f.config.Rules["contract-framework"]; ok {
		if f.isRuleEnabled("contract-framework") {
			if configMap, ok := contractConfig.(map[string]interface{}); ok {
				rule := rulesci.NewContractFrameworkRule(rulesci.ContractFrameworkRule{
					RequirePython:     f.getBoolFromMap(configMap, "require-python"),
					RequireRust:       f.getBoolFromMap(configMap, "require-rust"),
					RequireTypeScript: f.getBoolFromMap(configMap, "require-typescript"),
					RequireGo:         f.getBoolFromMap(configMap, "require-go"),
					RequireJava:       f.getBoolFromMap(configMap, "require-java"),
					RequireCSharp:     f.getBoolFromMap(configMap, "require-csharp"),
					RequireCPlusPlus:  f.getBoolFromMap(configMap, "require-cplusplus"),
					CustomFrameworks:  f.getStringSliceFromMap(configMap, "custom-frameworks"),
				})
				rulesList = append(rulesList, rule)
			}
		}
	}

	if specADRConfig, ok := f.config.Rules["spec-adr-enforcement"]; ok {
		if f.isRuleEnabled("spec-adr-enforcement") {
			if configMap, ok := specADRConfig.(map[string]interface{}); ok {
				rule := rulesci.NewSpecADRRule(rulesci.SpecADRRule{
					RequireSpecFolder:   f.getBoolFromMap(configMap, "require-spec-folder"),
					RequireADRFolder:    f.getBoolFromMap(configMap, "require-adr-folder"),
					EnforceSpecTemplate: f.getBoolFromMap(configMap, "enforce-spec-template"),
					EnforceADRTemplate:  f.getBoolFromMap(configMap, "enforce-adr-template"),
					SpecFolderPaths:     f.getStringSliceFromMap(configMap, "spec-folder-paths"),
					ADRFolderPaths:      f.getStringSliceFromMap(configMap, "adr-folder-paths"),
					SpecFilePatterns:    f.getStringSliceFromMap(configMap, "spec-file-patterns"),
					ADRFilePatterns:     f.getStringSliceFromMap(configMap, "adr-file-patterns"),
				})
				rulesList = append(rulesList, rule)
			}
		}
	}

	return rulesList
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

func (f *RuleFactory) getRuleConfig(ruleName string) (interface{}, bool) {
	if f.config == nil || f.config.Rules == nil {
		return nil, false
	}

	value, exists := f.config.Rules[ruleName]
	if !exists {
		return nil, false
	}

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

func (f *RuleFactory) getIntConfig(ruleName, key string) (int, bool) {
	config, ok := f.getRuleConfig(ruleName)
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

func (f *RuleFactory) getStringMapConfig(ruleName string) (map[string]string, bool) {
	config, ok := f.getRuleConfig(ruleName)
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

func (f *RuleFactory) getStringSliceConfig(ruleName string) ([]string, bool) {
	config, ok := f.getRuleConfig(ruleName)
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
