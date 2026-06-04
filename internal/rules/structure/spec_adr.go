package structure

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	yaml "gopkg.in/yaml.v3"
)

// SpecADRRule validates feature specifications and architecture decision records.
type SpecADRRule struct {
	ruleName             string
	RequireSpecFolder    *bool    `json:"require-spec-folder"`
	RequireADRFolder     *bool    `json:"require-adr-folder"`
	EnforceSpecTemplate  *bool    `json:"enforce-spec-template"`
	EnforceADRTemplate   *bool    `json:"enforce-adr-template"`
	SpecFolderPaths      []string `json:"spec-folder-paths"`
	ADRFolderPaths       []string `json:"adr-folder-paths"`
	SpecFilePatterns     []string `json:"spec-file-patterns"`
	ADRFilePatterns      []string `json:"adr-file-patterns"`
	SpecRequiredHeadings []string `json:"spec-required-headings"`
	ADRRequiredHeadings  []string `json:"adr-required-headings"`
	ADRRequiredMetadata  []string `json:"adr-required-metadata"`
}

// Name returns the rule name.
func (r *SpecADRRule) Name() string {
	if r.ruleName != "" {
		return r.ruleName
	}
	return "spec-adr"
}

// Check validates files and folders against the specification and ADR configuration.
func (r *SpecADRRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation

	// Check for Spec folder existence
	if r.RequireSpecFolder != nil && *r.RequireSpecFolder {
		if !r.hasFolderPath(dirs, r.SpecFolderPaths) {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    ".",
				Message: r.formatMissingSpecFolderMessage(),
			})
		}
	}

	// Check for ADR folder existence
	if r.RequireADRFolder != nil && *r.RequireADRFolder {
		if !r.hasFolderPath(dirs, r.ADRFolderPaths) {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    ".",
				Message: r.formatMissingADRFolderMessage(),
			})
		}
	}

	// Filter ignored files before validating
	activeFiles := rules.FilterIgnoredFiles(files, r.Name())

	// Validate spec files follow template
	if r.EnforceSpecTemplate != nil && *r.EnforceSpecTemplate {
		specViolations := r.validateSpecFiles(activeFiles)
		violations = append(violations, specViolations...)
	}

	// Validate ADR files follow template
	if r.EnforceADRTemplate != nil && *r.EnforceADRTemplate {
		adrViolations := r.validateADRFiles(activeFiles)
		violations = append(violations, adrViolations...)
	}

	return violations
}

// hasFolderPath checks if any of the expected folder paths exist.
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

