package structure

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

const errMissingMaxParam = "missing 'max' parameter"
const errInvalidConfig = "invalid configuration"

func init() {
	rules.Register("max-depth", func(ctx *rules.RuleContext) (rules.Rule, error) {
		max, ok := ctx.GetInt("max")
		if !ok {
			return nil, errors.New(errMissingMaxParam)
		}
		overrides := parseMaxDepthOverrides(ctx.Config["overrides"])
		if len(overrides) == 0 {
			return NewMaxDepthRule(max), nil
		}
		return NewMaxDepthRuleWithOverrides(max, overrides), nil
	})

	rules.Register("max-files-in-dir", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxFilesRule(max), nil
		}
		return nil, errors.New(errMissingMaxParam)
	})

	rules.Register("max-subdirs", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxSubdirsRule(max), nil
		}
		return nil, errors.New(errMissingMaxParam)
	})

	rules.Register("file-existence", func(ctx *rules.RuleContext) (rules.Rule, error) {
		reqs, ok := ctx.GetStringMap("")
		if !ok {
			return nil, fmt.Errorf("invalid configuration: file-existence expects map of pattern -> 'exists:N' / 'exists:N-M'")
		}
		if errs := ValidateFileExistenceConfig(reqs); len(errs) > 0 {
			return nil, fmt.Errorf("file-existence config errors: %s", strings.Join(errs, "; "))
		}
		return NewFileExistenceRule(reqs), nil
	})

	rules.Register("regex-match", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if patterns, ok := ctx.GetStringMap(""); ok {
			return NewRegexMatchRule(patterns), nil
		}
		return nil, errors.New(errInvalidConfig)
	})

	rules.Register("disallowed-patterns", func(ctx *rules.RuleContext) (rules.Rule, error) {
		patterns := extractStringList(ctx.Config, "patterns")
		if len(patterns) == 0 {
			patterns = extractStringList(ctx.Config, "")
		}
		if len(patterns) > 0 {
			return NewDisallowedPatternsRule(patterns), nil
		}
		return nil, errors.New(errInvalidConfig)
	})

	rules.Register("naming-convention", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if patterns, ok := ctx.GetStringMap(""); ok {
			return NewNamingConventionRule(patterns), nil
		}
		return nil, errors.New(errInvalidConfig)
	})

	rules.Register("uniqueness-constraints", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if constraints, ok := ctx.GetStringMap(""); ok {
			return NewUniquenessConstraintsRule(constraints), nil
		}
		return nil, errors.New(errInvalidConfig)
	})

	rules.Register("case-conflicts", func(_ *rules.RuleContext) (rules.Rule, error) {
		return NewCaseConflictsRule(), nil
	})

	rules.Register("disallow-empty-dirs", func(_ *rules.RuleContext) (rules.Rule, error) {
		return NewEmptyDirsRule(), nil
	})

	rules.Register("disallow-symlinks", func(_ *rules.RuleContext) (rules.Rule, error) {
		return NewSymlinksRule(), nil
	})

	rules.Register("disallow-deep-relative-imports", func(ctx *rules.RuleContext) (rules.Rule, error) {
		max := 3
		if v, ok := ctx.GetInt("max-parents"); ok && v > 0 {
			max = v
		}
		return NewDeepRelativeImportsRule(max), nil
	})

	rules.Register("spec-adr", func(ctx *rules.RuleContext) (rules.Rule, error) {
		return ParseSpecADRRule(ctx.Config)
	})

	rules.Register("spec-adr-enforcement", func(ctx *rules.RuleContext) (rules.Rule, error) {
		rule, err := ParseSpecADRRule(ctx.Config)
		if err != nil {
			return nil, err
		}
		rule.ruleName = "spec-adr-enforcement"
		return rule, nil
	})
}

// parseMaxDepthOverrides accepts:
//
//	overrides:
//	  "src/routes/**": 8
//	  "tests/**": 6
//
// or a list form:
//
//	overrides:
//	  - pattern: "src/routes/**"
//	    max: 8
// extractStringList extracts a list of strings from ctx.Config[key]
func extractStringList(config map[string]interface{}, key string) []string {
	list, ok := config[key].([]interface{})
	if !ok {
		return nil
	}
	var result []string
	for _, item := range list {
		if s, ok := item.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

func parseMaxDepthOverrides(raw interface{}) []MaxDepthOverride {
	if raw == nil {
		return nil
	}
	var overrides []MaxDepthOverride
	switch v := raw.(type) {
	case map[string]interface{}:
		for pattern, val := range v {
			if max := toInt(val); max > 0 {
				overrides = append(overrides, MaxDepthOverride{Pattern: pattern, Max: max})
			}
		}
	case []interface{}:
		for _, item := range v {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			pattern, _ := m["pattern"].(string)
			max := toInt(m["max"])
			if pattern != "" && max > 0 {
				overrides = append(overrides, MaxDepthOverride{Pattern: pattern, Max: max})
			}
		}
	}
	return overrides
}

func toInt(v interface{}) int {
	switch x := v.(type) {
	case int:
		return x
	case float64:
		return int(x)
	}
	return 0
}
