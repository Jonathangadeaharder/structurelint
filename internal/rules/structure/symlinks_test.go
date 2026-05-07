package structure

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestSymlinksRule_FlagsSymlink(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "real.go"},
		{Path: "link.go", IsSymlink: true},
	}
	v := NewSymlinksRule().Check(files, nil)
	if len(v) != 1 || v[0].Path != "link.go" {
		t.Fatalf("want 1 violation on 'link.go', got %+v", v)
	}
}
