package structure

import (
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// SymlinksRule flags any symbolic link inside the project tree. Symlinks
// committed to a repo can escape the project root, break on Windows, and
// confuse build tools. Most projects don't want them; teams that do can
// disable the rule.
type SymlinksRule struct{}

func (r *SymlinksRule) Name() string { return "disallow-symlinks" }

func (r *SymlinksRule) Check(files []walker.FileInfo, _ map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation
	for _, f := range files {
		if f.IsSymlink {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    f.Path,
				Message: "symbolic link not allowed in project tree",
				Suggestions: []string{
					"Replace the symlink with the actual file or directory",
					"If a link is required for tooling, exclude it via the `exclude` config",
				},
			})
		}
	}
	return violations
}

func NewSymlinksRule() *SymlinksRule { return &SymlinksRule{} }
