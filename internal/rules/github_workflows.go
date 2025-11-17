// Package rules provides GitHub workflow enforcement rules.
package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
	"gopkg.in/yaml.v3"
)

// GitHubWorkflowsRule enforces the presence and configuration of GitHub Actions workflows
// for test execution, security scanning, and code quality checks.
type GitHubWorkflowsRule struct {
	RequireTests    bool     `yaml:"require-tests"`
	RequireSecurity bool     `yaml:"require-security"`
	RequireQuality  bool     `yaml:"require-quality"`
	RequiredJobs    []string `yaml:"required-jobs"`
	RequiredTriggers []string `yaml:"required-triggers"`
	AllowMissing    []string `yaml:"allow-missing"`
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
	Name string `yaml:"name"`
	Uses string `yaml:"uses"`
	Run  string `yaml:"run"`
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
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    ".github/workflows",
				Message: "No test/CI workflow found. Add a workflow that runs tests on pull requests and pushes.",
			})
		}
	}

	if r.RequireSecurity {
		if !r.hasWorkflowType(workflows, "security", []string{"security", "scan", "codeql", "dependabot"}) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    ".github/workflows",
				Message: "No security scanning workflow found. Add CodeQL, dependency scanning, or other security checks.",
			})
		}
	}

	if r.RequireQuality {
		if !r.hasWorkflowType(workflows, "quality", []string{"quality", "lint", "format", "coverage"}) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    ".github/workflows",
				Message: "No code quality workflow found. Add linting, formatting, or coverage checks.",
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

	// Normalize workflows dir path
	normalizedDir := filepath.ToSlash(workflowsDir)
	if normalizedDir == "./.github/workflows" {
		normalizedDir = ".github/workflows"
	}

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

	return violations
}

// NewGitHubWorkflowsRule creates a new GitHubWorkflowsRule
func NewGitHubWorkflowsRule(config GitHubWorkflowsRule) *GitHubWorkflowsRule {
	return &GitHubWorkflowsRule{
		RequireTests:     config.RequireTests,
		RequireSecurity:  config.RequireSecurity,
		RequireQuality:   config.RequireQuality,
		RequiredJobs:     config.RequiredJobs,
		RequiredTriggers: config.RequiredTriggers,
		AllowMissing:     config.AllowMissing,
	}
}
