// Package rules provides GitHub workflow enforcement rules.
package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/lang"
	"github.com/structurelint/structurelint/internal/walker"
	"gopkg.in/yaml.v3"
)

// GitHubWorkflowsRule enforces the presence and configuration of GitHub Actions workflows
// for test execution, security scanning, and code quality checks.
type GitHubWorkflowsRule struct {
	RequireTests          bool     `yaml:"require-tests"`
	RequireSecurity       bool     `yaml:"require-security"`
	RequireQuality        bool     `yaml:"require-quality"`
	RequiredJobs          []string `yaml:"required-jobs"`
	RequiredTriggers      []string `yaml:"required-triggers"`
	AllowMissing          []string `yaml:"allow-missing"`
	RequireLogCommits     bool     `yaml:"require-log-commits"`
	RequireRepomixArtifact bool    `yaml:"require-repomix-artifact"`
}

// WorkflowFile represents a parsed GitHub Actions workflow file
type WorkflowFile struct {
	Name string                 `yaml:"name"`
	On   interface{}            `yaml:"on"`
	Jobs map[string]WorkflowJob `yaml:"jobs"`
}

// WorkflowJob represents a job in a GitHub Actions workflow
type WorkflowJob struct {
	Name  string           `yaml:"name"`
	RunsOn interface{}     `yaml:"runs-on"`
	Steps []WorkflowStep   `yaml:"steps"`
}

// WorkflowStep represents a step in a workflow job
type WorkflowStep struct {
	Name string                 `yaml:"name"`
	Uses string                 `yaml:"uses"`
	Run  string                 `yaml:"run"`
	If   string                 `yaml:"if"`
	With map[string]interface{} `yaml:"with"`
}

// Name returns the rule name
func (r *GitHubWorkflowsRule) Name() string {
	return "github-workflows"
}

// Check validates GitHub workflow requirements
func (r *GitHubWorkflowsRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Find the root directory
	rootDir := r.findRootDir(files)
	if rootDir == "" {
		// If we can't determine root, check if any file is in the root
		for _, file := range files {
			if file.ParentPath == "" || file.ParentPath == "." {
				rootDir = "."
				break
			}
		}
	}

	// Check if .github/workflows directory exists
	workflowsDir := filepath.Join(rootDir, ".github", "workflows")
	workflowFiles := r.findWorkflowFiles(files, workflowsDir)

	fmt.Printf("DEBUG: RootDir: %s, Total Files: %d, Workflow Files: %d\n", rootDir, len(files), len(workflowFiles))

	if len(workflowFiles) == 0 {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    ".github/workflows",
			Message: "GitHub workflows directory not found. Add CI/CD workflows for testing, security, and code quality.",
		})
		return violations
	}

	// Parse and validate workflows
	workflows := r.parseWorkflows(workflowFiles)

	// Check for required workflow types
	if r.RequireTests {
		if !r.hasWorkflowType(workflows, "test", []string{"test", "ci", "build"}) {
			autoFix := r.generateWorkflowAutoFix(rootDir, "test")
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    ".github/workflows",
				Message: "No test/CI workflow found. Add a workflow that runs tests on pull requests and pushes.",
				Suggestions: []string{
					"Create a CI workflow file in .github/workflows/",
					"Use the auto-generated workflow based on your project's language",
				},
				AutoFix: autoFix,
			})
		}
	}

	if r.RequireSecurity {
		if !r.hasWorkflowType(workflows, "security", []string{"security", "scan", "codeql", "dependabot"}) {
			autoFix := r.generateWorkflowAutoFix(rootDir, "security")
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    ".github/workflows",
				Message: "No security scanning workflow found. Add CodeQL, dependency scanning, or other security checks.",
				Suggestions: []string{
					"Create a security workflow file in .github/workflows/",
					"Use the auto-generated workflow with CodeQL and Trivy scanning",
				},
				AutoFix: autoFix,
			})
		}
	}

	if r.RequireQuality {
		if !r.hasWorkflowType(workflows, "quality", []string{"quality", "lint", "format", "coverage"}) {
			autoFix := r.generateWorkflowAutoFix(rootDir, "quality")
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    ".github/workflows",
				Message: "No code quality workflow found. Add linting, formatting, or coverage checks.",
				Suggestions: []string{
					"Create a quality workflow file in .github/workflows/",
					"Use the auto-generated workflow based on your project's language",
				},
				AutoFix: autoFix,
			})
		}
	}

	// Validate individual workflows
	for path, workflow := range workflows {
		workflowViolations := r.validateWorkflow(path, workflow)
		violations = append(violations, workflowViolations...)
	}

	return violations
}

