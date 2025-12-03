package rules

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestMaxSubdirsRule_WhenChecking(t *testing.T) {
	tests := []struct {
		name          string
		maxSubdirs    int
		dirs          map[string]*walker.DirInfo
		wantViolCount int
	}{
		{
			name:       "GivenSubdirCountWithinLimit_WhenChecking_ThenReturnsNoViolations",
			maxSubdirs: 10,
			dirs: map[string]*walker.DirInfo{
				"src": {SubdirCount: 5},
				"lib": {SubdirCount: 8},
			},
			wantViolCount: 0,
		},
		{
			name:       "GivenSubdirCountExceedsLimit_WhenChecking_ThenReturnsMultipleViolations",
			maxSubdirs: 5,
			dirs: map[string]*walker.DirInfo{
				"src":  {SubdirCount: 10},
				"lib":  {SubdirCount: 3},
				"test": {SubdirCount: 6},
			},
			wantViolCount: 2,
		},
		{
			name:       "GivenSubdirCountAtLimit_WhenChecking_ThenReturnsNoViolations",
			maxSubdirs: 5,
			dirs: map[string]*walker.DirInfo{
				"src": {SubdirCount: 5},
			},
			wantViolCount: 0,
		},
		{
			name:       "GivenRootDirectoryExceedsLimit_WhenChecking_ThenReturnsViolation",
			maxSubdirs: 3,
			dirs: map[string]*walker.DirInfo{
				"": {SubdirCount: 5},
			},
			wantViolCount: 1,
		},
		{
			name:          "WhenDirsEmpty_ThenReturnsNoViolations",
			maxSubdirs:    5,
			dirs:          map[string]*walker.DirInfo{},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			rule := NewMaxSubdirsRule(tt.maxSubdirs)

			// Act
			violations := rule.Check(nil, tt.dirs)

			// Assert
			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
			}

			for _, v := range violations {
				if v.Rule != "max-subdirs" {
					t.Errorf("violation rule = %v, want max-subdirs", v.Rule)
				}
			}
		})
	}
}

func TestMaxSubdirsRule_WhenGettingName(t *testing.T) {
	// Arrange
	rule := NewMaxSubdirsRule(10)

	// Act
	got := rule.Name()

	// Assert
	if got != "max-subdirs" {
		t.Errorf("Name() = %v, want max-subdirs", got)
	}
}
