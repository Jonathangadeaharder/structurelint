// Package rules provides linter configuration enforcement rules.
package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
	"gopkg.in/yaml.v3"
)

// LinterConfigRule enforces the presence of linter configurations for various languages.
// It checks for:
// - Python: mypy, black, ruff, pylint, flake8
// - TypeScript: ESLint, Prettier, TSConfig
// - Go: golangci-lint, gofmt, go vet
// - HTML: HTMLHint, html-validate, prettier
// - CSS: stylelint, prettier
// - SQL: sqlfluff, sqlfmt
// - Rust: clippy, rustfmt
type LinterConfigRule struct {
	RequirePython     bool     `yaml:"require-python"`
	RequireTypeScript bool     `yaml:"require-typescript"`
	RequireGo         bool     `yaml:"require-go"`
	RequireHTML       bool     `yaml:"require-html"`
	RequireCSS        bool     `yaml:"require-css"`
	RequireSQL        bool     `yaml:"require-sql"`
	RequireRust       bool     `yaml:"require-rust"`
	CustomLinters     []string `yaml:"custom-linters"`
}

// LinterConfig defines the expected configuration files and workflow steps for a language
type LinterConfig struct {
	Language      string
	ConfigFiles   []string // One of these files must exist
	WorkflowSteps []string // Keywords to look for in GitHub workflows
}

// Name returns the rule name
func (r *LinterConfigRule) Name() string {
	return "linter-config"
}

// Check validates linter configuration requirements
func (r *LinterConfigRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Detect which languages are present in the project
	languages := r.detectLanguages(files)

	// Define linter configurations for each language
	linterConfigs := r.getLinterConfigs()

	// Check each language
	if r.RequirePython && languages["python"] {
		pythonViolations := r.checkLanguageLinters(files, linterConfigs["python"])
		violations = append(violations, pythonViolations...)
	}

	if r.RequireTypeScript && languages["typescript"] {
		tsViolations := r.checkLanguageLinters(files, linterConfigs["typescript"])
		violations = append(violations, tsViolations...)
	}

	if r.RequireGo && languages["go"] {
		goViolations := r.checkLanguageLinters(files, linterConfigs["go"])
		violations = append(violations, goViolations...)
	}

	if r.RequireHTML && languages["html"] {
		htmlViolations := r.checkLanguageLinters(files, linterConfigs["html"])
		violations = append(violations, htmlViolations...)
	}

	if r.RequireCSS && languages["css"] {
		cssViolations := r.checkLanguageLinters(files, linterConfigs["css"])
		violations = append(violations, cssViolations...)
	}

	if r.RequireSQL && languages["sql"] {
		sqlViolations := r.checkLanguageLinters(files, linterConfigs["sql"])
		violations = append(violations, sqlViolations...)
	}

	if r.RequireRust && languages["rust"] {
		rustViolations := r.checkLanguageLinters(files, linterConfigs["rust"])
		violations = append(violations, rustViolations...)
	}

	return violations
}

// detectLanguages detects which programming languages are present in the project
func (r *LinterConfigRule) detectLanguages(files []walker.FileInfo) map[string]bool {
	languages := make(map[string]bool)

	for _, file := range files {
		if file.IsDir {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Path))
		switch ext {
		case ".py":
			languages["python"] = true
		case ".ts", ".tsx":
			languages["typescript"] = true
		case ".js", ".jsx":
			// Treat JavaScript as TypeScript for linter enforcement
			// since ESLint, Prettier, and other TS tools work for both
			languages["typescript"] = true
		case ".go":
			languages["go"] = true
		case ".html", ".htm":
			languages["html"] = true
		case ".css", ".scss", ".sass", ".less":
			languages["css"] = true
		case ".sql":
			languages["sql"] = true
		case ".rs":
			languages["rust"] = true
		}
	}

	return languages
}

