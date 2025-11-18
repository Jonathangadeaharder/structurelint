package rules

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

// TestSpecADRRule_SpecFolderRequirement tests specification folder enforcement
func TestSpecADRRule_SpecFolderRequirement(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(dir string) ([]walker.FileInfo, map[string]*walker.DirInfo, error)
		requireFolder  bool
		wantViolations bool
		description    string
	}{
		{
			name:          "Project with docs/specs folder",
			requireFolder: true,
			setupFiles: func(dir string) ([]walker.FileInfo, map[string]*walker.DirInfo, error) {
				specsDir := filepath.Join(dir, "docs", "specs")
				if err := os.MkdirAll(specsDir, 0755); err != nil {
					return nil, nil, err
				}

				return []walker.FileInfo{},
					map[string]*walker.DirInfo{
						"docs/specs": {},
					}, nil
			},
			wantViolations: false,
			description:    "Should pass when docs/specs folder exists",
		},
		{
			name:          "Project without specs folder",
			requireFolder: true,
			setupFiles: func(dir string) ([]walker.FileInfo, map[string]*walker.DirInfo, error) {
				return []walker.FileInfo{},
					map[string]*walker.DirInfo{
						"src": {},
					}, nil
			},
			wantViolations: true,
			description:    "Should fail when no specs folder exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			files, dirs, err := tt.setupFiles(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			rule := NewSpecADRRule(SpecADRRule{
				RequireSpecFolder: tt.requireFolder,
			})

			violations := rule.Check(files, dirs)

			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("%s: got violations=%v, want violations=%v\nViolations: %v",
					tt.description, hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// TestSpecADRRule_ADRFolderRequirement tests ADR folder enforcement
func TestSpecADRRule_ADRFolderRequirement(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(dir string) ([]walker.FileInfo, map[string]*walker.DirInfo, error)
		requireFolder  bool
		wantViolations bool
		description    string
	}{
		{
			name:          "Project with docs/adr folder",
			requireFolder: true,
			setupFiles: func(dir string) ([]walker.FileInfo, map[string]*walker.DirInfo, error) {
				adrDir := filepath.Join(dir, "docs", "adr")
				if err := os.MkdirAll(adrDir, 0755); err != nil {
					return nil, nil, err
				}

				return []walker.FileInfo{},
					map[string]*walker.DirInfo{
						"docs/adr": {},
					}, nil
			},
			wantViolations: false,
			description:    "Should pass when docs/adr folder exists",
		},
		{
			name:          "Project without ADR folder",
			requireFolder: true,
			setupFiles: func(dir string) ([]walker.FileInfo, map[string]*walker.DirInfo, error) {
				return []walker.FileInfo{},
					map[string]*walker.DirInfo{
						"src": {},
					}, nil
			},
			wantViolations: true,
			description:    "Should fail when no ADR folder exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			files, dirs, err := tt.setupFiles(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			rule := NewSpecADRRule(SpecADRRule{
				RequireADRFolder: tt.requireFolder,
			})

			violations := rule.Check(files, dirs)

			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("%s: got violations=%v, want violations=%v\nViolations: %v",
					tt.description, hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// TestSpecADRRule_SpecTemplateEnforcement tests feature specification template validation
func TestSpecADRRule_SpecTemplateEnforcement(t *testing.T) {
	tests := []struct {
		name              string
		setupFiles        func(dir string) ([]walker.FileInfo, error)
		enforceTemplate   bool
		wantViolationMsgs []string
		description       string
	}{
		{
			name:            "Valid feature specification",
			enforceTemplate: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				specPath := filepath.Join(dir, "feature-auth-spec.md")
				content := `# Feature Specification: User Authentication

**Feature Branch**: [###-user-auth]
**Created**: 2025-01-15
**Status**: Draft

## User Scenarios & Testing

### User Story 1 - Login (Priority: P1)

Users can log in with email and password

**Independent Test**: Can be tested by submitting login form

**Acceptance Scenarios**:
1. **Given** valid credentials, **When** user submits, **Then** user is authenticated

## Requirements

### Functional Requirements

- **FR-001**: System MUST validate email format
- **FR-002**: System MUST hash passwords

## Success Criteria

### Measurable Outcomes

- **SC-001**: Users can log in within 5 seconds
- **SC-002**: System handles 1000 concurrent logins
`
				if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: specPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolationMsgs: []string{},
			description:       "Should pass when spec has all required sections",
		},
		{
			name:            "Missing user scenarios section",
			enforceTemplate: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				specPath := filepath.Join(dir, "feature-incomplete-spec.md")
				content := `# Feature Specification: Incomplete Feature

## Requirements

### Functional Requirements

- **FR-001**: Something

## Success Criteria

- **SC-001**: Something
`
				if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: specPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolationMsgs: []string{"User Scenarios & Testing"},
			description:       "Should fail when missing user scenarios",
		},
		{
			name:            "Missing priorities in user stories",
			enforceTemplate: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				specPath := filepath.Join(dir, "feature-no-priority-spec.md")
				content := `# Feature Specification: No Priority Feature

## User Scenarios & Testing

### User Story 1 - Something

Users can do something

## Requirements

### Functional Requirements

- **FR-001**: System MUST do something

## Success Criteria

- **SC-001**: Users can do something
`
				if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: specPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolationMsgs: []string{"prioritized user stories"},
			description:       "Should fail when user stories lack priorities",
		},
		{
			name:            "Missing FR format",
			enforceTemplate: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				specPath := filepath.Join(dir, "feature-bad-fr-spec.md")
				content := `# Feature Specification: Bad FR Format

## User Scenarios & Testing

### User Story 1 - Something (Priority: P1)

Users can do something

## Requirements

### Functional Requirements

- System must do something (no FR number)

## Success Criteria

- **SC-001**: Users can do something
`
				if err := os.WriteFile(specPath, []byte(content), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: specPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolationMsgs: []string{"FR-### format"},
			description:       "Should fail when functional requirements lack FR-### format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			files, err := tt.setupFiles(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			rule := NewSpecADRRule(SpecADRRule{
				EnforceSpecTemplate: tt.enforceTemplate,
			})

			violations := rule.Check(files, make(map[string]*walker.DirInfo))

			// Check for expected violation messages
			for _, expectedMsg := range tt.wantViolationMsgs {
				found := false
				for _, v := range violations {
					if strings.Contains(v.Message, expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: expected violation message containing '%s' but not found. Got: %v",
						tt.description, expectedMsg, violations)
				}
			}

			// If no violations expected, ensure none found
			if len(tt.wantViolationMsgs) == 0 && len(violations) > 0 {
				t.Errorf("%s: expected no violations but got: %v",
					tt.description, violations)
			}
		})
	}
}

