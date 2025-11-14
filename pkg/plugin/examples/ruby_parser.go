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

// RubyParser is an example plugin that parses Ruby files
type RubyParser struct{}

// Name returns the plugin name
func (p *RubyParser) Name() string {
	return "ruby-parser"
}

// Version returns the plugin version
func (p *RubyParser) Version() string {
	return "1.0.0"
}

// SupportedExtensions returns file extensions this parser handles
func (p *RubyParser) SupportedExtensions() []string {
	return []string{".rb"}
}

// ParseImports extracts require statements from Ruby files
func (p *RubyParser) ParseImports(filePath string) ([]plugin.Import, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var imports []plugin.Import
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Match various require patterns:
	// require 'foo'
	// require "foo"
	// require_relative 'foo'
	// require_relative "../foo"
	requireRegex := regexp.MustCompile(`^\s*require\s+['"]([^'"]+)['"]`)
	requireRelativeRegex := regexp.MustCompile(`^\s*require_relative\s+['"]([^'"]+)['"]`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for require_relative (always relative)
		if matches := requireRelativeRegex.FindStringSubmatch(line); matches != nil {
			importPath := strings.TrimSpace(matches[1])
			imports = append(imports, plugin.Import{
				ImportPath: importPath,
				IsRelative: true,
				Line:       lineNum,
			})
			continue
		}

		// Check for regular require
		if matches := requireRegex.FindStringSubmatch(line); matches != nil {
			importPath := strings.TrimSpace(matches[1])
			// Relative if starts with . or /
			isRelative := strings.HasPrefix(importPath, ".") || strings.HasPrefix(importPath, "/")
			imports = append(imports, plugin.Import{
				ImportPath: importPath,
				IsRelative: isRelative,
				Line:       lineNum,
			})
		}
	}

	return imports, scanner.Err()
}

// ParseExports extracts exported symbols from Ruby files
func (p *RubyParser) ParseExports(filePath string) ([]plugin.Export, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var exports []plugin.Export
	scanner := bufio.NewScanner(file)
	lineNum := 0

	// Match public class/module/method definitions:
	// class Foo
	// module Bar
	// def foo
	classRegex := regexp.MustCompile(`^\s*class\s+(\w+)`)
	moduleRegex := regexp.MustCompile(`^\s*module\s+(\w+)`)
	methodRegex := regexp.MustCompile(`^\s*def\s+(\w+)`)

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// Check for class definitions
		if matches := classRegex.FindStringSubmatch(line); matches != nil {
			exports = append(exports, plugin.Export{
				Name: matches[1],
				Kind: "class",
				Line: lineNum,
			})
		}

		// Check for module definitions
		if matches := moduleRegex.FindStringSubmatch(line); matches != nil {
			exports = append(exports, plugin.Export{
				Name: matches[1],
				Kind: "module",
				Line: lineNum,
			})
		}

		// Check for method definitions
		if matches := methodRegex.FindStringSubmatch(line); matches != nil {
			// Skip private/protected methods (simplified - doesn't handle all cases)
			if !strings.Contains(line, "private") && !strings.Contains(line, "protected") {
				exports = append(exports, plugin.Export{
					Name: matches[1],
					Kind: "method",
					Line: lineNum,
				})
			}
		}
	}

	return exports, scanner.Err()
}

// Example usage:
// func init() {
//     plugin.RegisterParser(&RubyParser{})
// }
