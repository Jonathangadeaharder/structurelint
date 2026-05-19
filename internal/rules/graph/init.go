package graph

import (
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

const errImportGraphRequired = "import graph required"

func init() {
	rules.Register("enforce-layer-boundaries", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if ctx.ImportGraph == nil {
			return nil, fmt.Errorf(errImportGraphRequired)
		}
		return NewLayerBoundariesRule(ctx.ImportGraph), nil
	})

	rules.Register("disallow-orphaned-files", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if ctx.ImportGraph == nil {
			return nil, fmt.Errorf(errImportGraphRequired)
		}
		// TODO: Pass entrypoints from context
		return NewOrphanedFilesRule(ctx.ImportGraph, []string{}), nil
	})

	rules.Register("disallow-import-cycles", func(ctx *rules.RuleContext) (rules.Rule, error) {
		if ctx.ImportGraph == nil {
			return nil, fmt.Errorf(errImportGraphRequired)
		}
		return NewImportCyclesRule(ctx.ImportGraph), nil
	})
}