// findRootDir finds the root directory of the project
func (r *GitHubWorkflowsRule) findRootDir(files []walker.FileInfo) string {
	// Look for common root indicators
	for _, file := range files {
		if strings.HasSuffix(file.Path, "go.mod") ||
			strings.HasSuffix(file.Path, "package.json") ||
			strings.HasSuffix(file.Path, ".git") {
			return filepath.Dir(file.Path)
		}
	}
	return ""
}

// findWorkflowFiles finds all GitHub workflow files
func (r *GitHubWorkflowsRule) findWorkflowFiles(files []walker.FileInfo, workflowsDir string) []walker.FileInfo {
	var workflowFiles []walker.FileInfo
	_ = workflowsDir // Parameter kept for API compatibility but unused

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Check if file is in .github/workflows directory
		normalizedPath := filepath.ToSlash(file.Path)
		parentPath := filepath.ToSlash(file.ParentPath)

		if (strings.Contains(normalizedPath, ".github/workflows") ||
			strings.Contains(parentPath, ".github/workflows")) &&
			(strings.HasSuffix(file.Path, ".yml") || strings.HasSuffix(file.Path, ".yaml")) {
			workflowFiles = append(workflowFiles, file)
		}
	}

	return workflowFiles
}

// parseWorkflows parses workflow YAML files
func (r *GitHubWorkflowsRule) parseWorkflows(files []walker.FileInfo) map[string]*WorkflowFile {
	workflows := make(map[string]*WorkflowFile)

	for _, file := range files {
		data, err := os.ReadFile(file.Path)
		if err != nil {
			continue
		}

		var workflow WorkflowFile
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			continue
		}

		workflows[file.Path] = &workflow
	}

	return workflows
}

// hasWorkflowType checks if any workflow matches the required type
func (r *GitHubWorkflowsRule) hasWorkflowType(workflows map[string]*WorkflowFile, workflowType string, keywords []string) bool {
	for _, workflow := range workflows {
		workflowNameLower := strings.ToLower(workflow.Name)

		// Check if workflow name contains any of the keywords
		for _, keyword := range keywords {
			if strings.Contains(workflowNameLower, keyword) {
				return true
			}
		}

		// Check job names
		for jobName := range workflow.Jobs {
			jobNameLower := strings.ToLower(jobName)
			for _, keyword := range keywords {
				if strings.Contains(jobNameLower, keyword) {
					return true
				}
			}
		}
	}

	return false
}

// validateWorkflow validates an individual workflow file
func (r *GitHubWorkflowsRule) validateWorkflow(path string, workflow *WorkflowFile) []Violation {
	var violations []Violation

	// Check for workflow name
	if workflow.Name == "" {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    path,
			Message: "Workflow missing 'name' field. Add a descriptive name for the workflow.",
		})
	}

	// Check for triggers
	if workflow.On == nil {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    path,
			Message: "Workflow missing 'on' triggers. Specify when the workflow should run (e.g., on: [push, pull_request]).",
		})
	} else {
		violations = append(violations, r.validateTriggers(path, workflow)...)
	}

	// Check for jobs
	if len(workflow.Jobs) == 0 {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    path,
			Message: "Workflow has no jobs defined. Add at least one job to execute.",
		})
	}

	// Validate required jobs
	for _, requiredJob := range r.RequiredJobs {
		if !r.hasJob(workflow, requiredJob) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    path,
				Message: fmt.Sprintf("Required job '%s' not found in workflow.", requiredJob),
			})
		}
	}

	// Validate jobs
	for jobName, job := range workflow.Jobs {
		jobViolations := r.validateJob(path, jobName, job)
		violations = append(violations, jobViolations...)
	}

	return violations
}

