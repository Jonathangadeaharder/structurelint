package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestMaxFilesRule_Check(t *testing.T) {
	tests := []struct {
		name          string
		maxFiles      int
		dirs          map[string]*walker.DirInfo
		wantViolCount int
	}{
		{
			name:     "no violations when file count within limit",
			maxFiles: 10,
			dirs: map[string]*walker.DirInfo{
				"src": {FileCount: 5},
				"lib": {FileCount: 8},
			},
			wantViolCount: 0,
		},
		{
			name:     "violation when file count exceeds limit",
			maxFiles: 5,
			dirs: map[string]*walker.DirInfo{
				"src":  {FileCount: 10},
				"lib":  {FileCount: 3},
				"test": {FileCount: 6},
			},
			wantViolCount: 2,
		},
		{
			name:     "exact count at limit",
			maxFiles: 5,
			dirs: map[string]*walker.DirInfo{
				"src": {FileCount: 5},
			},
			wantViolCount: 0,
		},
		{
			name:     "root directory violation",
			maxFiles: 3,
			dirs: map[string]*walker.DirInfo{
				"": {FileCount: 5},
			},
			wantViolCount: 1,
		},
		{
			name:          "empty dirs",
			maxFiles:      5,
			dirs:          map[string]*walker.DirInfo{},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewMaxFilesRule(tt.maxFiles)
			violations := rule.Check(nil, tt.dirs)

			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
			}

			for _, v := range violations {
				if v.Rule != "max-files-in-dir" {
					t.Errorf("violation rule = %v, want max-files-in-dir", v.Rule)
				}
			}
		})
	}
}

func TestMaxFilesRule_Name(t *testing.T) {
	rule := NewMaxFilesRule(10)
	if got := rule.Name(); got != "max-files-in-dir" {
		t.Errorf("Name() = %v, want max-files-in-dir", got)
	}
}
