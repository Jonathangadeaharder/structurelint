# CI Quality Gates Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task.

**Goal:** Refactor `github-workflows` structurelint rule from flat checks into strategy-pattern per project type (SvelteKit, Python, Go, Rust) with registry pattern, DI, and testability.

**Architecture:** Strategy pattern per project type, registry for strategy discovery, FileReader interface for DI. Rule detects project type(s) from files, dispatches to matching strategies. Each strategy enforces project-specific CI gates, linter presence, coverage thresholds, and suppression limits.

**Tech Stack:** Go, `gopkg.in/yaml.v3`, testing with `testing` package (Go stdlib mocks)

---

### Task 1: Define shared types (types.go)

**Files:**
- Create: `internal/rules/ci/types.go`

- [ ] **Step 1: Write types.go**

```go
package ci

import "github.com/Jonathangadeaharder/structurelint/internal/rules"

type ProjectType string

const (
	SvelteKit ProjectType = "sveltekit"
	Python    ProjectType = "python"
	Go        ProjectType = "golang"
	Rust      ProjectType = "rust"
)

type CoverageThresholds struct {
	Branches   float64
	Lines      float64
	Functions  float64
	Statements float64
}

type CIGate struct {
	Name     string
	Required bool
	Hint     string
}

type LinterTool struct {
	Name     string
	Required bool
	Hint     string
}

type CheckResult struct {
	Rule    string
	Path    string
	Message string
	Fix     string
}

func (c CheckResult) ToViolation() rules.Violation {
	v := rules.Violation{
		Rule:    "github-workflows",
		Path:    c.Path,
		Message: c.Message,
	}
	if c.Fix != "" {
		v.Suggestions = []string{c.Fix}
	}
	return v
}
```

- [ ] **Step 2: Verify compiles**

Run: `go build ./internal/rules/ci/`
Expected: success

- [ ] **Step 3: Commit**

```bash
git add internal/rules/ci/types.go
git commit -m "feat(ci): define shared types for strategy pattern"
```

---

### Task 2: FileReader interface for DI (file_reader.go)

**Files:**
- Create: `internal/rules/ci/file_reader.go`
- Test: `internal/rules/ci/file_reader_test.go`

- [ ] **Step 1: Write file_reader.go**

```go
package ci

type FileReader interface {
	ReadFile(path string) ([]byte, error)
}

type OSFileReader struct{}

func (OSFileReader) ReadFile(path string) ([]byte, error) {
	return os.ReadFile(path)
}

type MockFileReader struct {
	Files map[string]string
}

func (m MockFileReader) ReadFile(path string) ([]byte, error) {
	content, ok := m.Files[path]
	if !ok {
		return nil, fmt.Errorf("file not found: %s", path)
	}
	return []byte(content), nil
}
```

- [ ] **Step 2: Write file_reader_test.go**

```go
package ci

import (
	"testing"
)

func TestMockFileReader(t *testing.T) {
	r := MockFileReader{Files: map[string]string{
		"test.txt": "hello",
	}}
	data, err := r.ReadFile("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello" {
		t.Fatalf("expected hello, got %s", data)
	}
	_, err = r.ReadFile("missing.txt")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
```

- [ ] **Step 3: Run test**

Run: `go test ./internal/rules/ci/ -run TestMockFileReader -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/rules/ci/file_reader.go internal/rules/ci/file_reader_test.go
git commit -m "feat(ci): add FileReader interface for DI"
```

---

### Task 3: Strategy interface (strategy.go)

**Files:**
- Create: `internal/rules/ci/strategy.go`

- [ ] **Step 1: Write strategy.go**

```go
package ci

type Strategy interface {
	ProjectType() ProjectType
	RequiredCoverage() CoverageThresholds
	RequiredCIGates() []CIGate
	RequiredLinters() []LinterTool
	CheckProjectConfig(files []FileInfo, reader FileReader) []CheckResult
	CheckWorkflowSteps(jobs map[string]JobInfo) []CheckResult
	CheckSuppressions(files []FileInfo, reader FileReader) []CheckResult
}
```

Add supporting types to types.go:

```go
type FileInfo struct {
	Path    string
	AbsPath string
	IsDir   bool
	Content string // populated by reader
}

type JobInfo struct {
	Name  string
	Steps []StepInfo
}

type StepInfo struct {
	Name            string
	Run             string
	ContinueOnError string
	Uses            string
	Line            int
}
```

- [ ] **Step 2: Verify compiles**

Run: `go build ./internal/rules/ci/`
Expected: success

- [ ] **Step 3: Commit**

```bash
git add internal/rules/ci/strategy.go
git commit -m "feat(ci): add Strategy interface"
```

---

### Task 4: Strategy registry (strategy_registry.go)

**Files:**
- Create: `internal/rules/ci/strategy_registry.go`
- Test: `internal/rules/ci/strategy_registry_test.go`

- [ ] **Step 1: Write strategy_registry.go**

```go
package ci

type StrategyRegistry struct {
	strategies map[ProjectType]Strategy
}

func NewStrategyRegistry() *StrategyRegistry {
	return &StrategyRegistry{strategies: make(map[ProjectType]Strategy)}
}

func (r *StrategyRegistry) Register(s Strategy) {
	r.strategies[s.ProjectType()] = s
}

func (r *StrategyRegistry) StrategiesFor(types []ProjectType) []Strategy {
	var out []Strategy
	for _, t := range types {
		if s, ok := r.strategies[t]; ok {
			out = append(out, s)
		}
	}
	return out
}
```

- [ ] **Step 2: Write strategy_registry_test.go**

