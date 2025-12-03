// Package rules provides rule implementations for structurelint.
package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// FileContentRule validates file content against templates or patterns
type FileContentRule struct {
	Templates   map[string]string // File pattern -> template name
	TemplateDir string            // Directory containing templates (e.g., ".structurelint/templates")
	RootPath    string            // Root path for resolving template directory
}

// Template represents a file content template
type Template struct {
	Name             string
	RequiredSections []string   // Required section headers (e.g., "## Overview", "## Usage")
	RequiredPatterns []string   // Regex patterns that must be present
	ForbiddenPatterns []string  // Regex patterns that must NOT be present
	MustStartWith    string     // Content must start with this pattern
	MustEndWith      string     // Content must end with this pattern
}

// Name returns the rule name
func (r *FileContentRule) Name() string {
	return "file-content"
}

// Check validates file content against templates
func (r *FileContentRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	var violations []Violation

	// Load templates
	templates := r.loadTemplates()

	for _, file := range files {
		if file.IsDir {
			continue
		}

		// Check if this file matches any template pattern
		for pattern, templateName := range r.Templates {
			if matchesGlobPattern(file.Path, pattern) {
				template, exists := templates[templateName]
				if !exists {
					violations = append(violations, Violation{
						Rule:    r.Name(),
						Path:    file.Path,
						Message: fmt.Sprintf("template '%s' not found", templateName),
					})
					continue
				}

				// Validate file content against template
				fileViolations := r.validateFileContent(file, template)
				violations = append(violations, fileViolations...)
			}
		}
	}

	return violations
}

// validateFileContent validates a file's content against a template
func (r *FileContentRule) validateFileContent(file walker.FileInfo, template Template) []Violation {
	var violations []Violation

	// Read file content
	content, err := os.ReadFile(file.AbsPath)
	if err != nil {
		violations = append(violations, Violation{
			Rule:    r.Name(),
			Path:    file.Path,
			Message: fmt.Sprintf("failed to read file: %v", err),
		})
		return violations
	}

	contentStr := string(content)

	// Validate different aspects
	violations = append(violations, r.checkRequiredSections(file, template, contentStr)...)
	violations = append(violations, r.checkPatterns(file, template.RequiredPatterns, contentStr, "missing")...)
	violations = append(violations, r.checkPatterns(file, template.ForbiddenPatterns, contentStr, "contains forbidden")...)
	violations = append(violations, r.checkStartEnd(file, template, contentStr)...)

	return violations
}

func (r *FileContentRule) checkRequiredSections(file walker.FileInfo, template Template, content string) []Violation {
	var violations []Violation
	for _, section := range template.RequiredSections {
		if !strings.Contains(content, section) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("missing required section: '%s'", section),
			})
		}
	}
	return violations
}

func (r *FileContentRule) checkPatterns(file walker.FileInfo, patterns []string, content string, violation_type string) []Violation {
	var violations []Violation
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue
		}

		matches := re.MatchString(content)
		shouldViolate := (violation_type == "missing" && !matches) || (violation_type == "contains forbidden" && matches)

		if shouldViolate {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("%s required pattern: '%s'", violation_type, pattern),
			})
		}
	}
	return violations
}

func (r *FileContentRule) checkStartEnd(file walker.FileInfo, template Template, content string) []Violation {
	var violations []Violation
	trimmed := strings.TrimSpace(content)

	if template.MustStartWith != "" {
		re, err := regexp.Compile(template.MustStartWith)
		if err == nil && !re.MatchString(trimmed) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("must start with pattern: '%s'", template.MustStartWith),
			})
		}
	}

	if template.MustEndWith != "" {
		re, err := regexp.Compile(template.MustEndWith)
		if err == nil && !re.MatchString(trimmed) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("must end with pattern: '%s'", template.MustEndWith),
			})
		}
	}

	return violations
}

// loadTemplates loads template definitions from the template directory
func (r *FileContentRule) loadTemplates() map[string]Template {
	templates := make(map[string]Template)

	templateDirPath := filepath.Join(r.RootPath, r.TemplateDir)

	// Check if template directory exists
	if _, err := os.Stat(templateDirPath); os.IsNotExist(err) {
		return templates
	}

	// Read all template files
	entries, err := os.ReadDir(templateDirPath)
	if err != nil {
		return templates
	}

	for _, entry := range entries {
		if entry.IsDir() || !r.isTemplateFile(entry.Name()) {
			continue
		}

		templatePath := filepath.Join(templateDirPath, entry.Name())
		if template, err := r.parseTemplateFile(templatePath); err == nil {
			templateName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			template.Name = templateName
			templates[templateName] = template
		}
	}

	return templates
}

func (r *FileContentRule) isTemplateFile(name string) bool {
	return strings.HasSuffix(name, ".yml") || strings.HasSuffix(name, ".yaml")
}

// parseTemplateFile parses a template file
func (r *FileContentRule) parseTemplateFile(path string) (Template, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return Template{}, err
	}

	return r.parseTemplateContent(string(content)), nil
}

func (r *FileContentRule) parseTemplateContent(content string) Template {
	template := Template{}
	lines := strings.Split(content, "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if r.shouldSkipLine(line) {
			continue
		}

		currentSection = r.updateSection(line, currentSection, &template)

		// Parse list items
		if strings.HasPrefix(line, "- ") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "- "))
			value = strings.Trim(value, `"'`)
			r.addToTemplateSection(&template, currentSection, value)
		}
	}

	return template
}

func (r *FileContentRule) shouldSkipLine(line string) bool {
	return line == "" || strings.HasPrefix(line, "#")
}

func (r *FileContentRule) updateSection(line, currentSection string, template *Template) string {
	switch {
	case strings.HasPrefix(line, "required-sections:"):
		return "required-sections"
	case strings.HasPrefix(line, "required-patterns:"):
		return "required-patterns"
	case strings.HasPrefix(line, "forbidden-patterns:"):
		return "forbidden-patterns"
	case strings.HasPrefix(line, "must-start-with:"):
		template.MustStartWith = strings.TrimSpace(strings.TrimPrefix(line, "must-start-with:"))
		return "must-start-with"
	case strings.HasPrefix(line, "must-end-with:"):
		template.MustEndWith = strings.TrimSpace(strings.TrimPrefix(line, "must-end-with:"))
		return "must-end-with"
	default:
		return currentSection
	}
}

func (r *FileContentRule) addToTemplateSection(template *Template, section, value string) {
	switch section {
	case "required-sections":
		template.RequiredSections = append(template.RequiredSections, value)
	case "required-patterns":
		template.RequiredPatterns = append(template.RequiredPatterns, value)
	case "forbidden-patterns":
		template.ForbiddenPatterns = append(template.ForbiddenPatterns, value)
	}
}


// NewFileContentRule creates a new FileContentRule
func NewFileContentRule(templates map[string]string, templateDir, rootPath string) *FileContentRule {
	return &FileContentRule{
		Templates:   templates,
		TemplateDir: templateDir,
		RootPath:    rootPath,
	}
}
