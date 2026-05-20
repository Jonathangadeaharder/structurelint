package quality

import (
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"github.com/stretchr/testify/assert"
)

func TestMaxCognitiveComplexityRule_Name(t *testing.T) {
	r := NewMaxCognitiveComplexityRule(10, nil)
	assert.Equal(t, "max-cognitive-complexity", r.Name())
}

func TestMaxCognitiveComplexityRule_Check_NoViolations(t *testing.T) {
	r := NewMaxCognitiveComplexityRule(100, nil)
	files := []walker.FileInfo{
		{AbsPath: filepath.Join("testdata", "simple.go"), Path: "testdata/simple.go"},
	}
	violations := r.Check(files, nil)
	assert.Empty(t, violations)
}

func TestMaxCognitiveComplexityRule_Check_WithViolation(t *testing.T) {
	// Create a temporary Go file with moderate complexity
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "complex.go")
	writeTestFile(t, goFile, `package main

func simple() {
    x := 1
    _ = x
}

func complex() {
    if true {
        if true {
            if true {
                println("deep")
            }
        }
    }
}
`)

	r := NewMaxCognitiveComplexityRule(1, nil) // Very low threshold
	files := []walker.FileInfo{
		{AbsPath: goFile, Path: "complex.go"},
	}
	violations := r.Check(files, nil)
	assert.NotEmpty(t, violations, "should find violations with max=1")
	hasNamed := false
	for _, v := range violations {
		if v.Path == "complex.go" {
			hasNamed = true
		}
	}
	assert.True(t, hasNamed, "expected violation for complex.go")
}

func TestMaxCognitiveComplexityRule_Check_IgnoreUnknownType(t *testing.T) {
	r := NewMaxCognitiveComplexityRule(10, nil)
	files := []walker.FileInfo{
		{AbsPath: "foo.bar", Path: "foo.bar"},
	}
	violations := r.Check(files, nil)
	assert.Empty(t, violations)
}

func TestMaxCognitiveComplexityRule_Check_FilePatterns(t *testing.T) {
	r := NewMaxCognitiveComplexityRule(10, []string{"src/**/*.go"})
	files := []walker.FileInfo{
		{AbsPath: "src/main.go", Path: "src/main.go"},
		{AbsPath: "main.go", Path: "main.go"},
	}
	violations := r.Check(files, nil)
	// Both files: src/main.go is a Go file but may not parse (no real file),
	// but main.go should be filtered by patterns.
	assert.Empty(t, violations)
}

func TestMaxCognitiveComplexityRule_Check_TestMaxSkipsTests(t *testing.T) {
	r := NewMaxCognitiveComplexityRule(10, nil)
	files := []walker.FileInfo{
		{AbsPath: "foo_test.go", Path: "foo_test.go"},
	}
	violations := r.Check(files, nil)
	assert.Empty(t, violations)
}

func TestMaxCognitiveComplexityRule_Check_TestMaxOverride(t *testing.T) {
	r := NewMaxCognitiveComplexityRule(10, nil).WithTestMax(20)
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "foo_test.go")
	writeTestFile(t, goFile, `package main

func TestFoo(t *testing.T) {
    if true {
        println("a")
    }
    if true {
        println("b")
    }
    if true {
        println("c")
    }
}
`)

	files := []walker.FileInfo{
		{AbsPath: goFile, Path: "foo_test.go"},
	}
	violations := r.Check(files, nil)
	assert.Empty(t, violations, "cognitive complexity of 3 should pass with test-max=20")
}

func TestNewMaxCognitiveComplexityRule(t *testing.T) {
	r := NewMaxCognitiveComplexityRule(15, []string{"*.go"})
	assert.Equal(t, 15, r.Max)
	assert.Len(t, r.FilePatterns, 1)
	assert.Equal(t, 0, r.TestMax)
}

func TestWithTestMax(t *testing.T) {
	r := NewMaxCognitiveComplexityRule(10, nil)
	assert.Equal(t, 0, r.TestMax)
	r2 := r.WithTestMax(25)
	assert.Equal(t, 25, r2.TestMax)
}

// writeTestFile writes content to a file, panicking on error.
func writeTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := writeFile(path, content); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
}
