package structure

import (
	"fmt"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/parser"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// DeepRelativeImportsRule flags relative imports that climb too many parent
// directories (e.g. `../../../foo`). Deep relative imports make code brittle
// to refactors and signal that the file should depend on a shared root via
// an absolute alias instead.
//
// Threshold defaults to 3 (`../../../` and deeper triggers).
type DeepRelativeImportsRule struct {
	MaxParents int
	parser     *parser.Parser
}

func (r *DeepRelativeImportsRule) Name() string { return "disallow-deep-relative-imports" }

func (r *DeepRelativeImportsRule) Check(files []walker.FileInfo, _ map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation
	if r.parser == nil {
		r.parser = parser.New("")
	}
	limit := r.MaxParents
	if limit <= 0 {
		limit = 3
	}

	for _, f := range files {
		if f.IsDir {
			continue
		}
		imports, err := r.parser.ParseFile(f.AbsPath)
		if err != nil {
			continue
		}
		for _, imp := range imports {
			if !imp.IsRelative {
				continue
			}
			parents := countLeadingParents(imp.ImportPath)
			if parents >= limit {
				violations = append(violations, rules.Violation{
					Rule:    r.Name(),
					Path:    f.Path,
					Message: fmt.Sprintf("relative import climbs %d parent directories (limit %d): %q", parents, limit, imp.ImportPath),
					Suggestions: []string{
						"Use an absolute import alias (e.g. configured via tsconfig paths or Go modules)",
						"Move the imported file closer, or extract the shared dependency to a common package",
					},
				})
			}
		}
	}
	return violations
}

func countLeadingParents(importPath string) int {
	parents := 0
	rest := strings.TrimPrefix(importPath, "./")
	for strings.HasPrefix(rest, "../") {
		parents++
		rest = rest[3:]
	}
	return parents
}

func NewDeepRelativeImportsRule(maxParents int) *DeepRelativeImportsRule {
	return &DeepRelativeImportsRule{MaxParents: maxParents}
}