// TestSpecADRRule_ADRTemplateEnforcement tests ADR template validation
func TestSpecADRRule_ADRTemplateEnforcement(t *testing.T) {
	tests := []struct {
		name              string
		setupFiles        func(dir string) ([]walker.FileInfo, error)
		enforceTemplate   bool
		wantViolationMsgs []string
		description       string
	}{
		{
			name:            "Valid ADR",
			enforceTemplate: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				adrPath := filepath.Join(dir, "ADR-001-use-postgres.md")
				content := `---
status: accepted
date: 2025-01-15
decision-makers: Tech Team
---

# Use PostgreSQL for Database

## Context and Problem Statement

We need to choose a database for our application.

## Considered Options

* PostgreSQL
* MySQL
* MongoDB

## Decision Outcome

Chosen option: "PostgreSQL", because it provides ACID compliance and excellent performance.

### Consequences

* Good, because we get strong consistency
* Bad, because learning curve for team

## Pros and Cons of the Options

### PostgreSQL

* Good, because ACID compliant
* Good, because mature ecosystem
`
				if err := os.WriteFile(adrPath, []byte(content), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: adrPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolationMsgs: []string{},
			description:       "Should pass when ADR has all required sections",
		},
		{
			name:            "Missing context section",
			enforceTemplate: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				adrPath := filepath.Join(dir, "ADR-002-incomplete.md")
				content := `---
status: proposed
date: 2025-01-15
---

# Incomplete ADR

## Considered Options

* Option 1
* Option 2

## Decision Outcome

Chosen option: "Option 1"
`
				if err := os.WriteFile(adrPath, []byte(content), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: adrPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolationMsgs: []string{"Context and Problem Statement"},
			description:       "Should fail when missing context section",
		},
		{
			name:            "Missing status field",
			enforceTemplate: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				adrPath := filepath.Join(dir, "ADR-003-no-status.md")
				content := `---
date: 2025-01-15
---

# ADR Without Status

## Context and Problem Statement

Some problem

## Considered Options

* Option 1

## Decision Outcome

Chosen option: "Option 1"
`
				if err := os.WriteFile(adrPath, []byte(content), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: adrPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolationMsgs: []string{"status"},
			description:       "Should fail when missing status field",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			files, err := tt.setupFiles(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			rule := NewSpecADRRule(SpecADRRule{
				EnforceADRTemplate: tt.enforceTemplate,
			})

			violations := rule.Check(files, make(map[string]*walker.DirInfo))

			// Check for expected violation messages
			for _, expectedMsg := range tt.wantViolationMsgs {
				found := false
				for _, v := range violations {
					if strings.Contains(v.Message, expectedMsg) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("%s: expected violation message containing '%s' but not found. Got: %v",
						tt.description, expectedMsg, violations)
				}
			}

			// If no violations expected, ensure none found
			if len(tt.wantViolationMsgs) == 0 && len(violations) > 0 {
				t.Errorf("%s: expected no violations but got: %v",
					tt.description, violations)
			}
		})
	}
}

// TestSpecADRRule_Name tests the rule name
func TestSpecADRRule_Name(t *testing.T) {
	rule := NewSpecADRRule(SpecADRRule{})
	name := rule.Name()

	if name != "spec-adr-enforcement" {
		t.Errorf("Expected rule name to be 'spec-adr-enforcement', got '%s'", name)
	}
}
