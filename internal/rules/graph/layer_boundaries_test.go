package graph

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
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
			"src/domain/user.ts": {"src/presentation/userComponent.ts"},
		},
		FileLayers: map[string]*config.Layer{
			"src/domain/user.ts":                &layers[0],
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
		{Name: "app", Path: "src/app/**", DependsOn: []string{"*"}},
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

func TestResolveDependencyToFile_ExactMatch(t *testing.T) {
	rule := NewLayerBoundariesRule(&graph.ImportGraph{})

	files := []walker.FileInfo{
		{Path: "src/domain/user.ts"},
		{Path: "src/application/service.ts"},
	}

	resolved := rule.resolveDependencyToFile("src/domain/user.ts", files)
	if resolved != "src/domain/user.ts" {
		t.Errorf("Expected 'src/domain/user.ts', got '%s'", resolved)
	}
}

func TestResolveDependencyToFile_WithExtension(t *testing.T) {
	rule := NewLayerBoundariesRule(&graph.ImportGraph{})

	files := []walker.FileInfo{
		{Path: "src/domain/user.ts"},
		{Path: "src/domain/product.tsx"},
		{Path: "src/services/auth.js"},
		{Path: "src/utils/helper.py"},
		{Path: "src/models/model.go"},
	}

	resolved := rule.resolveDependencyToFile("src/domain/user", files)
	if resolved != "src/domain/user.ts" {
		t.Errorf("Expected 'src/domain/user.ts', got '%s'", resolved)
	}

	resolved = rule.resolveDependencyToFile("src/domain/product", files)
	if resolved != "src/domain/product.tsx" {
		t.Errorf("Expected 'src/domain/product.tsx', got '%s'", resolved)
	}

	resolved = rule.resolveDependencyToFile("src/services/auth", files)
	if resolved != "src/services/auth.js" {
		t.Errorf("Expected 'src/services/auth.js', got '%s'", resolved)
	}

	resolved = rule.resolveDependencyToFile("src/utils/helper", files)
	if resolved != "src/utils/helper.py" {
		t.Errorf("Expected 'src/utils/helper.py', got '%s'", resolved)
	}

	resolved = rule.resolveDependencyToFile("src/models/model", files)
	if resolved != "src/models/model.go" {
		t.Errorf("Expected 'src/models/model.go', got '%s'", resolved)
	}
}

func TestResolveDependencyToFile_IndexFile(t *testing.T) {
	rule := NewLayerBoundariesRule(&graph.ImportGraph{})

	files := []walker.FileInfo{
		{Path: "src/components/index.ts"},
		{Path: "src/utils/index.tsx"},
		{Path: "src/helpers/index.js"},
		{Path: "src/services/index.jsx"},
	}

	resolved := rule.resolveDependencyToFile("src/components", files)
	if resolved != "src/components/index.ts" {
		t.Errorf("Expected 'src/components/index.ts', got '%s'", resolved)
	}

	resolved = rule.resolveDependencyToFile("src/utils", files)
	if resolved != "src/utils/index.tsx" {
		t.Errorf("Expected 'src/utils/index.tsx', got '%s'", resolved)
	}

	resolved = rule.resolveDependencyToFile("src/helpers", files)
	if resolved != "src/helpers/index.js" {
		t.Errorf("Expected 'src/helpers/index.js', got '%s'", resolved)
	}

	resolved = rule.resolveDependencyToFile("src/services", files)
	if resolved != "src/services/index.jsx" {
		t.Errorf("Expected 'src/services/index.jsx', got '%s'", resolved)
	}
}

func TestResolveDependencyToFile_GoPackage(t *testing.T) {
	rule := NewLayerBoundariesRule(&graph.ImportGraph{})

	files := []walker.FileInfo{
		{Path: "pkg/domain/user.go"},
		{Path: "pkg/domain/product.go"},
		{Path: "pkg/service/auth.go"},
	}

	resolved := rule.resolveDependencyToFile("pkg/domain", files)
	if resolved != "pkg/domain/user.go" {
		t.Errorf("Expected 'pkg/domain/user.go', got '%s'", resolved)
	}

	resolved = rule.resolveDependencyToFile("pkg/service", files)
	if resolved != "pkg/service/auth.go" {
		t.Errorf("Expected 'pkg/service/auth.go', got '%s'", resolved)
	}
}

func TestResolveDependencyToFile_NoMatch(t *testing.T) {
	rule := NewLayerBoundariesRule(&graph.ImportGraph{})

	files := []walker.FileInfo{
		{Path: "src/domain/user.ts"},
	}

	resolved := rule.resolveDependencyToFile("src/nonexistent/file", files)
	if resolved != "" {
		t.Errorf("Expected empty string for no match, got '%s'", resolved)
	}
}

func TestResolveDependencyToFile_PriorityOrder(t *testing.T) {
	rule := NewLayerBoundariesRule(&graph.ImportGraph{})

	files := []walker.FileInfo{
		{Path: "src/user"},
		{Path: "src/user.ts"},
		{Path: "src/user/index.ts"},
	}

	resolved := rule.resolveDependencyToFile("src/user", files)
	if resolved != "src/user" {
		t.Errorf("Expected exact match 'src/user', got '%s'", resolved)
	}
}

func TestLayerBoundariesRule_Name(t *testing.T) {
	rule := NewLayerBoundariesRule(&graph.ImportGraph{})

	if rule.Name() != "enforce-layer-boundaries" {
		t.Errorf("Expected rule name 'enforce-layer-boundaries', got '%s'", rule.Name())
	}
}
