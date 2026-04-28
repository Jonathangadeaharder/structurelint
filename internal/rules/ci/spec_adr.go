// Package ci provides CI/CD enforcement rules.
package ci

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// SpecADRRule enforces the presence and structure of specification and ADR documents.
// It checks for:
// - Designated folders for specs and ADRs (e.g., docs/specs, docs/adr)
// - Feature specification template compliance
// - ADR (Architecture Decision Record) template compliance
type SpecADRRule struct {
	RequireSpecFolder    bool     `yaml:"require-spec-folder"`
	RequireADRFolder     bool     `yaml:"require-adr-folder"`
	SpecFolderPaths      []string `yaml:"spec-folder-paths"`      // e.g., ["docs/specs", "specifications"]
	ADRFolderPaths       []string `yaml:"adr-folder-paths"`       // e.g., ["docs/adr", "docs/decisions"]
	EnforceSpecTemplate  bool     `yaml:"enforce-spec-template"`  // Validate spec files follow template
	EnforceADRTemplate   bool     `yaml:"enforce-adr-template"`   // Validate ADR files follow template
	SpecFilePatterns     []string `yaml:"spec-file-patterns"`     // e.g., ["*-spec.md", "feature-*.md"]
	ADRFilePatterns      []string `yaml:"adr-file-patterns"`      // e.g., ["ADR-*.md", "*-decision.md"]
}

// Name returns the rule name
func (r *SpecADRRule) Name() string {
	return "spec-adr-enforcement"
}

// Check validates specification and ADR requirements
func (r *SpecADRRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation

	// Use configured values or defaults (don't mutate receiver)
	specFolderPaths := r.SpecFolderPaths
	if len(specFolderPaths) == 0 {
		specFolderPaths = []string{"docs/specs", "specifications", "specs"}
	}
	adrFolderPaths := r.ADRFolderPaths
	if len(adrFolderPaths) == 0 {
		adrFolderPaths = []string{"docs/adr", "docs/decisions", "adr"}
	}
	specFilePatterns := r.SpecFilePatterns
	if len(specFilePatterns) == 0 {
		specFilePatterns = []string{"*-spec.md", "feature-*.md"}
	}
	adrFilePatterns := r.ADRFilePatterns
	if len(adrFilePatterns) == 0 {
		adrFilePatterns = []string{"ADR-*.md", "*-decision.md", "adr-*.md"}
	}

	// Check for spec folder existence
	if r.RequireSpecFolder {
		if !r.hasFolderPath(dirs, specFolderPaths) {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    ".",
				Message: r.formatMissingSpecFolderMessage(specFolderPaths),
			})
		}
	}

	// Check for ADR folder existence
	if r.RequireADRFolder {
		if !r.hasFolderPath(dirs, adrFolderPaths) {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    ".",
				Message: r.formatMissingADRFolderMessage(adrFolderPaths),
			})
		}
	}

	// Validate spec files follow template
	if r.EnforceSpecTemplate {
		specViolations := r.validateSpecFiles(files, specFilePatterns)
		violations = append(violations, specViolations...)
	}

	// Validate ADR files follow template
	if r.EnforceADRTemplate {
		adrViolations := r.validateADRFiles(files, adrFilePatterns)
		violations = append(violations, adrViolations...)
	}

	return violations
}

// hasFolderPath checks if any of the expected folder paths exist
func (r *SpecADRRule) hasFolderPath(dirs map[string]*walker.DirInfo, paths []string) bool {
	for dirPath := range dirs {
		normalizedPath := filepath.ToSlash(dirPath)
		for _, expectedPath := range paths {
			expectedNormalized := filepath.ToSlash(expectedPath)
			if normalizedPath == expectedNormalized || strings.HasSuffix(normalizedPath, "/"+expectedNormalized) {
				return true
			}
		}
	}
	return false
}

// validateSpecFiles validates that specification files follow the required template
func (r *SpecADRRule) validateSpecFiles(files []walker.FileInfo, specFilePatterns []string) []rules.Violation {
	var violations []rules.Violation

	// Required sections in feature specification template
	requiredSections := []string{
		"# Feature Specification:",
		"## User Scenarios & Testing",
		"## Requirements",
		"### Functional Requirements",
		"## Success Criteria",
	}

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Check if file matches spec patterns
		if !r.matchesPatterns(filepath.Base(file.Path), specFilePatterns) {
			continue
		}

		// Read file content
		content, err := os.ReadFile(file.Path)
		if err != nil {
			continue
		}

		contentStr := string(content)

		// Check for required sections
		missingSections := []string{}
		for _, section := range requiredSections {
			if !strings.Contains(contentStr, section) {
				missingSections = append(missingSections, section)
			}
		}

		if len(missingSections) > 0 {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: r.formatMissingSpecSectionsMessage(missingSections),
			})
		}

		// Validate that user stories have priorities
		if strings.Contains(contentStr, "## User Scenarios & Testing") {
			if !strings.Contains(contentStr, "(Priority: P") && !strings.Contains(contentStr, "(Priority:P") {
				violations = append(violations, rules.Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: "Feature specification must include prioritized user stories (e.g., 'Priority: P1', 'Priority: P2'). Each story should be independently testable.",
				})
			}
		}

		// Validate functional requirements format
		if strings.Contains(contentStr, "### Functional Requirements") {
			if !strings.Contains(contentStr, "FR-") {
				violations = append(violations, rules.Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: "Functional requirements must use FR-### format (e.g., 'FR-001: System MUST...').",
				})
			}
		}

		// Validate success criteria format
		if strings.Contains(contentStr, "## Success Criteria") {
			if !strings.Contains(contentStr, "SC-") {
				violations = append(violations, rules.Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: "Success criteria must use SC-### format (e.g., 'SC-001: Users can...').",
				})
			}
		}
	}

	return violations
}

