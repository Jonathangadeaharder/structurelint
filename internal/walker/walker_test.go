package walker

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWalker(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()

	// Create test directory structure
	// tmpDir/
	//   a.txt
	//   dir1/
	//     b.txt
	//     dir2/
	//       c.txt

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "a.txt"), []byte("test"), 0644))

	dir1 := filepath.Join(tmpDir, "dir1")
	require.NoError(t, os.MkdirAll(dir1, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir1, "b.txt"), []byte("test"), 0644))

	dir2 := filepath.Join(dir1, "dir2")
	require.NoError(t, os.MkdirAll(dir2, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(dir2, "c.txt"), []byte("test"), 0644))

	// Act
	w := New(tmpDir)
	require.NoError(t, w.Walk())

	files := w.GetFiles()

	// Assert
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

	assert.Equal(t, 3, fileCount)
	assert.Equal(t, 2, dirCount)
}

func TestWalker_Depth(t *testing.T) {
	tmpDir := t.TempDir()

	// Create nested structure
	deep := filepath.Join(tmpDir, "a", "b", "c")
	require.NoError(t, os.MkdirAll(deep, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(deep, "file.txt"), []byte("test"), 0644))

	w := New(tmpDir)
	require.NoError(t, w.Walk())

	maxDepth := w.GetMaxDepth()

	// Depth should be 3 (a/=1, b/=2, c/=3, file.txt is still depth 3)
	// Directories increment depth, files don't increment beyond their parent
	assert.Equal(t, 3, maxDepth)
}

func TestWalker_DirInfo(t *testing.T) {
	tmpDir := t.TempDir()

	// Create directory with multiple files and subdirs
	testDir := filepath.Join(tmpDir, "testdir")
	require.NoError(t, os.MkdirAll(testDir, 0755))

	// Add 3 files
	for i := 1; i <= 3; i++ {
		filename := filepath.Join(testDir, filepath.Base(tmpDir)+string(rune('a'+i-1))+".txt")
		require.NoError(t, os.WriteFile(filename, []byte("test"), 0644))
	}

	// Add 2 subdirectories
	for i := 1; i <= 2; i++ {
		subdir := filepath.Join(testDir, filepath.Base(tmpDir)+"sub"+string(rune('0'+i)))
		require.NoError(t, os.MkdirAll(subdir, 0755))
	}

	w := New(tmpDir)
	require.NoError(t, w.Walk())

	dirs := w.GetDirs()

	testDirInfo, ok := dirs["testdir"]
	require.True(t, ok, "Expected to find testdir in dir info")

	assert.Equal(t, 3, testDirInfo.FileCount)
	assert.Equal(t, 2, testDirInfo.SubdirCount)
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
		assert.Equal(t, tt.want, got, "MatchesPattern(%q, %q)", tt.path, tt.pattern)
	}
}
