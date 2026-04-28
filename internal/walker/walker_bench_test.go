package walker

import (
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkWalk(b *testing.B) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(root); err != nil {
		b.Fatalf("project root not found at %s: %v", root, err)
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := New(root)
		if err := w.Walk(); err != nil {
			b.Fatalf("Walk failed: %v", err)
		}
	}
}

func BenchmarkWalk_WithExcludes(b *testing.B) {
	root := filepath.Join("..", "..")
	if _, err := os.Stat(root); err != nil {
		b.Fatalf("project root not found at %s: %v", root, err)
	}

	excludes := []string{"*.md", "docs/**", "testdata/**"}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := New(root).WithExclude(excludes)
		if err := w.Walk(); err != nil {
			b.Fatalf("Walk failed: %v", err)
		}
	}
}

func BenchmarkWalk_MediumTree(b *testing.B) {
	tmpDir := b.TempDir()

	for i := 0; i < 10; i++ {
		dir := filepath.Join(tmpDir, "dir", string(rune('a'+i)), "sub")
		if err := os.MkdirAll(dir, 0755); err != nil {
			b.Fatal(err)
		}
		for j := 0; j < 5; j++ {
			f := filepath.Join(dir, "file"+string(rune('0'+j))+".go")
			if err := os.WriteFile(f, []byte("package main\n"), 0644); err != nil {
				b.Fatal(err)
			}
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		w := New(tmpDir)
		if err := w.Walk(); err != nil {
			b.Fatalf("Walk failed: %v", err)
		}
	}
}

func BenchmarkMatchesPattern(b *testing.B) {
	patterns := []string{"*.ts", "src/**/*.go", "docs/", "*.test.js"}
	paths := []string{
		"src/app.ts",
		"src/internal/walker/walker.go",
		"docs/README.md",
		"test/app.test.js",
		"package.json",
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		for _, p := range paths {
			for _, pat := range patterns {
				MatchesPattern(p, pat)
			}
		}
	}
}
