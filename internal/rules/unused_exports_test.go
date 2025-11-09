package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/parser"
)

func TestUnusedExportsRule_DetectsUnusedExports(t *testing.T) {
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/index.ts",
			"src/unused.ts",
		},
		Dependencies: map[string][]string{
			"src/index.ts": {}, // Doesn't import anything
		},
		Exports: map[string][]parser.Export{
			"src/unused.ts": {
				{
					SourceFile: "src/unused.ts",
					Names:      []string{"unusedFunction", "UnusedClass"},
					IsDefault:  false,
				},
			},
		},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if len(violations) > 0 && violations[0].Path != "src/unused.ts" {
		t.Errorf("Expected violation for src/unused.ts, got %s", violations[0].Path)
	}
}

func TestUnusedExportsRule_NoViolationWhenImported(t *testing.T) {
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/index.ts",
			"src/utils.ts",
		},
		Dependencies: map[string][]string{
			"src/index.ts": {"src/utils.ts"}, // Imports utils
		},
		Exports: map[string][]parser.Export{
			"src/utils.ts": {
				{
					SourceFile: "src/utils.ts",
					Names:      []string{"helperFunction"},
					IsDefault:  false,
				},
			},
		},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	if len(violations) != 0 {
		t.Errorf("Expected no violations when file is imported, got %d", len(violations))
	}
}

func TestUnusedExportsRule_NoViolationWithoutExports(t *testing.T) {
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/index.ts",
		},
		Dependencies: map[string][]string{},
		Exports:      map[string][]parser.Export{},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for files without exports, got %d", len(violations))
	}
}
