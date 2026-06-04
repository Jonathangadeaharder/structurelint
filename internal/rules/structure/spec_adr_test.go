package structure

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// TestSpecADRRule_FolderEnforcement tests that folders are correctly required when configured.
func TestSpecADRRule_FolderEnforcement(t *testing.T) {
	requireTrue := true
	requireFalse := false

	tests := []struct {
		name              string
		rule              *SpecADRRule
		dirs              map[string]*walker.DirInfo
		wantViolationMsgs []string
	}{
		{
			name: "All folders present - should pass",
			rule: NewSpecADRRule(SpecADRRule{
				RequireSpecFolder: &requireTrue,
				RequireADRFolder:  &requireTrue,
			}),
			dirs: map[string]*walker.DirInfo{
				"docs/specs": {},
				"docs/adr":   {},
			},
			wantViolationMsgs: []string{},
		},
		{
			name: "Spec folder missing - should fail",
			rule: NewSpecADRRule(SpecADRRule{
				RequireSpecFolder: &requireTrue,
				RequireADRFolder:  &requireFalse,
			}),
			dirs: map[string]*walker.DirInfo{
				"docs/adr": {},
			},
			wantViolationMsgs: []string{"No specification folder found"},
		},
		{
			name: "ADR folder missing - should fail",
			rule: NewSpecADRRule(SpecADRRule{
				RequireSpecFolder: &requireFalse,
				RequireADRFolder:  &requireTrue,
			}),
			dirs: map[string]*walker.DirInfo{
				"docs/specs": {},
			},
			wantViolationMsgs: []string{"No ADR (Architecture Decision Record) folder found"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := tt.rule.Check(nil, tt.dirs)
			for _, expectedMsg := range tt.wantViolationMsgs {
				found := false
				for _, v := range violations {
					if strings.Contains(v.Message, expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected violation containing '%s', got %v", expectedMsg, violations)
				}
			}
			if len(tt.wantViolationMsgs) == 0 && len(violations) > 0 {
				t.Errorf("Expected no violations, got %v", violations)
			}
		})
	}
}

// TestSpecADRRule_SpecTemplateEnforcement tests specification template validation.
func TestSpecADRRule_SpecTemplateEnforcement(t *testing.T) {
	requireTrue := true

	tests := []struct {
		name              string
		setupFiles        func(dir string) ([]walker.FileInfo, error)
		wantViolationMsgs []string
		description       string
	}{
		{
			name: "Valid spec",
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				specPath := filepath.Join(dir, "feature-valid-spec.md")
				content := `# Feature Specification: Valid Feature
**Feature Branch**: 123-valid-feature
**Created**: 2026-06-04
**Status**: Draft

## User Scenarios & Testing

### User Story 1 - Something (Priority: P1)
**Independent Test**: Test it.
**Acceptance Scenarios**:
1. Given some state, when action, then outcome.

## Requirements

### Functional Requirements

- **FR-001**: System MUST do something.

## Success Criteria

- **SC-001**: System successfully does something.
`
				if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				return []walker.FileInfo{{Path: specPath, ParentPath: dir, IsDir: false}}, nil
			},
			wantViolationMsgs: []string{},
			description:       "Should pass when spec matches template",
		},
		{
			name: "Missing required sections",
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				specPath := filepath.Join(dir, "feature-missing-sections-spec.md")
				content := `# Feature Specification: Missing Stuff
`
				if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				return []walker.FileInfo{{Path: specPath, ParentPath: dir, IsDir: false}}, nil
			},
			wantViolationMsgs: []string{"Feature specification is missing required sections"},
			description:       "Should fail when required template headings are missing",
		},
		{
			name: "Missing user stories priority",
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				specPath := filepath.Join(dir, "feature-no-priority-spec.md")
				content := `# Feature Specification: No Priority
## User Scenarios & Testing
### User Story 1 - Test Story (No priority here)
## Requirements
### Functional Requirements
- **FR-001**: Done
## Success Criteria
- **SC-001**: Done
`
				if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				return []walker.FileInfo{{Path: specPath, ParentPath: dir, IsDir: false}}, nil
			},
			wantViolationMsgs: []string{"prioritized user stories"},
			description:       "Should fail when user stories lack priority",
		},
		{
			name: "Missing FR prefix",
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				specPath := filepath.Join(dir, "feature-no-fr-spec.md")
				content := `# Feature Specification: No FR
## User Scenarios & Testing
### User Story 1 - Test (Priority: P1)
## Requirements
### Functional Requirements
- System MUST do something without FR prefix.
## Success Criteria
- **SC-001**: Done
`
				if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				return []walker.FileInfo{{Path: specPath, ParentPath: dir, IsDir: false}}, nil
			},
			wantViolationMsgs: []string{"Functional requirements must use FR-### format"},
			description:       "Should fail when functional requirements lack FR-### format",
		},
		{
			name: "Missing SC prefix",
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				specPath := filepath.Join(dir, "feature-no-sc-spec.md")
				content := `# Feature Specification: No SC
## User Scenarios & Testing
### User Story 1 - Test (Priority: P1)
## Requirements
### Functional Requirements
- **FR-001**: Done
## Success Criteria
- System works fine without SC.
`
				if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				return []walker.FileInfo{{Path: specPath, ParentPath: dir, IsDir: false}}, nil
			},
			wantViolationMsgs: []string{"Success criteria must use SC-### format"},
			description:       "Should fail when success criteria lack SC-### format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			files, err := tt.setupFiles(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			requireFalse := false
			rule := NewSpecADRRule(SpecADRRule{
				RequireSpecFolder:   &requireFalse,
				RequireADRFolder:    &requireFalse,
				EnforceSpecTemplate: &requireTrue,
			})

			violations := rule.Check(files, make(map[string]*walker.DirInfo))

			for _, expectedMsg := range tt.wantViolationMsgs {
				found := false
				for _, v := range violations {
					if strings.Contains(v.Message, expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: Expected violation containing '%s', got %v", tt.description, expectedMsg, violations)
				}
			}
			if len(tt.wantViolationMsgs) == 0 && len(violations) > 0 {
				t.Errorf("%s: Expected no violations, got %v", tt.description, violations)
			}
		})
	}
}