// validateTriggers validates workflow triggers
func (r *GitHubWorkflowsRule) validateTriggers(path string, workflow *WorkflowFile) []Violation {
	var violations []Violation

	triggers := r.extractTriggers(workflow.On)

	// Check for required triggers
	for _, required := range r.RequiredTriggers {
		found := false
		for _, trigger := range triggers {
			if trigger == required {
				found = true
				break
			}
		}
		if !found {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    path,
				Message: fmt.Sprintf("Required trigger '%s' not found. Add to workflow triggers.", required),
			})
		}
	}

	return violations
}

// extractTriggers extracts trigger names from the 'on' field
func (r *GitHubWorkflowsRule) extractTriggers(on interface{}) []string {
	var triggers []string

	switch v := on.(type) {
	case string:
		triggers = append(triggers, v)
	case []interface{}:
		for _, item := range v {
			if str, ok := item.(string); ok {
				triggers = append(triggers, str)
			}
		}
	case map[string]interface{}:
		for key := range v {
			triggers = append(triggers, key)
		}
	}

	return triggers
}

// hasJob checks if a workflow has a job with the given name or pattern
func (r *GitHubWorkflowsRule) hasJob(workflow *WorkflowFile, pattern string) bool {
	for jobName := range workflow.Jobs {
		if strings.Contains(strings.ToLower(jobName), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

// validateJob validates an individual job
func (r *GitHubWorkflowsRule) validateJob(path, jobName string, job WorkflowJob) []Violation {
	var violations []Violation

	// Check for runs-on
	if job.RunsOn == nil {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    path,
			Message: fmt.Sprintf("Job '%s' missing 'runs-on' field. Specify the runner environment.", jobName),
		})
	}

	// Check for steps
	if len(job.Steps) == 0 {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    path,
			Message: fmt.Sprintf("Job '%s' has no steps defined. Add steps to execute in this job.", jobName),
		})
	}

	// Check for logging requirements
	if r.RequireLogCommits && !r.hasLogCommitSteps(job) {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    path,
			Message: fmt.Sprintf("Job '%s' missing log commit steps. Add steps to commit and push execution logs to the triggering branch so agents can pull and review results. Example: use 'git add *.log && git commit -m \"Add logs\" && git push'.", jobName),
		})
	}

	if r.RequireRepomixArtifact && !r.hasRepomixArtifactSteps(job) {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    path,
			Message: fmt.Sprintf("Job '%s' missing repomix artifact steps. Add steps to run repomix and upload the codebase summary as an artifact for agent context. Example: run 'npx repomix' and use 'actions/upload-artifact@v4' to upload the output.", jobName),
		})
	}

	return violations
}

// hasLogCommitSteps checks if a job has steps that commit logs to the branch
func (r *GitHubWorkflowsRule) hasLogCommitSteps(job WorkflowJob) bool {
	hasLogCreation := false
	hasGitCommit := false
	hasGitPush := false

	for _, step := range job.Steps {
		runLower := strings.ToLower(step.Run)

		// Check for log file creation (tee or redirection to .log files)
		if strings.Contains(runLower, "tee") && strings.Contains(runLower, ".log") {
			hasLogCreation = true
		}
		if strings.Contains(runLower, ">.log") || strings.Contains(runLower, ">>.log") {
			hasLogCreation = true
		}

		// Check for git commit operations
		if strings.Contains(runLower, "git commit") || strings.Contains(runLower, "git add") {
			hasGitCommit = true
		}

		// Check for git push operations
		if strings.Contains(runLower, "git push") {
			hasGitPush = true
		}
	}

	// All three components are required: log creation, git commit, and git push
	return hasLogCreation && hasGitCommit && hasGitPush
}

