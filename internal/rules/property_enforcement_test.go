package rules

import (
	"strings"
	"testing"

	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/walker"
)

func TestPropertyEnforcementRule_DetectCycles(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {"a.go"}, // Creates cycle: A -> B -> C -> A
		},
	}
	config := PropertyEnforcementConfig{
		DetectCycles: true,
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	if len(violations) == 0 {
		t.Error("expected cycle detection to find violations, got none")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "cyclic") || strings.Contains(v.Message, "cycle") {
			found = true
			if v.Context == "" {
				t.Error("expected context to show the cycle path")
			}
		}
	}

	if !found {
		t.Error("expected at least one cyclic dependency violation")
	}
}

func TestPropertyEnforcementRule_NoCycles(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {}, // No cycle: A -> B -> C
		},
	}
	config := PropertyEnforcementConfig{
		DetectCycles: true,
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	cycleViolations := []Violation{}
	for _, v := range violations {
		if strings.Contains(v.Message, "cyclic") || strings.Contains(v.Message, "cycle") {
			cycleViolations = append(cycleViolations, v)
		}
	}

	if len(cycleViolations) > 0 {
		t.Errorf("expected no cycle violations, got %d: %v", len(cycleViolations), cycleViolations)
	}
}

func TestPropertyEnforcementRule_MaxDependenciesPerFile(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go", "c.go", "d.go", "e.go", "f.go"}, // 5 dependencies
			"b.go": {},
			"c.go": {},
		},
	}
	config := PropertyEnforcementConfig{
		MaxDependenciesPerFile: 3, // Limit to 3
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	if len(violations) == 0 {
		t.Error("expected max dependencies violation, got none")
	}

	found := false
	for _, v := range violations {
		if v.Path == "a.go" && strings.Contains(v.Message, "too many dependencies") {
			found = true
			if !strings.Contains(v.Actual, "5") {
				t.Errorf("expected actual to show 5 dependencies, got: %s", v.Actual)
			}
		}
	}

	if !found {
		t.Error("expected violation for a.go having too many dependencies")
	}
}

func TestPropertyEnforcementRule_MaxDependenciesPerFile_UnderLimit(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go", "c.go"}, // 2 dependencies
			"b.go": {},
		},
	}
	config := PropertyEnforcementConfig{
		MaxDependenciesPerFile: 5, // Limit to 5
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	for _, v := range violations {
		if strings.Contains(v.Message, "too many dependencies") {
			t.Errorf("unexpected max dependencies violation: %v", v)
		}
	}
}

func TestPropertyEnforcementRule_MaxDependencyDepth(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {"d.go"},
			"d.go": {"e.go"},
			"e.go": {}, // Deep chain: A -> B -> C -> D -> E (depth 4)
		},
	}
	config := PropertyEnforcementConfig{
		MaxDependencyDepth: 2, // Limit to depth 2
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	if len(violations) == 0 {
		t.Error("expected dependency depth violation, got none")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "dependency chain too deep") {
			found = true
		}
	}

	if !found {
		t.Error("expected at least one dependency depth violation")
	}
}

func TestPropertyEnforcementRule_MaxDependencyDepth_UnderLimit(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {}, // Shallow chain: A -> B (depth 1)
		},
	}
	config := PropertyEnforcementConfig{
		MaxDependencyDepth: 3, // Limit to depth 3
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	for _, v := range violations {
		if strings.Contains(v.Message, "dependency chain too deep") {
			t.Errorf("unexpected depth violation: %v", v)
		}
	}
}

func TestPropertyEnforcementRule_ForbiddenPatterns(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"internal/service.go": {"external/api.go"}, // Forbidden dependency
			"cmd/main.go":         {"internal/service.go"},
		},
	}
	config := PropertyEnforcementConfig{
		ForbiddenPatterns: []string{
			"internal/** -> external/**",
		},
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	if len(violations) == 0 {
		t.Error("expected forbidden pattern violation, got none")
	}

	found := false
	for _, v := range violations {
		if v.Path == "internal/service.go" && strings.Contains(v.Message, "forbidden dependency") {
			found = true
		}
	}

	if !found {
		t.Error("expected violation for internal/service.go importing external/api.go")
	}
}

func TestPropertyEnforcementRule_ForbiddenPatterns_NoViolation(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"cmd/main.go":         {"internal/service.go"},
			"internal/service.go": {"internal/db.go"}, // Allowed dependency
		},
	}
	config := PropertyEnforcementConfig{
		ForbiddenPatterns: []string{
			"internal/** -> external/**",
		},
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	for _, v := range violations {
		if strings.Contains(v.Message, "forbidden dependency") {
			t.Errorf("unexpected forbidden pattern violation: %v", v)
		}
	}
}

func TestPropertyEnforcementRule_MultipleChecks(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go", "c.go", "d.go", "e.go"}, // Too many dependencies (4)
			"b.go": {"a.go"},                         // Creates a cycle with a.go
			"c.go": {},
		},
	}
	config := PropertyEnforcementConfig{
		DetectCycles:           true,
		MaxDependenciesPerFile: 2, // Limit to 2
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	if len(violations) == 0 {
		t.Error("expected multiple violations, got none")
	}

	hasCycle := false
	hasMaxDeps := false

	for _, v := range violations {
		if strings.Contains(v.Message, "cyclic") || strings.Contains(v.Message, "cycle") {
			hasCycle = true
		}
		if strings.Contains(v.Message, "too many dependencies") {
			hasMaxDeps = true
		}
	}

	if !hasCycle {
		t.Error("expected cycle violation")
	}
	if !hasMaxDeps {
		t.Error("expected max dependencies violation")
	}
}

func TestPropertyEnforcementRule_NilGraph(t *testing.T) {
	// Arrange
	config := PropertyEnforcementConfig{
		DetectCycles: true,
	}
	rule := NewPropertyEnforcementRule(nil, config) // Nil graph

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	if len(violations) != 0 {
		t.Errorf("expected no violations with nil graph, got %d", len(violations))
	}
}

func TestPropertyEnforcementRule_ComplexCycle(t *testing.T) {
	// Arrange
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {"d.go"},
			"d.go": {"b.go"}, // Creates cycle: B -> C -> D -> B
			"e.go": {"f.go"}, // Unrelated chain
			"f.go": {},
		},
	}
	config := PropertyEnforcementConfig{
		DetectCycles: true,
	}
	rule := NewPropertyEnforcementRule(importGraph, config)

	// Act
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Assert
	if len(violations) == 0 {
		t.Error("expected cycle detection to find violations in complex cycle")
	}

	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "cyclic") || strings.Contains(v.Message, "cycle") {
			if v.Path == "b.go" || v.Path == "c.go" || v.Path == "d.go" {
				found = true
			}
		}
	}

	if !found {
		t.Error("expected cycle violation for files b.go, c.go, or d.go")
	}
}

func TestMatchPattern(t *testing.T) {
	// Arrange
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"internal/service.go", "internal/**", true},
		{"internal/db/repo.go", "internal/**", true},
		{"external/api.go", "internal/**", false},
		{"cmd/main.go", "cmd/**", true},
		{"test.go", "*.go", true},
		{"test.py", "*.go", false},
	}

	// Act & Assert
	for _, tt := range tests {
		got := matchPattern(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}