```go
package ci

import (
	"testing"
)

type mockStrategy struct {
	pt ProjectType
}

func (m mockStrategy) ProjectType() ProjectType { return m.pt }
func (m mockStrategy) RequiredCoverage() CoverageThresholds { return CoverageThresholds{} }
func (m mockStrategy) RequiredCIGates() []CIGate { return nil }
func (m mockStrategy) RequiredLinters() []LinterTool { return nil }
func (m mockStrategy) CheckProjectConfig(files []FileInfo, reader FileReader) []CheckResult { return nil }
func (m mockStrategy) CheckWorkflowSteps(jobs map[string]JobInfo) []CheckResult { return nil }
func (m mockStrategy) CheckSuppressions(files []FileInfo, reader FileReader) []CheckResult { return nil }

func TestStrategyRegistry(t *testing.T) {
	r := NewStrategyRegistry()
	s := mockStrategy{pt: SvelteKit}
	r.Register(s)

	got := r.StrategiesFor([]ProjectType{SvelteKit})
	if len(got) != 1 {
		t.Fatalf("expected 1 strategy, got %d", len(got))
	}

	got = r.StrategiesFor([]ProjectType{Python})
	if len(got) != 0 {
		t.Fatalf("expected 0 strategies for unregistered type, got %d", len(got))
	}
}
```

- [ ] **Step 3: Run test**

Run: `go test ./internal/rules/ci/ -run TestStrategyRegistry -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/rules/ci/strategy_registry.go internal/rules/ci/strategy_registry_test.go
git commit -m "feat(ci): add StrategyRegistry with register/lookup"
```

---

### Task 5: Project type detection (detector.go)

**Files:**
- Create: `internal/rules/ci/detector.go`
- Test: `internal/rules/ci/detector_test.go`

- [ ] **Step 1: Write detector.go**

```go
package ci

import (
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
		}
		// Check for svelte.config in project root
		if strings.Contains(path, "svelte.config") && strings.HasSuffix(base, ".js") || strings.HasSuffix(base, ".ts") {
			if d.svelteConfigExists(f) {
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

func (d *ProjectDetector) svelteConfigExists(f FileInfo) bool {
	return true // FileInfo already confirmed existence
}
```

- [ ] **Step 2: Write detector_test.go**

```go
package ci

import (
	"testing"
)

func TestDetectSvelteKit(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/package.json": `{"devDependencies": {"svelte": "^5.0.0"}}`,
		"/project/svelte.config.js": `import adapter from '@sveltejs/adapter-auto'`,
	}}
	detector := NewProjectDetector(reader)
	files := []FileInfo{
		{Path: "package.json", AbsPath: "/project/package.json"},
		{Path: "svelte.config.js", AbsPath: "/project/svelte.config.js"},
	}
	types := detector.Detect(files)
	if len(types) != 1 || types[0] != SvelteKit {
		t.Fatalf("expected SvelteKit, got %v", types)
	}
}

func TestDetectGo(t *testing.T) {
	detector := NewProjectDetector(nil)
	files := []FileInfo{
		{Path: "go.mod", AbsPath: "/project/go.mod"},
	}
	types := detector.Detect(files)
	if len(types) != 1 || types[0] != Go {
		t.Fatalf("expected Go, got %v", types)
	}
}

func TestDetectMultiple(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/package.json": `{"dependencies": {"svelte": "^5.0.0"}}`,
		"/project/pyproject.toml": `[project]`,
	}}
	detector := NewProjectDetector(reader)
	files := []FileInfo{
		{Path: "package.json", AbsPath: "/project/package.json"},
		{Path: "pyproject.toml", AbsPath: "/project/pyproject.toml"},
		{Path: "svelte.config.ts", AbsPath: "/project/svelte.config.ts"},
	}
	types := detector.Detect(files)
	if len(types) != 2 {
		t.Fatalf("expected 2 types, got %d: %v", len(types), types)
	}
	hasSK, hasPy := false, false
	for _, t := range types {
		if t == SvelteKit { hasSK = true }
		if t == Python { hasPy = true }
	}
	if !hasSK || !hasPy {
		t.Fatalf("expected SvelteKit and Python, got %v", types)
	}
}
```

- [ ] **Step 3: Run test**

Run: `go test ./internal/rules/ci/ -run TestDetect -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/rules/ci/detector.go internal/rules/ci/detector_test.go
git commit -m "feat(ci): add project type detection"
```

---

### Task 6: Shared CI checks (base.go)

**Files:**
- Create: `internal/rules/ci/base.go`
- Test: `internal/rules/ci/base_test.go`

Move existing shared checks from workflow_quality.go into base.go. These apply to ALL project types.

- [ ] **Step 1: Write base.go**

