package graph

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestNewBuilder(t *testing.T) {
	// Arrange
	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
	}

	// Act
	builder := NewBuilder("/test/path", layers)

	// Assert
	if builder == nil {
		t.Fatal("NewBuilder returned nil")
	}

	if builder.rootPath != "/test/path" {
		t.Errorf("Expected rootPath '/test/path', got '%s'", builder.rootPath)
	}

	if len(builder.layers) != 1 {
		t.Errorf("Expected 1 layer, got %d", len(builder.layers))
	}
}

func TestBuild_EmptyFileList(t *testing.T) {
	builder := NewBuilder("/test", []config.Layer{})
	graph, err := builder.Build([]walker.FileInfo{})

	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	if graph == nil {
		t.Fatal("Graph is nil")
	}

	if len(graph.AllFiles) != 0 {
		t.Errorf("Expected 0 files, got %d", len(graph.AllFiles))
	}
}

func TestBuild_WithLayers(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "graph-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	domainDir := filepath.Join(tmpDir, "src", "domain")
	if err := os.MkdirAll(domainDir, 0755); err != nil {
		t.Fatal(err)
	}

	userFile := filepath.Join(domainDir, "user.ts")
	if err := os.WriteFile(userFile, []byte("export class User {}"), 0644); err != nil {
		t.Fatal(err)
	}

	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
	}

	builder := NewBuilder(tmpDir, layers)

	files := []walker.FileInfo{
		{Path: "src/domain/user.ts", AbsPath: userFile, IsDir: false},
	}

	graph, err := builder.Build(files)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// Check that file is in AllFiles
	if len(graph.AllFiles) != 1 {
		t.Errorf("Expected 1 file, got %d", len(graph.AllFiles))
	}

	// Check that file is assigned to domain layer
	layer := graph.GetLayerForFile("src/domain/user.ts")
	if layer == nil {
		t.Fatal("File not assigned to any layer")
	}

	if layer.Name != "domain" {
		t.Errorf("Expected layer 'domain', got '%s'", layer.Name)
	}
}

func TestMatchesLayerPath_SimplePrefix(t *testing.T) {
	builder := NewBuilder("/test", []config.Layer{})

	// Simple prefix match
	if !builder.matchesLayerPath("src/domain/user.ts", "src/domain") {
		t.Error("Expected src/domain/user.ts to match src/domain")
	}

	if builder.matchesLayerPath("src/application/service.ts", "src/domain") {
		t.Error("Expected src/application/service.ts NOT to match src/domain")
	}
}

func TestMatchesLayerPath_Glob(t *testing.T) {
	builder := NewBuilder("/test", []config.Layer{})

	// Glob pattern with **
	if !builder.matchesLayerPath("src/domain/user.ts", "src/domain/**") {
		t.Error("Expected src/domain/user.ts to match src/domain/**")
	}

	if !builder.matchesLayerPath("src/domain/models/user.ts", "src/domain/**") {
		t.Error("Expected src/domain/models/user.ts to match src/domain/**")
	}

	if builder.matchesLayerPath("src/application/service.ts", "src/domain/**") {
		t.Error("Expected src/application/service.ts NOT to match src/domain/**")
	}
}

func TestMatchesLayerPath_GlobWithSuffix(t *testing.T) {
	builder := NewBuilder("/test", []config.Layer{})

	// Glob with suffix
	if !builder.matchesLayerPath("src/domain/user.model.ts", "src/**/*.model.ts") {
		t.Error("Expected src/domain/user.model.ts to match src/**/*.model.ts")
	}
}

func TestFindLayerForFile(t *testing.T) {
	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
		{Name: "infrastructure", Path: "src/infrastructure/**", DependsOn: []string{"domain"}},
	}

	builder := NewBuilder("/test", layers)

	// Test domain layer
	layer := builder.findLayerForFile("src/domain/user.ts")
	if layer == nil || layer.Name != "domain" {
		t.Error("Expected to find domain layer")
	}

	// Test application layer
	layer = builder.findLayerForFile("src/application/userService.ts")
	if layer == nil || layer.Name != "application" {
		t.Error("Expected to find application layer")
	}

	// Test no layer
	layer = builder.findLayerForFile("src/other/file.ts")
	if layer != nil {
		t.Errorf("Expected no layer, got %s", layer.Name)
	}
}

func TestCanLayerDependOn_AllowedDependency(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "presentation", Path: "src/presentation/**", DependsOn: []string{"application", "domain"}},
			{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	presentation := graph.FindLayerByName("presentation")
	application := graph.FindLayerByName("application")
	domain := graph.FindLayerByName("domain")

	// presentation can depend on application
	if !graph.CanLayerDependOn(presentation, application) {
		t.Error("Expected presentation to be able to depend on application")
	}

	// presentation can depend on domain
	if !graph.CanLayerDependOn(presentation, domain) {
		t.Error("Expected presentation to be able to depend on domain")
	}

	// application can depend on domain
	if !graph.CanLayerDependOn(application, domain) {
		t.Error("Expected application to be able to depend on domain")
	}
}

