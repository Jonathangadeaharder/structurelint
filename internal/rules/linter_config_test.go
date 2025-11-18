package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestLinterConfigRule_Name(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{}

	// Act
	name := rule.Name()

	// Assert
	if name != "linter-config" {
		t.Errorf("Expected rule name 'linter-config', got '%s'", name)
	}
}

func TestLinterConfigRule_PythonWithPyprojectToml(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: "pyproject.toml", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with pyproject.toml, got %d violations: %v", len(violations), violations)
	}
}

func TestLinterConfigRule_PythonWithoutConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) == 0 {
		t.Error("Expected violation for missing Python linter configuration")
	}
	if !containsMessage(violations, "No Python linter configuration found") {
		t.Error("Expected violation message about missing Python linter configuration")
	}
}

func TestLinterConfigRule_PythonWithFlake8Config(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: ".flake8", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with .flake8 config, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_TypeScriptWithESLintConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireTypeScript: true,
	}
	files := []walker.FileInfo{
		{Path: "main.ts", ParentPath: ".", IsDir: false},
		{Path: ".eslintrc.json", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with .eslintrc.json, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_TypeScriptWithoutConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireTypeScript: true,
	}
	files := []walker.FileInfo{
		{Path: "main.ts", ParentPath: ".", IsDir: false},
		{Path: "package.json", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) == 0 {
		t.Error("Expected violation for missing TypeScript linter configuration")
	}
	if !containsMessage(violations, "No TypeScript linter configuration found") {
		t.Error("Expected violation message about missing TypeScript linter configuration")
	}
}

func TestLinterConfigRule_GoWithGolangCIConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireGo: true,
	}
	files := []walker.FileInfo{
		{Path: "main.go", ParentPath: ".", IsDir: false},
		{Path: ".golangci.yml", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with .golangci.yml, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_GoWithoutConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireGo: true,
	}
	files := []walker.FileInfo{
		{Path: "main.go", ParentPath: ".", IsDir: false},
		{Path: "go.mod", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) == 0 {
		t.Error("Expected violation for missing Go linter configuration")
	}
	if !containsMessage(violations, "No Go linter configuration found") {
		t.Error("Expected violation message about missing Go linter configuration")
	}
}

func TestLinterConfigRule_PythonWithWorkflow(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	workflowContent := `
name: Python Linting
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Black
        run: black --check .
      - name: Run mypy
        run: mypy src/
`
	workflowPath := filepath.Join(workflowsDir, "python-lint.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}

	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: tmpDir, IsDir: false},
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with workflow containing black and mypy, got %d violations: %v", len(violations), violations)
	}
}

