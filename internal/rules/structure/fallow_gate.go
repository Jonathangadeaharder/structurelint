package structure

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// FallowGateRule validates that a fallow CI gate workflow exists
// for SvelteKit projects.
type FallowGateRule struct {
	rootDir string
}

// NewFallowGateRule creates a new FallowGateRule.
func NewFallowGateRule(rootDir string) *FallowGateRule {
	return &FallowGateRule{rootDir: rootDir}
}

func (r *FallowGateRule) Name() string { return "fallow-gate" }

func (r *FallowGateRule) Check(files []walker.FileInfo, _ map[string]*walker.DirInfo) []rules.Violation {
	// Skip if not a SvelteKit project
	if !r.isSvelteKit(files) {
		return nil
	}

	workflowsDir := filepath.Join(r.rootDir, ".github", "workflows")

	if !hasPrefixFile(workflowsDir, "fallow") {
		return []rules.Violation{
			{
				Rule:    r.Name(),
				Path:    ".github/workflows/fallow.yml",
				Message: "missing required fallow CI gate for SvelteKit project",
				Suggestions: []string{
					"Create .github/workflows/fallow.yml with fallow-rs/fallow@v2 action",
					"See: https://github.com/fallow-rs/fallow",
				},
			},
		}
	}

	return nil
}

// isSvelteKit checks if the project is a SvelteKit project by looking for
// svelte.config.js/ts or @sveltejs/kit in package.json.
func (r *FallowGateRule) isSvelteKit(files []walker.FileInfo) bool {
	hasSvelteConfig := false

	for _, f := range files {
		name := filepath.Base(f.Path)
		if name == "svelte.config.js" || name == "svelte.config.ts" {
			hasSvelteConfig = true
		}
		if name == "package.json" && r.hasSvelteKitDependency(f.AbsPath) {
			return true
		}
	}

	return hasSvelteConfig
}

func (r *FallowGateRule) hasSvelteKitDependency(pkgPath string) bool {
	data, err := os.ReadFile(pkgPath)
	if err != nil {
		return false
	}

	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false
	}

	if _, ok := pkg.Dependencies["@sveltejs/kit"]; ok {
		return true
	}
	if _, ok := pkg.DevDependencies["@sveltejs/kit"]; ok {
		return true
	}
	return false
}
