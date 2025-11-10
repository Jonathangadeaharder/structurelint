package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestMaxDepthRule_WhenChecking(t *testing.T) {
	tests := []struct {
		name          string
		maxDepth      int
		files         []walker.FileInfo
		wantViolCount int
	}{
		{
			name:     "GivenDepthWithinLimit_WhenChecking_ThenReturnsNoViolations",
			maxDepth: 5,
			files: []walker.FileInfo{
				{Path: "a/b/c/file.go", Depth: 3},
				{Path: "a/b/file.go", Depth: 2},
				{Path: "file.go", Depth: 0},
			},
			wantViolCount: 0,
		},
		{
			name:     "GivenDepthExceedsLimit_WhenChecking_ThenReturnsViolation",
			maxDepth: 2,
			files: []walker.FileInfo{
				{Path: "a/b/c/file.go", Depth: 3},
				{Path: "a/b/file.go", Depth: 2},
			},
			wantViolCount: 1,
		},
		{
			name:     "GivenMultipleExcessiveDepths_WhenChecking_ThenReturnsMultipleViolations",
			maxDepth: 1,
			files: []walker.FileInfo{
				{Path: "a/b/c/file.go", Depth: 3},
				{Path: "a/b/file.go", Depth: 2},
				{Path: "a/file.go", Depth: 1},
			},
			wantViolCount: 2,
		},
		{
			name:     "GivenDepthAtLimit_WhenChecking_ThenReturnsNoViolations",
			maxDepth: 3,
			files: []walker.FileInfo{
				{Path: "a/b/c/file.go", Depth: 3},
			},
			wantViolCount: 0,
		},
		{
			name:          "WhenFileListEmpty_ThenReturnsNoViolations",
			maxDepth:      5,
			files:         []walker.FileInfo{},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rule := NewMaxDepthRule(tt.maxDepth)

			// Act
			violations := rule.Check(tt.files, nil)

			// Assert
			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
			}

			for _, v := range violations {
				if v.Rule != "max-depth" {
					t.Errorf("violation rule = %v, want max-depth", v.Rule)
				}
			}
		})
	}
}

func TestMaxDepthRule_WhenGettingName(t *testing.T) {
	// Arrange
	rule := NewMaxDepthRule(5)

	// Act
	got := rule.Name()

	// Assert
	if got != "max-depth" {
		t.Errorf("Name() = %v, want max-depth", got)
	}
}
