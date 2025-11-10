package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestMaxSubdirsRule_Check(t *testing.T) {
	tests := []struct {
		name          string
		maxSubdirs    int
		dirs          map[string]*walker.DirInfo
		wantViolCount int
	}{
		{
			name:       "no violations when subdir count within limit",
			maxSubdirs: 10,
			dirs: map[string]*walker.DirInfo{
				"src": {SubdirCount: 5},
				"lib": {SubdirCount: 8},
			},
			wantViolCount: 0,
		},
		{
			name:       "violation when subdir count exceeds limit",
			maxSubdirs: 5,
			dirs: map[string]*walker.DirInfo{
				"src":  {SubdirCount: 10},
				"lib":  {SubdirCount: 3},
				"test": {SubdirCount: 6},
			},
			wantViolCount: 2,
		},
		{
			name:       "exact count at limit",
			maxSubdirs: 5,
			dirs: map[string]*walker.DirInfo{
				"src": {SubdirCount: 5},
			},
			wantViolCount: 0,
		},
		{
			name:       "root directory violation",
			maxSubdirs: 3,
			dirs: map[string]*walker.DirInfo{
				"": {SubdirCount: 5},
			},
			wantViolCount: 1,
		},
		{
			name:          "empty dirs",
			maxSubdirs:    5,
			dirs:          map[string]*walker.DirInfo{},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewMaxSubdirsRule(tt.maxSubdirs)
			violations := rule.Check(nil, tt.dirs)

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

func TestMaxSubdirsRule_Name(t *testing.T) {
	rule := NewMaxSubdirsRule(10)
	if got := rule.Name(); got != "max-subdirs" {
		t.Errorf("Name() = %v, want max-subdirs", got)
	}
}
