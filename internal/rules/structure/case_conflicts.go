package structure

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// CaseConflictsRule flags files (or directories) in the same directory whose
// names differ only by letter case. macOS and Windows filesystems are
// case-insensitive by default, so `Foo.ts` and `foo.ts` collide silently —
// fine on a developer's laptop, broken on Linux CI.
type CaseConflictsRule struct{}

func (r *CaseConflictsRule) Name() string { return "case-conflicts" }

func (r *CaseConflictsRule) Check(files []walker.FileInfo, _ map[string]*walker.DirInfo) []rules.Violation {
	type entry struct {
		original string
		isDir    bool
	}
	groups := make(map[string][]entry)

	for _, f := range files {
		dir := f.ParentPath
		base := filepath.Base(f.Path)
		key := dir + "\x00" + strings.ToLower(base)
		groups[key] = append(groups[key], entry{original: f.Path, isDir: f.IsDir})
	}

	var violations []rules.Violation
	for _, list := range groups {
		if len(list) < 2 {
			continue
		}
		paths := make([]string, 0, len(list))
		for _, e := range list {
			paths = append(paths, e.original)
		}
		sort.Strings(paths)
		for _, p := range paths[1:] {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    p,
				Message: fmt.Sprintf("name collides case-insensitively with %q", paths[0]),
				Suggestions: []string{
					"Rename one of the conflicting entries to a different stem",
					"Case-only differences break on case-insensitive filesystems (macOS, Windows)",
				},
			})
		}
	}
	return violations
}

func NewCaseConflictsRule() *CaseConflictsRule { return &CaseConflictsRule{} }
