//go:build !cgo

package parser

import (
	"testing"
)

func TestParserV2_NoCGO_Fallback(t *testing.T) {
	// Create a parser
	p := NewV2(".")

	// Verify it's not nil
	if p == nil {
		t.Fatal("NewV2 returned nil")
	}

	// Verify it has the fallback parser
	if p.v1 == nil {
		t.Error("NewV2 did not initialize fallback parser")
	}

	// Test ResolveImportPath (simple delegation)
	res := p.ResolveImportPath("src/main.go", "./utils")
	if res != "src\\utils" && res != "src/utils" {
		t.Errorf("ResolveImportPath failed: %s", res)
	}
}
