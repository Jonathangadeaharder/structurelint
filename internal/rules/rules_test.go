package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestMaxDepthRule(t *testing.T) {
	rule := NewMaxDepthRule(3)

	files := []walker.FileInfo{
		{Path: "a.txt", Depth: 1},
		{Path: "b/c.txt", Depth: 2},
		{Path: "d/e/f.txt", Depth: 3},
		{Path: "g/h/i/j.txt", Depth: 4}, // Violates
	}

	violations := rule.Check(files, nil)

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

	dirs := map[string]*walker.DirInfo{
		"dir1": {Path: "dir1", FileCount: 2},
		"dir2": {Path: "dir2", FileCount: 3}, // Violates
		"dir3": {Path: "dir3", FileCount: 1},
	}

	violations := rule.Check(nil, dirs)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if violations[0].Path != "dir2" {
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
