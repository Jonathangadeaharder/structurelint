package init

import (
	"strings"
	"testing"
)

func TestGenerateConfig(t *testing.T) {
	// Arrange
	// Act
	// Assert
	tests := []struct {
		name             string
		info             *ProjectInfo
		wantContains     []string
		wantNotContains  []string
	}{
		{
			name: "Go project with adjacent tests",
			info: &ProjectInfo{
				PrimaryLanguage: &LanguageInfo{
					Language:    "go",
					FileCount:   50,
					TestPattern: "adjacent",
				},
				MaxDepth:       5,
				MaxFilesInDir:  20,
				MaxSubdirs:     10,
				DocumentationStyle: "comprehensive",
			},
			wantContains: []string{
				"max-depth:",
				"max: 5",
				"test-adjacency:",
				"pattern: \"adjacent\"",
				"**/*.go",
				"file-existence:",
				"README.md",
			},
		},
		{
			name: "Python project with separate tests",
			info: &ProjectInfo{
				Languages: []LanguageInfo{
					{
						Language:    "python",
						FileCount:   30,
						TestPattern: "separate",
						TestDir:     "tests",
					},
				},
				PrimaryLanguage: &LanguageInfo{
					Language:    "python",
					FileCount:   30,
					TestPattern: "separate",
					TestDir:     "tests",
				},
				MaxDepth:       4,
				MaxFilesInDir:  15,
				MaxSubdirs:     8,
				DocumentationStyle: "minimal",
			},
			wantContains: []string{
				"pattern: \"separate\"",
				"test-dir: \"tests\"",
				"__pycache__",
			},
		},
		{
			name: "TypeScript project",
			info: &ProjectInfo{
				Languages: []LanguageInfo{
					{
						Language:           "typescript",
						FileCount:          40,
						TestPattern:        "adjacent",
						HasIntegrationDir:  true,
						IntegrationDir:     "tests/integration",
					},
				},
				PrimaryLanguage: &LanguageInfo{
					Language:           "typescript",
					FileCount:          40,
					TestPattern:        "adjacent",
					HasIntegrationDir:  true,
					IntegrationDir:     "tests/integration",
				},
				MaxDepth:       6,
				MaxFilesInDir:  25,
				MaxSubdirs:     12,
				DocumentationStyle: "none",
			},
			wantContains: []string{
				"node_modules",
				"dist",
				"tests/integration",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateConfig(tt.info)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("GenerateConfig() missing expected content: %q", want)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(got, notWant) {
					t.Errorf("GenerateConfig() contains unexpected content: %q", notWant)
				}
			}

			// Verify basic structure
			if !strings.Contains(got, "root: true") {
				t.Error("GenerateConfig() missing 'root: true'")
			}
			if !strings.Contains(got, "rules:") {
				t.Error("GenerateConfig() missing 'rules:'")
			}
		})
	}
}

func TestGenerateSummary(t *testing.T) {
	info := &ProjectInfo{
		Languages: []LanguageInfo{
			{
				Language:    "go",
				FileCount:   50,
				TestPattern: "adjacent",
			},
			{
				Language:  "python",
				FileCount: 20,
			},
		},
		PrimaryLanguage: &LanguageInfo{
			Language:    "go",
			FileCount:   50,
			TestPattern: "adjacent",
		},
		MaxDepth:           5,
		MaxFilesInDir:      20,
		MaxSubdirs:         10,
		DocumentationStyle: "comprehensive",
	}

	got := GenerateSummary(info)

	wantContains := []string{
		"Project Analysis Summary",
		"Languages Detected:",
		"go",
		"Test pattern: adjacent",
		"Project Structure:",
		"Max depth: 5",
		"Documentation:",
		"Comprehensive",
	}

	for _, want := range wantContains {
		if !strings.Contains(got, want) {
			t.Errorf("GenerateSummary() missing expected content: %q", want)
		}
	}
}

func Test_generateExemptions(t *testing.T) {
	tests := []struct {
		lang         string
		wantContains []string
	}{
		{
			lang:         "go",
			wantContains: []string{"cmd/**/*.go", "**/*_gen.go"},
		},
		{
			lang:         "python",
			wantContains: []string{"**/__init__.py", "setup.py"},
		},
		{
			lang:         "typescript",
			wantContains: []string{"**/*.d.ts", "**/index.ts"},
		},
		{
			lang:         "java",
			wantContains: []string{"**/Main.java"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			got := generateExemptions(tt.lang)

			for _, want := range tt.wantContains {
				if !strings.Contains(got, want) {
					t.Errorf("generateExemptions(%q) missing expected content: %q", tt.lang, want)
				}
			}
		})
	}
}