```go
package ci

import (
	"fmt"
	"regexp"
	"strings"
)

var maskingPattern = regexp.MustCompile(`\|\|\s*(true|echo\s+['"]?['"]?)\s*$`)

var qualityStepNamePatterns = []string{
	"lint", "typecheck", "type.check", "type check",
	"test", "pytest", "check", "quality",
	"coverage", "ruff", "pyright", "biome", "svelte-check",
}

func checkCommandMasking(jobs map[string]JobInfo) []CheckResult {
	var results []CheckResult
	for jobName, job := range jobs {
		for _, step := range job.Steps {
			if maskingPattern.MatchString(step.Run) {
				results = append(results, CheckResult{
					Path:    fmt.Sprintf(".github/workflows job=%q step=%q", jobName, step.Name),
					Message: fmt.Sprintf("Command masking on %q: %q", step.Name, strings.TrimSpace(step.Run)),
					Fix:     "Remove '|| true' or '|| echo \"\"' to let command failures propagate.",
				})
			}
		}
	}
	return results
}

func checkContinueOnError(jobs map[string]JobInfo) []CheckResult {
	var results []CheckResult
	for jobName, job := range jobs {
		for _, step := range job.Steps {
			if step.ContinueOnError != "true" && step.ContinueOnError != "yes" {
				continue
			}
			lower := strings.ToLower(step.Name)
			for _, p := range qualityStepNamePatterns {
				if strings.Contains(lower, p) {
					results = append(results, CheckResult{
						Path:    fmt.Sprintf(".github/workflows job=%q step=%q", jobName, step.Name),
						Message: fmt.Sprintf("continue-on-error on quality step %q", step.Name),
						Fix:     "Remove continue-on-error: true from quality check steps.",
					})
					break
				}
			}
		}
	}
	return results
}

func checkRequiredChecksAggregator(jobs map[string]JobInfo) []CheckResult {
	for name := range jobs {
		lower := strings.ToLower(name)
		if strings.Contains(lower, "required-checks") || strings.Contains(lower, "required.checks") {
			return nil
		}
	}
	return []CheckResult{{
		Message: "Workflow missing a required-checks aggregator job",
		Fix:     `Add a "required-checks" job that depends on all quality jobs and verifies results.`,
	}}
}
```

- [ ] **Step 2: Write base_test.go**

```go
package ci

import (
	"testing"
)

func TestCheckCommandMasking(t *testing.T) {
	jobs := map[string]JobInfo{
		"test": {
			Steps: []StepInfo{
				{Name: "Run tests", Run: "go test ./... || true"},
			},
		},
	}
	results := checkCommandMasking(jobs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestCheckCommandMaskingClean(t *testing.T) {
	jobs := map[string]JobInfo{
		"test": {
			Steps: []StepInfo{
				{Name: "Run tests", Run: "go test ./..."},
			},
		},
	}
	results := checkCommandMasking(jobs)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestCheckContinueOnError(t *testing.T) {
	jobs := map[string]JobInfo{
		"quality": {
			Steps: []StepInfo{
				{Name: "lint", ContinueOnError: "true", Run: "ruff check"},
			},
		},
	}
	results := checkContinueOnError(jobs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}

func TestCheckRequiredChecksAggregator(t *testing.T) {
	jobs := map[string]JobInfo{
		"test":        {},
		"required-checks": {},
	}
	results := checkRequiredChecksAggregator(jobs)
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestCheckRequiredChecksAggregatorMissing(t *testing.T) {
	jobs := map[string]JobInfo{
		"test": {},
		"lint": {},
	}
	results := checkRequiredChecksAggregator(jobs)
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
}
```

- [ ] **Step 3: Run tests**

Run: `go test ./internal/rules/ci/ -run "TestCheck" -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/rules/ci/base.go internal/rules/ci/base_test.go
git commit -m "feat(ci): add shared CI checks (masking, continue-on-error, aggregator)"
```

---

### Task 7: SvelteKit strategy (strategies/sveltekit.go)

**Files:**
- Create: `internal/rules/ci/strategies/sveltekit.go`
- Test: `internal/rules/ci/strategies/sveltekit_test.go`

- [ ] **Step 1: Write sveltekit.go**

```go
package strategies

import (
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

type SvelteKitStrategy struct {
	reader         ci.FileReader
	coverage       ci.CoverageThresholds
	requireVitestLinter bool
	requireSvelteuml    bool
}

func NewSvelteKitStrategy(reader ci.FileReader, cfg map[string]interface{}) *SvelteKitStrategy {
	s := &SvelteKitStrategy{reader: reader}
	s.coverage = ci.CoverageThresholds{
		Branches:   90,
		Lines:      80,
		Functions:  90,
		Statements: 90,
	}
	if cfg != nil {
		if v, ok := cfg["require-vitest-linter"].(bool); ok {
			s.requireVitestLinter = v
		}
		if v, ok := cfg["require-svelteuml"].(bool); ok {
			s.requireSvelteuml = v
		}
		if cv, ok := cfg["coverage"].(map[string]interface{}); ok {
			if b, ok := cv["branches"].(float64); ok { s.coverage.Branches = b }
			if l, ok := cv["lines"].(float64); ok { s.coverage.Lines = l }
			if f, ok := cv["functions"].(float64); ok { s.coverage.Functions = f }
			if st, ok := cv["statements"].(float64); ok { s.coverage.Statements = st }
		}
	}
	return s
}

func (s *SvelteKitStrategy) ProjectType() ci.ProjectType { return ci.SvelteKit }
func (s *SvelteKitStrategy) RequiredCoverage() ci.CoverageThresholds { return s.coverage }
func (s *SvelteKitStrategy) RequiredCIGates() []ci.CIGate {
	gates := []ci.CIGate{
		{Name: "svelte-check --fail-on-warnings", Required: true, Hint: "Add svelte-check with --fail-on-warnings"},
		{Name: "biome check", Required: true, Hint: "Add biome check to CI"},
		{Name: "vitest coverage", Required: true, Hint: "Add vitest run --coverage"},
		{Name: "build", Required: true, Hint: "Add pnpm build"},
	}
	if s.requireVitestLinter {
		gates = append(gates, ci.CIGate{
			Name:     "vitest-linter",
			Required: true,
			Hint:     "Add vitest-linter CI gate",
		})
	}
	if s.requireSvelteuml {
		gates = append(gates, ci.CIGate{
			Name:     "svelteuml",
			Required: true,
			Hint:     "Add svelteuml diagram generation to CI",
		})
	}
	return gates
}
func (s *SvelteKitStrategy) RequiredLinters() []ci.LinterTool {
	return []ci.LinterTool{
		{Name: "biome", Required: true, Hint: "Configure biome.json"},
	}
}
func (s *SvelteKitStrategy) CheckProjectConfig(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
func (s *SvelteKitStrategy) CheckWorkflowSteps(jobs map[string]ci.JobInfo) []ci.CheckResult {
	var results []ci.CheckResult
	for _, gates := range s.RequiredCIGates() {
		found := false
		for _, job := range jobs {
			for _, step := range job.Steps {
				runLower := strings.ToLower(step.Run)
				nameLower := strings.ToLower(step.Name)
				combined := runLower + " " + nameLower
				switch {
				case strings.Contains(gates.Name, "svelte-check"):
					if strings.Contains(combined, "svelte-check") {
						found = true
						if !strings.Contains(combined, "--fail-on-warnings") {
							results = append(results, ci.CheckResult{
								Message: "svelte-check without --fail-on-warnings",
								Fix:     `Add --fail-on-warnings to svelte-check command.`,
							})
						}
					}
				case strings.Contains(gates.Name, "vitest-linter"):
					if strings.Contains(combined, "vitest-linter") {
						found = true
					}
				case strings.Contains(gates.Name, "svelteuml"):
					if strings.Contains(combined, "svelteuml") {
						found = true
					}
				case strings.Contains(gates.Name, "biome"):
					if strings.Contains(combined, "biome") {
						found = true
					}
				case strings.Contains(gates.Name, "vitest"):
					if strings.Contains(combined, "vitest") && strings.Contains(combined, "coverage") {
						found = true
					}
				case strings.Contains(gates.Name, "build"):
					if (strings.Contains(runLower, "pnpm") || strings.Contains(runLower, "npm")) && strings.Contains(runLower, "build") {
						found = true
					}
				}
			}
		}
		if !found && gates.Required {
			results = append(results, ci.CheckResult{
				Message: "Missing required CI gate: " + gates.Name,
				Fix:     gates.Hint,
			})
		}
	}
	return results
}

func (s *SvelteKitStrategy) CheckSuppressions(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult {
	return nil
}
```

- [ ] **Step 2: Write sveltekit_test.go**

```go
package strategies

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

func TestSvelteKitRequiredGates(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewSvelteKitStrategy(reader, nil)
	gates := strat.RequiredCIGates()
	if len(gates) < 4 {
		t.Fatalf("expected at least 4 gates, got %d", len(gates))
	}
}

func TestSvelteKitChecksSvelteCheck(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewSvelteKitStrategy(reader, nil)
	jobs := map[string]ci.JobInfo{
		"quality": {
			Steps: []ci.StepInfo{
				{Name: "svelte-check", Run: "pnpm exec svelte-check --tsconfig tsconfig.json"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	found := false
	for _, r := range results {
		if strings.Contains(r.Message, "--fail-on-warnings") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected violation for missing --fail-on-warnings")
	}
}

func TestSvelteKitMissingRequiredGate(t *testing.T) {
	reader := ci.MockFileReader{}
	cfg := map[string]interface{}{"require-vitest-linter": true}
	strat := NewSvelteKitStrategy(reader, cfg)
	jobs := map[string]ci.JobInfo{
		"test": {
			Steps: []ci.StepInfo{
				{Name: "run tests", Run: "pnpm vitest run"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	// Should find vitest-linter missing
	found := false
	for _, r := range results {
		if strings.Contains(r.Message, "vitest-linter") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected violation for missing vitest-linter gate")
	}
}
```

Add import for `"strings"` in the test file.

- [ ] **Step 3: Verify compiles**

Run: `go build ./internal/rules/ci/...`
Expected: success

- [ ] **Step 4: Run test**

Run: `go test ./internal/rules/ci/strategies/ -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add internal/rules/ci/strategies/ 
git commit -m "feat(ci): add SvelteKit strategy"
```

---

### Task 8: Python strategy (strategies/python.go)

**Files:**
- Create: `internal/rules/ci/strategies/python.go`
- Test: `internal/rules/ci/strategies/python_test.go`

- [ ] **Step 1: Write python.go**

```go
package strategies

import (
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

type PythonStrategy struct {
	reader            ci.FileReader
	coverage          ci.CoverageThresholds
	requirePytestLinter bool
}

func NewPythonStrategy(reader ci.FileReader, cfg map[string]interface{}) *PythonStrategy {
	s := &PythonStrategy{reader: reader}
	s.coverage = ci.CoverageThresholds{
		Branches:  90,
		Lines:     80,
		Functions: 90,
		Statements: 80,
	}
	if cfg != nil {
		if v, ok := cfg["require-pytest-linter"].(bool); ok {
			s.requirePytestLinter = v
		}
		if cv, ok := cfg["coverage"].(map[string]interface{}); ok {
			if b, ok := cv["branches"].(float64); ok { s.coverage.Branches = b }
			if l, ok := cv["lines"].(float64); ok { s.coverage.Lines = l }
			if f, ok := cv["functions"].(float64); ok { s.coverage.Functions = f }
			if st, ok := cv["statements"].(float64); ok { s.coverage.Statements = st }
		}
	}
	return s
}

func (s *PythonStrategy) ProjectType() ci.ProjectType { return ci.Python }
func (s *PythonStrategy) RequiredCoverage() ci.CoverageThresholds { return s.coverage }
func (s *PythonStrategy) RequiredCIGates() []ci.CIGate {
	gates := []ci.CIGate{
		{Name: "ruff check", Required: true, Hint: "Add ruff check to CI"},
		{Name: "pyright", Required: true, Hint: "Add pyright type-checking to CI"},
		{Name: "pytest --cov-branch --cov-fail-under", Required: true, Hint: "Ensure pytest uses --cov-branch and --cov-fail-under=90"},
	}
	if s.requirePytestLinter {
		gates = append(gates, ci.CIGate{
			Name:     "pytest-linter",
			Required: true,
			Hint:     "Add pytest-linter CI gate",
		})
	}
	return gates
}
func (s *PythonStrategy) RequiredLinters() []ci.LinterTool {
	return []ci.LinterTool{
		{Name: "ruff", Required: true, Hint: "Configure ruff in pyproject.toml"},
		{Name: "pyright", Required: true, Hint: "Configure pyright in pyproject.toml or pyrightconfig.json"},
	}
}
func (s *PythonStrategy) CheckProjectConfig(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
func (s *PythonStrategy) CheckWorkflowSteps(jobs map[string]ci.JobInfo) []ci.CheckResult {
	var results []ci.CheckResult
	for _, job := range jobs {
		for _, step := range job.Steps {
			runLower := strings.ToLower(step.Run)
			if !strings.Contains(runLower, "pytest") {
				continue
			}
			if !strings.Contains(runLower, "--cov-branch") {
				results = append(results, ci.CheckResult{
					Message: "pytest command missing --cov-branch",
					Fix:     "Add --cov-branch to pytest command for branch coverage.",
				})
			}
			if !strings.Contains(runLower, "--cov-fail-under") {
				results = append(results, ci.CheckResult{
					Message: "pytest command missing --cov-fail-under",
					Fix:     "Add --cov-fail-under=90 to pytest command.",
				})
			}
		}
	}

	// Check missing ruff/pyright/pytest-linter gates
	foundRuff := false
	foundPyright := false
	foundPytestLinter := false
	for _, job := range jobs {
		for _, step := range job.Steps {
			combined := strings.ToLower(step.Run + " " + step.Name)
			if strings.Contains(combined, "ruff") {
				foundRuff = true
			}
			if strings.Contains(combined, "pyright") {
				foundPyright = true
			}
			if s.requirePytestLinter && strings.Contains(combined, "pytest-linter") {
				foundPytestLinter = true
			}
		}
	}
	if !foundRuff {
		results = append(results, ci.CheckResult{
			Message: "Missing ruff check in CI",
			Fix:     "Add ruff check to CI workflow.",
		})
	}
	if !foundPyright {
		results = append(results, ci.CheckResult{
			Message: "Missing pyright in CI",
			Fix:     "Add pyright type-checking to CI workflow.",
		})
	}
	if s.requirePytestLinter && !foundPytestLinter {
		results = append(results, ci.CheckResult{
			Message: "Missing pytest-linter CI gate",
			Fix:     "Add pytest-linter CI gate.",
		})
	}

	return results
}
func (s *PythonStrategy) CheckSuppressions(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult {
	var results []ci.CheckResult
	for _, f := range files {
		if !strings.HasSuffix(f.Path, ".py") {
			continue
		}
		data, err := reader.ReadFile(f.AbsPath)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		count := 0
		for _, line := range lines {
			trimmed := strings.TrimSpace(line)
			if strings.Contains(trimmed, "# noqa") || strings.Contains(trimmed, "# type: ignore") {
				count++
			}
		}
		if count > 0 {
			results = append(results, ci.CheckResult{
				Path:    f.Path,
				Message: "Python suppression comments exceed threshold",
				Fix:     "Reduce # noqa / # type: ignore comments. Address root causes.",
			})
		}
	}
	return results
}
```

- [ ] **Step 2: Write python_test.go**

```go
package strategies

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

func TestPythonCheckPytestCoverage(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewPythonStrategy(reader, nil)
	jobs := map[string]ci.JobInfo{
		"test": {
			Steps: []ci.StepInfo{
				{Name: "run tests", Run: "pytest"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	covBranch := false
	covFailUnder := false
	for _, r := range results {
		if strings.Contains(r.Message, "--cov-branch") {
			covBranch = true
		}
		if strings.Contains(r.Message, "--cov-fail-under") {
			covFailUnder = true
		}
	}
	if !covBranch || !covFailUnder {
		t.Fatal("expected violations for missing --cov-branch and --cov-fail-under")
	}
}

func TestPythonCheckPytestCoveragePass(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewPythonStrategy(reader, nil)
	jobs := map[string]ci.JobInfo{
		"test": {
			Steps: []ci.StepInfo{
				{Name: "test", Run: "pytest --cov --cov-branch --cov-fail-under=90"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	for _, r := range results {
		if strings.Contains(r.Message, "pytest") {
			t.Fatalf("unexpected violation: %s", r.Message)
		}
	}
}

func TestPythonMissingRuff(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewPythonStrategy(reader, nil)
	jobs := map[string]ci.JobInfo{
		"test": {
			Steps: []ci.StepInfo{
				{Name: "run tests", Run: "pytest"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	found := false
	for _, r := range results {
		if strings.Contains(r.Message, "ruff") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected violation for missing ruff")
	}
}
```

Add import for `"strings"` in the test file.

- [ ] **Step 3: Run test**

Run: `go test ./internal/rules/ci/strategies/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/rules/ci/strategies/python.go internal/rules/ci/strategies/python_test.go
git commit -m "feat(ci): add Python strategy"
```

---

### Task 9: Go strategy (strategies/golang.go)

**Files:**
- Create: `internal/rules/ci/strategies/golang.go`
- Test: `internal/rules/ci/strategies/golang_test.go`

- [ ] **Step 1: Write golang.go**

```go
package strategies

import (
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

type GoStrategy struct {
	reader   ci.FileReader
	coverage ci.CoverageThresholds
}

func NewGoStrategy(reader ci.FileReader, cfg map[string]interface{}) *GoStrategy {
	s := &GoStrategy{reader: reader}
	s.coverage = ci.CoverageThresholds{Lines: 90}
	if cfg != nil {
		if cv, ok := cfg["coverage"].(map[string]interface{}); ok {
			if l, ok := cv["lines"].(float64); ok { s.coverage.Lines = l }
		}
	}
	return s
}

func (s *GoStrategy) ProjectType() ci.ProjectType { return ci.Go }
func (s *GoStrategy) RequiredCoverage() ci.CoverageThresholds { return s.coverage }
func (s *GoStrategy) RequiredCIGates() []ci.CIGate {
	return []ci.CIGate{
		{Name: "go test -race", Required: true, Hint: "Add go test -race -covermode=atomic"},
		{Name: "golangci-lint", Required: true, Hint: "Add golangci-lint to CI"},
		{Name: "go vet", Required: true, Hint: "Add go vet to CI"},
	}
}
func (s *GoStrategy) RequiredLinters() []ci.LinterTool {
	return []ci.LinterTool{
		{Name: "golangci-lint", Required: true, Hint: "Configure .golangci.yml"},
	}
}
func (s *GoStrategy) CheckProjectConfig(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
func (s *GoStrategy) CheckWorkflowSteps(jobs map[string]ci.JobInfo) []ci.CheckResult {
	var results []ci.CheckResult
	foundTest := false
	foundLint := false
	foundVet := false

	for _, job := range jobs {
		for _, step := range job.Steps {
			combined := strings.ToLower(step.Run + " " + step.Name)
			if strings.Contains(combined, "go test") || strings.Contains(combined, "gotest") {
				foundTest = true
				if strings.Contains(step.Run, "-race") {
					// has race detection
				}
			}
			if strings.Contains(combined, "golangci") && strings.Contains(combined, "lint") {
				foundLint = true
			}
			if strings.Contains(combined, "go vet") {
				foundVet = true
			}
		}
	}
	if !foundTest {
		results = append(results, ci.CheckResult{
			Message: "Missing go test in CI",
			Fix:     "Add go test -race -covermode=atomic to CI.",
		})
	}
	if !foundLint {
		results = append(results, ci.CheckResult{
			Message: "Missing golangci-lint in CI",
			Fix:     "Add golangci-lint run to CI.",
		})
	}
	if !foundVet {
		results = append(results, ci.CheckResult{
			Message: "Missing go vet in CI",
			Fix:     "Add go vet to CI.",
		})
	}
	return results
}
func (s *GoStrategy) CheckSuppressions(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
```

- [ ] **Step 2: Write golang_test.go**

```go
package strategies

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

func TestGoMissingGates(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewGoStrategy(reader, nil)
	jobs := map[string]ci.JobInfo{
		"build": {
			Steps: []ci.StepInfo{
				{Name: "build", Run: "go build ./..."},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	foundTest := false
	foundLint := false
	foundVet := false
	for _, r := range results {
		if strings.Contains(r.Message, "go test") { foundTest = true }
		if strings.Contains(r.Message, "golangci-lint") { foundLint = true }
		if strings.Contains(r.Message, "go vet") { foundVet = true }
	}
	if !foundTest || !foundLint || !foundVet {
		t.Fatal("expected violations for all 3 missing Go gates")
	}
}

func TestGoAllGatesPresent(t *testing.T) {
	reader := ci.MockFileReader{}
	strat := NewGoStrategy(reader, nil)
	jobs := map[string]ci.JobInfo{
		"quality": {
			Steps: []ci.StepInfo{
				{Name: "lint", Run: "golangci-lint run ./..."},
				{Name: "vet", Run: "go vet ./..."},
				{Name: "test", Run: "go test -race -covermode=atomic ./..."},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	if len(results) > 0 {
		t.Fatalf("expected 0 violations, got %d: %v", len(results), results)
	}
}
```

Add import for `"strings"` in the test file.

- [ ] **Step 3: Run test**

Run: `go test ./internal/rules/ci/strategies/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/rules/ci/strategies/golang.go internal/rules/ci/strategies/golang_test.go
git commit -m "feat(ci): add Go strategy"
```

---

### Task 10: Rust strategy (strategies/rust.go)

**Files:**
- Create: `internal/rules/ci/strategies/rust.go`
- Test: `internal/rules/ci/strategies/rust_test.go`

- [ ] **Step 1: Write rust.go**

```go
package strategies

import (
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

type RustStrategy struct {
	reader              ci.FileReader
	coverage            ci.CoverageThresholds
	requireCargoTestLint bool
}

func NewRustStrategy(reader ci.FileReader, cfg map[string]interface{}) *RustStrategy {
	s := &RustStrategy{reader: reader}
	s.coverage = ci.CoverageThresholds{Lines: 90}
	if cfg != nil {
		if v, ok := cfg["require-cargo-test-lint"].(bool); ok {
			s.requireCargoTestLint = v
		}
		if cv, ok := cfg["coverage"].(map[string]interface{}); ok {
			if l, ok := cv["lines"].(float64); ok { s.coverage.Lines = l }
		}
	}
	return s
}

func (s *RustStrategy) ProjectType() ci.ProjectType { return ci.Rust }
func (s *RustStrategy) RequiredCoverage() ci.CoverageThresholds { return s.coverage }
func (s *RustStrategy) RequiredCIGates() []ci.CIGate {
	gates := []ci.CIGate{
		{Name: "cargo clippy", Required: true, Hint: "Add cargo clippy to CI"},
		{Name: "cargo fmt --check", Required: true, Hint: "Add cargo fmt --check to CI"},
		{Name: "cargo test", Required: true, Hint: "Add cargo test to CI"},
		{Name: "coverage", Required: true, Hint: "Add cargo llvm-cov or tarpaulin for coverage"},
	}
	if s.requireCargoTestLint {
		gates = append(gates, ci.CIGate{
			Name:     "cargo test-lint",
			Required: true,
			Hint:     "Add cargo-test-lint to CI",
		})
	}
	return gates
}
func (s *RustStrategy) RequiredLinters() []ci.LinterTool {
	return []ci.LinterTool{
		{Name: "clippy", Required: true, Hint: "Configure clippy in Cargo.toml"},
		{Name: "rustfmt", Required: true, Hint: "Configure rustfmt"},
	}
}
func (s *RustStrategy) CheckProjectConfig(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
func (s *RustStrategy) CheckWorkflowSteps(jobs map[string]ci.JobInfo) []ci.CheckResult {
	var results []ci.CheckResult
	foundClippy := false
	foundFmt := false
	foundTest := false
	foundCoverage := false
	foundTestLint := false

	for _, job := range jobs {
		for _, step := range job.Steps {
			combined := strings.ToLower(step.Run + " " + step.Name)
			if strings.Contains(combined, "clippy") {
				foundClippy = true
			}
			if strings.Contains(combined, "cargo fmt") || strings.Contains(combined, "rustfmt") {
				foundFmt = true
			}
			if strings.Contains(combined, "cargo test") && !strings.Contains(combined, "test-lint") {
				foundTest = true
			}
			if strings.Contains(combined, "llvm-cov") || strings.Contains(combined, "tarpaulin") || strings.Contains(combined, "--fail-under") {
				foundCoverage = true
			}
			if s.requireCargoTestLint && strings.Contains(combined, "test-lint") {
				foundTestLint = true
			}
		}
	}
	if !foundClippy {
		results = append(results, ci.CheckResult{
			Message: "Missing cargo clippy in CI",
			Fix:     "Add cargo clippy to CI workflow.",
		})
	}
	if !foundFmt {
		results = append(results, ci.CheckResult{Message: "Missing cargo fmt --check in CI", Fix: "Add cargo fmt --check to CI."})
	}
	if !foundTest {
		results = append(results, ci.CheckResult{Message: "Missing cargo test in CI", Fix: "Add cargo test to CI."})
	}
	if !foundCoverage {
		results = append(results, ci.CheckResult{Message: "Missing coverage gate in CI", Fix: "Add cargo llvm-cov or tarpaulin."})
	}
	if s.requireCargoTestLint && !foundTestLint {
		results = append(results, ci.CheckResult{Message: "Missing cargo test-lint in CI", Fix: "Add cargo-test-lint to CI."})
	}
	return results
}
func (s *RustStrategy) CheckSuppressions(files []ci.FileInfo, reader ci.FileReader) []ci.CheckResult { return nil }
```

- [ ] **Step 2: Write rust_test.go**

```go
package strategies

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci"
)

func TestRustMissingGates(t *testing.T) {
	reader := ci.MockFileReader{}
	cfg := map[string]interface{}{"require-cargo-test-lint": true}
	strat := NewRustStrategy(reader, cfg)
	jobs := map[string]ci.JobInfo{
		"build": {
			Steps: []ci.StepInfo{
				{Name: "build", Run: "cargo build"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	expected := []string{"clippy", "fmt", "cargo test", "coverage", "test-lint"}
	for _, e := range expected {
		found := false
		for _, r := range results {
			if strings.Contains(r.Message, e) {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("expected violation containing %q", e)
		}
	}
}

func TestRustAllGatesPresent(t *testing.T) {
	reader := ci.MockFileReader{}
	cfg := map[string]interface{}{"require-cargo-test-lint": true}
	strat := NewRustStrategy(reader, cfg)
	jobs := map[string]ci.JobInfo{
		"quality": {
			Steps: []ci.StepInfo{
				{Name: "clippy", Run: "cargo clippy -- -W clippy::all"},
				{Name: "fmt", Run: "cargo fmt --check"},
				{Name: "test", Run: "cargo test"},
				{Name: "coverage", Run: "cargo llvm-cov --fail-under-lines 90"},
				{Name: "test-lint", Run: "cargo test-lint"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	if len(results) > 0 {
		t.Fatalf("expected 0 violations, got %d: %v", len(results), results)
	}
}
```

Add import for `"strings"` in the test file.

- [ ] **Step 3: Run test**

Run: `go test ./internal/rules/ci/strategies/ -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add internal/rules/ci/strategies/rust.go internal/rules/ci/strategies/rust_test.go
git commit -m "feat(ci): add Rust strategy"
```

---

### Task 11: Refactor init.go and rule.go

**Files:**
- Modify: `internal/rules/ci/init.go`
- Modify: `internal/rules/ci/workflow_quality.go`
- Modify: `internal/linter/factory.go`

- [ ] **Step 1: Rewrite init.go**

```go
package ci

import (
	"fmt"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/strategies"
)

func init() {
	rules.Register("github-workflows", func(ctx *rules.RuleContext) (rules.Rule, error) {
		reader := OSFileReader{}
		registry := newDefaultRegistry(reader, ctx.Config)
		detector := NewProjectDetector(reader)
		return NewWorkflowQualityRule(registry, reader, detector, ctx.Config), nil
	})
}

func newDefaultRegistry(reader FileReader, cfg map[string]interface{}) *StrategyRegistry {
	registry := NewStrategyRegistry()

	svelteCfg := extractConfig("sveltekit", cfg)
	pythonCfg := extractConfig("python", cfg)
	goCfg := extractConfig("golang", cfg)
	rustCfg := extractConfig("rust", cfg)

	registry.Register(strategies.NewSvelteKitStrategy(reader, svelteCfg))
	registry.Register(strategies.NewPythonStrategy(reader, pythonCfg))
	registry.Register(strategies.NewGoStrategy(reader, goCfg))
	registry.Register(strategies.NewRustStrategy(reader, rustCfg))

	return registry
}

func extractConfig(pt string, fullCfg map[string]interface{}) map[string]interface{} {
	if fullCfg == nil {
		return nil
	}
	if v, ok := fullCfg[pt].(map[string]interface{}); ok {
		return v
	}
	return nil
}
```

- [ ] **Step 2: Rewrite workflow_quality.go**

Replace existing flat rule with strategy-dispatching version:

```go
package ci

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"gopkg.in/yaml.v3"
)

type WorkflowQualityRule struct {
	registry *StrategyRegistry
	reader   FileReader
	detector *ProjectDetector
	config   map[string]interface{}
}

func NewWorkflowQualityRule(registry *StrategyRegistry, reader FileReader, detector *ProjectDetector, config map[string]interface{}) *WorkflowQualityRule {
	return &WorkflowQualityRule{
		registry: registry,
		reader:   reader,
		detector: detector,
		config:   config,
	}
}

func (r *WorkflowQualityRule) Name() string {
	return "github-workflows"
}

func (r *WorkflowQualityRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	internalFiles := toInternalFiles(files)

	projectTypes := r.detector.Detect(internalFiles)
	if len(projectTypes) == 0 {
		return nil
	}

	strategies := r.registry.StrategiesFor(projectTypes)

	workflowFiles := filterWorkflowFiles(files)
	workflowJobs := r.parseAllWorkflows(workflowFiles)

	var results []CheckResult

	// Global checks
	globalCfg := extractConfig("global", r.config)
	if v, ok := globalCfg["disallow-command-masking"].(bool); ok && v {
		results = append(results, checkCommandMasking(workflowJobs)...)
	}
	if v, ok := globalCfg["disallow-continue-on-error-on-quality"].(bool); ok && v {
		results = append(results, checkContinueOnError(workflowJobs)...)
	}
	if v, ok := globalCfg["require-required-checks-aggregator"].(bool); ok && v {
		results = append(results, checkRequiredChecksAggregator(workflowJobs)...)
	}

	// Strategy-specific checks
	allFiles := toInternalFiles(files)
	for _, s := range strategies {
		results = append(results, s.CheckWorkflowSteps(workflowJobs)...)
		results = append(results, s.CheckProjectConfig(allFiles, r.reader)...)
		results = append(results, s.CheckSuppressions(allFiles, r.reader)...)
	}

	var violations []rules.Violation
	for _, cr := range results {
		violations = append(violations, cr.ToViolation())
	}
	return violations
}

func (r *WorkflowQualityRule) parseAllWorkflows(files []walker.FileInfo) map[string]JobInfo {
	jobs := make(map[string]JobInfo)
	for _, f := range files {
		for name, job := range parseWorkflowJobs(f, r.reader) {
			jobs[name] = job
		}
	}
	return jobs
}

func parseWorkflowJobs(f walker.FileInfo, reader FileReader) map[string]JobInfo {
	data, err := reader.ReadFile(f.AbsPath)
	if err != nil {
		return nil
	}
	var workflow yaml.Node
	if err := yaml.Unmarshal(data, &workflow); err != nil {
		return nil
	}
	rawJobs := findJobs(&workflow)
	jobs := make(map[string]JobInfo, len(rawJobs))
	for name, jobNode := range rawJobs {
		ji := JobInfo{Name: name}
		if jobNode.Kind != yaml.MappingNode {
			continue
		}
		for i := 0; i < len(jobNode.Content)-1; i += 2 {
			if jobNode.Content[i].Value == "steps" && jobNode.Content[i+1].Kind == yaml.SequenceNode {
				for _, stepNode := range jobNode.Content[i+1].Content {
					if stepNode.Kind != yaml.MappingNode {
						continue
					}
					si := StepInfo{Line: stepNode.Line}
					for j := 0; j < len(stepNode.Content)-1; j += 2 {
						key := stepNode.Content[j].Value
						val := stepNode.Content[j+1].Value
						switch key {
						case "name":
							si.Name = val
						case "run":
							si.Run = val
						case "continue-on-error":
							si.ContinueOnError = val
						case "uses":
							si.Uses = val
						}
					}
					ji.Steps = append(ji.Steps, si)
				}
			}
		}
		jobs[name] = ji
	}
	return jobs
}

func toInternalFiles(files []walker.FileInfo) []FileInfo {
	var out []FileInfo
	for _, f := range files {
		out = append(out, FileInfo{
			Path:    f.Path,
			AbsPath: f.AbsPath,
			IsDir:   f.IsDir,
		})
	}
	return out
}

func filterWorkflowFiles(files []walker.FileInfo) []walker.FileInfo {
	var out []walker.FileInfo
	for _, f := range files {
		if f.IsDir {
			continue
		}
		path := filepath.ToSlash(f.Path)
		if strings.Contains(path, ".github/workflows/") && (strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml")) {
			out = append(out, f)
		}
	}
	return out
}
```

- [ ] **Step 3: Update factory.go blank import** (verify already present)

- [ ] **Step 4: Verify compiles**

Run: `go build ./...`
Expected: success

- [ ] **Step 5: Run all tests**

Run: `go test ./internal/rules/ci/... -v`
Expected: PASS

- [ ] **Step 6: Clean up old unused code**

Remove from `workflow_quality.go`: `checkCommandMasking`, `checkContinueOnErrorOnQuality`, `checkRequiredChecksAggregator`, `checkPytestCoverage`, `checkSvelteCheckWarnings`, `checkPackageJSON`, `walkSteps`, `findJobs`, `findStepKey`, `maskingPattern`, `qualityStepNamePatterns`, `pytestRe`, `svelteCheckRe`, `violation`, `toRuleViolations`

- [ ] **Step 7: Commit**

```bash
git add internal/rules/ci/ internal/linter/factory.go
git commit -m "feat(ci): refactor to strategy pattern with registry and DI"
```

---

### Task 12: Write integration test

**Files:**
- Create: `internal/rules/ci/integration_test.go`

- [ ] **Step 1: Write integration_test.go**

```go
package ci

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/strategies"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func TestIntegrationSvelteKitProject(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/package.json":        `{"devDependencies": {"svelte": "^5.0.0"}}`,
		"/project/svelte.config.ts":     `import adapter from '@sveltejs/adapter-auto'`,
		"/project/.github/workflows/ci.yml": `
name: CI
on: [push]
jobs:
  required-checks:
    runs-on: ubuntu-latest
    steps:
      - run: echo "ok"
  quality:
    runs-on: ubuntu-latest
    steps:
      - name: Lint
        run: pnpm dlx @biomejs/biome check src/
      - name: Type check
        run: pnpm exec svelte-check --tsconfig tsconfig.json
      - name: Test
        run: pnpm vitest run --coverage
      - name: Build
        run: pnpm build
`,
	}}

	registry := NewStrategyRegistry()
	registry.Register(strategies.NewSvelteKitStrategy(reader, nil))

	detector := NewProjectDetector(reader)
	rule := NewWorkflowQualityRule(registry, reader, detector, map[string]interface{}{
		"global": map[string]interface{}{
			"disallow-command-masking":                true,
			"disallow-continue-on-error-on-quality":   true,
			"require-required-checks-aggregator":      true,
		},
	})

	files := []walker.FileInfo{
		{Path: "package.json", AbsPath: "/project/package.json"},
		{Path: "svelte.config.ts", AbsPath: "/project/svelte.config.ts"},
		{Path: ".github/workflows/ci.yml", AbsPath: "/project/.github/workflows/ci.yml"},
	}

	violations := rule.Check(files, nil)
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations, got %d:", len(violations))
		for _, v := range violations {
			t.Logf("  %s: %s", v.Path, v.Message)
		}
	}
}

func TestIntegrationPythonProjectMissingGates(t *testing.T) {
	reader := MockFileReader{Files: map[string]string{
		"/project/pyproject.toml":  `[project]`,
		"/project/.github/workflows/ci.yml": `
name: CI
on: [push]
jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        run: git checkout
`,
	}}

	registry := NewStrategyRegistry()
	registry.Register(strategies.NewPythonStrategy(reader, nil))

	detector := NewProjectDetector(reader)
	rule := NewWorkflowQualityRule(registry, reader, detector, map[string]interface{}{
		"global": map[string]interface{}{
			"require-required-checks-aggregator": true,
		},
	})

	files := []walker.FileInfo{
		{Path: "pyproject.toml", AbsPath: "/project/pyproject.toml"},
		{Path: ".github/workflows/ci.yml", AbsPath: "/project/.github/workflows/ci.yml"},
	}

	violations := rule.Check(files, nil)
	if len(violations) < 1 {
		t.Fatal("expected violations for missing quality gates")
	}
}
```

- [ ] **Step 2: Run integration test**

Run: `go test ./internal/rules/ci/ -run TestIntegration -v`
Expected: PASS

- [ ] **Step 3: Commit**

```bash
git add internal/rules/ci/integration_test.go
git commit -m "test(ci): add integration tests for strategy pattern"
```

---

### Task 13: Dogfood — add github-workflows rule to structurelint's own .structurelint.yml

- [ ] **Step 1: Add rule to .structurelint.yml**

```yaml
rules:
  github-workflows:
    global:
      disallow-command-masking: true
      disallow-continue-on-error-on-quality: true
      require-required-checks-aggregator: true
    golang:
      coverage:
        lines: 90
```

- [ ] **Step 2: Run structurelint on itself**

Run: `go run .`
Expected: changes found or no violations

- [ ] **Step 3: Commit**

```bash
git add .structurelint.yml
git commit -m "chore: enable github-workflows rule on self"
```