// validateSpecFiles validates that specification files follow the required template.
func (r *SpecADRRule) validateSpecFiles(files []walker.FileInfo) []rules.Violation {
	var violations []rules.Violation

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Check if file matches spec patterns
		if !r.matchesPatterns(file.Path, r.SpecFilePatterns) {
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
		for _, section := range r.SpecRequiredHeadings {
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

// validateADRFiles validates that ADR files follow the required template.
func (r *SpecADRRule) validateADRFiles(files []walker.FileInfo) []rules.Violation {
	var violations []rules.Violation

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Check if file matches ADR patterns
		if !r.matchesPatterns(file.Path, r.ADRFilePatterns) {
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
		for _, section := range r.ADRRequiredHeadings {
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

		// Validate required metadata fields exist in frontmatter
		frontmatter, err := parseFrontmatter(contentStr)
		if err != nil {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: "ADR must include valid YAML frontmatter (between '---' delimiters) containing required metadata fields.",
			})
			continue
		}

		for _, meta := range r.ADRRequiredMetadata {
			keyToFind := strings.ToLower(strings.TrimSuffix(meta, ":"))
			found := false
			for k := range frontmatter {
				if strings.ToLower(k) == keyToFind {
					found = true
					break
				}
			}
			if !found {
				violations = append(violations, rules.Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: fmt.Sprintf("ADR must include a '%s' field.", strings.TrimSuffix(meta, ":")),
				})
			}
		}
	}

	return violations
}

// parseFrontmatter extracts YAML frontmatter from markdown content.
func parseFrontmatter(content string) (map[string]interface{}, error) {
	trimmed := strings.TrimSpace(content)
	if !strings.HasPrefix(trimmed, "---") {
		return nil, fmt.Errorf("no frontmatter found")
	}

	lines := strings.Split(trimmed, "\n")
	if len(lines) < 2 {
		return nil, fmt.Errorf("too short for frontmatter")
	}

	if strings.TrimSpace(lines[0]) != "---" {
		return nil, fmt.Errorf("first line is not frontmatter start")
	}

	closingLineIdx := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			closingLineIdx = i
			break
		}
	}

	if closingLineIdx == -1 {
		return nil, fmt.Errorf("no closing frontmatter marker")
	}

	yamlBlock := strings.Join(lines[1:closingLineIdx], "\n")

	var data map[string]interface{}
	err := yaml.Unmarshal([]byte(yamlBlock), &data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// matchesPatterns checks if a file path matches any of the given patterns.
func (r *SpecADRRule) matchesPatterns(path string, patterns []string) bool {
	for _, pattern := range patterns {
		if r.matchSubpath(path, pattern) {
			return true
		}
		// Case-insensitive check
		if r.matchSubpath(strings.ToLower(path), strings.ToLower(pattern)) {
			return true
		}
		// Also match against basename for backward compatibility
		base := filepath.Base(path)
		if r.matchSubpath(base, pattern) {
			return true
		}
		if r.matchSubpath(strings.ToLower(base), strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func (r *SpecADRRule) matchSubpath(path, pattern string) bool {
	path = filepath.ToSlash(path)
	pattern = filepath.ToSlash(pattern)
	if walker.MatchesPattern(path, pattern) {
		return true
	}
	parts := strings.Split(path, "/")
	for i := 1; i < len(parts); i++ {
		subpath := strings.Join(parts[i:], "/")
		if walker.MatchesPattern(subpath, pattern) {
			return true
		}
	}
	return false
}

// formatMissingSpecFolderMessage creates an error message for missing spec folder.
func (r *SpecADRRule) formatMissingSpecFolderMessage() string {
	return "No specification folder found. Expected one of: " + strings.Join(r.SpecFolderPaths, ", ") + ". " +
		"Create a dedicated folder for feature specifications to maintain clear documentation structure."
}

// formatMissingADRFolderMessage creates an error message for missing ADR folder.
func (r *SpecADRRule) formatMissingADRFolderMessage() string {
	return "No ADR (Architecture Decision Record) folder found. Expected one of: " + strings.Join(r.ADRFolderPaths, ", ") + ". " +
		"Create a dedicated folder for ADRs to document architectural decisions."
}

// formatMissingSpecSectionsMessage creates an error message for missing spec sections.
func (r *SpecADRRule) formatMissingSpecSectionsMessage(missingSections []string) string {
	return "Feature specification is missing required sections:\n" +
		"  - " + strings.Join(missingSections, "\n  - ") + "\n\n" +
		"Each feature spec must include:\n" +
		"  1. User Scenarios & Testing (with prioritized, independently testable user stories)\n" +
		"  2. Functional Requirements (FR-### format)\n" +
		"  3. Success Criteria (SC-### format with measurable outcomes)"
}

// formatMissingADRSectionsMessage creates an error message for missing ADR sections.
func (r *SpecADRRule) formatMissingADRSectionsMessage(missingSections []string) string {
	return "ADR is missing required sections:\n" +
		"  - " + strings.Join(missingSections, "\n  - ") + "\n\n" +
		"Each ADR must include:\n" +
		"  1. Context and Problem Statement\n" +
		"  2. Considered Options\n" +
		"  3. Decision Outcome (with chosen option and justification)"
}

// NewSpecADRRule creates a new SpecADRRule.
func NewSpecADRRule(config SpecADRRule) *SpecADRRule {
	rule := &SpecADRRule{
		RequireSpecFolder:    config.RequireSpecFolder,
		RequireADRFolder:     config.RequireADRFolder,
		EnforceSpecTemplate:  config.EnforceSpecTemplate,
		EnforceADRTemplate:   config.EnforceADRTemplate,
		SpecFolderPaths:      config.SpecFolderPaths,
		ADRFolderPaths:       config.ADRFolderPaths,
		SpecFilePatterns:     config.SpecFilePatterns,
		ADRFilePatterns:      config.ADRFilePatterns,
		SpecRequiredHeadings: config.SpecRequiredHeadings,
		ADRRequiredHeadings:  config.ADRRequiredHeadings,
		ADRRequiredMetadata:  config.ADRRequiredMetadata,
	}

	// Helper for bool pointer defaults (default to true)
	defaultTrue := func(b *bool) *bool {
		if b == nil {
			t := true
			return &t
		}
		return b
	}

	rule.RequireSpecFolder = defaultTrue(rule.RequireSpecFolder)
	rule.RequireADRFolder = defaultTrue(rule.RequireADRFolder)
	rule.EnforceSpecTemplate = defaultTrue(rule.EnforceSpecTemplate)
	rule.EnforceADRTemplate = defaultTrue(rule.EnforceADRTemplate)

	// Set string slice defaults if empty
	if len(rule.SpecFolderPaths) == 0 {
		rule.SpecFolderPaths = []string{"docs/specs", "specifications"}
	}
	if len(rule.ADRFolderPaths) == 0 {
		rule.ADRFolderPaths = []string{"docs/adr", "docs/decisions"}
	}
	if len(rule.SpecFilePatterns) == 0 {
		rule.SpecFilePatterns = []string{"*-spec.md", "feature-*.md"}
	}
	if len(rule.ADRFilePatterns) == 0 {
		rule.ADRFilePatterns = []string{"ADR-*.md", "adr-*.md"}
	}
	if len(rule.SpecRequiredHeadings) == 0 {
		rule.SpecRequiredHeadings = []string{
			"# Feature Specification:",
			"## User Scenarios & Testing",
			"## Requirements",
			"### Functional Requirements",
			"## Success Criteria",
		}
	}
	if len(rule.ADRRequiredHeadings) == 0 {
		rule.ADRRequiredHeadings = []string{
			"## Context and Problem Statement",
			"## Considered Options",
			"## Decision Outcome",
		}
	}
	if len(rule.ADRRequiredMetadata) == 0 {
		rule.ADRRequiredMetadata = []string{
			"status:",
			"date:",
		}
	}

	return rule
}

// ParseSpecADRRule parses raw config into a SpecADRRule.
func ParseSpecADRRule(raw map[string]interface{}) (*SpecADRRule, error) {
	var ruleConfig SpecADRRule
	jsonBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(jsonBytes, &ruleConfig); err != nil {
		return nil, err
	}
	return NewSpecADRRule(ruleConfig), nil
}
