package strategies

import (
	"errors"
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

type mockPythonReader struct {
	files map[string]string
}

func (m mockPythonReader) ReadFile(path string) ([]byte, error) {
	if c, ok := m.files[path]; ok {
		return []byte(c), nil
	}
	return nil, errors.New("not found")
}

func TestPythonCheckPytestCoverage(t *testing.T) {
	strat := NewPythonStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
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
	strat := NewPythonStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
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
	strat := NewPythonStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
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

func TestPythonStrategyMethods(t *testing.T) {
	cfg := map[string]interface{}{
		"require-pytest-linter": true,
		"coverage": map[string]interface{}{
			"branches":   85.0,
			"lines":      75.0,
			"functions":  85.0,
			"statements": 75.0,
		},
	}
	strat := NewPythonStrategy(nil, cfg)

	if strat.ProjectType() != core.Python {
		t.Errorf("expected core.Python, got %v", strat.ProjectType())
	}

	cov := strat.RequiredCoverage()
	if cov.Branches != 85.0 || cov.Lines != 75.0 || cov.Functions != 85.0 || cov.Statements != 75.0 {
		t.Errorf("unexpected coverage thresholds: %v", cov)
	}

	gates := strat.RequiredCIGates()
	// Should have 4 gates since requirePytestLinter is true
	if len(gates) != 4 {
		t.Errorf("expected 4 CI gates, got %d", len(gates))
	}

	linters := strat.RequiredLinters()
	if len(linters) != 2 {
		t.Errorf("expected 2 linters, got %d", len(linters))
	}

	if res := strat.CheckProjectConfig(nil, nil); res != nil {
		t.Errorf("expected nil check project config, got %v", res)
	}

	// Test missing pytest-linter step check
	jobs := map[string]core.JobInfo{
		"quality": {
			Steps: []core.StepInfo{
				{Name: "ruff", Run: "ruff check"},
				{Name: "pyright", Run: "pyright"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	foundLinter := false
	for _, r := range results {
		if strings.Contains(r.Message, "pytest-linter") {
			foundLinter = true
		}
	}
	if !foundLinter {
		t.Error("expected violation for missing pytest-linter")
	}

	// Test suppressions check
	reader := mockPythonReader{files: map[string]string{
		"/app/main.py":   "print('hello')\n# noqa\n# type: ignore",
		"/app/helper.go": "# noqa",
		"/app/err.py":    "error",
	}}

	files := []core.FileInfo{
		{Path: "main.py", AbsPath: "/app/main.py"},
		{Path: "helper.go", AbsPath: "/app/helper.go"},
		{Path: "err.py", AbsPath: "/app/err.py"},
	}

	suppressions := strat.CheckSuppressions(files, reader)
	if len(suppressions) != 1 || suppressions[0].Path != "main.py" {
		t.Errorf("expected 1 suppression on main.py, got %v", suppressions)
	}
}
