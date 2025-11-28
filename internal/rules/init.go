package rules

import (
	"fmt"
)

func init() {
	// Simple Int Rules
	Register("max-depth", func(ctx *RuleContext) (Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxDepthRule(max), nil
		}
		return nil, fmt.Errorf("missing 'max' parameter")
	})

	Register("max-files-in-dir", func(ctx *RuleContext) (Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxFilesRule(max), nil
		}
		return nil, fmt.Errorf("missing 'max' parameter")
	})

	Register("max-subdirs", func(ctx *RuleContext) (Rule, error) {
		if max, ok := ctx.GetInt("max"); ok {
			return NewMaxSubdirsRule(max), nil
		}
		return nil, fmt.Errorf("missing 'max' parameter")
	})

	// Map Rules
	Register("file-existence", func(ctx *RuleContext) (Rule, error) {
		if reqs, ok := ctx.GetStringMap(""); ok {
			return NewFileExistenceRule(reqs), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	Register("regex-match", func(ctx *RuleContext) (Rule, error) {
		if patterns, ok := ctx.GetStringMap(""); ok {
			return NewRegexMatchRule(patterns), nil
		}
		return nil, fmt.Errorf("invalid configuration")
	})

	Register("disallowed-patterns", func(ctx *RuleContext) (Rule, error) {
		// This one is usually a list of strings
		var patterns []string
		if list, ok := ctx.Config["patterns"].([]interface{}); ok {
			for _, item := range list {
				if s, ok := item.(string); ok {
					patterns = append(patterns, s)
				}
			}
		} else if list, ok := ctx.Config[""].([]interface{}); ok {
			// Handle case where config is just the list
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

	// Graph Dependent Rules
	Register("enforce-layer-boundaries", func(ctx *RuleContext) (Rule, error) {
		if ctx.ImportGraph == nil {
			return nil, fmt.Errorf("import graph required")
		}
		return NewLayerBoundariesRule(ctx.ImportGraph), nil
	})

	Register("disallow-orphaned-files", func(ctx *RuleContext) (Rule, error) {
		if ctx.ImportGraph == nil {
			return nil, fmt.Errorf("import graph required")
		}
		// TODO: Pass entrypoints from context
		return NewOrphanedFilesRule(ctx.ImportGraph, []string{}), nil
	})

	Register("disallow-unused-exports", func(ctx *RuleContext) (Rule, error) {
		if ctx.ImportGraph == nil {
			return nil, fmt.Errorf("import graph required")
		}
		return NewUnusedExportsRule(ctx.ImportGraph), nil
	})
}
