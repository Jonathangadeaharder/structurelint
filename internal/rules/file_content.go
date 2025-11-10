// Package rules provides rule implementations for structurelint.
package rules

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/structurelint/structurelint/internal/walker"
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

	// Check required sections
	for _, section := range template.RequiredSections {
		if !strings.Contains(contentStr, section) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("missing required section: '%s'", section),
			})
		}
	}

	// Check required patterns
	for _, pattern := range template.RequiredPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue // Skip invalid regex
		}

		if !re.MatchString(contentStr) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("missing required pattern: '%s'", pattern),
			})
		}
	}

	// Check forbidden patterns
	for _, pattern := range template.ForbiddenPatterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			continue // Skip invalid regex
		}

		if re.MatchString(contentStr) {
			violations = append(violations, Violation{
				Rule:    r.Name(),
				Path:    file.Path,
				Message: fmt.Sprintf("contains forbidden pattern: '%s'", pattern),
			})
		}
	}

	// Check must start with
	if template.MustStartWith != "" {
		re, err := regexp.Compile(template.MustStartWith)
		if err == nil {
			if !re.MatchString(strings.TrimSpace(contentStr)) {
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: fmt.Sprintf("must start with pattern: '%s'", template.MustStartWith),
				})
			}
		}
	}

	// Check must end with
	if template.MustEndWith != "" {
		re, err := regexp.Compile(template.MustEndWith)
		if err == nil {
			if !re.MatchString(strings.TrimSpace(contentStr)) {
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Path:    file.Path,
					Message: fmt.Sprintf("must end with pattern: '%s'", template.MustEndWith),
				})
			}
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
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yml") && !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}

		templatePath := filepath.Join(templateDirPath, entry.Name())
		template, err := r.loadTemplate(templatePath)
		if err == nil {
			templateName := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
			templates[templateName] = template
		}
	}

	return templates
}

// loadTemplate loads a single template from a file
func (r *FileContentRule) loadTemplate(path string) (Template, error) {
	// For now, we'll use a simple format parser
	// In production, you'd want to use a proper YAML parser

	content, err := os.ReadFile(path)
	if err != nil {
		return Template{}, err
	}

	template := Template{
		Name: filepath.Base(path),
	}

	// Simple line-by-line parsing
	lines := strings.Split(string(content), "\n")
	currentSection := ""

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if strings.HasPrefix(line, "required-sections:") {
			currentSection = "required-sections"
			continue
		}
		if strings.HasPrefix(line, "required-patterns:") {
			currentSection = "required-patterns"
			continue
		}
		if strings.HasPrefix(line, "forbidden-patterns:") {
			currentSection = "forbidden-patterns"
			continue
		}
		if strings.HasPrefix(line, "must-start-with:") {
			currentSection = "must-start-with"
			template.MustStartWith = strings.TrimSpace(strings.TrimPrefix(line, "must-start-with:"))
			continue
		}
		if strings.HasPrefix(line, "must-end-with:") {
			currentSection = "must-end-with"
			template.MustEndWith = strings.TrimSpace(strings.TrimPrefix(line, "must-end-with:"))
			continue
		}

		// Parse list items
		if strings.HasPrefix(line, "- ") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "- "))
			value = strings.Trim(value, `"'`)

			switch currentSection {
			case "required-sections":
				template.RequiredSections = append(template.RequiredSections, value)
			case "required-patterns":
				template.RequiredPatterns = append(template.RequiredPatterns, value)
			case "forbidden-patterns":
				template.ForbiddenPatterns = append(template.ForbiddenPatterns, value)
			}
		}
	}

	return template, nil
}

// NewFileContentRule creates a new FileContentRule
func NewFileContentRule(templates map[string]string, templateDir, rootPath string) *FileContentRule {
	return &FileContentRule{
		Templates:   templates,
		TemplateDir: templateDir,
		RootPath:    rootPath,
	}
}
