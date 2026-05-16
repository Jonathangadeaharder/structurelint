package ci

import (
	"encoding/json"
	"path/filepath"
	"strings"
)

type ProjectDetector struct {
	reader FileReader
}

func NewProjectDetector(reader FileReader) *ProjectDetector {
	return &ProjectDetector{reader: reader}
}

func (d *ProjectDetector) Detect(files []FileInfo) []ProjectType {
	var types []ProjectType
	hasSvelteKit := false
	hasPython := false
	hasGo := false
	hasRust := false

	for _, f := range files {
		base := filepath.Base(f.Path)
		path := filepath.ToSlash(f.Path)
		switch {
		case base == "go.mod":
			hasGo = true
		case base == "Cargo.toml":
			hasRust = true
		case base == "pyproject.toml" || base == "setup.py" || base == "setup.cfg":
			hasPython = true
		case base == "package.json":
			isSK, _ := d.isSvelteKit(f)
			if isSK {
				hasSvelteKit = true
			}
		default:
			if strings.Contains(path, "svelte.config") && (strings.HasSuffix(base, ".js") || strings.HasSuffix(base, ".ts")) {
				hasSvelteKit = true
			}
		}
	}

	if hasSvelteKit {
		types = append(types, SvelteKit)
	}
	if hasPython {
		types = append(types, Python)
	}
	if hasGo {
		types = append(types, Go)
	}
	if hasRust {
		types = append(types, Rust)
	}
	return types
}

func (d *ProjectDetector) isSvelteKit(f FileInfo) (bool, error) {
	if d.reader == nil {
		return false, nil
	}
	data, err := d.reader.ReadFile(f.AbsPath)
	if err != nil {
		return false, err
	}
	var pkg struct {
		Dependencies    map[string]string `json:"dependencies"`
		DevDependencies map[string]string `json:"devDependencies"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return false, err
	}
	if _, ok := pkg.Dependencies["svelte"]; ok {
		return true, nil
	}
	if _, ok := pkg.DevDependencies["svelte"]; ok {
		return true, nil
	}
	return false, nil
}
