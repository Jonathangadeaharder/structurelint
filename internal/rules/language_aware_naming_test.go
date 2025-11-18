package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestLanguageAwareNamingConvention_Python_UsesSnakeCase(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/user_service.py", IsDir: false},   // Valid snake_case
		{Path: "src/UserService.py", IsDir: false},     // Invalid - should be snake_case
		{Path: "src/data_processor.py", IsDir: false}, // Valid snake_case
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (UserService.py), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}

	if len(violations) > 0 && violations[0].Path != "src/UserService.py" {
		t.Errorf("Expected violation for UserService.py, got %s", violations[0].Path)
	}
}

func TestLanguageAwareNamingConvention_JavaScript_UsesCamelCase(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/userService.js", IsDir: false},   // Valid camelCase
		{Path: "src/UserService.js", IsDir: false},   // Invalid - should be camelCase
		{Path: "src/dataProcessor.js", IsDir: false}, // Valid camelCase
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (UserService.js), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}

	if len(violations) > 0 && violations[0].Path != "src/UserService.js" {
		t.Errorf("Expected violation for UserService.js, got %s", violations[0].Path)
	}
}

func TestLanguageAwareNamingConvention_ReactComponents_UsePascalCase(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/components/UserProfile.jsx", IsDir: false}, // Valid PascalCase
		{Path: "src/components/userProfile.jsx", IsDir: false}, // Invalid - should be PascalCase
		{Path: "src/components/DataTable.tsx", IsDir: false},   // Valid PascalCase
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (userProfile.jsx), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}

	if len(violations) > 0 && violations[0].Path != "src/components/userProfile.jsx" {
		t.Errorf("Expected violation for userProfile.jsx, got %s", violations[0].Path)
	}
}

func TestLanguageAwareNamingConvention_Go_UsesPascalCase(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "internal/UserService.go", IsDir: false},  // Valid PascalCase
		{Path: "internal/user_service.go", IsDir: false}, // Invalid - should be PascalCase
		{Path: "internal/DataProcessor.go", IsDir: false}, // Valid PascalCase
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (user_service.go), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}

	if len(violations) > 0 && violations[0].Path != "internal/user_service.go" {
		t.Errorf("Expected violation for user_service.go, got %s", violations[0].Path)
	}
}

func TestLanguageAwareNamingConvention_UserPatternsOverrideDefaults(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/user_service.py", IsDir: false}, // Would normally require snake_case
	}

	// User wants PascalCase for Python (overriding default)
	userPatterns := map[string]string{
		"*.py": "PascalCase",
	}

	rule, err := NewLanguageAwareNamingConventionRule("", userPatterns)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Act
	violations := rule.Check(files, nil)

	// Assert - should violate because user_service.py is not PascalCase
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (user overrides should apply), got %d", len(violations))
	}
}

func TestLanguageAwareNamingConvention_Java_UsesPascalCase(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/main/java/UserService.java", IsDir: false},  // Valid PascalCase
		{Path: "src/main/java/userService.java", IsDir: false},  // Invalid
		{Path: "src/main/java/DataProcessor.java", IsDir: false}, // Valid PascalCase
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (userService.java), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestLanguageAwareNamingConvention_Rust_UsesSnakeCase(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/user_service.rs", IsDir: false},   // Valid snake_case
		{Path: "src/UserService.rs", IsDir: false},    // Invalid
		{Path: "src/data_processor.rs", IsDir: false}, // Valid snake_case
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (UserService.rs), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestLanguageAwareNamingConvention_MultiLanguageProject(t *testing.T) {
	// Arrange - Mix of Python, JavaScript, and Go
	files := []walker.FileInfo{
		{Path: "backend/user_service.py", IsDir: false},  // Valid Python snake_case
		{Path: "backend/UserService.py", IsDir: false},   // Invalid Python
		{Path: "frontend/userService.js", IsDir: false},  // Valid JS camelCase
		{Path: "frontend/UserService.js", IsDir: false},  // Invalid JS
		{Path: "internal/UserService.go", IsDir: false},  // Valid Go PascalCase
		{Path: "internal/user_service.go", IsDir: false}, // Invalid Go
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	// Act
	violations := rule.Check(files, nil)

	// Assert - Should catch 3 violations (one per language)
	if len(violations) != 3 {
		t.Errorf("Expected 3 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}
