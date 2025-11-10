package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestFileExistenceRule_Check(t *testing.T) {
	tests := []struct {
		name          string
		requirements  map[string]string
		files         []walker.FileInfo
		dirs          map[string]*walker.DirInfo
		wantViolCount int
	}{
		{
			name:         "file exists - satisfied",
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
			name:         "file missing - violation",
			requirements: map[string]string{"README.md": "exists:1"},
			files:        []walker.FileInfo{},
			dirs: map[string]*walker.DirInfo{
				"": {},
			},
			wantViolCount: 1,
		},
		{
			name:         "OR pattern - one match satisfies",
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
			name:         "range requirement - satisfied",
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
			name:         "range requirement - too many",
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
			name:         "multiple directories checked independently",
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
			rule := NewFileExistenceRule(tt.requirements)
			violations := rule.Check(tt.files, tt.dirs)

			if len(violations) != tt.wantViolCount {
				t.Errorf("Check() got %d violations, want %d", len(violations), tt.wantViolCount)
				for _, v := range violations {
					t.Logf("  - %s: %s", v.Path, v.Message)
				}
			}
		})
	}
}

func TestFileExistenceRule_parseCountSpec(t *testing.T) {
	rule := &FileExistenceRule{}

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
			gotMin, gotMax := rule.parseCountSpec(tt.spec)
			if gotMin != tt.wantMin || gotMax != tt.wantMax {
				t.Errorf("parseCountSpec(%q) = (%d, %d), want (%d, %d)",
					tt.spec, gotMin, gotMax, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestFileExistenceRule_Name(t *testing.T) {
	rule := NewFileExistenceRule(map[string]string{"README.md": "exists:1"})
	if got := rule.Name(); got != "file-existence" {
		t.Errorf("Name() = %v, want file-existence", got)
	}
}
