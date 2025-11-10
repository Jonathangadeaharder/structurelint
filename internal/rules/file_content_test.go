package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestFileContentRule_Check(t *testing.T) {
	// Create temp directory with templates
	tmpDir := t.TempDir()
	templateDir := filepath.Join(tmpDir, ".structurelint", "templates")
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a simple template
	readmeTemplate := `required-sections:
  - "# "
  - "## Overview"

required-patterns:
  - ".*"
`
	if err := os.WriteFile(filepath.Join(templateDir, "readme.yml"), []byte(readmeTemplate), 0644); err != nil {
		t.Fatal(err)
	}

	// Create test README file
	readmeContent := `# My Project

## Overview

This is a test project.
`
	readmePath := filepath.Join(tmpDir, "README.md")
	if err := os.WriteFile(readmePath, []byte(readmeContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name          string
		templates     map[string]string
		files         []walker.FileInfo
		wantViolCount int
	}{
		{
			name:      "valid README - no violations",
			templates: map[string]string{"**/README.md": "readme"},
			files: []walker.FileInfo{
				{Path: "README.md", AbsPath: readmePath, IsDir: false},
			},
			wantViolCount: 0,
		},
		{
			name:      "template not found",
			templates: map[string]string{"**/README.md": "nonexistent"},
			files: []walker.FileInfo{
				{Path: "README.md", AbsPath: readmePath, IsDir: false},
			},
			wantViolCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewFileContentRule(tt.templates, ".structurelint/templates", tmpDir)
			violations := rule.Check(tt.files, nil)

			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
				for _, v := range violations {
					t.Logf("  - %s: %s", v.Path, v.Message)
				}
			}
		})
	}
}

func TestFileContentRule_loadTemplate(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name                 string
		templateContent      string
		wantSections         int
		wantRequiredPatterns int
		wantForbiddenPatt    int
	}{
		{
			name: "template with sections and patterns",
			templateContent: `# Test template
required-sections:
  - "## Overview"
  - "## Usage"

required-patterns:
  - ".*"

forbidden-patterns:
  - "TODO"
  - "FIXME"
`,
			wantSections:         2,
			wantRequiredPatterns: 1,
			wantForbiddenPatt:    2,
		},
		{
			name: "template with only sections",
			templateContent: `required-sections:
  - "# Title"
  - "## Description"
`,
			wantSections:         2,
			wantRequiredPatterns: 0,
			wantForbiddenPatt:    0,
		},
		{
			name:                 "empty template",
			templateContent:      ``,
			wantSections:         0,
			wantRequiredPatterns: 0,
			wantForbiddenPatt:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			templatePath := filepath.Join(tmpDir, "test.yml")
			if err := os.WriteFile(templatePath, []byte(tt.templateContent), 0644); err != nil {
				t.Fatal(err)
			}

			rule := &FileContentRule{}
			template, err := rule.loadTemplate(templatePath)
			if err != nil {
				t.Fatalf("loadTemplate() error = %v", err)
			}

			if len(template.RequiredSections) != tt.wantSections {
				t.Errorf("got %d required sections, want %d", len(template.RequiredSections), tt.wantSections)
			}
			if len(template.RequiredPatterns) != tt.wantRequiredPatterns {
				t.Errorf("got %d required patterns, want %d", len(template.RequiredPatterns), tt.wantRequiredPatterns)
			}
			if len(template.ForbiddenPatterns) != tt.wantForbiddenPatt {
				t.Errorf("got %d forbidden patterns, want %d", len(template.ForbiddenPatterns), tt.wantForbiddenPatt)
			}
		})
	}
}

func TestFileContentRule_validateFileContent(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name          string
		template      Template
		fileContent   string
		wantViolCount int
	}{
		{
			name: "content satisfies all requirements",
			template: Template{
				RequiredSections: []string{"## Overview", "## Usage"},
			},
			fileContent: `# Project

## Overview

This is an overview.

## Usage

Usage instructions here.
`,
			wantViolCount: 0,
		},
		{
			name: "missing required section",
			template: Template{
				RequiredSections: []string{"## Overview", "## Installation"},
			},
			fileContent: `# Project

## Overview

This is an overview.
`,
			wantViolCount: 1,
		},
		{
			name: "contains forbidden pattern",
			template: Template{
				ForbiddenPatterns: []string{"TODO", "FIXME"},
			},
			fileContent: `# Project

TODO: Add more content here.
`,
			wantViolCount: 1,
		},
		{
			name: "missing required regex pattern",
			template: Template{
				RequiredPatterns: []string{`\d{4}-\d{2}-\d{2}`}, // Date pattern
			},
			fileContent: `# Project

No dates in this content.
`,
			wantViolCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temp file with content
			testFile := filepath.Join(tmpDir, "test.md")
			if err := os.WriteFile(testFile, []byte(tt.fileContent), 0644); err != nil {
				t.Fatal(err)
			}

			rule := &FileContentRule{}
			fileInfo := walker.FileInfo{
				Path:    "test.md",
				AbsPath: testFile,
				IsDir:   false,
			}

			violations := rule.validateFileContent(fileInfo, tt.template)

			if len(violations) != tt.wantViolCount {
				t.Errorf("validateFileContent() got %d violations, want %d", len(violations), tt.wantViolCount)
				for _, v := range violations {
					t.Logf("  - %s: %s", v.Path, v.Message)
				}
			}
		})
	}
}

func TestFileContentRule_Name(t *testing.T) {
	rule := NewFileContentRule(nil, "", "")
	if got := rule.Name(); got != "file-content" {
		t.Errorf("Name() = %v, want file-content", got)
	}
}
