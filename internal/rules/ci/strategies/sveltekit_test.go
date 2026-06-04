package strategies

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/rules/ci/core"
)

func TestSvelteKitRequiredGates(t *testing.T) {
	strat := NewSvelteKitStrategy(nil, nil)
	gates := strat.RequiredCIGates()
	if len(gates) < 4 {
		t.Fatalf("expected at least 4 gates, got %d", len(gates))
	}
}

func TestSvelteKitChecksSvelteCheck(t *testing.T) {
	strat := NewSvelteKitStrategy(nil, nil)
	jobs := map[string]core.JobInfo{
		"quality": {
			Steps: []core.StepInfo{
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
	cfg := map[string]interface{}{
		"require-vitest-linter": true,
		"require-svelteuml":     true,
	}
	strat := NewSvelteKitStrategy(nil, cfg)
	jobs := map[string]core.JobInfo{
		"test": {
			Steps: []core.StepInfo{
				{Name: "run tests", Run: "pnpm vitest run"},
			},
		},
	}
	results := strat.CheckWorkflowSteps(jobs)
	foundLinter := false
	foundUml := false
	for _, r := range results {
		if strings.Contains(r.Message, "vitest-linter") {
			foundLinter = true
		}
		if strings.Contains(r.Message, "svelteuml") {
			foundUml = true
		}
	}
	if !foundLinter {
		t.Error("expected violation for missing vitest-linter gate")
	}
	if !foundUml {
		t.Error("expected violation for missing svelteuml gate")
	}
}

func TestSvelteKitStrategyMethods(t *testing.T) {
	cfg := map[string]interface{}{
		"require-vitest-linter": true,
		"coverage": map[string]interface{}{
			"branches":   85.0,
			"lines":      75.0,
			"functions":  85.0,
			"statements": 75.0,
		},
	}
	strat := NewSvelteKitStrategy(nil, cfg)

	if strat.ProjectType() != core.SvelteKit {
		t.Errorf("expected core.SvelteKit, got %v", strat.ProjectType())
	}

	cov := strat.RequiredCoverage()
	if cov.Branches != 85.0 || cov.Lines != 75.0 || cov.Functions != 85.0 || cov.Statements != 75.0 {
		t.Errorf("unexpected coverage thresholds: %v", cov)
	}

	gates := strat.RequiredCIGates()
	if len(gates) != 5 { // 4 default + vitest-linter
		t.Errorf("expected 5 gates, got %d", len(gates))
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
