package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseTypeScriptImports(t *testing.T) {
	// Arrange
	// Create temp file
	tmpDir := t.TempDir()
	tsFile := filepath.Join(tmpDir, "test.ts")

	content := `import { foo } from './foo';
import bar from '../bar';
import * as baz from './baz';
const qux = require('./qux');
`

	if err := os.WriteFile(tsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := New(tmpDir)

	// Act
	imports, err := parser.ParseFile(tsFile)

	// Assert
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(imports) != 4 {
		t.Errorf("Expected 4 imports, got %d", len(imports))
	}

	// Check first import
	if imports[0].ImportPath != "./foo" {
		t.Errorf("Expected import path './foo', got %s", imports[0].ImportPath)
	}

	if !imports[0].IsRelative {
		t.Error("Expected relative import")
	}
}

func TestParseGoImports(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := `package main

import (
	"fmt"
	"github.com/user/repo"
)

import "strings"
`

	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := New(tmpDir)
	imports, err := parser.ParseFile(goFile)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(imports) != 3 {
		t.Errorf("Expected 3 imports, got %d", len(imports))
	}

	expectedPaths := map[string]bool{
		"fmt":                  true,
		"github.com/user/repo": true,
		"strings":              true,
	}

	for _, imp := range imports {
		if !expectedPaths[imp.ImportPath] {
			t.Errorf("Unexpected import path: %s", imp.ImportPath)
		}
	}
}

func TestParsePythonImports(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `import os
from typing import List
import json.decoder
`

	if err := os.WriteFile(pyFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := New(tmpDir)
	imports, err := parser.ParseFile(pyFile)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(imports) != 3 {
		t.Errorf("Expected 3 imports, got %d", len(imports))
	}
}

func TestParseTypeScriptExports(t *testing.T) {
	tmpDir := t.TempDir()
	tsFile := filepath.Join(tmpDir, "test.ts")

	content := `export const foo = 'bar';
export function hello() {}
export class MyClass {}
export { one, two };
export default something;
`

	if err := os.WriteFile(tsFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := New(tmpDir)
	exports, err := parser.ParseExports(tsFile)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(exports) != 5 {
		t.Errorf("Expected 5 export statements, got %d", len(exports))
	}

	// Check for default export
	hasDefault := false
	for _, exp := range exports {
		if exp.IsDefault {
			hasDefault = true
			break
		}
	}

	if !hasDefault {
		t.Error("Expected to find default export")
	}
}

func TestParseGoExports(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := `package main

func HelloWorld() {}
func privateFunc() {}
type PublicType struct{}
type privateType struct{}
const PublicConst = 42
var PrivateVar = "test"
`

	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := New(tmpDir)
	exports, err := parser.ParseExports(goFile)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Should only find exported (uppercase) symbols
	expectedExports := map[string]bool{
		"HelloWorld":  true,
		"PublicType":  true,
		"PublicConst": true,
		"PrivateVar":  true,
	}

	exportCount := 0
	for _, exp := range exports {
		for _, name := range exp.Names {
			if expectedExports[name] {
				exportCount++
			}
		}
	}

	if exportCount != 4 {
		t.Errorf("Expected 4 exported symbols, found %d", exportCount)
	}
}

func TestParsePythonExports(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `def public_function():
    pass

def _private_function():
    pass

class PublicClass:
    pass

__all__ = ['public_function', 'PublicClass']
`

	if err := os.WriteFile(pyFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	parser := New(tmpDir)
	exports, err := parser.ParseExports(pyFile)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(exports) == 0 {
		t.Fatal("Expected to find exports")
	}

	// Should prefer __all__ definition
	hasPublicFunction := false
	hasPublicClass := false

	for _, exp := range exports {
		for _, name := range exp.Names {
			if name == "public_function" {
				hasPublicFunction = true
			}
			if name == "PublicClass" {
				hasPublicClass = true
			}
		}
	}

	if !hasPublicFunction || !hasPublicClass {
		t.Error("Expected to find exports from __all__")
	}
}

func TestResolveImportPath(t *testing.T) {
	parser := New("/project")

	tests := []struct {
		sourceFile string
		importPath string
		expected   string
	}{
		{"src/app.ts", "./utils", "src/utils"},
		{"src/components/Button.tsx", "../hooks/useButton", "src/hooks/useButton"},
		{"src/app.ts", "react", "react"}, // External import
	}

	for _, tt := range tests {
		result := parser.ResolveImportPath(tt.sourceFile, tt.importPath)
		if result != tt.expected {
			t.Errorf("ResolveImportPath(%s, %s) = %s, want %s",
				tt.sourceFile, tt.importPath, result, tt.expected)
		}
	}
}
