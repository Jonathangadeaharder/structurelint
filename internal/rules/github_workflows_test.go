package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestGitHubWorkflowsRule_Name(t *testing.T) {
	// Arrange
	rule := &GitHubWorkflowsRule{}

	// Act
	name := rule.Name()

	// Assert
	if name != "github-workflows" {
		t.Errorf("Expected rule name 'github-workflows', got '%s'", name)
	}
}

func TestGitHubWorkflowsRule_MissingWorkflowsDirectory(t *testing.T) {
	// Arrange
	rule := &GitHubWorkflowsRule{
		RequireTests:    true,
		RequireSecurity: true,
		RequireQuality:  true,
	}
	files := []walker.FileInfo{
		{Path: "README.md", ParentPath: ".", IsDir: false},
		{Path: "main.go", ParentPath: ".", IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if len(violations) == 0 {
		t.Error("Expected violation for missing .github/workflows directory")
	}
	if !containsMessage(violations, "GitHub workflows directory not found") {
		t.Error("Expected violation message about missing workflows directory")
	}
}

func TestGitHubWorkflowsRule_MissingTestWorkflow(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Deploy
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - run: echo "Deploying"
`
	workflowPath := filepath.Join(workflowsDir, "deploy.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireTests: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "No test/CI workflow found") {
		t.Error("Expected violation for missing test workflow")
	}
}

func TestGitHubWorkflowsRule_ValidTestWorkflow(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: CI Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: go test ./...
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireTests: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if containsMessage(violations, "No test/CI workflow found") {
		t.Error("Should not have violation for missing test workflow when one exists")
	}
}

func TestGitHubWorkflowsRule_MissingSecurityWorkflow(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Tests
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Testing"
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireSecurity: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "No security scanning workflow found") {
		t.Error("Expected violation for missing security workflow")
	}
}

func TestGitHubWorkflowsRule_ValidSecurityWorkflow(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: CodeQL Security Scan
on: [push, pull_request]
jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: github/codeql-action/init@v2
      - uses: github/codeql-action/analyze@v2
`
	workflowPath := filepath.Join(workflowsDir, "security.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireSecurity: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if containsMessage(violations, "No security scanning workflow found") {
		t.Error("Should not have violation for missing security workflow when one exists")
	}
}

func TestGitHubWorkflowsRule_MissingQualityWorkflow(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Tests
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Testing"
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireQuality: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "No code quality workflow found") {
		t.Error("Expected violation for missing code quality workflow")
	}
}

func TestGitHubWorkflowsRule_ValidQualityWorkflow(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Code Quality
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: golangci/golangci-lint-action@v3
`
	workflowPath := filepath.Join(workflowsDir, "lint.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireQuality: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if containsMessage(violations, "No code quality workflow found") {
		t.Error("Should not have violation for missing quality workflow when one exists")
	}
}

func TestGitHubWorkflowsRule_WorkflowMissingName(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Testing"
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "Workflow missing 'name' field") {
		t.Error("Expected violation for workflow missing name")
	}
}

func TestGitHubWorkflowsRule_WorkflowMissingTriggers(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Test
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Testing"
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "Workflow missing 'on' triggers") {
		t.Error("Expected violation for workflow missing triggers")
	}
}

func TestGitHubWorkflowsRule_RequiredTriggers(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - run: echo "Testing"
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequiredTriggers: []string{"pull_request"},
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "Required trigger 'pull_request' not found") {
		t.Error("Expected violation for missing required trigger")
	}
}

func TestGitHubWorkflowsRule_JobMissingRunsOn(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Test
on: [push]
jobs:
  test:
    steps:
      - run: echo "Testing"
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "Job 'test' missing 'runs-on' field") {
		t.Error("Expected violation for job missing runs-on")
	}
}

func TestGitHubWorkflowsRule_JobMissingSteps(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "Job 'test' has no steps defined") {
		t.Error("Expected violation for job missing steps")
	}
}

func TestGitHubWorkflowsRule_CompleteValidWorkflow(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}

	testWorkflow := `
name: CI Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: go test ./...
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "test.yml"), []byte(testWorkflow), 0644); err != nil {
		t.Fatal(err)
	}

	securityWorkflow := `
name: Security Scan
on: [push, pull_request]
jobs:
  security:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: github/codeql-action/analyze@v2
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "security.yml"), []byte(securityWorkflow), 0644); err != nil {
		t.Fatal(err)
	}

	qualityWorkflow := `