func TestCanLayerDependOn_ForbiddenDependency(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "presentation", Path: "src/presentation/**", DependsOn: []string{"application"}},
			{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	presentation := graph.FindLayerByName("presentation")
	application := graph.FindLayerByName("application")
	domain := graph.FindLayerByName("domain")

	// domain CANNOT depend on application
	if graph.CanLayerDependOn(domain, application) {
		t.Error("Expected domain NOT to be able to depend on application")
	}

	// domain CANNOT depend on presentation
	if graph.CanLayerDependOn(domain, presentation) {
		t.Error("Expected domain NOT to be able to depend on presentation")
	}

	// application CANNOT depend on presentation
	if graph.CanLayerDependOn(application, presentation) {
		t.Error("Expected application NOT to be able to depend on presentation")
	}
}

func TestCanLayerDependOn_SameLayer(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	domain := graph.FindLayerByName("domain")

	// Layer can depend on itself
	if !graph.CanLayerDependOn(domain, domain) {
		t.Error("Expected layer to be able to depend on itself")
	}
}

func TestCanLayerDependOn_Wildcard(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "presentation", Path: "src/presentation/**", DependsOn: []string{"*"}},
			{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		},
	}

	presentation := graph.FindLayerByName("presentation")
	application := graph.FindLayerByName("application")
	domain := graph.FindLayerByName("domain")

	// presentation can depend on anything (wildcard)
	if !graph.CanLayerDependOn(presentation, application) {
		t.Error("Expected presentation to depend on application (wildcard)")
	}

	if !graph.CanLayerDependOn(presentation, domain) {
		t.Error("Expected presentation to depend on domain (wildcard)")
	}
}

func TestCanLayerDependOn_NilLayers(t *testing.T) {
	graph := &ImportGraph{}

	// Nil layers should return true (allow dependency)
	if !graph.CanLayerDependOn(nil, nil) {
		t.Error("Expected nil layers to allow dependency")
	}

	layer := &config.Layer{Name: "test", Path: "test/**", DependsOn: []string{}}

	if !graph.CanLayerDependOn(layer, nil) {
		t.Error("Expected layer->nil to allow dependency")
	}

	if !graph.CanLayerDependOn(nil, layer) {
		t.Error("Expected nil->layer to allow dependency")
	}
}

func TestGetLayerForFile(t *testing.T) {
	domain := &config.Layer{Name: "domain", Path: "src/domain/**", DependsOn: []string{}}

	graph := &ImportGraph{
		FileLayers: map[string]*config.Layer{
			"src/domain/user.ts": domain,
		},
	}

	layer := graph.GetLayerForFile("src/domain/user.ts")
	if layer == nil {
		t.Fatal("Expected to find layer")
	}

	if layer.Name != "domain" {
		t.Errorf("Expected domain layer, got %s", layer.Name)
	}

	// Test non-existent file
	layer = graph.GetLayerForFile("src/other/file.ts")
	if layer != nil {
		t.Error("Expected nil for non-existent file")
	}
}

func TestGetDependencies(t *testing.T) {
	graph := &ImportGraph{
		Dependencies: map[string][]string{
			"src/presentation/component.ts": {"src/application/service", "src/domain/user"},
		},
	}

	deps := graph.GetDependencies("src/presentation/component.ts")
	if len(deps) != 2 {
		t.Errorf("Expected 2 dependencies, got %d", len(deps))
	}

	// Test non-existent file
	deps = graph.GetDependencies("src/other/file.ts")
	if len(deps) != 0 {
		t.Errorf("Expected 0 dependencies, got %d", len(deps))
	}
}

func TestFindLayerByName(t *testing.T) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
			{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
		},
	}

	// Find existing layer
	layer := graph.FindLayerByName("domain")
	if layer == nil {
		t.Fatal("Expected to find domain layer")
	}
	if layer.Name != "domain" {
		t.Errorf("Expected domain layer, got %s", layer.Name)
	}

	// Find non-existent layer
	layer = graph.FindLayerByName("nonexistent")
	if layer != nil {
		t.Error("Expected nil for non-existent layer")
	}
}

func TestBuild_IncomingReferences(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "graph-ref-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	srcDir := filepath.Join(tmpDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatal(err)
	}

	userFile := filepath.Join(srcDir, "user.ts")
	serviceFile := filepath.Join(srcDir, "service.ts")

	// user.ts is imported by service.ts
	if err := os.WriteFile(userFile, []byte("export class User {}"), 0644); err != nil {
		t.Fatal(err)
	}

	// service.ts imports user.ts
	if err := os.WriteFile(serviceFile, []byte("import { User } from './user'"), 0644); err != nil {
		t.Fatal(err)
	}

	builder := NewBuilder(tmpDir, []config.Layer{})

	files := []walker.FileInfo{
		{Path: filepath.Join("src", "user.ts"), AbsPath: userFile, IsDir: false},
		{Path: filepath.Join("src", "service.ts"), AbsPath: serviceFile, IsDir: false},
	}

	graph, err := builder.Build(files)
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	// service.ts should have imported user
	deps := graph.GetDependencies(filepath.Join("src", "service.ts"))
	if len(deps) == 0 {
		t.Error("Expected service.ts to have dependencies")
	}

	// user.ts should have incoming references
	if graph.IncomingRefs[filepath.Join("src", "user.ts")] < 1 {
		t.Error("Expected user.ts to have at least 1 incoming reference")
	}
}
