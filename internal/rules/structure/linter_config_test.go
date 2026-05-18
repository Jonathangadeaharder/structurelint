package structure

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// writeFile is a test helper that writes content to a file inside dir.
func writeFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("failed to write %s: %v", name, err)
	}
}

// walkDir is a test helper that walks a temp directory and returns files and dirs.
func walkDir(t *testing.T, dir string) ([]walker.FileInfo, map[string]*walker.DirInfo) {
	t.Helper()
	w := walker.New(dir)
	if err := w.Walk(); err != nil {
		t.Fatalf("walk failed: %v", err)
	}
	return w.GetFiles(), w.GetDirs()
}

func TestLinterConfig_EmptyProject_MissingSemgrep(t *testing.T) {
	dir := t.TempDir()
	rule := NewLinterConfigRule(dir)
	files, dirs := walkDir(t, dir)

	v := rule.Check(files, dirs)
	if len(v) != 1 {
		t.Fatalf("want 1 violation (missing .semgrep.yml), got %d: %+v", len(v), v)
	}
	if v[0].Path != ".semgrep.yml" {
		t.Fatalf("violation should target .semgrep.yml, got path=%q", v[0].Path)
	}
}

func TestLinterConfig_ValidPythonProject_ZeroViolations(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "pyproject.toml", `[tool.pyright]
typeCheckingMode = "strict"

[tool.ruff]
line-length = 100
`)
	writeFile(t, dir, ".semgrep.yml", "rules: []\n")

	rule := NewLinterConfigRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)
	if len(v) != 0 {
		t.Fatalf("want 0 violations, got %d: %+v", len(v), v)
	}
}

func TestLinterConfig_PythonMissingPyright(t *testing.T) {
	dir := t.TempDir()
	// pyproject.toml has [tool.ruff] but no [tool.pyright]
	writeFile(t, dir, "pyproject.toml", `[tool.ruff]
line-length = 100
`)
	writeFile(t, dir, ".semgrep.yml", "rules: []\n")

	rule := NewLinterConfigRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)
	if len(v) != 1 {
		t.Fatalf("want 1 violation (missing [tool.pyright]), got %d: %+v", len(v), v)
	}
	if v[0].Path != "pyproject.toml" {
		t.Fatalf("violation should target pyproject.toml, got path=%q", v[0].Path)
	}
}

func TestLinterConfig_PythonWrongPyrightMode(t *testing.T) {
	dir := t.TempDir()
	// pyright exists but typeCheckingMode is "basic", not "strict"
	writeFile(t, dir, "pyproject.toml", `[tool.pyright]
typeCheckingMode = "basic"

[tool.ruff]
line-length = 100
`)
	writeFile(t, dir, ".semgrep.yml", "rules: []\n")

	rule := NewLinterConfigRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)
	if len(v) != 1 {
		t.Fatalf("want 1 violation (wrong typeCheckingMode), got %d: %+v", len(v), v)
	}
	if v[0].Path != "pyproject.toml" {
		t.Fatalf("violation should target pyproject.toml, got path=%q", v[0].Path)
	}
}

func TestLinterConfig_PythonMissingRuff(t *testing.T) {
	dir := t.TempDir()
	// pyproject.toml has [tool.pyright] but no [tool.ruff]
	writeFile(t, dir, "pyproject.toml", `[tool.pyright]
typeCheckingMode = "strict"
`)
	writeFile(t, dir, ".semgrep.yml", "rules: []\n")

	rule := NewLinterConfigRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)
	if len(v) != 1 {
		t.Fatalf("want 1 violation (missing [tool.ruff]), got %d: %+v", len(v), v)
	}
	if v[0].Path != "pyproject.toml" {
		t.Fatalf("violation should target pyproject.toml, got path=%q", v[0].Path)
	}
}

func TestLinterConfig_ValidTSProject_ZeroViolations(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "tsconfig.json", `{
  "compilerOptions": {
    "strict": true
  }
}
`)
	writeFile(t, dir, ".semgrep.yml", "rules: []\n")

	rule := NewLinterConfigRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)
	if len(v) != 0 {
		t.Fatalf("want 0 violations, got %d: %+v", len(v), v)
	}
}

func TestLinterConfig_TSWithoutStrict(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, dir, "tsconfig.json", `{
  "compilerOptions": {
    "noImplicitAny": true
  }
}
`)
	writeFile(t, dir, ".semgrep.yml", "rules: []\n")

	rule := NewLinterConfigRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)
	if len(v) != 1 {
		t.Fatalf("want 1 violation (missing strict), got %d: %+v", len(v), v)
	}
	if v[0].Path != "tsconfig.json" {
		t.Fatalf("violation should target tsconfig.json, got path=%q", v[0].Path)
	}
}