// TestSpecADRRule_ADRTemplateEnforcement tests ADR template validation.
func TestSpecADRRule_ADRTemplateEnforcement(t *testing.T) {
	requireTrue := true

	tests := []struct {
		name              string
		setupFiles        func(dir string) ([]walker.FileInfo, error)
		wantViolationMsgs []string
		description       string
	}{
		{
			name: "Valid ADR",
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				adrPath := filepath.Join(dir, "ADR-001-valid.md")
				content := `---
status: accepted
date: 2026-06-04
---
# ADR-001: Valid
## Context and Problem Statement
Context description.
## Considered Options
Options description.
## Decision Outcome
Outcome description.
`
				if err := os.WriteFile(adrPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				return []walker.FileInfo{{Path: adrPath, ParentPath: dir, IsDir: false}}, nil
			},
			wantViolationMsgs: []string{},
			description:       "Should pass when ADR is fully valid",
		},
		{
			name: "Missing status metadata",
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				adrPath := filepath.Join(dir, "ADR-002-missing-status.md")
				content := `---
date: 2026-06-04
---
# ADR-002: Incomplete
## Context and Problem Statement
Context description.
## Considered Options
Options description.
## Decision Outcome
Outcome description.
`
				if err := os.WriteFile(adrPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				return []walker.FileInfo{{Path: adrPath, ParentPath: dir, IsDir: false}}, nil
			},
			wantViolationMsgs: []string{"ADR must include a 'status' field"},
			description:       "Should fail when status metadata is missing",
		},
		{
			name: "Missing headings",
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				adrPath := filepath.Join(dir, "ADR-003-missing-headings.md")
				content := `---
status: accepted
date: 2026-06-04
---
# ADR-003: No headings
`
				if err := os.WriteFile(adrPath, []byte(content), 0644); err != nil {
					return nil, err
				}
				return []walker.FileInfo{{Path: adrPath, ParentPath: dir, IsDir: false}}, nil
			},
			wantViolationMsgs: []string{"ADR is missing required sections"},
			description:       "Should fail when required headings are missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			files, err := tt.setupFiles(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			requireFalse := false
			rule := NewSpecADRRule(SpecADRRule{
				RequireSpecFolder:   &requireFalse,
				RequireADRFolder:    &requireFalse,
				EnforceADRTemplate: &requireTrue,
			})

			violations := rule.Check(files, make(map[string]*walker.DirInfo))

			for _, expectedMsg := range tt.wantViolationMsgs {
				found := false
				for _, v := range violations {
					if strings.Contains(v.Message, expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: Expected violation containing '%s', got %v", tt.description, expectedMsg, violations)
				}
			}
			if len(tt.wantViolationMsgs) == 0 && len(violations) > 0 {
				t.Errorf("%s: Expected no violations, got %v", tt.description, violations)
			}
		})
	}
}

// TestSpecADRRule_CustomConfig tests that custom headers and metadata are enforced.
func TestSpecADRRule_CustomConfig(t *testing.T) {
	requireTrue := true
	tmpDir := t.TempDir()

	adrPath := filepath.Join(tmpDir, "ADR-100-custom.md")
	content := `---
author: Jonathan
date: 2026-06-04
---
# ADR-100
## Custom Header One
Details.
`
	if err := os.WriteFile(adrPath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	requireFalse := false
	rule := NewSpecADRRule(SpecADRRule{
		RequireSpecFolder:   &requireFalse,
		RequireADRFolder:    &requireFalse,
		EnforceADRTemplate:  &requireTrue,
		ADRRequiredHeadings: []string{"## Custom Header One"},
		ADRRequiredMetadata: []string{"author:"},
	})

	violations := rule.Check([]walker.FileInfo{{Path: adrPath, ParentPath: tmpDir, IsDir: false}}, make(map[string]*walker.DirInfo))
	if len(violations) > 0 {
		t.Errorf("Expected no violations with custom config, got %v", violations)
	}
}
