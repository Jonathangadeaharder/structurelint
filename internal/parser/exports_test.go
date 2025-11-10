package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParser_parseGoExports(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantExports int
		wantNames   []string
	}{
		{
			name: "exported function",
			content: `package main

func PublicFunc() {}
func privateFunc() {}
`,
			wantExports: 1,
			wantNames:   []string{"PublicFunc"},
		},
		{
			name: "exported type",
			content: `package main

type PublicType struct {}
type privateType struct {}
`,
			wantExports: 1,
			wantNames:   []string{"PublicType"},
		},
		{
			name: "exported const and var",
			content: `package main

const PublicConst = "value"
var PublicVar = 123
const privateConst = "value"
`,
			wantExports: 2,
			wantNames:   []string{"PublicConst", "PublicVar"},
		},
		{
			name: "mixed exports",
			content: `package main

func ExportedFunc() {}
type ExportedType struct {}
const ExportedConst = 1
`,
			wantExports: 3,
			wantNames:   []string{"ExportedFunc", "ExportedType", "ExportedConst"},
		},
		{
			name:        "no exports",
			content:     `package main

func privateFunc() {}
`,
			wantExports: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.go")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			// Act
			p := New(tmpDir)
			exports, err := p.parseGoExports(testFile)

			// Assert
			if err != nil {
				t.Fatalf("parseGoExports() error = %v", err)
			}

			totalExports := 0
			for _, exp := range exports {
				totalExports += len(exp.Names)
			}

			if totalExports != tt.wantExports {
				t.Errorf("got %d exports, want %d", totalExports, tt.wantExports)
			}

			if tt.wantNames != nil {
				gotNames := make(map[string]bool)
				for _, exp := range exports {
					for _, name := range exp.Names {
						gotNames[name] = true
					}
				}
				for _, wantName := range tt.wantNames {
					if !gotNames[wantName] {
						t.Errorf("missing expected export: %s", wantName)
					}
				}
			}
		})
	}
}

func TestParser_parseTypeScriptJavaScriptExports(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantExports int
		hasDefault  bool
	}{
		{
			name: "named exports",
			content: `export const foo = 1;
export function bar() {}
export class Baz {}
`,
			wantExports: 3,
			hasDefault:  false,
		},
		{
			name: "default export",
			content: `const MyComponent = () => {};
export default MyComponent;
`,
			wantExports: 1,
			hasDefault:  true,
		},
		{
			name: "export list",
			content: `const a = 1;
const b = 2;
export { a, b };
`,
			wantExports: 1, // One export entry with multiple names
			hasDefault:  false,
		},
		{
			name: "mixed exports",
			content: `export const value = 123;
export default function main() {}
`,
			wantExports: 2,
			hasDefault:  true,
		},
		{
			name:        "no exports",
			content:     `const internal = 123;`,
			wantExports: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.ts")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			p := New(tmpDir)
			exports, err := p.parseTypeScriptJavaScriptExports(testFile)
			if err != nil {
				t.Fatalf("parseTypeScriptJavaScriptExports() error = %v", err)
			}

			if len(exports) != tt.wantExports {
				t.Errorf("got %d export entries, want %d", len(exports), tt.wantExports)
			}

			hasDefault := false
			for _, exp := range exports {
				if exp.IsDefault {
					hasDefault = true
					break
				}
			}

			if hasDefault != tt.hasDefault {
				t.Errorf("got hasDefault=%v, want %v", hasDefault, tt.hasDefault)
			}
		})
	}
}

func TestParser_parsePythonExports(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantExports int
		wantNames   []string
	}{
		{
			name: "with __all__",
			content: `__all__ = ["foo", "bar"]

def foo():
    pass

def bar():
    pass

def _private():
    pass
`,
			wantExports: 1,
			wantNames:   []string{"foo", "bar"},
		},
		{
			name: "without __all__ - public functions",
			content: `def public_func():
    pass

def _private_func():
    pass

class PublicClass:
    pass
`,
			wantExports: 1,
			wantNames:   []string{"public_func", "PublicClass"},
		},
		{
			name: "only private functions",
			content: `def _private():
    pass
`,
			wantExports: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			testFile := filepath.Join(tmpDir, "test.py")
			if err := os.WriteFile(testFile, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			p := New(tmpDir)
			exports, err := p.parsePythonExports(testFile)
			if err != nil {
				t.Fatalf("parsePythonExports() error = %v", err)
			}

			if len(exports) != tt.wantExports {
				t.Errorf("got %d export entries, want %d", len(exports), tt.wantExports)
			}

			if tt.wantNames != nil && len(exports) > 0 {
				gotNames := exports[0].Names
				if len(gotNames) != len(tt.wantNames) {
					t.Errorf("got %d names, want %d", len(gotNames), len(tt.wantNames))
				}
			}
		})
	}
}
