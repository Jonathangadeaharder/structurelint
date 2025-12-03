package rules

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/graph"
)

func TestOrphanedFilesRule_GivenOrphanedFile_WhenChecking_ThenDetectsOrphan(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/index.ts",
			"src/used.ts",
			"src/orphaned.ts",
		},
		IncomingRefs: map[string]int{
			"src/index.ts":    0, // Entrypoint, should not be flagged
			"src/used.ts":     2, // Used by others
			"src/orphaned.ts": 0, // Orphaned!
		},
	}
	rule := NewOrphanedFilesRule(importGraph, []string{"src/index.ts"})

	// Act
	violations := rule.Check(nil, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if len(violations) > 0 && violations[0].Path != "src/orphaned.ts" {
		t.Errorf("Expected violation for src/orphaned.ts, got %s", violations[0].Path)
	}
}

func TestOrphanedFilesRule_GivenEntrypoints_WhenChecking_ThenRespectsEntrypoints(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/main.ts",
			"src/app.ts",
		},
		IncomingRefs: map[string]int{
			"src/main.ts": 0,
			"src/app.ts":  0,
		},
	}
	rule := NewOrphanedFilesRule(importGraph, []string{"src/main.ts", "src/app.ts"})

	// Act
	violations := rule.Check(nil, nil)

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations for entrypoints, got %d", len(violations))
	}
}

func TestOrphanedFilesRule_GivenConfigFiles_WhenChecking_ThenExcludesConfigFiles(t *testing.T) {
	// Arrange
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

	// Act
	violations := rule.Check(nil, nil)

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations for config/doc files, got %d", len(violations))
	}
}

func TestOrphanedFilesRule_GivenTestFiles_WhenChecking_ThenExcludesTestFiles(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"src/app.test.ts",
			"src/app_test.go",
			"tests/integration.spec.ts",
		},
		IncomingRefs: map[string]int{},
	}
	rule := NewOrphanedFilesRule(importGraph, []string{})

	// Act
	violations := rule.Check(nil, nil)

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations for test files, got %d", len(violations))
	}
}

func TestOrphanedFilesRule_GivenEntryPointPatterns_WhenChecking_ThenRespectsPatterns(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"scripts/deploy.py",
			"scripts/migrate.py",
			"cli/commands/init.ts",
			"src/orphaned.ts",
		},
		IncomingRefs: map[string]int{
			"scripts/deploy.py":   0,
			"scripts/migrate.py":  0,
			"cli/commands/init.ts": 0,
			"src/orphaned.ts":     0,
		},
	}
	rule := NewOrphanedFilesRule(importGraph, []string{}).
		WithEntryPointPatterns([]string{
			"scripts/**",
			"cli/**/*.ts",
		})

	// Act
	violations := rule.Check(nil, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (orphaned.ts), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Path)
		}
	}

	if len(violations) > 0 && violations[0].Path != "src/orphaned.ts" {
		t.Errorf("Expected violation for src/orphaned.ts, got %s", violations[0].Path)
	}
}

func TestOrphanedFilesRule_GivenWildcardPattern_WhenChecking_ThenMatchesCorrectly(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		AllFiles: []string{
			"backend/main.py",
			"frontend/main.ts",
			"src/lib.py",
		},
		IncomingRefs: map[string]int{
			"backend/main.py":  0,
			"frontend/main.ts": 0,
			"src/lib.py":       0,
		},
	}
	rule := NewOrphanedFilesRule(importGraph, []string{}).
		WithEntryPointPatterns([]string{"**/main.*"})

	// Act
	violations := rule.Check(nil, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (lib.py), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Path)
		}
	}

	if len(violations) > 0 && violations[0].Path != "src/lib.py" {
		t.Errorf("Expected violation for src/lib.py, got %s", violations[0].Path)
	}
}