name: Code Quality
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: golangci/golangci-lint-action@v3
`
	if err := os.WriteFile(filepath.Join(workflowsDir, "lint.yml"), []byte(qualityWorkflow), 0644); err != nil {
		t.Fatal(err)
	}

	rule := &GitHubWorkflowsRule{
		RequireTests:    true,
		RequireSecurity: true,
		RequireQuality:  true,
	}

	files := []walker.FileInfo{
		{Path: filepath.Join(workflowsDir, "test.yml"), ParentPath: workflowsDir, IsDir: false},
		{Path: filepath.Join(workflowsDir, "security.yml"), ParentPath: workflowsDir, IsDir: false},
		{Path: filepath.Join(workflowsDir, "lint.yml"), ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if containsMessage(violations, "No test/CI workflow found") ||
		containsMessage(violations, "No security scanning workflow found") ||
		containsMessage(violations, "No code quality workflow found") {
		t.Errorf("Should not have violations for complete valid workflow setup, got: %v", violations)
	}
}

func TestGitHubWorkflowsRule_MissingLogCommits(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: go test ./...
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireLogCommits: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "Job 'test' missing log commit steps") {
		t.Errorf("Expected violation for missing log commit steps, got %d violations: %+v", len(violations), violations)
	}
}

func TestGitHubWorkflowsRule_ValidLogCommits(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: go test ./... 2>&1 | tee test.log
      - name: Commit logs
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "github-actions[bot]@users.noreply.github.com"
          git add *.log
          git commit -m "Add test execution logs"
          git push
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireLogCommits: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if containsMessage(violations, "Job 'test' missing log commit steps") {
		t.Error("Should not have violation for missing log commit steps when they exist")
	}
}

func TestGitHubWorkflowsRule_MissingRepomixArtifact(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: go test ./...
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireRepomixArtifact: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if !containsMessage(violations, "Job 'test' missing repomix artifact steps") {
		t.Error("Expected violation for missing repomix artifact steps")
	}
}

func TestGitHubWorkflowsRule_ValidRepomixArtifact(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0755); err != nil {
		t.Fatal(err)
	}
	workflowContent := `
name: Test
on: [push]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run tests
        run: go test ./...
      - name: Generate repomix summary
        run: npx repomix
      - name: Upload repomix artifact
        uses: actions/upload-artifact@v4
        with:
          name: repomix-output
          path: repomix-output.txt
          retention-days: 7
`
	workflowPath := filepath.Join(workflowsDir, "test.yml")
	if err := os.WriteFile(workflowPath, []byte(workflowContent), 0644); err != nil {
		t.Fatal(err)
	}
	rule := &GitHubWorkflowsRule{
		RequireRepomixArtifact: true,
	}
	files := []walker.FileInfo{
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
		{Path: "go.mod", ParentPath: tmpDir, IsDir: false},
	}

	// Act
	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	// Assert
	if containsMessage(violations, "Job 'test' missing repomix artifact steps") {
		t.Error("Should not have violation for missing repomix artifact steps when they exist")
	}
}

// Helper function to check if violations contain a specific message
func containsMessage(violations []Violation, message string) bool {
	for _, v := range violations {
		if v.Message == message ||
			(len(message) > 0 && len(v.Message) > 0 && v.Message[:len(message)] == message) {
			return true
		}
	}
	return false
}
