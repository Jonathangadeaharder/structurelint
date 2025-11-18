package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

// TestContractFrameworkRule_Python tests Python contract framework enforcement
func TestContractFrameworkRule_Python(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(dir string) ([]walker.FileInfo, error)
		requirePython  bool
		wantViolations bool
		description    string
	}{
		{
			name:          "Python with icontract",
			requirePython: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create Python file with icontract
				userPyPath := filepath.Join(dir, "user.py")
				if err := os.WriteFile(userPyPath, []byte(`
from icontract import require, ensure

@require(lambda name: len(name) > 0)
@ensure(lambda result: result.id > 0)
def create_user(name: str):
    return User(id=1, name=name)
`), 0644); err != nil {
					return nil, err
				}

				// Create requirements.txt with icontract
				reqPath := filepath.Join(dir, "requirements.txt")
				if err := os.WriteFile(reqPath, []byte(`
icontract==2.6.2
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: userPyPath, ParentPath: dir, IsDir: false},
					{Path: reqPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: false,
			description:    "Should pass when Python project uses icontract",
		},
		{
			name:          "Python with deal",
			requirePython: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create Python file with deal
				calcPyPath := filepath.Join(dir, "calculator.py")
				if err := os.WriteFile(calcPyPath, []byte(`
import deal

@deal.pre(lambda a, b: b != 0)
@deal.post(lambda result: result > 0)
def divide(a: int, b: int) -> float:
    return a / b
`), 0644); err != nil {
					return nil, err
				}

				// Create pyproject.toml with deal
				pyprojectPath := filepath.Join(dir, "pyproject.toml")
				if err := os.WriteFile(pyprojectPath, []byte(`
[tool.poetry.dependencies]
python = "^3.9"
deal = "^4.24.0"
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: calcPyPath, ParentPath: dir, IsDir: false},
					{Path: pyprojectPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: false,
			description:    "Should pass when Python project uses deal",
		},
		{
			name:          "Python without contract framework",
			requirePython: true,
			setupFiles: func(dir string) ([]walker.FileInfo, error) {
				// Create Python file without contracts
				appPyPath := filepath.Join(dir, "app.py")
				if err := os.WriteFile(appPyPath, []byte(`
def calculate(a, b):
    return a + b
`), 0644); err != nil {
					return nil, err
				}

				return []walker.FileInfo{
					{Path: appPyPath, ParentPath: dir, IsDir: false},
				}, nil
			},
			wantViolations: true,
			description:    "Should fail when Python project lacks contract framework",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a temporary directory with test files
			tmpDir := t.TempDir()
			files, err := tt.setupFiles(tmpDir)
			if err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			// When: we check the rule
			rule := NewContractFrameworkRule(ContractFrameworkRule{
				RequirePython: tt.requirePython,
			})

			violations := rule.Check(files, make(map[string]*walker.DirInfo))

			// Then: violations should match expectations
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("%s: got violations=%v, want violations=%v\nViolations: %v",
					tt.description, hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// TestContractFrameworkRule_TypeScript tests TypeScript contract framework enforcement
func TestContractFrameworkRule_TypeScript(t *testing.T) {
	tests := []struct {
		name              string
		setupFiles        func(dir string) error
		requireTypeScript bool
		wantViolations    bool
		description       string
	}{
		{
			name:              "TypeScript with Zod",
			requireTypeScript: true,
			setupFiles: func(dir string) error {
				// Create TypeScript file with Zod
				if err := os.WriteFile(filepath.Join(dir, "user.ts"), []byte(`
import { z } from "zod";

const UserSchema = z.object({
  id: z.number(),
  name: z.string(),
  email: z.string().email(),
});

export function createUser(data: unknown) {
  const user = UserSchema.parse(data);
  return user;
}
`), 0644); err != nil {
					return err
				}

				// Create package.json with Zod
				return os.WriteFile(filepath.Join(dir, "package.json"), []byte(`
{
  "dependencies": {
    "zod": "^3.22.0"
  }
}
`), 0644)
			},
			wantViolations: false,
			description:    "Should pass when TypeScript project uses Zod",
		},
		{
			name:              "TypeScript without contract framework",
			requireTypeScript: true,
			setupFiles: func(dir string) error {
				// Create TypeScript file without contracts
				return os.WriteFile(filepath.Join(dir, "utils.ts"), []byte(`
export function add(a: number, b: number): number {
  return a + b;
}
`), 0644)
			},
			wantViolations: true,
			description:    "Should fail when TypeScript project lacks contract framework",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a temporary directory with test files
			tmpDir := t.TempDir()
			if err := tt.setupFiles(tmpDir); err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			// When: we walk the directory and check the rule
			w := walker.New(tmpDir)
			if err := w.Walk(); err != nil {
				t.Fatalf("Failed to walk directory: %v", err)
			}

			rule := NewContractFrameworkRule(ContractFrameworkRule{
				RequireTypeScript: tt.requireTypeScript,
			})

			violations := rule.Check(w.GetFiles(), w.GetDirs())

			// Then: violations should match expectations
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("%s: got violations=%v, want violations=%v\nViolations: %v",
					tt.description, hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// TestContractFrameworkRule_Go tests Go contract framework enforcement
func TestContractFrameworkRule_Go(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(dir string) error
		requireGo      bool
		wantViolations bool
		description    string
	}{
		{
			name:      "Go with gocontracts",
			requireGo: true,
			setupFiles: func(dir string) error {
				// Create Go file with gocontracts
				if err := os.WriteFile(filepath.Join(dir, "calculator.go"), []byte(`
package calculator

//go:generate gocontracts

// Add adds two numbers
// Requires: b > 0
// Ensures: result > a
func Add(a, b int) int {
	return a + b
}
`), 0644); err != nil {
					return err
				}

				// Create go.mod with gocontracts
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(`
module example.com/calculator

go 1.21

require github.com/s-kostyaev/gocontracts v0.1.0
`), 0644)
			},
			wantViolations: false,
			description:    "Should pass when Go project uses gocontracts",
		},
		{
			name:      "Go with standard assertions",
			requireGo: true,
			setupFiles: func(dir string) error {
				// Create Go file with standard assertions
				if err := os.WriteFile(filepath.Join(dir, "validator.go"), []byte(`
package validator

func Validate(input string) error {
	if len(input) == 0 {
		panic("input cannot be empty")
	}
	return nil
}
`), 0644); err != nil {
					return err
				}

				// Create go.mod
				return os.WriteFile(filepath.Join(dir, "go.mod"), []byte(`
module example.com/validator

go 1.21
`), 0644)
			},
			wantViolations: false,
			description:    "Should pass when Go project uses standard assertions",
		},
		{
			name:      "Go without contract framework",
			requireGo: true,
			setupFiles: func(dir string) error {
				// Create Go file without contracts
				return os.WriteFile(filepath.Join(dir, "utils.go"), []byte(`
package utils

func Add(a, b int) int {
	return a + b
}
`), 0644)
			},
			wantViolations: true,
			description:    "Should fail when Go project lacks contract framework",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a temporary directory with test files
			tmpDir := t.TempDir()
			if err := tt.setupFiles(tmpDir); err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			// When: we walk the directory and check the rule
			w := walker.New(tmpDir)
			if err := w.Walk(); err != nil {
				t.Fatalf("Failed to walk directory: %v", err)
			}

			rule := NewContractFrameworkRule(ContractFrameworkRule{
				RequireGo: tt.requireGo,
			})

			violations := rule.Check(w.GetFiles(), w.GetDirs())

			// Then: violations should match expectations
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("%s: got violations=%v, want violations=%v\nViolations: %v",
					tt.description, hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// TestContractFrameworkRule_Java tests Java contract framework enforcement
func TestContractFrameworkRule_Java(t *testing.T) {
	tests := []struct {
		name           string
		setupFiles     func(dir string) error
		requireJava    bool
		wantViolations bool
		description    string
	}{
		{
			name:        "Java with Jakarta Bean Validation",
			requireJava: true,
			setupFiles: func(dir string) error {
				// Create Java file with Jakarta validation
				if err := os.WriteFile(filepath.Join(dir, "User.java"), []byte(`
import jakarta.validation.constraints.*;

public class User {
    @NotNull
    @Size(min = 1, max = 100)
    private String name;

    @Email
    private String email;

    @Min(18)
    private int age;
}
`), 0644); err != nil {
					return err
				}

				// Create pom.xml with Jakarta validation
				return os.WriteFile(filepath.Join(dir, "pom.xml"), []byte(`
<dependencies>
    <dependency>
        <groupId>jakarta.validation</groupId>
        <artifactId>jakarta.validation-api</artifactId>
        <version>3.0.2</version>
    </dependency>
</dependencies>
`), 0644)
			},
			wantViolations: false,
			description:    "Should pass when Java project uses Jakarta Bean Validation",
		},
		{
			name:        "Java without contract framework",
			requireJava: true,
			setupFiles: func(dir string) error {
				// Create Java file without contracts
				return os.WriteFile(filepath.Join(dir, "Calculator.java"), []byte(`
public class Calculator {
    public int add(int a, int b) {
        return a + b;
    }
}
`), 0644)
			},
			wantViolations: true,
			description:    "Should fail when Java project lacks contract framework",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a temporary directory with test files
			tmpDir := t.TempDir()
			if err := tt.setupFiles(tmpDir); err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			// When: we walk the directory and check the rule
			w := walker.New(tmpDir)
			if err := w.Walk(); err != nil {
				t.Fatalf("Failed to walk directory: %v", err)
			}

			rule := NewContractFrameworkRule(ContractFrameworkRule{
				RequireJava: tt.requireJava,
			})

			violations := rule.Check(w.GetFiles(), w.GetDirs())

			// Then: violations should match expectations
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("%s: got violations=%v, want violations=%v\nViolations: %v",
					tt.description, hasViolations, tt.wantViolations, violations)
			}
		})
	}
}

// TestContractFrameworkRule_CSharp tests C# contract framework enforcement and deprecated warnings
func TestContractFrameworkRule_CSharp(t *testing.T) {
	tests := []struct {
		name              string
		setupFiles        func(dir string) error
		requireCSharp     bool
		wantViolations    bool
		wantDeprecation   bool
		description       string
	}{
		{
			name:          "C# with nullable reference types",
			requireCSharp: true,
			setupFiles: func(dir string) error {
				// Create C# file with nullable reference types
				if err := os.WriteFile(filepath.Join(dir, "User.cs"), []byte(`
#nullable enable

public class User
{
    public string Name { get; set; }

    public User(string name)
    {
        ArgumentNullException.ThrowIfNull(name);
        Name = name;
    }
}
`), 0644); err != nil {
					return err
				}

				// Create .csproj file
				return os.WriteFile(filepath.Join(dir, "MyProject.csproj"), []byte(`
<Project Sdk="Microsoft.NET.Sdk">
    <PropertyGroup>
        <TargetFramework>net8.0</TargetFramework>
        <Nullable>enable</Nullable>
    </PropertyGroup>
</Project>
`), 0644)
			},
			wantViolations:  false,
			wantDeprecation: false,
			description:     "Should pass when C# project uses nullable reference types",
		},
		{
			name:          "C# with deprecated Code Contracts",
			requireCSharp: true,
			setupFiles: func(dir string) error {
				// Create C# file with deprecated Code Contracts
				return os.WriteFile(filepath.Join(dir, "Calculator.cs"), []byte(`
using System.Diagnostics.Contracts;

public class Calculator
{
    public int Divide(int a, int b)
    {
        Contract.Requires(b != 0);
        return a / b;
    }
}
`), 0644)
			},
			wantViolations:  true,
			wantDeprecation: true,
			description:     "Should warn when C# project uses deprecated Code Contracts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Given: a temporary directory with test files
			tmpDir := t.TempDir()
			if err := tt.setupFiles(tmpDir); err != nil {
				t.Fatalf("Failed to setup test files: %v", err)
			}

			// When: we walk the directory and check the rule
			w := walker.New(tmpDir)
			if err := w.Walk(); err != nil {
				t.Fatalf("Failed to walk directory: %v", err)
			}

			rule := NewContractFrameworkRule(ContractFrameworkRule{
				RequireCSharp: tt.requireCSharp,
			})

			violations := rule.Check(w.GetFiles(), w.GetDirs())

			// Then: violations should match expectations
			hasViolations := len(violations) > 0
			if hasViolations != tt.wantViolations {
				t.Errorf("%s: got violations=%v, want violations=%v\nViolations: %v",
					tt.description, hasViolations, tt.wantViolations, violations)
			}

			// Check for deprecation warnings
			if tt.wantDeprecation {
				foundDeprecation := false
				for _, v := range violations {
					if v.Message == "DEPRECATED: Microsoft Code Contracts is deprecated and unsupported in .NET Core+. Use nullable reference types (#nullable enable) and ArgumentNullException.ThrowIfNull instead." {
						foundDeprecation = true
						break
					}
				}
				if !foundDeprecation {
					t.Errorf("%s: expected deprecation warning but none found", tt.description)
				}
			}
		})
	}
}

// TestContractFrameworkRule_Name tests the rule name
func TestContractFrameworkRule_Name(t *testing.T) {
	// Given: a contract framework rule
	rule := NewContractFrameworkRule(ContractFrameworkRule{})

	// When: we get the name
	name := rule.Name()

	// Then: it should be "contract-framework"
	if name != "contract-framework" {
		t.Errorf("Expected rule name to be 'contract-framework', got '%s'", name)
	}
}
