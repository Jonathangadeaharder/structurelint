// Package rules provides design-by-contract framework enforcement rules.
package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// ContractFrameworkRule enforces the use of established design-by-contract frameworks.
// It checks for:
// - Python: icontract or deal
// - Rust: Native contracts (assertions, type system)
// - TypeScript: Zod for runtime validation
// - Go: gocontracts or standard assertions
// - Java: Jakarta Bean Validation 3.0 (JSR-380)
// - C#: Nullable reference types (Code Contracts is deprecated)
// - C++: Native contracts (targeting C++26)
type ContractFrameworkRule struct {
	RequirePython     bool     `yaml:"require-python"`
	RequireRust       bool     `yaml:"require-rust"`
	RequireTypeScript bool     `yaml:"require-typescript"`
	RequireGo         bool     `yaml:"require-go"`
	RequireJava       bool     `yaml:"require-java"`
	RequireCSharp     bool     `yaml:"require-csharp"`
	RequireCPlusPlus  bool     `yaml:"require-cplusplus"`
	CustomFrameworks  []string `yaml:"custom-frameworks"`
}

// ContractFrameworkConfig defines the expected contract frameworks for a language
type ContractFrameworkConfig struct {
	Language           string
	PackageFiles       []string // Files that should contain the dependency
	ImportPatterns     []string // Import/using statements to look for
	RecommendedTools   []string // List of recommended frameworks
	DeprecatedWarnings []string // Warnings about deprecated tools
}

// Name returns the rule name
func (r *ContractFrameworkRule) Name() string {
	return "contract-framework"
}

// Check validates contract framework requirements
func (r *ContractFrameworkRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Detect which languages are present in the project
	languages := r.detectLanguages(files)

	// Define contract framework configurations for each language
	frameworkConfigs := r.getFrameworkConfigs()

	// Check each language
	if r.RequirePython && languages["python"] {
		pythonViolations := r.checkLanguageContracts(files, frameworkConfigs["python"])
		violations = append(violations, pythonViolations...)
	}

	if r.RequireRust && languages["rust"] {
		rustViolations := r.checkLanguageContracts(files, frameworkConfigs["rust"])
		violations = append(violations, rustViolations...)
	}

	if r.RequireTypeScript && languages["typescript"] {
		tsViolations := r.checkLanguageContracts(files, frameworkConfigs["typescript"])
		violations = append(violations, tsViolations...)
	}

	if r.RequireGo && languages["go"] {
		goViolations := r.checkLanguageContracts(files, frameworkConfigs["go"])
		violations = append(violations, goViolations...)
	}

	if r.RequireJava && languages["java"] {
		javaViolations := r.checkLanguageContracts(files, frameworkConfigs["java"])
		violations = append(violations, javaViolations...)
	}

	if r.RequireCSharp && languages["csharp"] {
		csharpViolations := r.checkLanguageContracts(files, frameworkConfigs["csharp"])
		violations = append(violations, csharpViolations...)
	}

	if r.RequireCPlusPlus && languages["cplusplus"] {
		cppViolations := r.checkLanguageContracts(files, frameworkConfigs["cplusplus"])
		violations = append(violations, cppViolations...)
	}

	return violations
}

// detectLanguages detects which programming languages are present in the project
func (r *ContractFrameworkRule) detectLanguages(files []walker.FileInfo) map[string]bool {
	languages := make(map[string]bool)

	for _, file := range files {
		if file.IsDir {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Path))
		switch ext {
		case ".py":
			languages["python"] = true
		case ".rs":
			languages["rust"] = true
		case ".ts", ".tsx":
			languages["typescript"] = true
		case ".js", ".jsx":
			// Treat JavaScript as TypeScript for contract enforcement
			languages["typescript"] = true
		case ".go":
			languages["go"] = true
		case ".java":
			languages["java"] = true
		case ".cs":
			languages["csharp"] = true
		case ".cpp", ".cc", ".cxx", ".hpp", ".h":
			languages["cplusplus"] = true
		}
	}

	return languages
}

