package structure

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestCaseConflictsRule_DetectsCollision(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "src/Foo.ts", ParentPath: "src"},
		{Path: "src/foo.ts", ParentPath: "src"},
		{Path: "src/Bar.ts", ParentPath: "src"},
	}
	v := NewCaseConflictsRule().Check(files, nil)
	if len(v) != 1 {
		t.Fatalf("want 1 violation, got %d (%+v)", len(v), v)
	}
}

func TestCaseConflictsRule_NoCollision(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "a.ts", ParentPath: ""},
		{Path: "b.ts", ParentPath: ""},
	}
	if v := NewCaseConflictsRule().Check(files, nil); len(v) != 0 {
		t.Errorf("want no violations, got %d", len(v))
	}
}

func TestCaseConflictsRule_ScopedPerDirectory(t *testing.T) {
	files := []walker.FileInfo{
		{Path: "src/Foo.ts", ParentPath: "src"},
		{Path: "lib/foo.ts", ParentPath: "lib"},
	}
	if v := NewCaseConflictsRule().Check(files, nil); len(v) != 0 {
		t.Errorf("collisions should be per-directory; got %d violations", len(v))
	}
}
