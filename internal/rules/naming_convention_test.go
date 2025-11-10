package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestNamingConventionRule_Check(t *testing.T) {
	tests := []struct {
		name          string
		patterns      map[string]string
		files         []walker.FileInfo
		wantViolCount int
	}{
		{
			name:     "camelCase valid",
			patterns: map[string]string{"*.js": "camelCase"},
			files: []walker.FileInfo{
				{Path: "myFile.js"},
				{Path: "anotherFile.js"},
			},
			wantViolCount: 0,
		},
		{
			name:     "camelCase invalid - starts with uppercase",
			patterns: map[string]string{"*.js": "camelCase"},
			files: []walker.FileInfo{
				{Path: "MyFile.js"},
			},
			wantViolCount: 1,
		},
		{
			name:     "PascalCase valid",
			patterns: map[string]string{"*.tsx": "PascalCase"},
			files: []walker.FileInfo{
				{Path: "MyComponent.tsx"},
				{Path: "Button.tsx"},
			},
			wantViolCount: 0,
		},
		{
			name:     "PascalCase invalid - starts with lowercase",
			patterns: map[string]string{"*.tsx": "PascalCase"},
			files: []walker.FileInfo{
				{Path: "myComponent.tsx"},
			},
			wantViolCount: 1,
		},
		{
			name:     "kebab-case valid",
			patterns: map[string]string{"*.css": "kebab-case"},
			files: []walker.FileInfo{
				{Path: "my-styles.css"},
				{Path: "button-component.css"},
			},
			wantViolCount: 0,
		},
		{
			name:     "kebab-case invalid - has uppercase",
			patterns: map[string]string{"*.css": "kebab-case"},
			files: []walker.FileInfo{
				{Path: "MyStyles.css"},
			},
			wantViolCount: 1,
		},
		{
			name:     "snake_case valid",
			patterns: map[string]string{"*.py": "snake_case"},
			files: []walker.FileInfo{
				{Path: "my_module.py"},
				{Path: "test_utils.py"},
			},
			wantViolCount: 0,
		},
		{
			name:     "snake_case invalid - has uppercase",
			patterns: map[string]string{"*.py": "snake_case"},
			files: []walker.FileInfo{
				{Path: "MyModule.py"},
			},
			wantViolCount: 1,
		},
		{
			name:     "lowercase valid",
			patterns: map[string]string{"*.txt": "lowercase"},
			files: []walker.FileInfo{
				{Path: "readme.txt"},
			},
			wantViolCount: 0,
		},
		{
			name:     "multiple patterns",
			patterns: map[string]string{
				"*.js":  "camelCase",
				"*.tsx": "PascalCase",
			},
			files: []walker.FileInfo{
				{Path: "myFile.js"},
				{Path: "MyComponent.tsx"},
			},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNamingConventionRule(tt.patterns)
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

func Test_isCamelCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"myFile", true},
		{"myLongFileName", true},
		{"a", true},
		{"MyFile", false},
		{"my-file", false},
		{"my_file", false},
		{"my file", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isCamelCase(tt.input); got != tt.want {
				t.Errorf("isCamelCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_isPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"MyFile", true},
		{"MyLongFileName", true},
		{"A", true},
		{"myFile", false},
		{"My-File", false},
		{"My_File", false},
		{"My File", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isPascalCase(tt.input); got != tt.want {
				t.Errorf("isPascalCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_isKebabCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"my-file", true},
		{"my-long-file-name", true},
		{"file", true},
		{"MyFile", false},
		{"my_file", false},
		{"my file", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isKebabCase(tt.input); got != tt.want {
				t.Errorf("isKebabCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func Test_isSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"my_file", true},
		{"my_long_file_name", true},
		{"file", true},
		{"MyFile", false},
		{"my-file", false},
		{"my file", false},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			if got := isSnakeCase(tt.input); got != tt.want {
				t.Errorf("isSnakeCase(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestNamingConventionRule_Name(t *testing.T) {
	rule := NewNamingConventionRule(map[string]string{"*.js": "camelCase"})
	if got := rule.Name(); got != "naming-convention" {
		t.Errorf("Name() = %v, want naming-convention", got)
	}
}
