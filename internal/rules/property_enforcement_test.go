package rules

import (
	"strings"
	"testing"

	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/walker"
)

func TestPropertyEnforcementRule_DetectCycles(t *testing.T) {
	// Create a graph with a cycle: A -> B -> C -> A
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {"a.go"},
		},
	}

	config := PropertyEnforcementConfig{
		DetectCycles: true,
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

	if len(violations) == 0 {
		t.Error("expected cycle detection to find violations, got none")
	}

	// Check that the violation mentions a cycle
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
	// Create a graph without cycles: A -> B -> C
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {},
		},
	}

	config := PropertyEnforcementConfig{
		DetectCycles: true,
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Filter for cycle-related violations only
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
	// Create a graph where one file has too many dependencies
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go", "c.go", "d.go", "e.go", "f.go"},
			"b.go": {},
			"c.go": {},
		},
	}

	config := PropertyEnforcementConfig{
		MaxDependenciesPerFile: 3,
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

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
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go", "c.go"},
			"b.go": {},
		},
	}

	config := PropertyEnforcementConfig{
		MaxDependenciesPerFile: 5,
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Should have no violations for max dependencies
	for _, v := range violations {
		if strings.Contains(v.Message, "too many dependencies") {
			t.Errorf("unexpected max dependencies violation: %v", v)
		}
	}
}

func TestPropertyEnforcementRule_MaxDependencyDepth(t *testing.T) {
	// Create a deep dependency chain: A -> B -> C -> D -> E (depth 4)
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {"d.go"},
			"d.go": {"e.go"},
			"e.go": {},
		},
	}

	config := PropertyEnforcementConfig{
		MaxDependencyDepth: 2,
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

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
	// Create a shallow dependency chain: A -> B (depth 1)
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {},
		},
	}

	config := PropertyEnforcementConfig{
		MaxDependencyDepth: 3,
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Should have no depth violations
	for _, v := range violations {
		if strings.Contains(v.Message, "dependency chain too deep") {
			t.Errorf("unexpected depth violation: %v", v)
		}
	}
}

func TestPropertyEnforcementRule_ForbiddenPatterns(t *testing.T) {
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"internal/service.go": {"external/api.go"},
			"cmd/main.go":         {"internal/service.go"},
		},
	}

	config := PropertyEnforcementConfig{
		ForbiddenPatterns: []string{
			"internal/** -> external/**",
		},
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

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
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"cmd/main.go":         {"internal/service.go"},
			"internal/service.go": {"internal/db.go"},
		},
	}

	config := PropertyEnforcementConfig{
		ForbiddenPatterns: []string{
			"internal/** -> external/**",
		},
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

	// Should have no forbidden pattern violations
	for _, v := range violations {
		if strings.Contains(v.Message, "forbidden dependency") {
			t.Errorf("unexpected forbidden pattern violation: %v", v)
		}
	}
}

func TestPropertyEnforcementRule_MultipleChecks(t *testing.T) {
	// Create a graph with multiple violations
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go", "c.go", "d.go", "e.go"}, // Too many dependencies
			"b.go": {"a.go"},                         // Creates a cycle
			"c.go": {},
		},
	}

	config := PropertyEnforcementConfig{
		DetectCycles:           true,
		MaxDependenciesPerFile: 2,
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

	if len(violations) == 0 {
		t.Error("expected multiple violations, got none")
	}

	// Should have both cycle and max dependencies violations
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
	config := PropertyEnforcementConfig{
		DetectCycles: true,
	}

	rule := NewPropertyEnforcementRule(nil, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

	if len(violations) != 0 {
		t.Errorf("expected no violations with nil graph, got %d", len(violations))
	}
}

func TestPropertyEnforcementRule_ComplexCycle(t *testing.T) {
	// Create a more complex cycle: A -> B -> C -> D -> B
	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"a.go": {"b.go"},
			"b.go": {"c.go"},
			"c.go": {"d.go"},
			"d.go": {"b.go"}, // Cycle back to b
			"e.go": {"f.go"}, // Unrelated files
			"f.go": {},
		},
	}

	config := PropertyEnforcementConfig{
		DetectCycles: true,
	}

	rule := NewPropertyEnforcementRule(importGraph, config)
	violations := rule.Check([]walker.FileInfo{}, nil)

	if len(violations) == 0 {
		t.Error("expected cycle detection to find violations in complex cycle")
	}

	// Check that cycle involves b, c, or d
	found := false
	for _, v := range violations {
		if strings.Contains(v.Message, "cyclic") || strings.Contains(v.Message, "cycle") {
			// The violation should be for one of the files in the cycle
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

	for _, tt := range tests {
		got := matchPattern(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("matchPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}
