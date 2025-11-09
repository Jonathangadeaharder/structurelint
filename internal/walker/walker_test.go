package walker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWalker(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test directory structure
	// tmpDir/
	//   a.txt
	//   dir1/
	//     b.txt
	//     dir2/
	//       c.txt

	if err := os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	dir1 := filepath.Join(tmpDir, "dir1")
	if err := os.MkdirAll(dir1, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir1, "b.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	dir2 := filepath.Join(dir1, "dir2")
	if err := os.MkdirAll(dir2, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(dir2, "c.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	w := New(tmpDir)
	if err := w.Walk(); err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	files := w.GetFiles()

	// Should find 3 files and 2 directories
	fileCount := 0
	dirCount := 0
	for _, f := range files {
		if f.IsDir {
			dirCount++
		} else {
			fileCount++
		}
	}

	if fileCount != 3 {
		t.Errorf("Expected 3 files, got %d", fileCount)
	}

	if dirCount != 2 {
		t.Errorf("Expected 2 directories, got %d", dirCount)
	}
}

func TestWalker_Depth(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure
	deep := filepath.Join(tmpDir, "a", "b", "c")
	if err := os.MkdirAll(deep, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(deep, "file.txt"), []byte("test"), 0644); err != nil {
		t.Fatal(err)
	}

	w := New(tmpDir)
	if err := w.Walk(); err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	maxDepth := w.GetMaxDepth()

	// Depth should be 3 (a/=1, b/=2, c/=3, file.txt is still depth 3)
	// Directories increment depth, files don't increment beyond their parent
	if maxDepth != 3 {
		t.Errorf("Expected max depth 3, got %d", maxDepth)
	}
}

func TestWalker_DirInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory with multiple files and subdirs
	testDir := filepath.Join(tmpDir, "testdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Add 3 files
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(testDir, filepath.Base(tmpDir)+string(rune('a'+i-1))+".txt")
		if err := os.WriteFile(filename, []byte("test"), 0644); err != nil {
			t.Fatal(err)
		}
	}

	// Add 2 subdirectories
	for i := 1; i <= 2; i++ {
		subdir := filepath.Join(testDir, filepath.Base(tmpDir)+"sub"+string(rune('0'+i)))
		if err := os.MkdirAll(subdir, 0755); err != nil {
			t.Fatal(err)
		}
	}

	w := New(tmpDir)
	if err := w.Walk(); err != nil {
		t.Fatalf("Walk failed: %v", err)
	}

	dirs := w.GetDirs()

	testDirInfo, ok := dirs["testdir"]
	if !ok {
		t.Fatal("Expected to find testdir in dir info")
	}

	if testDirInfo.FileCount != 3 {
		t.Errorf("Expected 3 files, got %d", testDirInfo.FileCount)
	}

	if testDirInfo.SubdirCount != 2 {
		t.Errorf("Expected 2 subdirs, got %d", testDirInfo.SubdirCount)
	}
}

func TestMatchesPattern(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"file.ts", "*.ts", true},
		{"file.js", "*.ts", false},
		{"src/app.ts", "src/**/*.ts", true},
		{"src/components/Button.tsx", "src/**/*.tsx", true},
		{"test/app.ts", "src/**/*.ts", false},
		{"components/Button", "components/*/", true},
	}

	for _, tt := range tests {
		got := MatchesPattern(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("MatchesPattern(%q, %q) = %v, want %v",
				tt.path, tt.pattern, got, tt.want)
		}
	}
}