// getFrameworkConfigs returns the expected contract framework configurations for each language
func (r *ContractFrameworkRule) getFrameworkConfigs() map[string]ContractFrameworkConfig {
	return map[string]ContractFrameworkConfig{
		"python": {
			Language: "Python",
			PackageFiles: []string{
				"requirements.txt",
				"pyproject.toml",
				"Pipfile",
				"poetry.lock",
			},
			ImportPatterns: []string{
				"import icontract",
				"from icontract import",
				"import deal",
				"from deal import",
				"@icontract.require",
				"@icontract.ensure",
				"@deal.pre",
				"@deal.post",
			},
			RecommendedTools: []string{
				"icontract (rich error messages, strong inheritance support)",
				"deal (static analysis, formal verification)",
			},
		},
		"rust": {
			Language: "Rust",
			PackageFiles: []string{
				"Cargo.toml",
			},
			ImportPatterns: []string{
				"assert!",
				"debug_assert!",
				"assert_eq!",
				"assert_ne!",
				// Type system enforcement is implicit
			},
			RecommendedTools: []string{
				"Native assertions (assert!, debug_assert!)",
				"Type system (Result<T, E>, Option<T>)",
				"Note: Native contract support is in development for Rust",
			},
		},
		"typescript": {
			Language: "TypeScript",
			PackageFiles: []string{
				"package.json",
				"package-lock.json",
				"yarn.lock",
				"pnpm-lock.yaml",
			},
			ImportPatterns: []string{
				"import * as z from \"zod\"",
				"import { z } from \"zod\"",
				"from \"zod\"",
				".parse(",
				".safeParse(",
				"z.object(",
				"z.string(",
				"z.number(",
			},
			RecommendedTools: []string{
				"Zod (industry-standard for runtime validation and type-safe schemas)",
			},
		},
		"go": {
			Language: "Go",
			PackageFiles: []string{
				"go.mod",
				"go.sum",
			},
			ImportPatterns: []string{
				"github.com/s-kostyaev/gocontracts",
				"//go:generate gocontracts",
				// Standard Go assertions in tests
				"if .* != .* {",
				"panic(",
			},
			RecommendedTools: []string{
				"gocontracts (generates checks from doc comments)",
				"Standard Go assertions and interfaces",
			},
		},
		"java": {
			Language: "Java",
			PackageFiles: []string{
				"pom.xml",
				"build.gradle",
				"build.gradle.kts",
			},
			ImportPatterns: []string{
				"import jakarta.validation",
				"import javax.validation",
				"@Valid",
				"@NotNull",
				"@NotEmpty",
				"@NotBlank",
				"@Size",
				"@Min",
				"@Max",
				"@Pattern",
			},
			RecommendedTools: []string{
				"Jakarta Bean Validation 3.0 (JSR-380)",
				"Hibernate Validator (implementation with Spring Boot integration)",
			},
		},
		"csharp": {
			Language: "C#",
			PackageFiles: []string{
				"*.csproj",
				"*.sln",
			},
			ImportPatterns: []string{
				"#nullable enable",
				"ArgumentNullException.ThrowIfNull",
				"Debug.Assert",
				"Contract.Requires", // Deprecated but we check for it to warn
			},
			RecommendedTools: []string{
				"Nullable reference types (enable with #nullable enable)",
				"ArgumentNullException.ThrowIfNull for runtime checks",
				"Note: Microsoft Code Contracts is deprecated in .NET Core+",
			},
			DeprecatedWarnings: []string{
				"Microsoft Code Contracts is deprecated and unsupported in .NET Core+",
			},
		},
		"cplusplus": {
			Language: "C++",
			PackageFiles: []string{
				"CMakeLists.txt",
				"Makefile",
			},
			ImportPatterns: []string{
				"assert(",
				"static_assert(",
				"#include <cassert>",
				// C++26 contracts are not yet available
			},
			RecommendedTools: []string{
				"Standard assertions (assert, static_assert)",
				"Note: Native contract support targeting C++26",
				"Avoid third-party libraries until standard is available",
			},
		},
	}
}

// checkLanguageContracts checks if the required contract frameworks are configured for a language
func (r *ContractFrameworkRule) checkLanguageContracts(files []walker.FileInfo, config ContractFrameworkConfig) []Violation {
	var violations []Violation

	// Check for package files (dependency declarations)
	hasPackageFile := r.hasPackageFile(files, config.PackageFiles)

	// Check for import patterns (actual usage in code)
	hasImportPattern := r.hasImportPattern(files, config.ImportPatterns, config.Language)

	// If neither package files nor import patterns are found, report a violation
	if !hasPackageFile && !hasImportPattern {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    ".",
			Message: r.formatMissingContractMessage(config),
		})
	}

	// Check for deprecated frameworks and warn
	r.checkDeprecatedFrameworks(files, config, &violations)

	return violations
}

