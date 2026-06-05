//go:build cgo

package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParserV2_CGO_ParseFile(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := `package main
import "fmt"
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewV2(tmpDir)
	imports, err := p.ParseFile(goFile)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}

	if len(imports) != 1 {
		t.Fatalf("got %d imports, want 1", len(imports))
	}

	if imports[0].ImportPath != "fmt" {
		t.Errorf("import path = %q, want %q", imports[0].ImportPath, "fmt")
	}
}

func TestParserV2_CGO_ParseExports(t *testing.T) {
	tmpDir := t.TempDir()
	goFile := filepath.Join(tmpDir, "test.go")

	content := `package main
func ExportedFunc() {}
`
	if err := os.WriteFile(goFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewV2(tmpDir)
	exports, err := p.ParseExports(goFile)
	if err != nil {
		t.Fatalf("ParseExports() error = %v", err)
	}

	if len(exports) == 0 {
		t.Fatalf("expected at least one export, got 0")
	}
	found := false
	for _, e := range exports {
		for _, n := range e.Names {
			if n == "ExportedFunc" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Fatalf("expected export %q not found in %+v", "ExportedFunc", exports)
	}
}

func TestParserV2_CGO_UnsupportedExtension(t *testing.T) {
	tmpDir := t.TempDir()
	unsupportedFile := filepath.Join(tmpDir, "test.unknown")

	if err := os.WriteFile(unsupportedFile, []byte("some content"), 0644); err != nil {
		t.Fatal(err)
	}

	p := NewV2(tmpDir)
	imports, err := p.ParseFile(unsupportedFile)
	if err != nil {
		t.Fatalf("ParseFile() error = %v", err)
	}
	if len(imports) != 0 {
		t.Errorf("expected 0 imports, got %d", len(imports))
	}

	exports, err := p.ParseExports(unsupportedFile)
	if err != nil {
		t.Fatalf("ParseExports() error = %v", err)
	}
	if len(exports) != 0 {
		t.Errorf("expected 0 exports, got %d", len(exports))
	}
}

func TestParserV2_CGO_ResolveImportPath(t *testing.T) {
	p := NewV2(".")
	res := p.ResolveImportPath("src/main.go", "./utils")
	if res != "src/utils" {
		t.Errorf("ResolveImportPath failed: %s", res)
	}
}