// hasRepomixArtifactSteps checks if a job has steps that run repomix and upload the artifact
func (r *GitHubWorkflowsRule) hasRepomixArtifactSteps(job WorkflowJob) bool {
	hasRepomixRun := false
	hasArtifactUpload := false

	for _, step := range job.Steps {
		runLower := strings.ToLower(step.Run)

		// Check for repomix execution (npx repomix, npm repomix, etc.)
		if strings.Contains(runLower, "repomix") {
			hasRepomixRun = true
		}

		// Check if step uses upload-artifact action
		if strings.Contains(step.Uses, "upload-artifact") {
			// Check if it uploads repomix output (typically .xml, .txt, or repomix-output)
			if step.With != nil {
				if path, ok := step.With["path"]; ok {
					pathStr := strings.ToLower(fmt.Sprintf("%v", path))
					if strings.Contains(pathStr, "repomix") ||
						strings.Contains(pathStr, "output.xml") ||
						strings.Contains(pathStr, "output.txt") {
						hasArtifactUpload = true
					}
				}
			}
		}
	}

	// Both repomix execution and artifact upload are required
	return hasRepomixRun && hasArtifactUpload
}

// NewGitHubWorkflowsRule creates a new GitHubWorkflowsRule
func NewGitHubWorkflowsRule(config GitHubWorkflowsRule) *GitHubWorkflowsRule {
	return &GitHubWorkflowsRule{
		RequireTests:          config.RequireTests,
		RequireSecurity:       config.RequireSecurity,
		RequireQuality:        config.RequireQuality,
		RequiredJobs:          config.RequiredJobs,
		RequiredTriggers:      config.RequiredTriggers,
		AllowMissing:          config.AllowMissing,
		RequireLogCommits:     config.RequireLogCommits,
		RequireRepomixArtifact: config.RequireRepomixArtifact,
	}
}

// generateWorkflowAutoFix generates auto-fix workflows based on detected languages
func (r *GitHubWorkflowsRule) generateWorkflowAutoFix(rootDir string, workflowType string) *AutoFix {
	// Detect languages in the project
	languages, err := lang.DetectInDirectory(rootDir)
	if err != nil || len(languages) == 0 {
		// If detection fails, return nil (no auto-fix)
		return nil
	}

	// Get the primary language
	primaryLang := languages[0].Language

	var content string
	var fileName string

	switch workflowType {
	case "test":
		content = r.generateTestWorkflow(primaryLang, rootDir)
		fileName = "ci.yml"
	case "security":
		content = r.generateSecurityWorkflow(primaryLang)
		fileName = "security.yml"
	case "quality":
		content = r.generateQualityWorkflow(primaryLang)
		fileName = "quality.yml"
	default:
		return nil
	}

	if content == "" {
		return nil
	}

	return &AutoFix{
		FilePath: filepath.Join(rootDir, ".github", "workflows", fileName),
		Content:  content,
	}
}

