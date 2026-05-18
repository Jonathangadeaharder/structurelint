package structure

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// CIGatesRule validates that CI workflow files exist in .github/workflows/.
type CIGatesRule struct {
	rootDir string
}

// NewCIGatesRule creates a new CIGatesRule.
func NewCIGatesRule(rootDir string) *CIGatesRule {
	return &CIGatesRule{rootDir: rootDir}
}

func (r *CIGatesRule) Name() string { return "ci-gates" }

func (r *CIGatesRule) Check(files []walker.FileInfo, _ map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation

	workflowsDir := filepath.Join(r.rootDir, ".github", "workflows")

	// Check for pr-gate file.
	if !hasPrefixFile(workflowsDir, "pr-gate") {
		violations = append(violations, rules.Violation{
			Rule:    r.Name(),
			Path:    ".github/workflows/pr-gate.yml",
			Message: "missing required CI workflow: pr-gate.yml",
			Suggestions: []string{
				"Create .github/workflows/pr-gate.yml with PR quality checks",
				"See https://docs.github.com/en/actions/using-workflows",
			},
		})
	}

	// Check for merge-gate file.
	if !hasPrefixFile(workflowsDir, "merge-gate") {
		violations = append(violations, rules.Violation{
			Rule:    r.Name(),
			Path:    ".github/workflows/merge-gate.yml",
			Message: "missing required CI workflow: merge-gate.yml",
			Suggestions: []string{
				"Create .github/workflows/merge-gate.yml with merge quality checks",
				"See https://docs.github.com/en/actions/using-workflows",
			},
		})
	}

	return violations
}

// hasPrefixFile checks if any file in dir has a name starting with the given prefix.
func hasPrefixFile(dir, prefix string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := filepath.Ext(name)
		if (ext == ".yml" || ext == ".yaml") && strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
