package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestMaxDepthRule_Check(t *testing.T) {
	tests := []struct {
		name          string
		maxDepth      int
		files         []walker.FileInfo
		wantViolCount int
	}{
		{
			name:     "no violations when depth within limit",
			maxDepth: 5,
			files: []walker.FileInfo{
				{Path: "a/b/c/file.go", Depth: 3},
				{Path: "a/b/file.go", Depth: 2},
				{Path: "file.go", Depth: 0},
			},
			wantViolCount: 0,
		},
		{
			name:     "violation when depth exceeds limit",
			maxDepth: 2,
			files: []walker.FileInfo{
				{Path: "a/b/c/file.go", Depth: 3},
				{Path: "a/b/file.go", Depth: 2},
			},
			wantViolCount: 1,
		},
		{
			name:     "multiple violations",
			maxDepth: 1,
			files: []walker.FileInfo{
				{Path: "a/b/c/file.go", Depth: 3},
				{Path: "a/b/file.go", Depth: 2},
				{Path: "a/file.go", Depth: 1},
			},
			wantViolCount: 2,
		},
		{
			name:     "exact depth at limit",
			maxDepth: 3,
			files: []walker.FileInfo{
				{Path: "a/b/c/file.go", Depth: 3},
			},
			wantViolCount: 0,
		},
		{
			name:          "empty file list",
			maxDepth:      5,
			files:         []walker.FileInfo{},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewMaxDepthRule(tt.maxDepth)
			violations := rule.Check(tt.files, nil)

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

func TestMaxDepthRule_Name(t *testing.T) {
	rule := NewMaxDepthRule(5)
	if got := rule.Name(); got != "max-depth" {
		t.Errorf("Name() = %v, want max-depth", got)
	}
}
