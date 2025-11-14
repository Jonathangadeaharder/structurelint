// Package examples provides example plugin implementations for structurelint.
//
// @structurelint:no-test Example plugin implementations for documentation
package examples

import (
	"bufio"
	"os"
	"regexp"
	"strings"

	"github.com/structurelint/structurelint/pkg/plugin"
)

// RustParser is an example plugin that parses Rust files
type RustParser struct{}

// Name returns the plugin name
func (p *RustParser) Name() string {
	return "rust-parser"
}

// Version returns the plugin version
func (p *RustParser) Version() string {
	return "1.0.0"
}

// SupportedExtensions returns file extensions this parser handles
func (p *RustParser) SupportedExtensions() []string {
	return []string{".rs"}
}

// ParseImports extracts import statements from Rust files
func (p *RustParser) ParseImports(filePath string) ([]plugin.Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var imports []plugin.Import
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Match: use foo::bar;
	// Match: use foo::bar::{baz, qux};
	// Match: use super::foo;
	// Match: use crate::foo;
	useRegex := regexp.MustCompile(`^\s*use\s+([^;]+);`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		if matches := useRegex.FindStringSubmatch(line); matches != nil {
			importPath := strings.TrimSpace(matches[1])

			// Determine if relative (starts with super::, self::, or crate::)
			isRelative := strings.HasPrefix(importPath, "super::") ||
				strings.HasPrefix(importPath, "self::") ||
				strings.HasPrefix(importPath, "crate::")

			// Remove the curly braces part for simplicity (e.g., foo::{bar, baz} -> foo)
			if idx := strings.Index(importPath, "{"); idx != -1 {
				importPath = strings.TrimSpace(importPath[:idx])
				importPath = strings.TrimSuffix(importPath, "::")
			}

			imports = append(imports, plugin.Import{
				ImportPath: importPath,
				IsRelative: isRelative,
				Line:       lineNum,
			})
		}
	}

	return imports, scanner.Err()
}

// ParseExports extracts exported symbols from Rust files
func (p *RustParser) ParseExports(filePath string) ([]plugin.Export, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var exports []plugin.Export
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Match public declarations:
	// pub fn foo()
	// pub struct Bar
	// pub enum Baz
	// pub const QUX: i32
	// pub trait Trait
	pubFnRegex := regexp.MustCompile(`^\s*pub\s+fn\s+(\w+)`)
	pubStructRegex := regexp.MustCompile(`^\s*pub\s+struct\s+(\w+)`)
	pubEnumRegex := regexp.MustCompile(`^\s*pub\s+enum\s+(\w+)`)
	pubConstRegex := regexp.MustCompile(`^\s*pub\s+const\s+(\w+)`)
	pubTraitRegex := regexp.MustCompile(`^\s*pub\s+trait\s+(\w+)`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for public functions
		if matches := pubFnRegex.FindStringSubmatch(line); matches != nil {
			exports = append(exports, plugin.Export{
				Name: matches[1],
				Kind: "function",
				Line: lineNum,
			})
		}

		// Check for public structs
		if matches := pubStructRegex.FindStringSubmatch(line); matches != nil {
			exports = append(exports, plugin.Export{
				Name: matches[1],
				Kind: "struct",
				Line: lineNum,
			})
		}

		// Check for public enums
		if matches := pubEnumRegex.FindStringSubmatch(line); matches != nil {
			exports = append(exports, plugin.Export{
				Name: matches[1],
				Kind: "enum",
				Line: lineNum,
			})
		}

		// Check for public constants
		if matches := pubConstRegex.FindStringSubmatch(line); matches != nil {
			exports = append(exports, plugin.Export{
				Name: matches[1],
				Kind: "const",
				Line: lineNum,
			})
		}

		// Check for public traits
		if matches := pubTraitRegex.FindStringSubmatch(line); matches != nil {
			exports = append(exports, plugin.Export{
				Name: matches[1],
				Kind: "trait",
				Line: lineNum,
			})
		}
	}

	return exports, scanner.Err()
}

// Example usage:
// func init() {
//     plugin.RegisterParser(&RustParser{})
// }
