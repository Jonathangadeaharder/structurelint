package init

import (
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func Test_extensionToLanguage(t *testing.T) {
	// Arrange
	// Act
	// Assert
	tests := []struct {
		ext  string
		want string
	}{
		{".go", "go"},
		{".py", "python"},
		{".ts", "typescript"},
		{".tsx", "typescript"},
		{".js", "javascript"},
		{".jsx", "javascript"},
		{".java", "java"},
		{".rs", "rust"},
		{".rb", "ruby"},
		{".cpp", "cpp"},
		{".unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.ext, func(t *testing.T) {
			if got := extensionToLanguage(tt.ext); got != tt.want {
				t.Errorf("extensionToLanguage(%q) = %q, want %q", tt.ext, got, tt.want)
			}
		})
	}
}

func Test_isTestFile(t *testing.T) {
	tests := []struct {
		path string
		lang string
		want bool
	}{
		{"main_test.go", "go", true},
		{"main.go", "go", false},
		{"test_module.py", "python", true},
		{"module.py", "python", false},
		{"component.test.ts", "typescript", true},
		{"component.ts", "typescript", false},
		{"FileTest.java", "java", true},
		{"File.java", "java", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isTestFile(tt.path, tt.lang); got != tt.want {
				t.Errorf("isTestFile(%q, %q) = %v, want %v", tt.path, tt.lang, got, tt.want)
			}
		})
	}
}

func Test_getTestFilePatterns(t *testing.T) {
	tests := []struct {
		lang         string
		wantContains string
	}{
		{"go", "_test"},
		{"python", "test_"},
		{"typescript", ".test"},
		{"javascript", ".spec"},
		{"java", "Test"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			patterns := getTestFilePatterns(tt.lang)
			found := false
			for _, p := range patterns {
				if p == tt.wantContains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("getTestFilePatterns(%q) should contain %q, got %v", tt.lang, tt.wantContains, patterns)
			}
		})
	}
}

func Test_getSourcePatterns(t *testing.T) {
	tests := []struct {
		lang         string
		wantContains string
	}{
		{"go", "**/*.go"},
		{"python", "**/*.py"},
		{"typescript", "**/*.ts"},
		{"javascript", "**/*.js"},
	}

	for _, tt := range tests {
		t.Run(tt.lang, func(t *testing.T) {
			patterns := getSourcePatterns(tt.lang)
			found := false
			for _, p := range patterns {
				if p == tt.wantContains {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("getSourcePatterns(%q) should contain %q, got %v", tt.lang, tt.wantContains, patterns)
			}
		})
	}
}

func Test_calculateMaxDepth(t *testing.T) {
	tests := []struct {
		name      string
		files     []walker.FileInfo
		wantMin   int
		wantMax   int
		wantExact int
	}{
		{
			name: "shallow project",
			files: []walker.FileInfo{
				{Depth: 1},
				{Depth: 2},
			},
			wantExact: 4, // 2 + 2 buffer
		},
		{
			name: "deep project",
			files: []walker.FileInfo{
				{Depth: 8},
				{Depth: 9},
			},
			wantExact: 10, // capped at 10
		},
		{
			name: "very shallow",
			files: []walker.FileInfo{
				{Depth: 1},
			},
			wantExact: 4, // minimum of 4
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMaxDepth(tt.files)
			if tt.wantExact > 0 && got != tt.wantExact {
				t.Errorf("calculateMaxDepth() = %d, want %d", got, tt.wantExact)
			}
		})
	}
}

func Test_calculateMaxFilesInDir(t *testing.T) {
	tests := []struct {
		name      string
		dirs      map[string]*walker.DirInfo
		wantExact int
	}{
		{
			name: "moderate file counts",
			dirs: map[string]*walker.DirInfo{
				"src": {FileCount: 10},
				"lib": {FileCount: 15},
			},
			wantExact: 20, // 15 * 1.2 = 18, but minimum is 20
		},
		{
			name: "very few files",
			dirs: map[string]*walker.DirInfo{
				"src": {FileCount: 5},
			},
			wantExact: 20, // minimum of 20
		},
		{
			name: "many files",
			dirs: map[string]*walker.DirInfo{
				"src": {FileCount: 100},
			},
			wantExact: 100, // capped at 100
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calculateMaxFilesInDir(tt.dirs)
			if got != tt.wantExact {
				t.Errorf("calculateMaxFilesInDir() = %d, want %d", got, tt.wantExact)
			}
		})
	}
}

func Test_detectDocumentationStyle(t *testing.T) {
	tests := []struct {
		name  string
		files []walker.FileInfo
		want  string
	}{
		{
			name: "comprehensive documentation",
			files: []walker.FileInfo{
				{Path: "src", IsDir: true},
				{Path: "lib", IsDir: true},
				{Path: "README.md", IsDir: false},
				{Path: "src/README.md", IsDir: false},
			},
			want: "comprehensive", // 2 READMEs / 2 dirs > 0.5
		},
		{
			name: "minimal documentation",
			files: []walker.FileInfo{
				{Path: "src", IsDir: true},
				{Path: "lib", IsDir: true},
				{Path: "tests", IsDir: true},
				{Path: "README.md", IsDir: false},
			},
			want: "minimal", // 1 README / 3 dirs < 0.5
		},
		{
			name: "no documentation",
			files: []walker.FileInfo{
				{Path: "src", IsDir: true},
				{Path: "lib", IsDir: true},
			},
			want: "none",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := detectDocumentationStyle(tt.files)
			if got != tt.want {
				t.Errorf("detectDocumentationStyle() = %q, want %q", got, tt.want)
			}
		})
	}
}
