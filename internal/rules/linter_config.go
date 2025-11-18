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
// - HTML: HTMLHint, html-validate, html-eslint, prettier
// - CSS: stylelint, prettier
// - SQL: sqlfluff, sqlfmt
// - Rust: clippy, rustfmt
// - Markdown: markdownlint
// - Java: Checkstyle, PMD, SpotBugs
// - C++: clang-format, clang-tidy, cppcheck
// - C#: EditorConfig, StyleCop, dotnet format
type LinterConfigRule struct {
	RequirePython     bool     `yaml:"require-python"`
	RequireTypeScript bool     `yaml:"require-typescript"`
	RequireGo         bool     `yaml:"require-go"`
	RequireHTML       bool     `yaml:"require-html"`
	RequireCSS        bool     `yaml:"require-css"`
	RequireSQL        bool     `yaml:"require-sql"`
	RequireRust       bool     `yaml:"require-rust"`
	RequireMarkdown   bool     `yaml:"require-markdown"`
	RequireJava       bool     `yaml:"require-java"`
	RequireCpp        bool     `yaml:"require-cpp"`
	RequireCSharp     bool     `yaml:"require-csharp"`
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

	if r.RequireMarkdown && languages["markdown"] {
		markdownViolations := r.checkLanguageLinters(files, linterConfigs["markdown"])
		violations = append(violations, markdownViolations...)
	}

	if r.RequireJava && languages["java"] {
		javaViolations := r.checkLanguageLinters(files, linterConfigs["java"])
		violations = append(violations, javaViolations...)
	}

	if r.RequireCpp && languages["cpp"] {
		cppViolations := r.checkLanguageLinters(files, linterConfigs["cpp"])
		violations = append(violations, cppViolations...)
	}

	if r.RequireCSharp && languages["csharp"] {
		csharpViolations := r.checkLanguageLinters(files, linterConfigs["csharp"])
		violations = append(violations, csharpViolations...)
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
		case ".md", ".markdown":
			languages["markdown"] = true
		case ".java":
			languages["java"] = true
		case ".cpp", ".cc", ".cxx", ".c++", ".hpp", ".hh", ".hxx", ".h++":
			languages["cpp"] = true
		case ".cs":
			languages["csharp"] = true
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
				".eslintrc",           // html-eslint via ESLint plugin
				".eslintrc.json",
				".eslintrc.js",
				".eslintrc.yml",
				"eslint.config.js",
				".prettierrc",         // Prettier (also handles HTML)
				".prettierrc.json",
				".prettierrc.js",
				"prettier.config.js",
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"htmlhint", "html-validate", "html-eslint", "eslint", "prettier"},
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
		"markdown": {
			Language: "Markdown",
			ConfigFiles: []string{
				".markdownlint.json",     // markdownlint config (JSON)
				".markdownlint.yaml",     // markdownlint config (YAML)
				".markdownlint.yml",
				".markdownlintrc",        // markdownlint config (rc format)
				".markdownlint-cli2.jsonc", // markdownlint-cli2 config
				".markdownlint-cli2.yaml",
				".markdownlint-cli2.cjs",
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"markdownlint", "markdownlint-cli", "markdownlint-cli2", "remark-lint"},
		},
		"java": {
			Language: "Java",
			ConfigFiles: []string{
				"checkstyle.xml",         // Checkstyle config
				"checkstyle-config.xml",
				".checkstyle.xml",
				"pmd.xml",                // PMD config
				"pmd-ruleset.xml",
				".pmd.xml",
				"spotbugs.xml",           // SpotBugs config
				"spotbugs-exclude.xml",
				".spotbugs.xml",
				"pom.xml",                // Maven with linter plugins
				"build.gradle",           // Gradle with linter plugins
				"build.gradle.kts",
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"checkstyle", "pmd", "spotbugs", "maven checkstyle", "gradle check"},
		},
		"cpp": {
			Language: "C++",
			ConfigFiles: []string{
				".clang-format",          // clang-format config
				"_clang-format",
				".clang-tidy",            // clang-tidy config
				"_clang-tidy",
				".cppcheck",              // cppcheck config
				"cppcheck.xml",
				"compile_commands.json",  // Compilation database for clang-tidy
				"CMakeLists.txt",         // CMake with linter integration
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"clang-format", "clang-tidy", "cppcheck", "cpplint"},
		},
		"csharp": {
			Language: "C#",
			ConfigFiles: []string{
				".editorconfig",          // EditorConfig (used by dotnet format and Roslyn)
				"stylecop.json",          // StyleCop Analyzers config
				".stylecop.json",
				"Directory.Build.props",  // MSBuild properties (can include analyzer config)
				"omnisharp.json",         // OmniSharp config (includes formatting)
				".github/workflows/*.yml", // Workflow with linting steps
			},
			WorkflowSteps: []string{"dotnet format", "stylecop", "roslyn analyzers", "csharpier"},
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
			Rule:        r.Name(),
			Path:        ".",
			Message:     r.formatMissingLinterMessage(config),
			Expected:    fmt.Sprintf("Linter configuration for %s", config.Language),
			Actual:      "No linter configuration found",
			Suggestions: r.generateAutoFixSuggestions(config),
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

// generateAutoFixSuggestions generates best practice configuration suggestions for a language
func (r *LinterConfigRule) generateAutoFixSuggestions(config LinterConfig) []string {
	var suggestions []string

	switch config.Language {
	case "Python":
		suggestions = []string{
			"Create pyproject.toml with: [tool.black], [tool.ruff], [tool.mypy] sections",
			"Install: pip install black ruff mypy",
			"Add pre-commit hook: pip install pre-commit && pre-commit install",
			"Example pyproject.toml: https://black.readthedocs.io/en/stable/usage_and_configuration/the_basics.html",
		}
	case "TypeScript":
		suggestions = []string{
			"Create .eslintrc.json with @typescript-eslint/recommended config",
			"Create .prettierrc with recommended settings",
			"Install: npm install --save-dev eslint @typescript-eslint/parser @typescript-eslint/eslint-plugin prettier",
			"Example config: https://typescript-eslint.io/getting-started",
		}
	case "Go":
		suggestions = []string{
			"Create .golangci.yml with recommended linters",
			"Install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest",
			"Enable linters: govet, errcheck, staticcheck, gosimple, unused",
			"Example config: https://golangci-lint.run/usage/configuration/",
		}
	case "HTML":
		suggestions = []string{
			"Create .htmlhintrc or .htmlvalidate.json for HTML linting",
			"For html-eslint: npm install --save-dev @html-eslint/eslint-plugin @html-eslint/parser",
			"Add to .eslintrc: extends: ['plugin:@html-eslint/recommended']",
			"Example config: https://html-eslint.org/docs/getting-started",
		}
	case "CSS":
		suggestions = []string{
			"Create .stylelintrc.json with stylelint-config-standard",
			"Install: npm install --save-dev stylelint stylelint-config-standard",
			"Add extends: 'stylelint-config-standard' to .stylelintrc.json",
			"Example config: https://stylelint.io/user-guide/get-started",
		}
	case "SQL":
		suggestions = []string{
			"Create .sqlfluff config file",
			"Install: pip install sqlfluff",
			"Add [sqlfluff] section to .sqlfluff or pyproject.toml",
			"Example config: https://docs.sqlfluff.com/en/stable/configuration.html",
		}
	case "Rust":
		suggestions = []string{
			"Create rustfmt.toml for formatting rules",
			"Create clippy.toml for linting rules",
			"Enable in CI: cargo fmt --check && cargo clippy -- -D warnings",
			"Example config: https://rust-lang.github.io/rustfmt/",
		}
	case "Markdown":
		suggestions = []string{
			"Create .markdownlint.json with recommended rules",
			"Install: npm install --save-dev markdownlint-cli2",
			"Example config: { 'default': true, 'MD013': { 'line_length': 120 } }",
			"Docs: https://github.com/DavidAnson/markdownlint",
		}
	case "Java":
		suggestions = []string{
			"Create checkstyle.xml with Google or Sun style guide",
			"Add to Maven: maven-checkstyle-plugin, pmd-maven-plugin, spotbugs-maven-plugin",
			"Add to Gradle: id 'checkstyle', id 'pmd', id 'com.github.spotbugs'",
			"Example: https://checkstyle.org/google_style.html",
		}
	case "C++":
		suggestions = []string{
			"Create .clang-format based on LLVM, Google, or Mozilla style",
			"Create .clang-tidy with checks: -*, modernize-*, readability-*, performance-*",
			"Generate compile_commands.json: cmake -DCMAKE_EXPORT_COMPILE_COMMANDS=ON",
			"Example: https://clang.llvm.org/docs/ClangFormatStyleOptions.html",
		}
	case "C#":
		suggestions = []string{
			"Create .editorconfig with dotnet_diagnostic rules",
			"Create stylecop.json for StyleCop.Analyzers",
			"Add to .csproj: <PackageReference Include='StyleCop.Analyzers' />",
			"Example: https://docs.microsoft.com/en-us/dotnet/fundamentals/code-analysis/",
		}
	default:
		suggestions = []string{
			fmt.Sprintf("Add linter configuration for %s", config.Language),
			"Create appropriate config file from the expected list",
			"Add linting steps to GitHub workflows",
		}
	}

	return suggestions
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
		RequireMarkdown:   config.RequireMarkdown,
		RequireJava:       config.RequireJava,
		RequireCpp:        config.RequireCpp,
		RequireCSharp:     config.RequireCSharp,
		CustomLinters:     config.CustomLinters,
	}
}
