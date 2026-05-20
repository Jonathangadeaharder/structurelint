package quality

import (
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"github.com/stretchr/testify/assert"
)

func TestMaxHalsteadEffortRule_Name(t *testing.T) {
	r := NewMaxHalsteadEffortRule(1000, nil)
	assert.Equal(t, "max-halstead-effort", r.Name())
}

func TestMaxHalsteadEffortRule_Check_NoViolations(t *testing.T) {
	r := NewMaxHalsteadEffortRule(999999, nil)
	files := []walker.FileInfo{
		{AbsPath: filepath.Join("testdata", "simple.go"), Path: "testdata/simple.go"},
	}
	violations := r.Check(files, nil)
	assert.Empty(t, violations)
}

func TestMaxHalsteadEffortRule_Check_WithViolation(t *testing.T) {
	// Create a Go file with enough operators/operands to exceed a tiny threshold
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "busy.go")
	writeTestFile(t, goFile, `package main

func busy() {
    a, b, c, d, e := 1, 2, 3, 4, 5
    x := a + b*c - d/e + a*b - c*d + e*a - b*c + d*e
    _ = x
}
`)

	r := NewMaxHalsteadEffortRule(1, nil) // Very low threshold
	files := []walker.FileInfo{
		{AbsPath: goFile, Path: "busy.go"},
	}
	violations := r.Check(files, nil)
	assert.NotEmpty(t, violations, "should find violations with max=1")
}

func TestMaxHalsteadEffortRule_Check_IgnoreUnknownType(t *testing.T) {
	r := NewMaxHalsteadEffortRule(1000, nil)
	files := []walker.FileInfo{
		{AbsPath: "foo.bar", Path: "foo.bar"},
	}
	violations := r.Check(files, nil)
	assert.Empty(t, violations)
}

func TestMaxHalsteadEffortRule_Check_FilePatterns(t *testing.T) {
	r := NewMaxHalsteadEffortRule(1000, []string{"src/**/*.go"})
	files := []walker.FileInfo{
		{AbsPath: "src/main.go", Path: "src/main.go"},
		{AbsPath: "main.go", Path: "main.go"},
	}
	violations := r.Check(files, nil)
	assert.Empty(t, violations)
}

func TestMaxHalsteadEffortRule_Check_SkipsTestFiles(t *testing.T) {
	// The rule does NOT skip test files by default; it just applies the threshold.
	// This test verifies test files ARE analyzed.
	r := NewMaxHalsteadEffortRule(0.1, nil) // Very low threshold
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "foo_test.go")
	writeTestFile(t, testFile, `package main

func TestFoo(t *testing.T) {
    x := 1 + 2
    _ = x
}
`)

	files := []walker.FileInfo{
		{AbsPath: testFile, Path: "foo_test.go"},
	}
	violations := r.Check(files, nil)
	assert.NotEmpty(t, violations, "test files should be analyzed and violate low threshold")
}

func TestNewMaxHalsteadEffortRule(t *testing.T) {
	r := NewMaxHalsteadEffortRule(500000, []string{"*.go"})
	assert.Equal(t, 500000.0, r.Max)
	assert.Len(t, r.FilePatterns, 1)
}
