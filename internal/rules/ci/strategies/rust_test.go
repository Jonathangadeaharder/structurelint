package strategies

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

func TestRustMissingGates(t *testing.T) {
	cfg := map[string]interface{}{"require-cargo-test-lint": true}
	strat := NewRustStrategy(nil, cfg)
	jobs := map[string]core.JobInfo{
		"build": {
			Steps: []core.StepInfo{
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
	cfg := map[string]interface{}{"require-cargo-test-lint": true}
	strat := NewRustStrategy(nil, cfg)
	jobs := map[string]core.JobInfo{
		"quality": {
			Steps: []core.StepInfo{
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

func TestRustStrategyMethods(t *testing.T) {
	cfg := map[string]interface{}{
		"require-cargo-test-lint": true,
		"coverage": map[string]interface{}{
			"lines": 85.0,
		},
	}
	strat := NewRustStrategy(nil, cfg)

	if strat.ProjectType() != core.Rust {
		t.Errorf("expected core.Rust, got %v", strat.ProjectType())
	}

	cov := strat.RequiredCoverage()
	if cov.Lines != 85.0 {
		t.Errorf("expected lines coverage 85.0, got %f", cov.Lines)
	}

	gates := strat.RequiredCIGates()
	if len(gates) != 5 {
		t.Errorf("expected 5 gates, got %d", len(gates))
	}

	linters := strat.RequiredLinters()
	if len(linters) != 2 {
		t.Errorf("expected 2 linters, got %d", len(linters))
	}

	if res := strat.CheckProjectConfig(nil, nil); res != nil {
		t.Errorf("expected nil check project config, got %v", res)
	}

	if res := strat.CheckSuppressions(nil, nil); res != nil {
		t.Errorf("expected nil check suppressions, got %v", res)
	}
}
