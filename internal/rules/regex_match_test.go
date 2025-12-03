package rules

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestRegexMatchRule_BasicMatch(t *testing.T) {
	// Arrange
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

	// Act
	violations := rule.Check(files, nil)

	// Assert
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

// Mutation killing tests - target specific lived mutations

func TestRegexMatchRule_EmptyWildcardArray(t *testing.T) {
	// Test when pattern has no wildcards - tests line 97 conditions
	rule := NewRegexMatchRule(map[string]string{
		"src/domain/user.ts": "^user$", // No wildcards, exact path match
	})

	files := []walker.FileInfo{
		{Path: "src/domain/user.ts", IsDir: false},      // Should match (filename is "user")
		{Path: "src/domain/admin.ts", IsDir: false},     // Pattern doesn't apply
		{Path: "src/application/user.ts", IsDir: false}, // Pattern doesn't apply
	}

	violations := rule.Check(files, nil)

	// Should have 0 violations - the one matching file passes regex
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations with no wildcards, got %d", len(violations))
	}
}

func TestRegexMatchRule_SingleWildcardBoundary(t *testing.T) {
	// Tests line 97:11 - part == "*" condition
	rule := NewRegexMatchRule(map[string]string{
		"src/*/file.ts": "^${0}File$", // Single wildcard
	})

	files := []walker.FileInfo{
		{Path: "src/domain/domainFile.ts", IsDir: false}, // ${0} = domain, should match
		{Path: "src/app/appFile.ts", IsDir: false},       // ${0} = app, should match
		{Path: "src/user/file.ts", IsDir: false},         // ${0} = user, should NOT match (no "File" suffix)
	}

	violations := rule.Check(files, nil)

	// Should have 1 violation for the mismatched file
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
	}

	if len(violations) > 0 && violations[0].Path != "src/user/file.ts" {
		t.Errorf("Expected violation for src/user/file.ts, got %s", violations[0].Path)
	}
}

func TestRegexMatchRule_DoubleWildcardBoundary(t *testing.T) {
	// Tests line 97:26 - part == "**" condition
	rule := NewRegexMatchRule(map[string]string{
		"src/**/file.ts": "^file$", // Double wildcard (glob)
	})

	files := []walker.FileInfo{
		{Path: "src/file.ts", IsDir: false},               // Matches
		{Path: "src/domain/file.ts", IsDir: false},        // Matches
		{Path: "src/domain/models/file.ts", IsDir: false}, // Matches
		{Path: "src/domain/wrong.ts", IsDir: false},       // Doesn't match pattern
	}

	violations := rule.Check(files, nil)

	// All matching files should pass (filename is "file")
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations with ** wildcard, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s", v.Path)
		}
	}
}

func TestRegexMatchRule_IndexOutOfBounds(t *testing.T) {
	// Tests line 105:10 - idx < len(pathParts) boundary condition
	// Pattern has wildcard at position that doesn't exist in some file paths
	rule := NewRegexMatchRule(map[string]string{
		"src/*/*/file.ts": "^${0}-${1}$", // Expects 3 path segments before filename
	})

	files := []walker.FileInfo{
		{Path: "src/domain/models/file.ts", IsDir: false}, // Has enough segments: domain-models
		{Path: "src/app/wrong.ts", IsDir: false},          // Shorter path - pattern won't match anyway
	}

	violations := rule.Check(files, nil)

	// At least one violation should occur since paths don't match the expected pattern
	if len(violations) < 1 {
		t.Errorf("Expected at least 1 violation, got %d", len(violations))
	}
}

func TestRegexMatchRule_WildcardIndexBoundary(t *testing.T) {
	// Tests the boundary condition when wildcard index equals array length
	rule := NewRegexMatchRule(map[string]string{
		"src/*/file.ts": "^${0}File$",
	})

	files := []walker.FileInfo{
		{Path: "src/domain/domainFile.ts", IsDir: false}, // idx=1, pathParts=["src","domain","domainFile.ts"], 1 < 3 = true
		{Path: "src/x/xFile.ts", IsDir: false},           // Short path
	}

	violations := rule.Check(files, nil)

	// Both should pass (domainFile matches pattern, xFile matches pattern)
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations, got %d", len(violations))
	}
}

func TestRegexMatchRule_MultipleWildcardsEdgeCase(t *testing.T) {
	// Tests multiple wildcards with boundary conditions
	rule := NewRegexMatchRule(map[string]string{
		"*/*/*/file.ts": "^${0}-${1}-${2}$",
	})

	files := []walker.FileInfo{
		{Path: "a/b/c/a-b-c.ts", IsDir: false},     // Should match
		{Path: "x/y/z/file.ts", IsDir: false},      // Should NOT match (wrong filename)
		{Path: "short/path.ts", IsDir: false},      // Pattern won't apply
	}

	violations := rule.Check(files, nil)

	// Should have 1 violation for x/y/z/file.ts (filename is "file" but should be "x-y-z")
	foundViolation := false
	for _, v := range violations {
		if v.Path == "x/y/z/file.ts" {
			foundViolation = true
		}
	}

	if !foundViolation {
		t.Error("Expected violation for x/y/z/file.ts")
	}
}

func TestRegexMatchRule_NoSubstitutionNeeded(t *testing.T) {
	// Test when regex pattern has no ${n} placeholders
	rule := NewRegexMatchRule(map[string]string{
		"src/*/file.ts": "^file$", // No ${0} substitution
	})

	files := []walker.FileInfo{
		{Path: "src/domain/file.ts", IsDir: false},
		{Path: "src/app/file.ts", IsDir: false},
		{Path: "src/wrong/other.ts", IsDir: false}, // Doesn't match pattern
	}

	violations := rule.Check(files, nil)

	// All should pass (filename is "file")
	if len(violations) != 0 {
		t.Errorf("Expected 0 violations, got %d", len(violations))
	}
}
