package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestRegexMatchRule_BasicMatch(t *testing.T) {
	rule := NewRegexMatchRule(map[string]string{
		"src/*.ts": "^[a-z]+$", // Must be lowercase letters only
	})

	files := []walker.FileInfo{
		{Path: "src/user.ts", IsDir: false},      // Should match
		{Path: "src/admin.ts", IsDir: false},     // Should match
		{Path: "src/UserAuth.ts", IsDir: false},  // Should NOT match (has uppercase)
		{Path: "src/file-1.ts", IsDir: false},    // Should NOT match (has dash and number)
		{Path: "other/test.ts", IsDir: false},    // Should be ignored (wrong directory)
	}

	violations := rule.Check(files, nil)

	// Should have 2 violations: UserAuth.ts and file-1.ts
	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Path)
		}
	}

	// Check that violations are for the right files
	violationPaths := make(map[string]bool)
	for _, v := range violations {
		violationPaths[v.Path] = true
	}

	if !violationPaths["src/UserAuth.ts"] {
		t.Error("Expected violation for src/UserAuth.ts")
	}
	if !violationPaths["src/file-1.ts"] {
		t.Error("Expected violation for src/file-1.ts")
	}
}

func TestRegexMatchRule_Negation(t *testing.T) {
	rule := NewRegexMatchRule(map[string]string{
		"src/*.ts": "regex:!^test", // Must NOT start with "test"
	})

	files := []walker.FileInfo{
		{Path: "src/user.ts", IsDir: false},       // Should match (doesn't start with test)
		{Path: "src/testUser.ts", IsDir: false},   // Should NOT match (starts with test)
		{Path: "src/testHelper.ts", IsDir: false}, // Should NOT match (starts with test)
	}

	violations := rule.Check(files, nil)

	// Should have 2 violations
	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Path)
		}
	}

	// Check violations are for test* files
	for _, v := range violations {
		if v.Path != "src/testUser.ts" && v.Path != "src/testHelper.ts" {
			t.Errorf("Unexpected violation for %s", v.Path)
		}
	}
}

func TestRegexMatchRule_DirectorySubstitution(t *testing.T) {
	rule := NewRegexMatchRule(map[string]string{
		"components/*/Button.tsx": "^${0}Button$", // Filename must be <directory>Button
	})

	files := []walker.FileInfo{
		{Path: "components/atoms/atomsButton.tsx", IsDir: false},         // Should match
		{Path: "components/molecules/moleculesButton.tsx", IsDir: false}, // Should match
		{Path: "components/atoms/UserButton.tsx", IsDir: false},          // Should NOT match
		{Path: "components/organisms/Button.tsx", IsDir: false},          // Should NOT match (missing prefix)
	}

	violations := rule.Check(files, nil)

	// Should have at least 1 violation
	if len(violations) < 1 {
		t.Errorf("Expected at least 1 violation, got %d", len(violations))
	}

	// Verify at least one expected violation is present
	foundViolation := false
	for _, v := range violations {
		if v.Path == "components/atoms/UserButton.tsx" || v.Path == "components/organisms/Button.tsx" {
			foundViolation = true
			break
		}
	}

	if !foundViolation {
		t.Error("Expected violation for UserButton.tsx or Button.tsx")
	}
}

func TestRegexMatchRule_InvalidRegex(t *testing.T) {
	rule := NewRegexMatchRule(map[string]string{
		"src/*.ts": "[[[invalid", // Invalid regex
	})

	files := []walker.FileInfo{
		{Path: "src/user.ts", IsDir: false},
	}

	violations := rule.Check(files, nil)

	// Should have 0 violations (invalid regex is skipped)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations for invalid regex, got %d", len(violations))
	}
}

func TestRegexMatchRule_RegexPrefix(t *testing.T) {
	rule := NewRegexMatchRule(map[string]string{
		"src/*.ts": "regex:^[a-z]+$", // With "regex:" prefix
	})

	files := []walker.FileInfo{
		{Path: "src/user.ts", IsDir: false},     // Should match
		{Path: "src/UserAuth.ts", IsDir: false}, // Should NOT match
	}

	violations := rule.Check(files, nil)

	// Should have 1 violation
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if violations[0].Path != "src/UserAuth.ts" {
		t.Errorf("Expected violation for src/UserAuth.ts, got %s", violations[0].Path)
	}
}

func TestRegexMatchRule_IgnoresDirectories(t *testing.T) {
	rule := NewRegexMatchRule(map[string]string{
		"src/*": "^[a-z]+$",
	})

	files := []walker.FileInfo{
		{Path: "src/user.ts", IsDir: false},
		{Path: "src/UserDir", IsDir: true}, // Directory - should be ignored
	}

	violations := rule.Check(files, nil)

	// Should have 0 violations (directory is ignored, user.ts matches)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations, got %d", len(violations))
	}
}

func TestRegexMatchRule_FileExtensionHandling(t *testing.T) {
	// Regex should match filename without extension
	rule := NewRegexMatchRule(map[string]string{
		"src/*.test.ts": "^test[A-Z][a-zA-Z]+$", // Must be testCamelCase
	})

	files := []walker.FileInfo{
		{Path: "src/testUser.test.ts", IsDir: false},   // Should match
		{Path: "src/testHelper.test.ts", IsDir: false}, // Should match
		{Path: "src/userTest.test.ts", IsDir: false},   // Should NOT match
		{Path: "src/test.test.ts", IsDir: false},       // Should NOT match (no uppercase)
	}

	violations := rule.Check(files, nil)

	// Should have violations (at least for files that don't match)
	if len(violations) < 1 {
		t.Errorf("Expected at least 1 violation, got %d", len(violations))
	}

	// Check that userTest.test.ts has a violation (wrong pattern)
	foundUserTestViolation := false
	for _, v := range violations {
		if v.Path == "src/userTest.test.ts" {
			foundUserTestViolation = true
			break
		}
	}

	if !foundUserTestViolation {
		t.Error("Expected violation for src/userTest.test.ts")
	}
}

func TestRegexMatchRule_NoViolations(t *testing.T) {
	rule := NewRegexMatchRule(map[string]string{
		"src/*.ts": "^[a-z]+$",
	})

	files := []walker.FileInfo{
		{Path: "src/user.ts", IsDir: false},
		{Path: "src/admin.ts", IsDir: false},
		{Path: "src/helper.ts", IsDir: false},
	}

	violations := rule.Check(files, nil)

	if len(violations) != 0 {
		t.Errorf("Expected 0 violations, got %d", len(violations))
	}
}

func TestRegexMatchRule_Name(t *testing.T) {
	rule := NewRegexMatchRule(map[string]string{})

	if rule.Name() != "regex-match" {
		t.Errorf("Expected rule name 'regex-match', got '%s'", rule.Name())
	}
}
