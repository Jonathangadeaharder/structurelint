package ci

import (
	"errors"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

type ErrorFileReader struct{}

func (ErrorFileReader) ReadFile(path string) ([]byte, error) {
	return nil, errors.New("read error")
}

type StringFileReader struct {
	content string
}

func (s StringFileReader) ReadFile(path string) ([]byte, error) {
	return []byte(s.content), nil
}

func TestWorkflowQualityRule_ParseErrors(t *testing.T) {
	// Test file read error
	file := walker.FileInfo{Path: ".github/workflows/ci.yml", AbsPath: "/project/.github/workflows/ci.yml"}
	jobs := parseWorkflowJobs(file, ErrorFileReader{})
	if jobs != nil {
		t.Errorf("expected nil jobs on file read error, got %v", jobs)
	}

	// Test invalid yaml
	invalidYamlReader := StringFileReader{content: `invalid: {yaml:`}
	jobs = parseWorkflowJobs(file, invalidYamlReader)
	if jobs != nil {
		t.Errorf("expected nil jobs on invalid yaml, got %v", jobs)
	}

	// Test non-mapping document root
	listRootReader := StringFileReader{content: `- item1`}
	jobs = parseWorkflowJobs(file, listRootReader)
	if len(jobs) != 0 {
		t.Errorf("expected 0 jobs on non-mapping document root, got %v", jobs)
	}

	// Test non-mapping job node
	nonMappingJobReader := StringFileReader{content: `
jobs:
  list-job: [1, 2, 3]
`}
	jobs = parseWorkflowJobs(file, nonMappingJobReader)
	if len(jobs) != 0 {
		t.Fatalf("expected 0 job entries for non-mapping job, got %d", len(jobs))
	}

	// Test non-mapping step node
	nonMappingStepReader := StringFileReader{content: `
jobs:
  job1:
    steps:
      - "non-mapping-step"
`}
	jobs = parseWorkflowJobs(file, nonMappingStepReader)
	if len(jobs["job1"].Steps) != 0 {
		t.Errorf("expected 0 steps for non-mapping step list, got %v", jobs["job1"].Steps)
	}
}