// generateTestWorkflow generates a test/CI workflow based on language
func (r *GitHubWorkflowsRule) generateTestWorkflow(language lang.Language, rootDir string) string {
	switch language {
	case lang.Go:
		return `name: CI - Tests

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Run Tests
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Cache Go modules
        uses: actions/cache@v4
        with:
          path: ~/go/pkg/mod
          key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
          restore-keys: |
            ${{ runner.os }}-go-

      - name: Download dependencies
        run: go mod download

      - name: Run tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.out
          flags: unittests
          name: codecov-umbrella
          fail_ci_if_error: false
`

	case lang.Python:
		return `name: CI - Tests

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Run Tests
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ['3.9', '3.10', '3.11']

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Python ${{ matrix.python-version }}
        uses: actions/setup-python@v5
        with:
          python-version: ${{ matrix.python-version }}

      - name: Cache pip packages
        uses: actions/cache@v4
        with:
          path: ~/.cache/pip
          key: ${{ runner.os }}-pip-${{ hashFiles('**/requirements.txt', '**/pyproject.toml') }}
          restore-keys: |
            ${{ runner.os }}-pip-

      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install pytest pytest-cov
          if [ -f requirements.txt ]; then pip install -r requirements.txt; fi
          if [ -f pyproject.toml ]; then pip install -e .[dev]; fi

      - name: Run tests
        run: pytest --cov=. --cov-report=xml

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage.xml
          flags: unittests
          name: codecov-umbrella
          fail_ci_if_error: false
`

	case lang.TypeScript, lang.JavaScript, lang.React:
		return `name: CI - Tests

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Run Tests
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: |
          if [ -f "yarn.lock" ]; then
            npm install -g yarn
            yarn install --frozen-lockfile
          elif [ -f "pnpm-lock.yaml" ]; then
            npm install -g pnpm
            pnpm install --frozen-lockfile
          else
            npm ci
          fi

      - name: Run tests
        run: npm test -- --coverage

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./coverage/coverage-final.json
          flags: unittests
          name: codecov-umbrella
          fail_ci_if_error: false
`

	case lang.Rust:
		return `name: CI - Tests

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Run Tests
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Rust
        uses: dtolnay/rust-toolchain@stable

      - name: Cache cargo registry
        uses: actions/cache@v4
        with:
          path: ~/.cargo/registry
          key: ${{ runner.os }}-cargo-registry-${{ hashFiles('**/Cargo.lock') }}

      - name: Cache cargo index
        uses: actions/cache@v4
        with:
          path: ~/.cargo/git
          key: ${{ runner.os }}-cargo-index-${{ hashFiles('**/Cargo.lock') }}

      - name: Cache cargo build
        uses: actions/cache@v4
        with:
          path: target
          key: ${{ runner.os }}-cargo-build-target-${{ hashFiles('**/Cargo.lock') }}

      - name: Run tests
        run: cargo test --verbose

      - name: Run tests with coverage
        run: |
          cargo install cargo-tarpaulin
          cargo tarpaulin --out Xml

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          file: ./cobertura.xml
          flags: unittests
          name: codecov-umbrella
          fail_ci_if_error: false
`

	case lang.Java:
		return `name: CI - Tests

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  test:
    name: Run Tests
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up JDK 17
        uses: actions/setup-java@v4
        with:
          java-version: '17'
          distribution: 'temurin'

      - name: Setup Gradle
        if: hashFiles('**/*.gradle*', '**/gradle-wrapper.properties') != ''
        uses: gradle/actions/setup-gradle@v3

      - name: Cache Maven packages
        if: hashFiles('**/pom.xml') != ''
        uses: actions/cache@v4
        with:
          path: ~/.m2
          key: ${{ runner.os }}-m2-${{ hashFiles('**/pom.xml') }}
          restore-keys: ${{ runner.os }}-m2

      - name: Build and test with Gradle
        if: hashFiles('**/*.gradle*', '**/gradle-wrapper.properties') != ''
        run: ./gradlew build test

      - name: Build and test with Maven
        if: hashFiles('**/pom.xml') != ''
        run: |
          mvn -B package --file pom.xml
          mvn test

      - name: Generate coverage report (Maven)
        if: hashFiles('**/pom.xml') != ''
        run: mvn jacoco:report || echo "JaCoCo not configured, skipping coverage"
        continue-on-error: true

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./target/site/jacoco/jacoco.xml,./build/reports/jacoco/test/jacocoTestReport.xml
          flags: unittests
          name: codecov-umbrella
          fail_ci_if_error: false
`

	default:
		return ""
	}
}

