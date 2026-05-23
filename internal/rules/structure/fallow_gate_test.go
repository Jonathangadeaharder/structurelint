package structure

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFallowGate_NonSvelteKit_NoViolations(t *testing.T) {
	dir := t.TempDir()
	rule := NewFallowGateRule(dir)
	files, dirs := walkDir(t, dir)

	v := rule.Check(files, dirs)
	if len(v) != 0 {
		t.Fatalf("want 0 violations for non-SvelteKit project, got %d: %+v", len(v), v)
	}
}

func TestFallowGate_SvelteKitWithoutFallow_HasViolation(t *testing.T) {
	dir := t.TempDir()

	// Create svelte.config.js in root
	writeFile(t, dir, "svelte.config.js", "export default {};\n")
	// Create .github/workflows dir but no fallow file
	workflowsDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		t.Fatalf("mkdir .github/workflows: %v", err)
	}
	writeFile(t, workflowsDir, "pr-gate.yml", "name: PR Gate\n")

	rule := NewFallowGateRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)

	if len(v) != 1 {
		t.Fatalf("want 1 violation (fallow missing), got %d: %+v", len(v), v)
	}
	if v[0].Path != ".github/workflows/fallow.yml" {
		t.Fatalf("violation should target fallow.yml, got path=%q", v[0].Path)
	}
}

func TestFallowGate_SvelteKitWithFallow_NoViolations(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "svelte.config.js", "export default {};\n")
	workflowsDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		t.Fatalf("mkdir .github/workflows: %v", err)
	}
	writeFile(t, workflowsDir, "fallow.yml", "name: Fallow\n")

	rule := NewFallowGateRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)

	if len(v) != 0 {
		t.Fatalf("want 0 violations (fallow present), got %d: %+v", len(v), v)
	}
}

func TestFallowGate_SvelteKitPackageJSON_DetectsDependency(t *testing.T) {
	dir := t.TempDir()

	// Create package.json with @sveltejs/kit dependency
	pkg := `{
		"name": "test-app",
		"dependencies": {
			"@sveltejs/kit": "^2.0.0"
		}
	}`
	writeFile(t, dir, "package.json", pkg)

	rule := NewFallowGateRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)

	if len(v) != 1 {
		t.Fatalf("want 1 violation (fallow missing, sveltekit detected via package.json), got %d: %+v", len(v), v)
	}
}

func TestFallowGate_SvelteKitDevDependency_Detects(t *testing.T) {
	dir := t.TempDir()

	pkg := `{
		"name": "test-app",
		"devDependencies": {
			"@sveltejs/kit": "^2.0.0"
		}
	}`
	writeFile(t, dir, "package.json", pkg)

	rule := NewFallowGateRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)

	if len(v) != 1 {
		t.Fatalf("want 1 violation (sveltekit in devDeps), got %d: %+v", len(v), v)
	}
}

func TestFallowGate_FallowPrefixMatching(t *testing.T) {
	dir := t.TempDir()

	writeFile(t, dir, "svelte.config.ts", "export default {};\n")
	workflowsDir := filepath.Join(dir, ".github", "workflows")
	if err := os.MkdirAll(workflowsDir, 0o755); err != nil {
		t.Fatalf("mkdir .github/workflows: %v", err)
	}
	writeFile(t, workflowsDir, "fallow-pr-gate.yml", "name: Fallow PR Gate\n")

	rule := NewFallowGateRule(dir)
	files, dirs := walkDir(t, dir)
	v := rule.Check(files, dirs)

	if len(v) != 0 {
		t.Fatalf("want 0 violations (prefix matched fallow-), got %d: %+v", len(v), v)
	}
}
