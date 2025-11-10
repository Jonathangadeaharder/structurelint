package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestDisallowedPatternsRule_WhenChecking(t *testing.T) {
	tests := []struct {
		name          string
		patterns      []string
		files         []walker.FileInfo
		wantViolCount int
		wantPaths     []string
	}{
		{
			name:     "GivenSimplePattern_WhenMatching_ThenFindsViolation",
			patterns: []string{"*.tmp"},
			files: []walker.FileInfo{
				{Path: "file.tmp"},
				{Path: "file.go"},
			},
			wantViolCount: 1,
			wantPaths:     []string{"file.tmp"},
		},
		{
			name:     "GivenGlobPatternWithDoubleStar_WhenMatching_ThenFindsMultipleViolations",
			patterns: []string{"**/*.log"},
			files: []walker.FileInfo{
				{Path: "logs/app.log"},
				{Path: "src/test.log"},
				{Path: "file.go"},
			},
			wantViolCount: 2,
		},
		{
			name:     "GivenBasenamePattern_WhenMatchingAcrossDirectories_ThenFindsAllMatches",
			patterns: []string{".DS_Store"},
			files: []walker.FileInfo{
				{Path: ".DS_Store"},
				{Path: "src/.DS_Store"},
				{Path: "file.go"},
			},
			wantViolCount: 2, // Matches basename in both locations
		},
		{
			name:     "GivenNegationPattern_WhenEvaluating_ThenAllowsExceptions",
			patterns: []string{"*.md", "!README.md"},
			files: []walker.FileInfo{
				{Path: "README.md"},
				{Path: "CHANGELOG.md"},
				{Path: "docs/guide.md"},
			},
			wantViolCount: 2,
		},
		{
			name:     "GivenDirectoryPattern_WhenMatching_ThenFindsViolation",
			patterns: []string{"node_modules/**"},
			files: []walker.FileInfo{
				{Path: "node_modules/package/index.js"},
				{Path: "src/index.js"},
			},
			wantViolCount: 1,
		},
		{
			name:     "GivenMultiplePatterns_WhenMatching_ThenFindsAllViolations",
			patterns: []string{"*.tmp", "*.bak", ".DS_Store"},
			files: []walker.FileInfo{
				{Path: "file.tmp"},
				{Path: "backup.bak"},
				{Path: ".DS_Store"},
				{Path: "file.go"},
			},
			wantViolCount: 3,
		},
		{
			name:          "WhenNoMatches_ThenReturnsNoViolations",
			patterns:      []string{"*.tmp"},
			files:         []walker.FileInfo{{Path: "file.go"}},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rule := NewDisallowedPatternsRule(tt.patterns)

			// Act
			violations := rule.Check(tt.files, nil)

			// Assert
			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
				for _, v := range violations {
					t.Logf("  - %s: %s", v.Path, v.Message)
				}
			}

			if tt.wantPaths != nil {
				gotPaths := make([]string, len(violations))
				for i, v := range violations {
					gotPaths[i] = v.Path
				}
				if len(gotPaths) != len(tt.wantPaths) {
					t.Errorf("got paths %v, want %v", gotPaths, tt.wantPaths)
				}
			}
		})
	}
}

func TestDisallowedPatternsRule_WhenGettingName(t *testing.T) {
	// Arrange
	rule := NewDisallowedPatternsRule([]string{"*.tmp"})

	// Act
	got := rule.Name()

	// Assert
	if got != "disallowed-patterns" {
		t.Errorf("Name() = %v, want disallowed-patterns", got)
	}
}

func Test_WhenMatchingGlobPattern(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{"WhenExactMatch_ThenReturnsTrue", "file.tmp", "file.tmp", true},
		{"WhenGlobMatches_ThenReturnsTrue", "file.tmp", "*.tmp", true},
		{"WhenGlobDoesNotMatch_ThenReturnsFalse", "file.go", "*.tmp", false},
		{"WhenDoubleStarPrefix_ThenMatchesNestedPath", "src/test/file.go", "**/file.go", true},
		{"WhenDoubleStarMiddle_ThenMatchesPath", "src/test/file.go", "src/**/file.go", true},
		{"WhenDoubleStarNoMatch_ThenReturnsFalse", "other/file.go", "src/**/*.go", false},
		{"WhenFullPathPattern_ThenMatches", "src/components/Button.tsx", "src/components/*.tsx", true},
		{"WhenBasenameOnly_ThenMatchesInAnyDirectory", "deep/nested/file.tmp", "*.tmp", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange & Act
			got := matchesGlobPattern(tt.path, tt.pattern)

			// Assert
			if got != tt.want {
				t.Errorf("matchesGlobPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}