func TestLinterConfigRule_NoPythonFiles(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.go", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	// Should have no violations since there are no Python files
	if len(violations) != 0 {
		t.Errorf("Expected no violations when no Python files exist, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_MultipleLanguages(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequirePython:     true,
		RequireTypeScript: true,
		RequireGo:         true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: "app.ts", ParentPath: ".", IsDir: false},
		{Path: "server.go", ParentPath: ".", IsDir: false},
		{Path: "pyproject.toml", ParentPath: ".", IsDir: false},
		{Path: ".eslintrc.json", ParentPath: ".", IsDir: false},
		{Path: ".golangci.yml", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with all linter configs present, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_PrettierConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireTypeScript: true,
	}
	files := []walker.FileInfo{
		{Path: "main.ts", ParentPath: ".", IsDir: false},
		{Path: ".prettierrc", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with .prettierrc, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_TsconfigJson(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireTypeScript: true,
	}
	files := []walker.FileInfo{
		{Path: "main.ts", ParentPath: ".", IsDir: false},
		{Path: "tsconfig.json", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with tsconfig.json, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_HTMLWithHTMLHintConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireHTML: true,
	}
	files := []walker.FileInfo{
		{Path: "index.html", ParentPath: ".", IsDir: false},
		{Path: ".htmlhintrc", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with .htmlhintrc, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_HTMLWithoutConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireHTML: true,
	}
	files := []walker.FileInfo{
		{Path: "index.html", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) == 0 {
		t.Error("Expected violation for missing HTML linter configuration")
	}
	if !containsMessage(violations, "No HTML linter configuration found") {
		t.Error("Expected violation message about missing HTML linter configuration")
	}
}

func TestLinterConfigRule_CSSWithStylelintConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireCSS: true,
	}
	files := []walker.FileInfo{
		{Path: "styles.css", ParentPath: ".", IsDir: false},
		{Path: ".stylelintrc.json", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with .stylelintrc.json, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_CSSWithoutConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireCSS: true,
	}
	files := []walker.FileInfo{
		{Path: "styles.css", ParentPath: ".", IsDir: false},
		{Path: "package.json", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) == 0 {
		t.Error("Expected violation for missing CSS linter configuration")
	}
	if !containsMessage(violations, "No CSS linter configuration found") {
		t.Error("Expected violation message about missing CSS linter configuration")
	}
}

func TestLinterConfigRule_SQLWithSQLFluffConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireSQL: true,
	}
	files := []walker.FileInfo{
		{Path: "query.sql", ParentPath: ".", IsDir: false},
		{Path: ".sqlfluff", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with .sqlfluff, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_SQLWithoutConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireSQL: true,
	}
	files := []walker.FileInfo{
		{Path: "query.sql", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) == 0 {
		t.Error("Expected violation for missing SQL linter configuration")
	}
	if !containsMessage(violations, "No SQL linter configuration found") {
		t.Error("Expected violation message about missing SQL linter configuration")
	}
}

func TestLinterConfigRule_RustWithRustfmtConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireRust: true,
	}
	files := []walker.FileInfo{
		{Path: "main.rs", ParentPath: ".", IsDir: false},
		{Path: "rustfmt.toml", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with rustfmt.toml, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_RustWithoutConfig(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireRust: true,
	}
	files := []walker.FileInfo{
		{Path: "main.rs", ParentPath: ".", IsDir: false},
		{Path: "Cargo.toml", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) == 0 {
		t.Error("Expected violation for missing Rust linter configuration")
	}
	if !containsMessage(violations, "No Rust linter configuration found") {
		t.Error("Expected violation message about missing Rust linter configuration")
	}
}

func TestLinterConfigRule_NoHTMLFiles(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequireHTML: true,
	}
	files := []walker.FileInfo{
		{Path: "main.go", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	// Should have no violations since there are no HTML files
	if len(violations) != 0 {
		t.Errorf("Expected no violations when no HTML files exist, got %d violations", len(violations))
	}
}

func TestLinterConfigRule_AllLanguages(t *testing.T) {
	// Arrange
	rule := &LinterConfigRule{
		RequirePython:     true,
		RequireTypeScript: true,
		RequireGo:         true,
		RequireHTML:       true,
		RequireCSS:        true,
		RequireSQL:        true,
		RequireRust:       true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: "app.ts", ParentPath: ".", IsDir: false},
		{Path: "server.go", ParentPath: ".", IsDir: false},
		{Path: "index.html", ParentPath: ".", IsDir: false},
		{Path: "styles.css", ParentPath: ".", IsDir: false},
		{Path: "query.sql", ParentPath: ".", IsDir: false},
		{Path: "lib.rs", ParentPath: ".", IsDir: false},
		{Path: "pyproject.toml", ParentPath: ".", IsDir: false},
		{Path: ".eslintrc.json", ParentPath: ".", IsDir: false},
		{Path: ".golangci.yml", ParentPath: ".", IsDir: false},
		{Path: ".htmlhintrc", ParentPath: ".", IsDir: false},
		{Path: ".stylelintrc.json", ParentPath: ".", IsDir: false},
		{Path: ".sqlfluff", ParentPath: ".", IsDir: false},
		{Path: "rustfmt.toml", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations with all linter configs present, got %d violations", len(violations))
	}
}
