package rules

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestUniquenessConstraintsRule_GivenSingletonConstraint_WhenMultipleMatches_ThenViolation(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/services/user_service.py", IsDir: false},
		{Path: "src/services/user_service_clean.py", IsDir: false},
		{Path: "src/services/user_controller.py", IsDir: false}, // Does NOT match *_service*.py pattern
	}

	constraints := map[string]string{
		"*_service*.py": "singleton",
	}

	rule := NewUniquenessConstraintsRule(constraints)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (user_service files), got %d", len(violations))
	}

	if len(violations) > 0 {
		if !containsPath(violations[0].Message, "user_service.py") {
			t.Errorf("Violation message should mention user_service.py, got: %s", violations[0].Message)
		}
		if !containsPath(violations[0].Message, "user_service_clean.py") {
			t.Errorf("Violation message should mention user_service_clean.py, got: %s", violations[0].Message)
		}
	}
}

func TestUniquenessConstraintsRule_GivenSingletonConstraint_WhenOneMatch_ThenNoViolation(t *testing.T) {
	// Arrange - Each directory has only ONE service file
	files := []walker.FileInfo{
		{Path: "src/user/user_service.py", IsDir: false},
		{Path: "src/user/user_controller.py", IsDir: false},
		{Path: "src/product/product_service.py", IsDir: false},
	}

	constraints := map[string]string{
		"*_service.py": "singleton",
	}

	rule := NewUniquenessConstraintsRule(constraints)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Message)
		}
	}
}

func TestUniquenessConstraintsRule_GivenMultipleDirectories_WhenViolationsInSeparateDirs_ThenMultipleViolations(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/auth/auth_service.py", IsDir: false},
		{Path: "src/auth/auth_service_v2.py", IsDir: false},
		{Path: "src/billing/billing_service.py", IsDir: false},
		{Path: "src/billing/billing_service_new.py", IsDir: false},
	}

	constraints := map[string]string{
		"*_service*.py": "singleton",
	}

	rule := NewUniquenessConstraintsRule(constraints)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 2 {
		t.Errorf("Expected 2 violations (one per directory), got %d", len(violations))
	}
}

func TestUniquenessConstraintsRule_GivenJavaRepositoryPattern_WhenDuplicateRepositories_ThenViolation(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/main/java/com/example/UserRepository.java", IsDir: false},
		{Path: "src/main/java/com/example/UserRepository2.java", IsDir: false},
		{Path: "src/main/java/com/example/ProductRepository.java", IsDir: false},
	}

	constraints := map[string]string{
		"*Repository*.java": "singleton",
	}

	rule := NewUniquenessConstraintsRule(constraints)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (UserRepository files), got %d", len(violations))
	}
}

func TestUniquenessConstraintsRule_GivenMultiplePatterns_WhenViolations_ThenDetectsAll(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/services/user_service.py", IsDir: false},
		{Path: "src/services/user_service_v2.py", IsDir: false},
		{Path: "src/repositories/user_repository.py", IsDir: false},
		{Path: "src/repositories/user_repository_impl.py", IsDir: false},
	}

	constraints := map[string]string{
		"*_service*.py":     "singleton",
		"*_repository*.py": "singleton",
	}

	rule := NewUniquenessConstraintsRule(constraints)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 2 {
		t.Errorf("Expected 2 violations (service + repository), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestUniquenessConstraintsRule_GivenDifferentDirectories_WhenSamePatternInEach_ThenNoViolation(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/auth/auth_service.py", IsDir: false},
		{Path: "src/billing/billing_service.py", IsDir: false},
		{Path: "src/user/user_service.py", IsDir: false},
	}

	constraints := map[string]string{
		"*_service.py": "singleton",
	}

	rule := NewUniquenessConstraintsRule(constraints)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations (each service in different directory), got %d", len(violations))
	}
}

func TestUniquenessConstraintsRule_GivenExactPattern_WhenMultipleMatches_ThenViolation(t *testing.T) {
	// Arrange
	files := []walker.FileInfo{
		{Path: "src/config.py", IsDir: false},
		{Path: "src/config_old.py", IsDir: false},
	}

	constraints := map[string]string{
		"config*.py": "singleton",
	}

	rule := NewUniquenessConstraintsRule(constraints)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}
}

// Helper function to check if a message contains a path
func containsPath(message, path string) bool {
	// Check if the message contains the path or just the filename
	return strings.Contains(message, path) || strings.Contains(message, filepath.Base(path))
}
