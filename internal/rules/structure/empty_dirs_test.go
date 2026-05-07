package structure

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestEmptyDirsRule(t *testing.T) {
	dirs := map[string]*walker.DirInfo{
		"src":         {Path: "src", FileCount: 3},
		"empty":       {Path: "empty"},
		"has-subdirs": {Path: "has-subdirs", SubdirCount: 1},
	}
	v := NewEmptyDirsRule().Check(nil, dirs)
	if len(v) != 1 || v[0].Path != "empty" {
		t.Fatalf("want 1 violation on 'empty', got %+v", v)
	}
}
