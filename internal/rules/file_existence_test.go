package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestFileExistenceRule_WhenChecking(t *testing.T) {
	tests := []struct {
		name          string
		requirements  map[string]string
		files         []walker.FileInfo
		dirs          map[string]*walker.DirInfo
		wantViolCount int
	}{
		{
			name:         "GivenFileExists_WhenChecking_ThenReturnsSatisfied",
			requirements: map[string]string{"README.md": "exists:1"},
			files: []walker.FileInfo{
				{Path: "README.md", ParentPath: ""},
			},
			dirs: map[string]*walker.DirInfo{
				"": {},
			},
			wantViolCount: 0,
		},
		{
			name:         "GivenFileMissing_WhenChecking_ThenReturnsViolation",
			requirements: map[string]string{"README.md": "exists:1"},
			files:        []walker.FileInfo{},
			dirs: map[string]*walker.DirInfo{
				"": {},
			},
			wantViolCount: 1,
		},
		{
			name:         "GivenORPattern_WhenOneMatches_ThenReturnsSatisfied",
			requirements: map[string]string{"index.ts|index.js": "exists:1"},
			files: []walker.FileInfo{
				{Path: "index.js", ParentPath: ""},
			},
			dirs: map[string]*walker.DirInfo{
				"": {},
			},
			wantViolCount: 0,
		},
		{
			name:         "GivenRangeRequirement_WhenCountWithinRange_ThenReturnsSatisfied",
			requirements: map[string]string{"*.md": "exists:1-5"},
			files: []walker.FileInfo{
				{Path: "README.md", ParentPath: ""},
				{Path: "CHANGELOG.md", ParentPath: ""},
			},
			dirs: map[string]*walker.DirInfo{
				"": {},
			},
			wantViolCount: 0,
		},
		{
			name:         "GivenRangeRequirement_WhenTooManyFiles_ThenReturnsViolation",
			requirements: map[string]string{"*.md": "exists:1-2"},
			files: []walker.FileInfo{
				{Path: "README.md", ParentPath: ""},
				{Path: "CHANGELOG.md", ParentPath: ""},
				{Path: "CONTRIBUTING.md", ParentPath: ""},
			},
			dirs: map[string]*walker.DirInfo{
				"": {},
			},
			wantViolCount: 1,
		},
		{
			name:         "GivenMultipleDirectories_WhenChecking_ThenEachCheckedIndependently",
			requirements: map[string]string{"README.md": "exists:1"},
			files: []walker.FileInfo{
				{Path: "README.md", ParentPath: ""},
				{Path: "src/utils.go", ParentPath: "src"},
			},
			dirs: map[string]*walker.DirInfo{
				"":    {},
				"src": {},
			},
			wantViolCount: 1, // src/ missing README.md
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rule := NewFileExistenceRule(tt.requirements)

			// Act
			violations := rule.Check(tt.files, tt.dirs)

			// Assert
			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
				for _, v := range violations {
					t.Logf("  - %s: %s", v.Path, v.Message)
				}
			}
		})
	}
}

func TestFileExistenceRule_WhenParsingCountSpec(t *testing.T) {
	tests := []struct {
		spec    string
		wantMin int
		wantMax int
	}{
		{"1", 1, 1},
		{"0", 0, 0},
		{"5", 5, 5},
		{"1-5", 1, 5},
		{"0-10", 0, 10},
		{"2-2", 2, 2},
	}

	for _, tt := range tests {
		t.Run(tt.spec, func(t *testing.T) {
			// Arrange
			rule := &FileExistenceRule{}

			// Act
			gotMin, gotMax := rule.parseCountSpec(tt.spec)

			// Assert
			if gotMin != tt.wantMin || gotMax != tt.wantMax {
				t.Errorf("parseCountSpec(%q) = (%d, %d), want (%d, %d)",
					tt.spec, gotMin, gotMax, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFileExistenceRule_WhenGettingName(t *testing.T) {
	// Arrange
	rule := NewFileExistenceRule(map[string]string{"README.md": "exists:1"})

	// Act
	got := rule.Name()

	// Assert
	if got != "file-existence" {
		t.Errorf("Name() = %v, want file-existence", got)
	}
}
