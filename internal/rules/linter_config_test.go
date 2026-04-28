package rules

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLinterConfigRule_Name(t *testing.T) {
	rule := &LinterConfigRule{}
	assert.Equal(t, "linter-config", rule.Name())
}

func TestLinterConfigRule_PythonWithPyprojectToml(t *testing.T) {
	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: "pyproject.toml", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_PythonWithoutConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.NotEmpty(t, violations)
	assert.True(t, containsMessage(violations, "No Python linter configuration found"))
}

func TestLinterConfigRule_PythonWithFlake8Config(t *testing.T) {
	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: ".flake8", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_TypeScriptWithESLintConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireTypeScript: true,
	}
	files := []walker.FileInfo{
		{Path: "main.ts", ParentPath: ".", IsDir: false},
		{Path: ".eslintrc.json", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_TypeScriptWithoutConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireTypeScript: true,
	}
	files := []walker.FileInfo{
		{Path: "main.ts", ParentPath: ".", IsDir: false},
		{Path: "package.json", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.NotEmpty(t, violations)
	assert.True(t, containsMessage(violations, "No TypeScript linter configuration found"))
}

func TestLinterConfigRule_GoWithGolangCIConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireGo: true,
	}
	files := []walker.FileInfo{
		{Path: "main.go", ParentPath: ".", IsDir: false},
		{Path: ".golangci.yml", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_GoWithoutConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireGo: true,
	}
	files := []walker.FileInfo{
		{Path: "main.go", ParentPath: ".", IsDir: false},
		{Path: "go.mod", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.NotEmpty(t, violations)
	assert.True(t, containsMessage(violations, "No Go linter configuration found"))
}

func TestLinterConfigRule_PythonWithWorkflow(t *testing.T) {
	tmpDir := t.TempDir()
	workflowsDir := filepath.Join(tmpDir, ".github", "workflows")
	require.NoError(t, os.MkdirAll(workflowsDir, 0755))

	workflowContent := `
name: Python Linting
on: [push, pull_request]
jobs:
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Run Black
        run: black --check .
      - name: Run mypy
        run: mypy src/
`
	workflowPath := filepath.Join(workflowsDir, "python-lint.yml")
	require.NoError(t, os.WriteFile(workflowPath, []byte(workflowContent), 0644))

	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: tmpDir, IsDir: false},
		{Path: workflowPath, ParentPath: workflowsDir, IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_NoPythonFiles(t *testing.T) {
	rule := &LinterConfigRule{
		RequirePython: true,
	}
	files := []walker.FileInfo{
		{Path: "main.go", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_MultipleLanguages(t *testing.T) {
	rule := &LinterConfigRule{
		RequirePython:     true,
		RequireTypeScript: true,
		RequireGo:         true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: "app.ts", ParentPath: ".", IsDir: false},
		{Path: "server.go", ParentPath: ".", IsDir: false},
		{Path: "pyproject.toml", ParentPath: ".", IsDir: false},
		{Path: ".eslintrc.json", ParentPath: ".", IsDir: false},
		{Path: ".golangci.yml", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_PrettierConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireTypeScript: true,
	}
	files := []walker.FileInfo{
		{Path: "main.ts", ParentPath: ".", IsDir: false},
		{Path: ".prettierrc", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_TsconfigJson(t *testing.T) {
	rule := &LinterConfigRule{
		RequireTypeScript: true,
	}
	files := []walker.FileInfo{
		{Path: "main.ts", ParentPath: ".", IsDir: false},
		{Path: "tsconfig.json", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_JavaScriptFilesRequireTypeScriptLinters(t *testing.T) {
	rule := &LinterConfigRule{
		RequireTypeScript: true,
	}
	files := []walker.FileInfo{
		{Path: "app.js", ParentPath: ".", IsDir: false},
		{Path: ".eslintrc.json", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_HTMLWithHTMLHintConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireHTML: true,
	}
	files := []walker.FileInfo{
		{Path: "index.html", ParentPath: ".", IsDir: false},
		{Path: ".htmlhintrc", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_HTMLWithoutConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireHTML: true,
	}
	files := []walker.FileInfo{
		{Path: "index.html", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.NotEmpty(t, violations)
	assert.True(t, containsMessage(violations, "No HTML linter configuration found"))
}

func TestLinterConfigRule_CSSWithStylelintConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireCSS: true,
	}
	files := []walker.FileInfo{
		{Path: "styles.css", ParentPath: ".", IsDir: false},
		{Path: ".stylelintrc.json", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_CSSWithoutConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireCSS: true,
	}
	files := []walker.FileInfo{
		{Path: "styles.css", ParentPath: ".", IsDir: false},
		{Path: "package.json", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.NotEmpty(t, violations)
	assert.True(t, containsMessage(violations, "No CSS linter configuration found"))
}

func TestLinterConfigRule_SQLWithSQLFluffConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireSQL: true,
	}
	files := []walker.FileInfo{
		{Path: "query.sql", ParentPath: ".", IsDir: false},
		{Path: ".sqlfluff", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_SQLWithoutConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireSQL: true,
	}
	files := []walker.FileInfo{
		{Path: "query.sql", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.NotEmpty(t, violations)
	assert.True(t, containsMessage(violations, "No SQL linter configuration found"))
}

func TestLinterConfigRule_RustWithRustfmtConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireRust: true,
	}
	files := []walker.FileInfo{
		{Path: "main.rs", ParentPath: ".", IsDir: false},
		{Path: "rustfmt.toml", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_RustWithoutConfig(t *testing.T) {
	rule := &LinterConfigRule{
		RequireRust: true,
	}
	files := []walker.FileInfo{
		{Path: "main.rs", ParentPath: ".", IsDir: false},
		{Path: "Cargo.toml", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.NotEmpty(t, violations)
	assert.True(t, containsMessage(violations, "No Rust linter configuration found"))
}

func TestLinterConfigRule_NoHTMLFiles(t *testing.T) {
	rule := &LinterConfigRule{
		RequireHTML: true,
	}
	files := []walker.FileInfo{
		{Path: "main.go", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_AllLanguages(t *testing.T) {
	rule := &LinterConfigRule{
		RequirePython:     true,
		RequireTypeScript: true,
		RequireGo:         true,
		RequireHTML:       true,
		RequireCSS:        true,
		RequireSQL:        true,
		RequireRust:       true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: "app.ts", ParentPath: ".", IsDir: false},
		{Path: "server.go", ParentPath: ".", IsDir: false},
		{Path: "index.html", ParentPath: ".", IsDir: false},
		{Path: "styles.css", ParentPath: ".", IsDir: false},
		{Path: "query.sql", ParentPath: ".", IsDir: false},
		{Path: "lib.rs", ParentPath: ".", IsDir: false},
		{Path: "pyproject.toml", ParentPath: ".", IsDir: false},
		{Path: ".eslintrc.json", ParentPath: ".", IsDir: false},
		{Path: ".golangci.yml", ParentPath: ".", IsDir: false},
		{Path: ".htmlhintrc", ParentPath: ".", IsDir: false},
		{Path: ".stylelintrc.json", ParentPath: ".", IsDir: false},
		{Path: ".sqlfluff", ParentPath: ".", IsDir: false},
		{Path: "rustfmt.toml", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_LanguageConfigurations(t *testing.T) {
	tests := []struct {
		name            string
		rule            *LinterConfigRule
		files           []walker.FileInfo
		expectViolation bool
		expectedMessage string
	}{
		{
			name: "Markdown with config",
			rule: &LinterConfigRule{RequireMarkdown: true},
			files: []walker.FileInfo{
				{Path: "README.md", ParentPath: ".", IsDir: false},
				{Path: ".markdownlint.json", ParentPath: ".", IsDir: false},
			},
			expectViolation: false,
		},
		{
			name: "Markdown without config",
			rule: &LinterConfigRule{RequireMarkdown: true},
			files: []walker.FileInfo{
				{Path: "README.md", ParentPath: ".", IsDir: false},
				{Path: "package.json", ParentPath: ".", IsDir: false},
			},
			expectViolation: true,
			expectedMessage: "No Markdown linter configuration found",
		},
		{
			name: "Java with Checkstyle config",
			rule: &LinterConfigRule{RequireJava: true},
			files: []walker.FileInfo{
				{Path: "src/Main.java", ParentPath: ".", IsDir: false},
				{Path: "checkstyle.xml", ParentPath: ".", IsDir: false},
			},
			expectViolation: false,
		},
		{
			name: "Java with pom.xml",
			rule: &LinterConfigRule{RequireJava: true},
			files: []walker.FileInfo{
				{Path: "src/Main.java", ParentPath: ".", IsDir: false},
				{Path: "pom.xml", ParentPath: ".", IsDir: false},
			},
			expectViolation: false,
		},
		{
			name: "Java without config",
			rule: &LinterConfigRule{RequireJava: true},
			files: []walker.FileInfo{
				{Path: "src/Main.java", ParentPath: ".", IsDir: false},
				{Path: "README.md", ParentPath: ".", IsDir: false},
			},
			expectViolation: true,
			expectedMessage: "No Java linter configuration found",
		},
		{
			name: "C++ with clang-format config",
			rule: &LinterConfigRule{RequireCpp: true},
			files: []walker.FileInfo{
				{Path: "src/main.cpp", ParentPath: ".", IsDir: false},
				{Path: ".clang-format", ParentPath: ".", IsDir: false},
			},
			expectViolation: false,
		},
		{
			name: "C++ with clang-tidy config",
			rule: &LinterConfigRule{RequireCpp: true},
			files: []walker.FileInfo{
				{Path: "src/main.cpp", ParentPath: ".", IsDir: false},
				{Path: ".clang-tidy", ParentPath: ".", IsDir: false},
			},
			expectViolation: false,
		},
		{
			name: "C++ without config",
			rule: &LinterConfigRule{RequireCpp: true},
			files: []walker.FileInfo{
				{Path: "src/main.cpp", ParentPath: ".", IsDir: false},
				{Path: "README.md", ParentPath: ".", IsDir: false},
			},
			expectViolation: true,
			expectedMessage: "No C++ linter configuration found",
		},
		{
			name: "C# with EditorConfig",
			rule: &LinterConfigRule{RequireCSharp: true},
			files: []walker.FileInfo{
				{Path: "src/Program.cs", ParentPath: ".", IsDir: false},
				{Path: ".editorconfig", ParentPath: ".", IsDir: false},
			},
			expectViolation: false,
		},
		{
			name: "C# with StyleCop config",
			rule: &LinterConfigRule{RequireCSharp: true},
			files: []walker.FileInfo{
				{Path: "src/Program.cs", ParentPath: ".", IsDir: false},
				{Path: "stylecop.json", ParentPath: ".", IsDir: false},
			},
			expectViolation: false,
		},
		{
			name: "C# without config",
			rule: &LinterConfigRule{RequireCSharp: true},
			files: []walker.FileInfo{
				{Path: "src/Program.cs", ParentPath: ".", IsDir: false},
				{Path: "README.md", ParentPath: ".", IsDir: false},
			},
			expectViolation: true,
			expectedMessage: "No C# linter configuration found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			violations := tt.rule.Check(tt.files, make(map[string]*walker.DirInfo))

			if tt.expectViolation {
				assert.NotEmpty(t, violations, "Expected violation but got none")
				if tt.expectedMessage != "" {
					assert.True(t, containsMessage(violations, tt.expectedMessage),
						"Expected message '%s' but got: %v", tt.expectedMessage, violations)
				}
			} else {
				assert.Empty(t, violations)
			}
		})
	}
}

func TestLinterConfigRule_AllLanguagesIncludingNew(t *testing.T) {
	rule := &LinterConfigRule{
		RequirePython:     true,
		RequireTypeScript: true,
		RequireGo:         true,
		RequireHTML:       true,
		RequireCSS:        true,
		RequireSQL:        true,
		RequireRust:       true,
		RequireMarkdown:   true,
		RequireJava:       true,
		RequireCpp:        true,
		RequireCSharp:     true,
	}
	files := []walker.FileInfo{
		{Path: "main.py", ParentPath: ".", IsDir: false},
		{Path: "app.ts", ParentPath: ".", IsDir: false},
		{Path: "server.go", ParentPath: ".", IsDir: false},
		{Path: "index.html", ParentPath: ".", IsDir: false},
		{Path: "styles.css", ParentPath: ".", IsDir: false},
		{Path: "query.sql", ParentPath: ".", IsDir: false},
		{Path: "lib.rs", ParentPath: ".", IsDir: false},
		{Path: "README.md", ParentPath: ".", IsDir: false},
		{Path: "Main.java", ParentPath: ".", IsDir: false},
		{Path: "main.cpp", ParentPath: ".", IsDir: false},
		{Path: "Program.cs", ParentPath: ".", IsDir: false},
		{Path: "pyproject.toml", ParentPath: ".", IsDir: false},
		{Path: ".eslintrc.json", ParentPath: ".", IsDir: false},
		{Path: ".golangci.yml", ParentPath: ".", IsDir: false},
		{Path: ".htmlhintrc", ParentPath: ".", IsDir: false},
		{Path: ".stylelintrc.json", ParentPath: ".", IsDir: false},
		{Path: ".sqlfluff", ParentPath: ".", IsDir: false},
		{Path: "rustfmt.toml", ParentPath: ".", IsDir: false},
		{Path: ".markdownlint.json", ParentPath: ".", IsDir: false},
		{Path: "checkstyle.xml", ParentPath: ".", IsDir: false},
		{Path: ".clang-format", ParentPath: ".", IsDir: false},
		{Path: ".editorconfig", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	assert.Empty(t, violations)
}

func TestLinterConfigRule_SuggestionsPresent(t *testing.T) {
	rule := &LinterConfigRule{
		RequireMarkdown: true,
	}
	files := []walker.FileInfo{
		{Path: "README.md", ParentPath: ".", IsDir: false},
	}

	violations := rule.Check(files, make(map[string]*walker.DirInfo))

	require.NotEmpty(t, violations, "Expected violation for missing Markdown linter configuration")
	require.NotEmpty(t, violations[0].Suggestions, "Expected suggestions to be present in violation")
	assert.NotEmpty(t, violations[0].Expected, "Expected 'Expected' field to be populated in violation")
	assert.NotEmpty(t, violations[0].Actual, "Expected 'Actual' field to be populated in violation")
}
