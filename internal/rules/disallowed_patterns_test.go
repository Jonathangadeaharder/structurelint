package rules

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestDisallowedPatternsRule_Check(t *testing.T) {
	tests := []struct {
		name          string
		patterns      []string
		files         []walker.FileInfo
		wantViolCount int
		wantPaths     []string
	}{
		{
			name:     "simple pattern match",
			patterns: []string{"*.tmp"},
			files: []walker.FileInfo{
				{Path: "file.tmp"},
				{Path: "file.go"},
			},
			wantViolCount: 1,
			wantPaths:     []string{"file.tmp"},
		},
		{
			name:     "glob pattern with **",
			patterns: []string{"**/*.log"},
			files: []walker.FileInfo{
				{Path: "logs/app.log"},
				{Path: "src/test.log"},
				{Path: "file.go"},
			},
			wantViolCount: 2,
		},
		{
			name:     "exact match - basename matching",
			patterns: []string{".DS_Store"},
			files: []walker.FileInfo{
				{Path: ".DS_Store"},
				{Path: "src/.DS_Store"},
				{Path: "file.go"},
			},
			wantViolCount: 2, // Matches basename in both locations
		},
		{
			name:     "negation pattern allows exceptions",
			patterns: []string{"*.md", "!README.md"},
			files: []walker.FileInfo{
				{Path: "README.md"},
				{Path: "CHANGELOG.md"},
				{Path: "docs/guide.md"},
			},
			wantViolCount: 2,
		},
		{
			name:     "directory pattern",
			patterns: []string{"node_modules/**"},
			files: []walker.FileInfo{
				{Path: "node_modules/package/index.js"},
				{Path: "src/index.js"},
			},
			wantViolCount: 1,
		},
		{
			name:     "multiple patterns",
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
			name:          "no violations",
			patterns:      []string{"*.tmp"},
			files:         []walker.FileInfo{{Path: "file.go"}},
			wantViolCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewDisallowedPatternsRule(tt.patterns)
			violations := rule.Check(tt.files, nil)

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

func TestDisallowedPatternsRule_Name(t *testing.T) {
	rule := NewDisallowedPatternsRule([]string{"*.tmp"})
	if got := rule.Name(); got != "disallowed-patterns" {
		t.Errorf("Name() = %v, want disallowed-patterns", got)
	}
}

func Test_matchesGlobPattern(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		pattern string
		want    bool
	}{
		{"exact match", "file.tmp", "file.tmp", true},
		{"glob match", "file.tmp", "*.tmp", true},
		{"glob no match", "file.go", "*.tmp", false},
		{"double star prefix", "src/test/file.go", "**/file.go", true},
		{"double star middle", "src/test/file.go", "src/**/file.go", true},
		{"double star no match", "other/file.go", "src/**/*.go", false},
		{"full path pattern", "src/components/Button.tsx", "src/components/*.tsx", true},
		{"basename only", "deep/nested/file.tmp", "*.tmp", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := matchesGlobPattern(tt.path, tt.pattern); got != tt.want {
				t.Errorf("matchesGlobPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
			}
		})
	}
}
