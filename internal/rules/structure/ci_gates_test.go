package structure

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCIGates_EmptyDir_TwoViolations(t *testing.T) {
	dir := t.TempDir()
	rule := NewCIGatesRule(dir)
	files, dirs := walkDir(t, dir)

	v := rule.Check(files, dirs)
	if len(v) != 2 {
		t.Fatalf("want 2 violations (pr-gate + merge-gate missing), got %d: %+v", len(v), v)
	}

	// Both violations should target correct paths.
	paths := make(map[string]bool)
	for _, vi := range v {
		paths[vi.Path] = true
	}
	if !paths[".github/workflows/pr-gate.yml"] {
		t.Error("expected violation for .github/workflows/pr-gate.yml")
	}
	if !paths[".github/workflows/merge-gate.yml"] {
		t.Error("expected violation for .github/workflows/merge-gate.yml")
	}
}

func TestCIGates_BothPresent_ZeroViolations(t *testing.T) {
	dir := t.TempDir()

	workflowsDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		t.Fatalf("mkdir .github/workflows: %v", err)
	}

	writeFile(t, workflowsDir, "pr-gate.yml", "name: PR Gate\n")
	writeFile(t, workflowsDir, "merge-gate.yml", "name: Merge Gate\n")

	rule := NewCIGatesRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)

	if len(v) != 0 {
		t.Fatalf("want 0 violations, got %d: %+v", len(v), v)
	}
}

func TestCIGates_OnlyPRGate_OneViolation(t *testing.T) {
	dir := t.TempDir()

	workflowsDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		t.Fatalf("mkdir .github/workflows: %v", err)
	}

	writeFile(t, workflowsDir, "pr-gate.yml", "name: PR Gate\n")

	rule := NewCIGatesRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)

	if len(v) != 1 {
		t.Fatalf("want 1 violation (merge-gate missing), got %d: %+v", len(v), v)
	}
	if v[0].Path != ".github/workflows/merge-gate.yml" {
		t.Fatalf("violation should target merge-gate.yml, got path=%q", v[0].Path)
	}
}

func TestCIGates_OnlyMergeGate_OneViolation(t *testing.T) {
	dir := t.TempDir()

	workflowsDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		t.Fatalf("mkdir .github/workflows: %v", err)
	}

	writeFile(t, workflowsDir, "merge-gate.yml", "name: Merge Gate\n")

	rule := NewCIGatesRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)

	if len(v) != 1 {
		t.Fatalf("want 1 violation (pr-gate missing), got %d: %+v", len(v), v)
	}
	if v[0].Path != ".github/workflows/pr-gate.yml" {
		t.Fatalf("violation should target pr-gate.yml, got path=%q", v[0].Path)
	}
}

func TestCIGates_PrefixMatching(t *testing.T) {
	dir := t.TempDir()

	workflowsDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		t.Fatalf("mkdir .github/workflows: %v", err)
	}

	// Files named differently but starting with the prefix.
	writeFile(t, workflowsDir, "pr-gate-extra.yml", "name: PR Gate Extra\n")
	writeFile(t, workflowsDir, "merge-gate-v2.yml", "name: Merge Gate V2\n")

	rule := NewCIGatesRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)

	if len(v) != 0 {
		t.Fatalf("want 0 violations (prefix matched), got %d: %+v", len(v), v)
	}
}
