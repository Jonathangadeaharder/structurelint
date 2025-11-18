package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadGitignorePatterns(t *testing.T) {
	tests := []struct {
		name          string
		gitignore     string
		wantPatterns  []string
		wantErr       bool
	}{
		{
			name: "basic patterns",
			gitignore: `# Comment
node_modules
*.log
dist/
`,
			wantPatterns: []string{
				"**/node_modules",
				"**/*.log",
				"dist/**",
			},
		},
		{
			name: "root-specific patterns",
			gitignore: `/build
/target
`,
			wantPatterns: []string{
				"build",
				"target",
			},
		},
		{
			name: "mixed patterns",
			gitignore: `# Build outputs
/bin
/obj
*.exe
temp/
.DS_Store
`,
			wantPatterns: []string{
				"bin",
				"obj",
				"**/*.exe",
				"temp/**",
				"**/.DS_Store",
			},
		},
		{
			name: "empty and comments only",
			gitignore: `# Just comments

# More comments
`,
			wantPatterns: []string{},
		},
		{
			name: "negation patterns (skipped)",
			gitignore: `*.log
!important.log
`,
			wantPatterns: []string{
				"**/*.log",
				// negation patterns are skipped
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory with .gitignore
			tmpDir := t.TempDir()
			gitignorePath := filepath.Join(tmpDir, ".gitignore")

			err := os.WriteFile(gitignorePath, []byte(tt.gitignore), 0644)
			if err != nil {
				t.Fatalf("Failed to create test .gitignore: %v", err)
			}

			got, err := LoadGitignorePatterns(tmpDir)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadGitignorePatterns() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if len(got) != len(tt.wantPatterns) {
				t.Errorf("LoadGitignorePatterns() got %d patterns, want %d\nGot: %v\nWant: %v",
					len(got), len(tt.wantPatterns), got, tt.wantPatterns)
				return
			}

			for i, pattern := range tt.wantPatterns {
				if got[i] != pattern {
					t.Errorf("LoadGitignorePatterns() pattern[%d] = %v, want %v", i, got[i], pattern)
				}
			}
		})
	}
}

func TestLoadGitignorePatternsNoFile(t *testing.T) {
	// Test with directory that has no .gitignore
	tmpDir := t.TempDir()

	got, err := LoadGitignorePatterns(tmpDir)
	if err != nil {
		t.Errorf("LoadGitignorePatterns() should not error when .gitignore doesn't exist, got: %v", err)
	}

	if len(got) != 0 {
		t.Errorf("LoadGitignorePatterns() should return empty list when .gitignore doesn't exist, got: %v", got)
	}
}

func TestNormalizeGitignorePattern(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"node_modules", "**/node_modules"},
		{"*.log", "**/*.log"},
		{"/build", "build"},
		{"/target/", "target/**"},
		{"dist/", "dist/**"},
		{".DS_Store", "**/.DS_Store"},
		{"src/temp", "src/temp"},
		{"", ""},
		{"  spaces  ", "**/spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeGitignorePattern(tt.input)
			if got != tt.want {
				t.Errorf("normalizeGitignorePattern(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestMergeWithGitignore(t *testing.T) {
	tests := []struct {
		name               string
		existing           []string
		gitignorePatterns  []string
		want               []string
	}{
		{
			name:     "merge without duplicates",
			existing: []string{"testdata/**", ".git/**"},
			gitignorePatterns: []string{"**/node_modules", "**/*.log"},
			want: []string{"testdata/**", ".git/**", "**/node_modules", "**/*.log"},
		},
		{
			name:     "remove duplicates",
			existing: []string{"testdata/**", "**/node_modules"},
			gitignorePatterns: []string{"**/node_modules", "**/*.log"},
			want: []string{"testdata/**", "**/node_modules", "**/*.log"},
		},
		{
			name:     "empty existing",
			existing: []string{},
			gitignorePatterns: []string{"**/node_modules", "**/*.log"},
			want: []string{"**/node_modules", "**/*.log"},
		},
		{
			name:     "empty gitignore",
			existing: []string{"testdata/**", ".git/**"},
			gitignorePatterns: []string{},
			want: []string{"testdata/**", ".git/**"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MergeWithGitignore(tt.existing, tt.gitignorePatterns)

			if len(got) != len(tt.want) {
				t.Errorf("MergeWithGitignore() got %d patterns, want %d\nGot: %v\nWant: %v",
					len(got), len(tt.want), got, tt.want)
				return
			}

			for i, pattern := range tt.want {
				if got[i] != pattern {
					t.Errorf("MergeWithGitignore() pattern[%d] = %v, want %v", i, got[i], pattern)
				}
			}
		})
	}
}
