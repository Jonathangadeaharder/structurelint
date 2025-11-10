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

// Mutation-killing tests for boundary conditions

func TestUnusedExportsRule_EmptyExportNames(t *testing.T) {
	// Tests line 67:24 - CONDITIONALS_BOUNDARY: len(exportNames) > 0
	// Edge case: Export with empty Names array
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/empty.ts",
		},
		Dependencies: map[string][]string{},
		Exports: map[string][]parser.Export{
			"src/empty.ts": {
				{
					SourceFile: "src/empty.ts",
					Names:      []string{}, // Empty names array
					IsDefault:  false,
				},
			},
		},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	// Empty export names should not cause violation (nothing to report)
	if len(violations) != 0 {
		t.Errorf("Expected no violations for empty export names, got %d", len(violations))
	}
}

func TestUnusedExportsRule_FormatNames_EmptyArray(t *testing.T) {
	// Tests line 101:16 - CONDITIONALS_NEGATION: len(names) == 0
	// This tests the formatNames function with empty array
	importGraph := &graph.ImportGraph{
		AllFiles: []string{"src/file.ts"},
		Dependencies: map[string][]string{},
		Exports: map[string][]parser.Export{
			"src/file.ts": {
				{
					SourceFile: "src/file.ts",
					Names:      []string{}, // Will trigger formatNames with empty array
					IsDefault:  false,
				},
			},
		},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	// Should handle empty names gracefully (no violation)
	if len(violations) != 0 {
		t.Errorf("Expected no violations with empty names, got %d", len(violations))
	}
}

func TestUnusedExportsRule_FormatNames_SingleName(t *testing.T) {
	// Tests line 104:16 - CONDITIONALS_NEGATION: len(names) == 1
	// Edge case: Exactly one export name
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/single.ts",
		},
		Dependencies: map[string][]string{},
		Exports: map[string][]parser.Export{
			"src/single.ts": {
				{
					SourceFile: "src/single.ts",
					Names:      []string{"singleExport"}, // Exactly 1 name
					IsDefault:  false,
				},
			},
		},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation for single unused export, got %d", len(violations))
	}

	// Check that message contains the single export name
	if len(violations) > 0 && !contains(violations[0].Message, "singleExport") {
		t.Errorf("Expected violation message to contain 'singleExport', got: %s", violations[0].Message)
	}
}

func TestUnusedExportsRule_FormatNames_TwoNames(t *testing.T) {
	// Tests boundary between single name and multiple names
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/two.ts",
		},
		Dependencies: map[string][]string{},
		Exports: map[string][]parser.Export{
			"src/two.ts": {
				{
					SourceFile: "src/two.ts",
					Names:      []string{"exportOne", "exportTwo"}, // Exactly 2 names
					IsDefault:  false,
				},
			},
		},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation for two unused exports, got %d", len(violations))
	}

	// Check that message contains both names
	if len(violations) > 0 {
		msg := violations[0].Message
		if !contains(msg, "exportOne") || !contains(msg, "exportTwo") {
			t.Errorf("Expected violation message to contain both export names, got: %s", msg)
		}
	}
}

func TestUnusedExportsRule_FormatNames_ThreeNames(t *testing.T) {
	// Tests line 107:16 - CONDITIONALS_BOUNDARY: len(names) <= 3
	// Edge case: Exactly 3 names (boundary)
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/three.ts",
		},
		Dependencies: map[string][]string{},
		Exports: map[string][]parser.Export{
			"src/three.ts": {
				{
					SourceFile: "src/three.ts",
					Names:      []string{"one", "two", "three"}, // Exactly 3 names
					IsDefault:  false,
				},
			},
		},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation for three unused exports, got %d", len(violations))
	}

	// With 3 names, should list all three
	if len(violations) > 0 {
		msg := violations[0].Message
		if !contains(msg, "one") || !contains(msg, "two") || !contains(msg, "three") {
			t.Errorf("Expected violation message to contain all three names, got: %s", msg)
		}
	}
}

func TestUnusedExportsRule_FormatNames_FourNames(t *testing.T) {
	// Tests line 107:16 - CONDITIONALS_BOUNDARY: len(names) <= 3
	// Edge case: More than 3 names (should show first 3 + "and N more")
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/many.ts",
		},
		Dependencies: map[string][]string{},
		Exports: map[string][]parser.Export{
			"src/many.ts": {
				{
					SourceFile: "src/many.ts",
					Names:      []string{"one", "two", "three", "four"}, // 4 names
					IsDefault:  false,
				},
			},
		},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation for four unused exports, got %d", len(violations))
	}

	// With 4 names, should show first 3 and "and 1 more"
	if len(violations) > 0 {
		msg := violations[0].Message
		if !contains(msg, "one") || !contains(msg, "two") || !contains(msg, "three") {
			t.Errorf("Expected violation message to contain first three names, got: %s", msg)
		}
		if !contains(msg, "more") {
			t.Errorf("Expected violation message to indicate more exports, got: %s", msg)
		}
	}
}

func TestUnusedExportsRule_FormatNames_ManyNames(t *testing.T) {
	// Tests line 107:16 - well beyond the boundary
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/lots.ts",
		},
		Dependencies: map[string][]string{},
		Exports: map[string][]parser.Export{
			"src/lots.ts": {
				{
					SourceFile: "src/lots.ts",
					Names:      []string{"a", "b", "c", "d", "e", "f", "g", "h"}, // 8 names
					IsDefault:  false,
				},
			},
		},
	}

	rule := NewUnusedExportsRule(importGraph)
	violations := rule.Check(nil, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation for many unused exports, got %d", len(violations))
	}

	// With 8 names, should show "a, b, c and 5 more"
	if len(violations) > 0 {
		msg := violations[0].Message
		if !contains(msg, "5 more") {
			t.Errorf("Expected violation message to show '5 more', got: %s", msg)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