// validateADRFiles validates that ADR files follow the required template
func (r *SpecADRRule) validateADRFiles(files []walker.FileInfo, adrFilePatterns []string) []rules.Violation {
	var violations []rules.Violation

	// Required sections in ADR template
	requiredSections := []string{
		"## Context and Problem Statement",
		"## Considered Options",
		"## Decision Outcome",
	}

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Check if file matches ADR patterns
		if !r.matchesPatterns(filepath.Base(file.Path), adrFilePatterns) {
			continue
		}

		// Read file content
		content, err := os.ReadFile(file.Path)
		if err != nil {
			continue
		}

		contentStr := string(content)

		// Check for required sections
		missingSections := []string{}
		for _, section := range requiredSections {
			if !strings.Contains(contentStr, section) {
				missingSections = append(missingSections, section)
			}
		}

		if len(missingSections) > 0 {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: r.formatMissingADRSectionsMessage(missingSections),
			})
		}

		// Validate status field exists
		if !strings.Contains(contentStr, "status:") && !strings.Contains(contentStr, "Status:") {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: "ADR must include a 'status' field (e.g., 'status: proposed', 'status: accepted').",
			})
		}

		// Validate date field exists
		if !strings.Contains(contentStr, "date:") && !strings.Contains(contentStr, "Date:") {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: "ADR must include a 'date' field in YYYY-MM-DD format.",
			})
		}
	}

	return violations
}

// matchesPatterns checks if a filename matches any of the given patterns
func (r *SpecADRRule) matchesPatterns(filename string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, filename)
		if err == nil && matched {
			return true
		}

		// Also check case-insensitive match
		matched, err = filepath.Match(strings.ToLower(pattern), strings.ToLower(filename))
		if err == nil && matched {
			return true
		}
	}
	return false
}

// formatMissingSpecFolderMessage creates an error message for missing spec folder
func (r *SpecADRRule) formatMissingSpecFolderMessage(specFolderPaths []string) string {
	return "No specification folder found. Expected one of: " + strings.Join(specFolderPaths, ", ") + ". " +
		"Create a dedicated folder for feature specifications to maintain clear documentation structure."
}

// formatMissingADRFolderMessage creates an error message for missing ADR folder
func (r *SpecADRRule) formatMissingADRFolderMessage(adrFolderPaths []string) string {
	return "No ADR (Architecture Decision Record) folder found. Expected one of: " + strings.Join(adrFolderPaths, ", ") + ". " +
		"Create a dedicated folder for ADRs to document architectural decisions."
}

// formatMissingSpecSectionsMessage creates an error message for missing spec sections
func (r *SpecADRRule) formatMissingSpecSectionsMessage(missingSections []string) string {
	return "Feature specification is missing required sections:\n" +
		"  - " + strings.Join(missingSections, "\n  - ") + "\n\n" +
		"Each feature spec must include:\n" +
		"  1. User Scenarios & Testing (with prioritized, independently testable user stories)\n" +
		"  2. Functional Requirements (FR-### format)\n" +
		"  3. Success Criteria (SC-### format with measurable outcomes)"
}

// formatMissingADRSectionsMessage creates an error message for missing ADR sections
func (r *SpecADRRule) formatMissingADRSectionsMessage(missingSections []string) string {
	return "ADR is missing required sections:\n" +
		"  - " + strings.Join(missingSections, "\n  - ") + "\n\n" +
		"Each ADR must include:\n" +
		"  1. Context and Problem Statement\n" +
		"  2. Considered Options\n" +
		"  3. Decision Outcome (with chosen option and justification)"
}

// NewSpecADRRule creates a new SpecADRRule
func NewSpecADRRule(config SpecADRRule) *SpecADRRule {
	return &SpecADRRule{
		RequireSpecFolder:    config.RequireSpecFolder,
		RequireADRFolder:     config.RequireADRFolder,
		SpecFolderPaths:      config.SpecFolderPaths,
		ADRFolderPaths:       config.ADRFolderPaths,
		EnforceSpecTemplate:  config.EnforceSpecTemplate,
		EnforceADRTemplate:   config.EnforceADRTemplate,
		SpecFilePatterns:     config.SpecFilePatterns,
		ADRFilePatterns:      config.ADRFilePatterns,
	}
}