// getLinterConfigs returns the expected linter configurations for each language
func (r *LinterConfigRule) getLinterConfigs() map[string]LinterConfig {
	return map[string]LinterConfig{
		"python": {
			Language: "Python",
			ConfigFiles: []string{
				"pyproject.toml",      // Modern Python tool config (black, ruff, mypy, etc.)
				".flake8",             // flake8 config
				"setup.cfg",           // Legacy tool config
				".pylintrc",           // pylint config
				"mypy.ini",            // mypy config
				"ruff.toml",           // ruff config
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"mypy", "black", "ruff", "pylint", "flake8"},
		},
		"typescript": {
			Language: "TypeScript",
			ConfigFiles: []string{
				".eslintrc",           // ESLint config (any format)
				".eslintrc.json",
				".eslintrc.js",
				".eslintrc.yml",
				"eslint.config.js",    // ESLint flat config
				".prettierrc",         // Prettier config (any format)
				".prettierrc.json",
				".prettierrc.js",
				"prettier.config.js",
				"tsconfig.json",       // TypeScript compiler config
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"eslint", "prettier", "tsc"},
		},
		"go": {
			Language: "Go",
			ConfigFiles: []string{
				".golangci.yml",       // golangci-lint config
				".golangci.yaml",
				"golangci.yml",
				"golangci.yaml",
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"golangci-lint", "gofmt", "go vet", "go fmt"},
		},
		"html": {
			Language: "HTML",
			ConfigFiles: []string{
				".htmlhintrc",         // HTMLHint config
				".htmlvalidate.json",  // html-validate config
				".prettierrc",         // Prettier (also handles HTML)
				".prettierrc.json",
				".prettierrc.js",
				"prettier.config.js",
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"htmlhint", "html-validate", "prettier"},
		},
		"css": {
			Language: "CSS",
			ConfigFiles: []string{
				".stylelintrc",        // Stylelint config (any format)
				".stylelintrc.json",
				".stylelintrc.js",
				"stylelint.config.js",
				".prettierrc",         // Prettier (also handles CSS)
				".prettierrc.json",
				".prettierrc.js",
				"prettier.config.js",
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"stylelint", "prettier"},
		},
		"sql": {
			Language: "SQL",
			ConfigFiles: []string{
				".sqlfluff",           // SQLFluff config
				"setup.cfg",           // SQLFluff can use setup.cfg
				"pyproject.toml",      // SQLFluff can use pyproject.toml
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"sqlfluff", "sqlfmt", "sql-lint"},
		},
		"rust": {
			Language: "Rust",
			ConfigFiles: []string{
				"rustfmt.toml",        // rustfmt config
				".rustfmt.toml",
				"clippy.toml",         // clippy config
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"clippy", "rustfmt", "cargo clippy", "cargo fmt"},
		},
	}
}

// checkLanguageLinters checks if the required linters are configured for a language
func (r *LinterConfigRule) checkLanguageLinters(files []walker.FileInfo, config LinterConfig) []Violation {
	var violations []Violation

	// Check for configuration files
	hasConfigFile := r.hasConfigFile(files, config.ConfigFiles)

	// Check for workflow steps
	hasWorkflowStep := r.hasWorkflowStep(files, config.WorkflowSteps)

	// If neither config files nor workflow steps are found, report a violation
	if !hasConfigFile && !hasWorkflowStep {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    ".",
			Message: r.formatMissingLinterMessage(config),
		})
	}

	return violations
}

// hasConfigFile checks if any of the expected config files exist
func (r *LinterConfigRule) hasConfigFile(files []walker.FileInfo, configFiles []string) bool {
	for _, file := range files {
		if file.IsDir {
			continue
		}

		filename := filepath.Base(file.Path)
		for _, expectedFile := range configFiles {
			// Skip workflow patterns for now
			if strings.Contains(expectedFile, "workflows") {
				continue
			}

			// Handle exact matches and pattern matches
			if filename == expectedFile {
				return true
			}

			// Handle pattern matching (e.g., .eslintrc*)
			if strings.Contains(expectedFile, "*") {
				matched, err := filepath.Match(expectedFile, filename)
				if err != nil {
					// This should not happen with hardcoded patterns
					// but skip this pattern if there's an error
					continue
				}
				if matched {
					return true
				}
			}

			// Special handling for .eslintrc (can have various extensions)
			if strings.HasPrefix(expectedFile, ".eslintrc") && strings.HasPrefix(filename, ".eslintrc") {
				return true
			}

			// Special handling for .prettierrc (can have various extensions)
			if strings.HasPrefix(expectedFile, ".prettierrc") && strings.HasPrefix(filename, ".prettierrc") {
				return true
			}
		}
	}

	return false
}

