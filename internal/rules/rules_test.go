package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestMaxDepthRule(t *testing.T) {
	// Arrange
	rule := NewMaxDepthRule(3)

	files := []walker.FileInfo{
		{Path: "a.txt", Depth: 1},
		{Path: "b/c.txt", Depth: 2},
		{Path: "d/e/f.txt", Depth: 3},
		{Path: "g/h/i/j.txt", Depth: 4}, // Violates
	}

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if violations[0].Path != "g/h/i/j.txt" {
		t.Errorf("Expected violation for g/h/i/j.txt, got %s", violations[0].Path)
	}

	if violations[0].Rule != "max-depth" {
		t.Errorf("Expected rule 'max-depth', got %s", violations[0].Rule)
	}
}

func TestMaxDepthRule_NoViolations(t *testing.T) {
	rule := NewMaxDepthRule(5)

	files := []walker.FileInfo{
		{Path: "a.txt", Depth: 1},
		{Path: "b/c.txt", Depth: 2},
		{Path: "d/e/f.txt", Depth: 3},
	}

	violations := rule.Check(files, nil)

	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
	}
}

func TestMaxFilesRule(t *testing.T) {
	rule := NewMaxFilesRule(2)

	files := []walker.FileInfo{
		{Path: "dir1/file1.go", ParentPath: "dir1", IsDir: false},
		{Path: "dir1/file2.go", ParentPath: "dir1", IsDir: false},
		{Path: "dir2/file1.go", ParentPath: "dir2", IsDir: false},
		{Path: "dir2/file2.go", ParentPath: "dir2", IsDir: false},
		{Path: "dir2/file3.go", ParentPath: "dir2", IsDir: false}, // Violates (3 > 2)
		{Path: "dir2/file_test.go", ParentPath: "dir2", IsDir: false}, // Test file - doesn't count
		{Path: "dir3/file1.go", ParentPath: "dir3", IsDir: false},
	}

	dirs := map[string]*walker.DirInfo{
		"dir1": {Path: "dir1"},
		"dir2": {Path: "dir2"}, // Has 3 non-test files
		"dir3": {Path: "dir3"},
	}

	violations := rule.Check(files, dirs)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if len(violations) > 0 && violations[0].Path != "dir2" {
		t.Errorf("Expected violation for dir2, got %s", violations[0].Path)
	}
}

func TestMaxSubdirsRule(t *testing.T) {
	rule := NewMaxSubdirsRule(3)

	dirs := map[string]*walker.DirInfo{
		"dir1": {Path: "dir1", SubdirCount: 2},
		"dir2": {Path: "dir2", SubdirCount: 4}, // Violates
		"dir3": {Path: "dir3", SubdirCount: 3},
	}

	violations := rule.Check(nil, dirs)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if violations[0].Path != "dir2" {
		t.Errorf("Expected violation for dir2, got %s", violations[0].Path)
	}
}

func TestNamingConventionRule_CamelCase(t *testing.T) {
	rule := NewNamingConventionRule(map[string]string{
		"*.ts": "camelCase",
	})

	files := []walker.FileInfo{
		{Path: "validName.ts"},
		{Path: "InvalidName.ts"},     // Violates (PascalCase)
		{Path: "another-invalid.ts"}, // Violates (kebab-case)
	}

	violations := rule.Check(files, nil)

	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
	}
}

func TestNamingConventionRule_PascalCase(t *testing.T) {
	rule := NewNamingConventionRule(map[string]string{
		"*.tsx": "PascalCase",
	})

	files := []walker.FileInfo{
		{Path: "ValidComponent.tsx"},
		{Path: "invalidComponent.tsx"}, // Violates (camelCase)
	}

	violations := rule.Check(files, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if violations[0].Path != "invalidComponent.tsx" {
		t.Errorf("Expected violation for invalidComponent.tsx, got %s", violations[0].Path)
	}
}

func TestNamingConventionRule_KebabCase(t *testing.T) {
	rule := NewNamingConventionRule(map[string]string{
		"*.css": "kebab-case",
	})

	files := []walker.FileInfo{
		{Path: "valid-name.css"},
		{Path: "InvalidName.css"},  // Violates (has uppercase)
		{Path: "invalid_name.css"}, // Violates (snake_case)
	}

	violations := rule.Check(files, nil)

	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
	}
}

func TestDisallowedPatternsRule(t *testing.T) {
	rule := NewDisallowedPatternsRule([]string{
		"src/utils/**",
		".DS_Store",
	})

	files := []walker.FileInfo{
		{Path: "src/components/Button.tsx"},
		{Path: "src/utils/helper.ts"}, // Violates
		{Path: ".DS_Store"},           // Violates
	}

	violations := rule.Check(files, nil)

	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
	}
}

func TestDisallowedPatternsRule_WithNegation(t *testing.T) {
	// Test negation patterns (!) to allow exceptions
	rule := NewDisallowedPatternsRule([]string{
		"internal/**/*.md",   // Disallow all .md files in internal/
		"!**/README.md",      // Except README.md files
	})

	files := []walker.FileInfo{
		{Path: "internal/config/README.md"},     // Should be allowed (negation)
		{Path: "internal/linter/README.md"},     // Should be allowed (negation)
		{Path: "internal/config/DESIGN.md"},     // Should violate
		{Path: "internal/linter/ARCHITECTURE.md"}, // Should violate
		{Path: "README.md"},                      // Not in internal/, so no violation
		{Path: "docs/GUIDE.md"},                  // Not in internal/, so no violation
	}

	violations := rule.Check(files, nil)

	// Should have 2 violations: DESIGN.md and ARCHITECTURE.md
	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}

	// Check that the right files violated
	violatedPaths := make(map[string]bool)
	for _, v := range violations {
		violatedPaths[v.Path] = true
	}

	if !violatedPaths["internal/config/DESIGN.md"] {
		t.Errorf("Expected violation for internal/config/DESIGN.md")
	}

	if !violatedPaths["internal/linter/ARCHITECTURE.md"] {
		t.Errorf("Expected violation for internal/linter/ARCHITECTURE.md")
	}
}

func TestFileExistenceRule(t *testing.T) {
	rule := NewFileExistenceRule(map[string]string{
		"index.ts": "exists:1",
	})

	files := []walker.FileInfo{
		{Path: "dir1/index.ts", ParentPath: "dir1"},
		{Path: "dir2/other.ts", ParentPath: "dir2"},
		{Path: "dir3/index.ts", ParentPath: "dir3"},
	}

	dirs := map[string]*walker.DirInfo{
		"dir1": {Path: "dir1"},
		"dir2": {Path: "dir2"}, // Violates - no index.ts
		"dir3": {Path: "dir3"},
	}

	violations := rule.Check(files, dirs)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if violations[0].Path != "dir2" {
		t.Errorf("Expected violation for dir2, got %s", violations[0].Path)
	}
}
