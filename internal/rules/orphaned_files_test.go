package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/graph"
)

func TestOrphanedFilesRule_DetectsOrphans(t *testing.T) {
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/index.ts",
			"src/used.ts",
			"src/orphaned.ts",
		},
		IncomingRefs: map[string]int{
			"src/index.ts": 0,      // Entrypoint, should not be flagged
			"src/used.ts":  2,      // Used by others
			"src/orphaned.ts": 0,   // Orphaned!
		},
	}

	rule := NewOrphanedFilesRule(importGraph, []string{"src/index.ts"})
	violations := rule.Check(nil, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if len(violations) > 0 && violations[0].Path != "src/orphaned.ts" {
		t.Errorf("Expected violation for src/orphaned.ts, got %s", violations[0].Path)
	}
}

func TestOrphanedFilesRule_RespectsEntrypoints(t *testing.T) {
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/main.ts",
			"src/app.ts",
		},
		IncomingRefs: map[string]int{
			"src/main.ts": 0,
			"src/app.ts": 0,
		},
	}

	rule := NewOrphanedFilesRule(importGraph, []string{"src/main.ts", "src/app.ts"})
	violations := rule.Check(nil, nil)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for entrypoints, got %d", len(violations))
	}
}

func TestOrphanedFilesRule_ExcludesConfigFiles(t *testing.T) {
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"package.json",
			"tsconfig.json",
			".structurelint.yml",
			"README.md",
		},
		IncomingRefs: map[string]int{},
	}

	rule := NewOrphanedFilesRule(importGraph, []string{})
	violations := rule.Check(nil, nil)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for config/doc files, got %d", len(violations))
	}
}

func TestOrphanedFilesRule_ExcludesTestFiles(t *testing.T) {
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/app.test.ts",
			"src/app_test.go",
			"tests/integration.spec.ts",
		},
		IncomingRefs: map[string]int{},
	}

	rule := NewOrphanedFilesRule(importGraph, []string{})
	violations := rule.Check(nil, nil)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for test files, got %d", len(violations))
	}
}
