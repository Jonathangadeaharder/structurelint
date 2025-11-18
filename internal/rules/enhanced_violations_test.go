package rules

import (
	"strings"
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestViolation_FormatDetailed_WithAllFields(t *testing.T) {
	// Arrange
	violation := Violation{
		Rule:    "naming-convention",
		Path:    "src/components/button.tsx",
		Message: "does not match naming convention 'PascalCase'",
		Expected: "PascalCase",
		Actual:   "camelCase",
		Suggestions: []string{
			"Rename to 'Button.tsx'",
			"Add to exclude patterns if intentional",
		},
		Context: "Pattern: src/components/**/*.tsx",
	}

	// Act
	formatted := violation.FormatDetailed()

	// Assert
	if !strings.Contains(formatted, "src/components/button.tsx") {
		t.Error("Should contain path")
	}
	if !strings.Contains(formatted, "Expected: PascalCase") {
		t.Error("Should contain expected value")
	}
	if !strings.Contains(formatted, "Actual: camelCase") {
		t.Error("Should contain actual value")
	}
	if !strings.Contains(formatted, "Rename to 'Button.tsx'") {
		t.Error("Should contain first suggestion")
	}
	if !strings.Contains(formatted, "Context: Pattern: src/components/**/*.tsx") {
		t.Error("Should contain context")
	}
}

func TestViolation_FormatDetailed_WithoutOptionalFields(t *testing.T) {
	// Arrange
	violation := Violation{
		Rule:    "max-depth",
		Path:    "src/deeply/nested/structure/file.py",
		Message: "exceeds maximum depth of 4",
	}

	// Act
	formatted := violation.FormatDetailed()

	// Assert
	expected := "src/deeply/nested/structure/file.py: exceeds maximum depth of 4"
	if formatted != expected {
		t.Errorf("Expected '%s', got '%s'", expected, formatted)
	}
}

func TestNamingConventionRule_EnhancedViolations_CamelToPascal(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "userProfile.tsx", IsDir: false},
	}

	rule := NewNamingConventionRule(map[string]string{
		"*.tsx": "PascalCase",
	})

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.Expected != "PascalCase" {
		t.Errorf("Expected 'PascalCase', got '%s'", v.Expected)
	}
	if v.Actual != "camelCase" {
		t.Errorf("Expected actual to be 'camelCase', got '%s'", v.Actual)
	}
	if len(v.Suggestions) == 0 {
		t.Error("Expected suggestions to be provided")
	}
	if !strings.Contains(v.Suggestions[0], "UserProfile.tsx") {
		t.Errorf("Expected suggestion to include 'UserProfile.tsx', got '%s'", v.Suggestions[0])
	}
}

func TestNamingConventionRule_EnhancedViolations_PascalToCamel(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "StringHelper.ts", IsDir: false},
	}

	rule := NewNamingConventionRule(map[string]string{
		"*.ts": "camelCase",
	})

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.Expected != "camelCase" {
		t.Errorf("Expected 'camelCase', got '%s'", v.Expected)
	}
	if v.Actual != "PascalCase" {
		t.Errorf("Expected actual to be 'PascalCase', got '%s'", v.Actual)
	}
	if !strings.Contains(v.Suggestions[0], "stringHelper.ts") {
		t.Errorf("Expected suggestion to include 'stringHelper.ts', got '%s'", v.Suggestions[0])
	}
}

func TestNamingConventionRule_EnhancedViolations_SnakeToCamel(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/services/user_service.js", IsDir: false},
	}

	rule := NewNamingConventionRule(map[string]string{
		"*.js": "camelCase",
	})

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}

	v := violations[0]
	if v.Expected != "camelCase" {
		t.Errorf("Expected 'camelCase', got '%s'", v.Expected)
	}
	if v.Actual != "snake_case" {
		t.Errorf("Expected actual to be 'snake_case', got '%s'", v.Actual)
	}
	if !strings.Contains(v.Suggestions[0], "userService.js") {
		t.Errorf("Expected suggestion to include 'userService.js', got '%s'", v.Suggestions[0])
	}
}

func TestSplitIntoWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"userService", []string{"user", "Service"}},
		{"UserService", []string{"User", "Service"}},
		{"user_service", []string{"user", "service"}},
		{"user-service", []string{"user", "service"}},
	}

	for _, tt := range tests {
		result := splitIntoWords(tt.input)
		if len(result) != len(tt.expected) {
			t.Errorf("splitIntoWords(%q): expected %d words, got %d (%v)", tt.input, len(tt.expected), len(result), result)
			continue
		}
		for i, word := range result {
			if word != tt.expected[i] {
				t.Errorf("splitIntoWords(%q): expected word %d to be %q, got %q", tt.input, i, tt.expected[i], word)
			}
		}
	}
}

func TestConvertToConvention(t *testing.T) {
	rule := &NamingConventionRule{}

	tests := []struct {
		input      string
		convention string
		expected   string
	}{
		{"UserService", "camelCase", "userService"},
		{"userService", "PascalCase", "UserService"},
		{"user_service", "camelCase", "userService"},
		{"user_service", "PascalCase", "UserService"},
		{"UserService", "snake_case", "user_service"},
		{"UserService", "kebab-case", "user-service"},
	}

	for _, tt := range tests {
		result := rule.convertToConvention(tt.input, tt.convention)
		if result != tt.expected {
			t.Errorf("convertToConvention(%q, %q): expected %q, got %q", tt.input, tt.convention, tt.expected, result)
		}
	}
}

func TestDetectConvention(t *testing.T) {
	rule := &NamingConventionRule{}

	tests := []struct {
		input    string
		expected string
	}{
		{"userService", "camelCase"},
		{"UserService", "PascalCase"},
		{"user_service", "snake_case"},
		{"user-service", "kebab-case"},
		{"simpleword", "camelCase"}, // Single lowercase word is camelCase
		{"CONSTANT", "PascalCase"}, // Single uppercase word is PascalCase
	}

	for _, tt := range tests {
		result := rule.detectConvention(tt.input)
		if result != tt.expected {
			t.Errorf("detectConvention(%q): expected %q, got %q", tt.input, tt.expected, result)
		}
	}
}

func TestViolation_Context(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/button.tsx", IsDir: false},
	}

	rule := NewNamingConventionRule(map[string]string{
		"*.tsx": "PascalCase",
	})

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Fatalf("Expected 1 violation, got %d", len(violations))
	}

	if violations[0].Context == "" {
		t.Error("Expected context to be set")
	}
	if !strings.Contains(violations[0].Context, "*.tsx") {
		t.Errorf("Expected context to contain pattern, got: %s", violations[0].Context)
	}
}
