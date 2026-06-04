package strategies

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

func TestGoMissingGates(t *testing.T) {
	strat := NewGoStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"build": {
			Steps: []core.StepInfo{
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
	strat := NewGoStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"quality": {
			Steps: []core.StepInfo{
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

func TestGoStrategyMethods(t *testing.T) {
	// Test config initialization
	cfg := map[string]interface{}{
		"coverage": map[string]interface{}{
			"lines": 85.0,
		},
	}
	strat := NewGoStrategy(nil, cfg)
	if strat.ProjectType() != core.Go {
		t.Errorf("expected core.Go, got %v", strat.ProjectType())
	}

	cov := strat.RequiredCoverage()
	if cov.Lines != 85.0 {
		t.Errorf("expected lines coverage threshold 85.0, got %f", cov.Lines)
	}

	gates := strat.RequiredCIGates()
	if len(gates) != 3 {
		t.Errorf("expected 3 CI gates, got %d", len(gates))
	}

	linters := strat.RequiredLinters()
	if len(linters) != 1 {
		t.Errorf("expected 1 linter, got %d", len(linters))
	}

	if res := strat.CheckProjectConfig(nil, nil); res != nil {
		t.Errorf("expected nil check project config, got %v", res)
	}

	if res := strat.CheckSuppressions(nil, nil); res != nil {
		t.Errorf("expected nil check suppressions, got %v", res)
	}
}
