package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/config"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func BenchmarkBuildGraph(b *testing.B) {
	tmpDir := b.TempDir()

	files := createFixture(b, tmpDir)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder(tmpDir, nil)
		_, err := builder.Build(files)
		if err != nil {
			b.Fatalf("Build failed: %v", err)
		}
	}
}

func BenchmarkBuildGraph_WithLayers(b *testing.B) {
	tmpDir := b.TempDir()

	files := createFixture(b, tmpDir)

	layers := []config.Layer{
		{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
		{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
		{Name: "infrastructure", Path: "src/infrastructure/**", DependsOn: []string{"domain", "application"}},
		{Name: "presentation", Path: "src/presentation/**", DependsOn: []string{"application"}},
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder(tmpDir, layers)
		_, err := builder.Build(files)
		if err != nil {
			b.Fatalf("Build failed: %v", err)
		}
	}
}

func BenchmarkBuildGraph_SmallProject(b *testing.B) {
	tmpDir := b.TempDir()
	files := createSmallFixture(b, tmpDir)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder(tmpDir, nil)
		_, err := builder.Build(files)
		if err != nil {
			b.Fatalf("Build failed: %v", err)
		}
	}
}

func BenchmarkBuildGraph_LargeProject(b *testing.B) {
	tmpDir := b.TempDir()
	files := createLargeFixture(b, tmpDir)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		builder := NewBuilder(tmpDir, nil)
		_, err := builder.Build(files)
		if err != nil {
			b.Fatalf("Build failed: %v", err)
		}
	}
}

func BenchmarkCanLayerDependOn(b *testing.B) {
	graph := &ImportGraph{
		Layers: []config.Layer{
			{Name: "presentation", Path: "src/presentation/**", DependsOn: []string{"application", "domain"}},
			{Name: "application", Path: "src/application/**", DependsOn: []string{"domain"}},
			{Name: "domain", Path: "src/domain/**", DependsOn: []string{}},
			{Name: "infrastructure", Path: "src/infrastructure/**", DependsOn: []string{"domain"}},
			{Name: "shared", Path: "src/shared/**", DependsOn: []string{"*"}},
		},
	}

	presentation := graph.FindLayerByName("presentation")
	domain := graph.FindLayerByName("domain")
	shared := graph.FindLayerByName("shared")

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		graph.CanLayerDependOn(presentation, domain)
		graph.CanLayerDependOn(domain, presentation)
		graph.CanLayerDependOn(shared, domain)
	}
}

func BenchmarkFindLayerByName(b *testing.B) {
	layers := make([]config.Layer, 20)
	for i := range layers {
		layers[i] = config.Layer{
			Name:     fmt.Sprintf("layer_%d", i),
			Path:     fmt.Sprintf("src/layer_%d/**", i),
			DependsOn: []string{},
		}
	}

	graph := &ImportGraph{Layers: layers}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for j := range layers {
			graph.FindLayerByName(layers[j].Name)
		}
	}
}

func createFixture(b testing.TB, root string) []walker.FileInfo {
	b.Helper()

	structure := map[string]string{
		"src/domain/user.ts":              "export class User { id: string; name: string; }",
		"src/domain/product.ts":           "export class Product { id: string; price: number; }",
		"src/domain/order.ts":             "import { User } from './user';\nimport { Product } from './product';\nexport class Order { user: User; products: Product[]; }",
		"src/application/userService.ts":  "import { User } from '../domain/user';\nexport class UserService { findById(id: string): User { return new User(); } }",
		"src/application/orderService.ts": "import { Order } from '../domain/order';\nimport { UserService } from './userService';\nexport class OrderService { create(user: User): Order { return new Order(); } }",
		"src/infrastructure/db.ts":        "import { User } from '../domain/user';\nimport { Product } from '../domain/product';\nexport class Database { findUser(id: string): User { return new User(); } }",
		"src/infrastructure/api.ts":       "import { UserService } from '../application/userService';\nexport class ApiClient { getUser(id: string): any { return {}; } }",
		"src/presentation/index.ts":       "import { OrderService } from '../application/orderService';\nimport { ApiClient } from '../infrastructure/api';\nexport function main() {}",
		"src/presentation/components.tsx": "import React from 'react';\nimport { User } from '../domain/user';\nexport const Component = () => null;",
		"src/shared/types.ts":             "export interface Config { debug: boolean; }",
		"src/shared/constants.ts":         "export const API_URL = 'http://localhost:3000';",
	}

	return writeFixtureFiles(b, root, structure)
}

func createSmallFixture(b testing.TB, root string) []walker.FileInfo {
	b.Helper()

	structure := map[string]string{
		"src/main.ts":     "import { Config } from './config';\nexport function start(cfg: Config) {}",
		"src/config.ts":   "export interface Config { port: number; }",
		"src/utils.ts":    "export function formatDate(d: Date): string { return d.toISOString(); }",
		"src/index.ts":    "import { start } from './main';\nimport { formatDate } from './utils';\nexport { start, formatDate };",
	}

	return writeFixtureFiles(b, root, structure)
}

func createLargeFixture(b testing.TB, root string) []walker.FileInfo {
	b.Helper()

	structure := make(map[string]string)
	layerNames := []string{"domain", "application", "infrastructure", "presentation", "shared"}

	for i := 0; i < 50; i++ {
		layer := layerNames[i%len(layerNames)]
		filePath := fmt.Sprintf("src/%s/module_%d.ts", layer, i)

		imports := ""
		if i > 0 {
			prevLayer := layerNames[(i-1)%len(layerNames)]
			imports = fmt.Sprintf("import { Module_%d } from '../%s/module_%d';\n", i-1, prevLayer, i-1)
		}

		structure[filePath] = imports + fmt.Sprintf("export class Module_%d { value: number = %d; }", i, i)
	}

	return writeFixtureFiles(b, root, structure)
}

func writeFixtureFiles(b testing.TB, root string, files map[string]string) []walker.FileInfo {
	b.Helper()

	var result []walker.FileInfo
	for relPath, content := range files {
		absPath := filepath.Join(root, relPath)
		dir := filepath.Dir(absPath)

		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatal(err)
		}

		if err := os.WriteFile(absPath, []byte(content), 0644); err != nil {
			b.Fatal(err)
		}

		result = append(result, walker.FileInfo{
			Path:    relPath,
			AbsPath: absPath,
			IsDir:   false,
		})
	}

	return result
}