// hasPackageFile checks if any of the expected package files exist and contain contract dependencies
func (r *ContractFrameworkRule) hasPackageFile(files []walker.FileInfo, packageFiles []string) bool {
	for _, file := range files {
		if file.IsDir {
			continue
		}

		filename := filepath.Base(file.Path)
		for _, expectedFile := range packageFiles {
			// Handle glob patterns like *.csproj
			if strings.Contains(expectedFile, "*") {
				matched, err := filepath.Match(expectedFile, filename)
				if err != nil {
					continue
				}
				if matched {
					// Check if file contains contract-related dependencies
					if r.fileContainsContractDependency(file.Path) {
						return true
					}
				}
			} else if filename == expectedFile {
				// Check if file contains contract-related dependencies
				if r.fileContainsContractDependency(file.Path) {
					return true
				}
			}
		}
	}

	return false
}

// fileContainsContractDependency checks if a dependency file contains contract-related packages
func (r *ContractFrameworkRule) fileContainsContractDependency(path string) bool {
	content, err := os.ReadFile(path)
	if err != nil {
		return false
	}

	contentStr := strings.ToLower(string(content))

	// Contract-related package names
	contractPackages := []string{
		"icontract",
		"deal",
		"zod",
		"gocontracts",
		"jakarta.validation",
		"javax.validation",
		"hibernate-validator",
	}

	for _, pkg := range contractPackages {
		if strings.Contains(contentStr, pkg) {
			return true
		}
	}

	return false
}

// hasImportPattern checks if any source files contain contract-related imports
func (r *ContractFrameworkRule) hasImportPattern(files []walker.FileInfo, patterns []string, language string) bool {
	// Determine which file extensions to check based on language
	var extensions []string
	switch language {
	case "Python":
		extensions = []string{".py"}
	case "Rust":
		extensions = []string{".rs"}
	case "TypeScript":
		extensions = []string{".ts", ".tsx", ".js", ".jsx"}
	case "Go":
		extensions = []string{".go"}
	case "Java":
		extensions = []string{".java"}
	case "C#":
		extensions = []string{".cs"}
	case "C++":
		extensions = []string{".cpp", ".cc", ".cxx", ".hpp", ".h"}
	}

	for _, file := range files {
		if file.IsDir {
			continue
		}

		ext := strings.ToLower(filepath.Ext(file.Path))
		validExt := false
		for _, validExtension := range extensions {
			if ext == validExtension {
				validExt = true
				break
			}
		}
		if !validExt {
			continue
		}

		content, err := os.ReadFile(file.Path)
		if err != nil {
			continue
		}

		contentStr := string(content)
		for _, pattern := range patterns {
			if strings.Contains(contentStr, pattern) {
				return true
			}
		}
	}

	return false
}

// checkDeprecatedFrameworks checks for usage of deprecated contract frameworks
func (r *ContractFrameworkRule) checkDeprecatedFrameworks(files []walker.FileInfo, config ContractFrameworkConfig, violations *[]Violation) {
	// Check for deprecated Code Contracts in C#
	if config.Language == "C#" {
		for _, file := range files {
			if file.IsDir || !strings.HasSuffix(file.Path, ".cs") {
				continue
			}

			content, err := os.ReadFile(file.Path)
			if err != nil {
				continue
			}

			if strings.Contains(string(content), "System.Diagnostics.Contracts") {
				*violations = append(*violations, Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: "DEPRECATED: Microsoft Code Contracts is deprecated and unsupported in .NET Core+. Use nullable reference types (#nullable enable) and ArgumentNullException.ThrowIfNull instead.",
				})
			}
		}
	}
}

// formatMissingContractMessage creates a detailed error message for missing contract framework
func (r *ContractFrameworkRule) formatMissingContractMessage(config ContractFrameworkConfig) string {
	message := fmt.Sprintf(
		"No design-by-contract framework found for %s. Recommended frameworks:\n",
		config.Language,
	)

	for _, tool := range config.RecommendedTools {
		message += fmt.Sprintf("  - %s\n", tool)
	}

	message += "\nDesign-by-contract enables better component contracts, runtime validation, and improved code reliability."

	if len(config.DeprecatedWarnings) > 0 {
		message += "\n\nIMPORTANT NOTES:\n"
		for _, warning := range config.DeprecatedWarnings {
			message += fmt.Sprintf("  - %s\n", warning)
		}
	}

	return strings.TrimSpace(message)
}

// NewContractFrameworkRule creates a new ContractFrameworkRule
func NewContractFrameworkRule(config ContractFrameworkRule) *ContractFrameworkRule {
	return &ContractFrameworkRule{
		RequirePython:     config.RequirePython,
		RequireRust:       config.RequireRust,
		RequireTypeScript: config.RequireTypeScript,
		RequireGo:         config.RequireGo,
		RequireJava:       config.RequireJava,
		RequireCSharp:     config.RequireCSharp,
		RequireCPlusPlus:  config.RequireCPlusPlus,
		CustomFrameworks:  config.CustomFrameworks,
	}
}