// hasWorkflowStep checks if any GitHub workflow contains the expected linter steps
func (r *LinterConfigRule) hasWorkflowStep(files []walker.FileInfo, linterKeywords []string) bool {
	// Find workflow files
	workflowFiles := r.findWorkflowFiles(files)

	// Parse workflows and check for linter steps
	for _, workflowFile := range workflowFiles {
		data, err := os.ReadFile(workflowFile.Path)
		if err != nil {
			continue
		}

		// Parse the workflow file
		var workflow WorkflowFile
		if err := yaml.Unmarshal(data, &workflow); err != nil {
			continue
		}

		// Check each job for linter steps
		for _, job := range workflow.Jobs {
			for _, step := range job.Steps {
				stepContent := strings.ToLower(step.Name + " " + step.Run + " " + step.Uses)
				for _, keyword := range linterKeywords {
					if strings.Contains(stepContent, strings.ToLower(keyword)) {
						return true
					}
				}
			}
		}
	}

	return false
}

// findWorkflowFiles finds all GitHub workflow files
func (r *LinterConfigRule) findWorkflowFiles(files []walker.FileInfo) []walker.FileInfo {
	var workflowFiles []walker.FileInfo

	for _, file := range files {
		if file.IsDir {
			continue
		}

		normalizedPath := filepath.ToSlash(file.Path)
		if strings.Contains(normalizedPath, ".github/workflows") &&
			(strings.HasSuffix(file.Path, ".yml") || strings.HasSuffix(file.Path, ".yaml")) {
			workflowFiles = append(workflowFiles, file)
		}
	}

	return workflowFiles
}

// formatMissingLinterMessage creates a detailed error message for missing linter configuration
func (r *LinterConfigRule) formatMissingLinterMessage(config LinterConfig) string {
	var configFileNames []string
	var toolNames []string

	// Extract non-workflow config files
	for _, cf := range config.ConfigFiles {
		if !strings.Contains(cf, "workflows") {
			configFileNames = append(configFileNames, cf)
		}
	}

	// Extract tool names from workflow steps
	for _, step := range config.WorkflowSteps {
		toolNames = append(toolNames, step)
	}

	// Format config files display
	var configFilesDisplay string
	if len(configFileNames) == 0 {
		configFilesDisplay = "(no standard config files)"
	} else {
		// Show up to 5 config file names
		displayCount := min(5, len(configFileNames))
		configFilesDisplay = strings.Join(configFileNames[:displayCount], ", ")
		if len(configFileNames) > 5 {
			configFilesDisplay += ", ..."
		}
	}

	message := fmt.Sprintf(
		"No %s linter configuration found. Expected one of: %s, or a GitHub workflow running: %s",
		config.Language,
		configFilesDisplay,
		strings.Join(toolNames, ", "),
	)

	return message
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// NewLinterConfigRule creates a new LinterConfigRule
func NewLinterConfigRule(config LinterConfigRule) *LinterConfigRule {
	return &LinterConfigRule{
		RequirePython:     config.RequirePython,
		RequireTypeScript: config.RequireTypeScript,
		RequireGo:         config.RequireGo,
		RequireHTML:       config.RequireHTML,
		RequireCSS:        config.RequireCSS,
		RequireSQL:        config.RequireSQL,
		RequireRust:       config.RequireRust,
		CustomLinters:     config.CustomLinters,
	}
}
