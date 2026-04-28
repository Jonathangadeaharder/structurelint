package structure

import (
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

func init() {
	rules.Register("max-depth", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxDepthRule(max), nil
		}
		return nil, fmt.Errorf("missing 'max' parameter")
	})

	rules.Register("max-files-in-dir", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxFilesRule(max), nil
		}
		return nil, fmt.Errorf("missing 'max' parameter")
	})

	rules.Register("max-subdirs", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxSubdirsRule(max), nil
		}
		return nil, fmt.Errorf("missing 'max' parameter")
	})

	rules.Register("file-existence", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if reqs, ok := ctx.GetStringMap(""); ok {
			return NewFileExistenceRule(reqs), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	rules.Register("regex-match", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if patterns, ok := ctx.GetStringMap(""); ok {
			return NewRegexMatchRule(patterns), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	rules.Register("disallowed-patterns", func(ctx *rules.RuleContext) (rules.Rule, error) {
		var patterns []string
		if list, ok := ctx.Config["patterns"].([]interface{}); ok {
			for _, item := range list {
				if s, ok := item.(string); ok {
					patterns = append(patterns, s)
				}
			}
		} else if list, ok := ctx.Config[""].([]interface{}); ok {
			for _, item := range list {
				if s, ok := item.(string); ok {
					patterns = append(patterns, s)
				}
			}
		}
		if len(patterns) > 0 {
			return NewDisallowedPatternsRule(patterns), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	rules.Register("naming-convention", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if patterns, ok := ctx.GetStringMap(""); ok {
			return NewNamingConventionRule(patterns), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	rules.Register("uniqueness-constraints", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if constraints, ok := ctx.GetStringMap(""); ok {
			return NewUniquenessConstraintsRule(constraints), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})
}
