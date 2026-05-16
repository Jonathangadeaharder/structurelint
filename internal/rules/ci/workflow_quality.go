package ci

import (
	"path/filepath"
	"strings"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

type WorkflowQualityRule struct {
	requirePytestCoverage           bool
	requireSvelteCheckWarnings      bool
	disallowCommandMasking          bool
	disallowContinueOnErrorOnQuality bool
	requireRequiredChecksAggregator bool
}

func NewWorkflowQualityRule(cfg map[string]interface{}) *WorkflowQualityRule {
	r := &WorkflowQualityRule{}
	if v, ok := cfg["require-pytest-coverage"].(bool); ok {
		r.requirePytestCoverage = v
	}
	if v, ok := cfg["require-svelte-check-fail-on-warnings"].(bool); ok {
		r.requireSvelteCheckWarnings = v
	}
	if v, ok := cfg["disallow-command-masking"].(bool); ok {
		r.disallowCommandMasking = v
	}
	if v, ok := cfg["disallow-continue-on-error-on-quality"].(bool); ok {
		r.disallowContinueOnErrorOnQuality = v
	}
	if v, ok := cfg["require-required-checks-aggregator"].(bool); ok {
		r.requireRequiredChecksAggregator = v
	}
	return r
}

func (r *WorkflowQualityRule) Name() string {
	return "github-workflows"
}

func (r *WorkflowQualityRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation

	workflowFiles := filterWorkflowFiles(files)

	for _, f := range workflowFiles {
		violations = append(violations, r.checkWorkflow(f)...)
	}

	return violations
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

func filterPackageJSONFiles(files []walker.FileInfo) []walker.FileInfo {
	var out []walker.FileInfo
	for _, f := range files {
		if f.IsDir {
			continue
		}
		if filepath.Base(f.Path) == "package.json" {
			out = append(out, f)
		}
	}
	return out
}

func (r *WorkflowQualityRule) checkWorkflow(f walker.FileInfo) []rules.Violation {
	return nil
}
