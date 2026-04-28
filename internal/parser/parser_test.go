package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	require.NoError(t, os.WriteFile(tsFile, []byte(content), 0644))

	parser := New(tmpDir)

	// Act
	imports, err := parser.ParseFile(tsFile)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 4, len(imports))

	// Check first import
	assert.Equal(t, "./foo", imports[0].ImportPath)
	assert.True(t, imports[0].IsRelative)
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

	require.NoError(t, os.WriteFile(goFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(goFile)

	require.NoError(t, err)
	assert.Equal(t, 3, len(imports))

	expectedPaths := map[string]bool{
		"fmt":                  true,
		"github.com/user/repo": true,
		"strings":              true,
	}

	for _, imp := range imports {
		assert.True(t, expectedPaths[imp.ImportPath], "Unexpected import path: %s", imp.ImportPath)
	}
}

func TestParsePythonImports(t *testing.T) {
	tmpDir := t.TempDir()
	pyFile := filepath.Join(tmpDir, "test.py")

	content := `import os
from typing import List
import json.decoder
`

	require.NoError(t, os.WriteFile(pyFile, []byte(content), 0644))

	parser := New(tmpDir)
	imports, err := parser.ParseFile(pyFile)

	require.NoError(t, err)
	assert.Equal(t, 3, len(imports))
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

	require.NoError(t, os.WriteFile(tsFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(tsFile)

	require.NoError(t, err)
	assert.Equal(t, 5, len(exports))

	// Check for default export
	hasDefault := false
	for _, exp := range exports {
		if exp.IsDefault {
			hasDefault = true
			break
		}
	}

	assert.True(t, hasDefault)
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

	require.NoError(t, os.WriteFile(goFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(goFile)

	require.NoError(t, err)

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

	assert.Equal(t, 4, exportCount)
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

	require.NoError(t, os.WriteFile(pyFile, []byte(content), 0644))

	parser := New(tmpDir)
	exports, err := parser.ParseExports(pyFile)

	require.NoError(t, err)
	require.NotEmpty(t, exports, "Expected to find exports")

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

	assert.True(t, hasPublicFunction)
	assert.True(t, hasPublicClass)
}

func TestResolveImportPath(t *testing.T) {
	parser := New("/project")

	tests := []struct {
		sourceFile string
		importPath string
		expected   string
	}{
		{"src/app.ts", "./utils", filepath.Join("src", "utils")},
		{"src/components/Button.tsx", "../hooks/useButton", filepath.Join("src", "hooks", "useButton")},
		{"src/app.ts", "react", "react"}, // External import
	}

	for _, tt := range tests {
		result := parser.ResolveImportPath(tt.sourceFile, tt.importPath)
		assert.Equal(t, tt.expected, result, "ResolveImportPath(%s, %s)", tt.sourceFile, tt.importPath)
	}
}
