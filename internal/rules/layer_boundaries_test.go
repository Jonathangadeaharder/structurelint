package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/config"
	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/walker"
)

func TestLayerBoundariesRule_ValidDependencies(t *testing.T) {
	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		{Name: "app", Path: "src/app/**", DependsOn: []string{"domain"}},
	}

	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/app/service.ts": {"src/domain/user.ts"},
		},
		FileLayers: map[string]*config.Layer{
			"src/app/service.ts":  &layers[1],
			"src/domain/user.ts":  &layers[0],
		},
		Layers: layers,
	}

	rule := NewLayerBoundariesRule(importGraph)
	files := []walker.FileInfo{
		{Path: "src/app/service.ts"},
		{Path: "src/domain/user.ts"},
	}

	violations := rule.Check(files, nil)

	if len(violations) != 0 {
		t.Errorf("Expected no violations for valid dependency, got %d", len(violations))
	}
}

func TestLayerBoundariesRule_InvalidDependency(t *testing.T) {
	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		{Name: "presentation", Path: "src/presentation/**", DependsOn: []string{}},
	}

	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/domain/user.ts": {"src/presentation/userComponent.ts"}, // Invalid!
		},
		FileLayers: map[string]*config.Layer{
			"src/domain/user.ts":               &layers[0],
			"src/presentation/userComponent.ts": &layers[1],
		},
		Layers: layers,
	}

	rule := NewLayerBoundariesRule(importGraph)
	files := []walker.FileInfo{
		{Path: "src/domain/user.ts"},
		{Path: "src/presentation/userComponent.ts"},
	}

	violations := rule.Check(files, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if len(violations) > 0 && violations[0].Path != "src/domain/user.ts" {
		t.Errorf("Expected violation for src/domain/user.ts, got %s", violations[0].Path)
	}
}

func TestLayerBoundariesRule_WildcardDependency(t *testing.T) {
	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		{Name: "app", Path: "src/app/**", DependsOn: []string{"*"}}, // Can depend on anything
	}

	importGraph := &graph.ImportGraph{
		Dependencies: map[string][]string{
			"src/app/service.ts": {"src/domain/user.ts"},
		},
		FileLayers: map[string]*config.Layer{
			"src/app/service.ts": &layers[1],
			"src/domain/user.ts": &layers[0],
		},
		Layers: layers,
	}

	rule := NewLayerBoundariesRule(importGraph)
	files := []walker.FileInfo{
		{Path: "src/app/service.ts"},
		{Path: "src/domain/user.ts"},
	}

	violations := rule.Check(files, nil)

	if len(violations) != 0 {
		t.Errorf("Expected no violations with wildcard dependency, got %d", len(violations))
	}
}