// generateSecurityWorkflow generates a security scanning workflow
func (r *GitHubWorkflowsRule) generateSecurityWorkflow(language lang.Language) string {
	// Map language to CodeQL language identifier
	codeqlLang := mapLanguageToCodeQL(language)

	return fmt.Sprintf(`name: Security Scanning

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]
  schedule:
    - cron: '0 0 * * 1'  # Weekly on Monday

jobs:
  codeql:
    name: CodeQL Analysis
    runs-on: ubuntu-latest
    permissions:
      actions: read
      contents: read
      security-events: write

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Initialize CodeQL
        uses: github/codeql-action/init@v3
        with:
          languages: %s

      - name: Autobuild
        uses: github/codeql-action/autobuild@v3

      - name: Perform CodeQL Analysis
        uses: github/codeql-action/analyze@v3

  dependency-scan:
    name: Dependency Scanning
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run Trivy vulnerability scanner
        uses: aquasecurity/trivy-action@0.28.0
        with:
          scan-type: 'fs'
          scan-ref: '.'
          format: 'sarif'
          output: 'trivy-results.sarif'

      - name: Upload Trivy results to GitHub Security
        uses: github/codeql-action/upload-sarif@v3
        with:
          sarif_file: 'trivy-results.sarif'
`, codeqlLang)
}

// mapLanguageToCodeQL maps internal language to CodeQL language identifier
func mapLanguageToCodeQL(language lang.Language) string {
	switch language {
	case lang.Go:
		return "go"
	case lang.Python:
		return "python"
	case lang.TypeScript, lang.JavaScript, lang.React:
		return "javascript"
	case lang.Java:
		return "java"
	case lang.CSharp:
		return "csharp"
	case lang.Ruby:
		return "ruby"
	default:
		return "go" // Default fallback
	}
}

// generateQualityWorkflow generates a code quality workflow based on language
func (r *GitHubWorkflowsRule) generateQualityWorkflow(language lang.Language) string {
	switch language {
	case lang.Go:
		return `name: Code Quality

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  lint:
    name: Lint Code
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Run golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
          args: --timeout=5m

      - name: Run go vet
        run: go vet ./...

      - name: Run go fmt
        run: |
          if [ -n "$(gofmt -s -l .)" ]; then
            echo "Go code is not formatted:"
            gofmt -s -d .
            exit 1
          fi
`

	case lang.Python:
		return `name: Code Quality

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  lint:
    name: Lint Code
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: '3.11'

      - name: Install dependencies
        run: |
          python -m pip install --upgrade pip
          pip install flake8 black pylint mypy

      - name: Run flake8
        run: flake8 . --count --select=E9,F63,F7,F82 --show-source --statistics

      - name: Run black
        run: black --check .

      - name: Run pylint
        run: pylint **/*.py
        continue-on-error: true

      - name: Run mypy
        run: mypy .
        continue-on-error: true
`

	case lang.TypeScript, lang.JavaScript, lang.React:
		return `name: Code Quality

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  lint:
    name: Lint Code
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Node.js
        uses: actions/setup-node@v4
        with:
          node-version: '20'
          cache: 'npm'

      - name: Install dependencies
        run: |
          if [ -f "yarn.lock" ]; then
            npm install -g yarn
            yarn install --frozen-lockfile
          elif [ -f "pnpm-lock.yaml" ]; then
            npm install -g pnpm
            pnpm install --frozen-lockfile
          else
            npm ci
          fi

      - name: Run ESLint
        run: npm run lint || npx eslint .

      - name: Run Prettier
        run: npx prettier --check .

      - name: Type check
        run: npm run type-check || npx tsc --noEmit
        continue-on-error: true
`

	case lang.Rust:
		return `name: Code Quality

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  lint:
    name: Lint Code
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Set up Rust
        uses: dtolnay/rust-toolchain@stable
        with:
          components: rustfmt, clippy

      - name: Run rustfmt
        run: cargo fmt -- --check

      - name: Run clippy
        run: cargo clippy -- -D warnings
`

	default:
		return `name: Code Quality

on:
  push:
    branches: [ main, master ]
  pull_request:
    branches: [ main, master ]

jobs:
  lint:
    name: Lint Code
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Run basic checks
        run: echo "Configure linting for your language"
`
	}
}
