package structure

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestLanguageAwareNamingConvention_Python_UsesSnakeCase(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "src/user_service.py", IsDir: false},
		{Path: "src/UserService.py", IsDir: false},
		{Path: "src/data_processor.py", IsDir: false},
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	violations := rule.Check(files, nil)

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
	files := []walker.FileInfo{
		{Path: "src/userService.js", IsDir: false},
		{Path: "src/UserService.js", IsDir: false},
		{Path: "src/dataProcessor.js", IsDir: false},
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	violations := rule.Check(files, nil)

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
	files := []walker.FileInfo{
		{Path: "src/components/UserProfile.jsx", IsDir: false},
		{Path: "src/components/userProfile.jsx", IsDir: false},
		{Path: "src/components/DataTable.tsx", IsDir: false},
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	violations := rule.Check(files, nil)

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
	files := []walker.FileInfo{
		{Path: "internal/UserService.go", IsDir: false},
		{Path: "internal/user_service.go", IsDir: false},
		{Path: "internal/DataProcessor.go", IsDir: false},
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	violations := rule.Check(files, nil)

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
	files := []walker.FileInfo{
		{Path: "src/user_service.py", IsDir: false},
	}

	userPatterns := map[string]string{
		"*.py": "PascalCase",
	}

	rule, err := NewLanguageAwareNamingConventionRule("", userPatterns)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	violations := rule.Check(files, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (user overrides should apply), got %d", len(violations))
	}
}

func TestLanguageAwareNamingConvention_Java_UsesPascalCase(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "src/main/java/UserService.java", IsDir: false},
		{Path: "src/main/java/userService.java", IsDir: false},
		{Path: "src/main/java/DataProcessor.java", IsDir: false},
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	violations := rule.Check(files, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (userService.java), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestLanguageAwareNamingConvention_Rust_UsesSnakeCase(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "src/user_service.rs", IsDir: false},
		{Path: "src/UserService.rs", IsDir: false},
		{Path: "src/data_processor.rs", IsDir: false},
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	violations := rule.Check(files, nil)

	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (UserService.rs), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestLanguageAwareNamingConvention_MultiLanguageProject(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "backend/user_service.py", IsDir: false},
		{Path: "backend/UserService.py", IsDir: false},
		{Path: "frontend/userService.js", IsDir: false},
		{Path: "frontend/UserService.js", IsDir: false},
		{Path: "internal/UserService.go", IsDir: false},
		{Path: "internal/user_service.go", IsDir: false},
	}

	rule, err := NewLanguageAwareNamingConventionRule("", nil)
	if err != nil {
		t.Fatalf("Failed to create rule: %v", err)
	}

	violations := rule.Check(files, nil)

	if len(violations) != 3 {
		t.Errorf("Expected 3 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}
