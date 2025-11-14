package rules

import (
	"fmt"
	"os"
	"strings"

	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/walker"
)

// UnusedExportsRule detects exported symbols that are never imported
type UnusedExportsRule struct {
	Graph *graph.ImportGraph
}

// Name returns the rule name
func (r *UnusedExportsRule) Name() string {
	return "disallow-unused-exports"
}

// Check validates that all exports are used somewhere
func (r *UnusedExportsRule) Check(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Violation {
	if r.Graph == nil {
		return []Violation{}
	}

	var violations []Violation

	// Build a map of which symbols are imported from which files
	importedSymbols := make(map[string]map[string]bool) // file -> set of imported symbol names

	for sourceFile, imports := range r.Graph.Dependencies {
		for _, importPath := range imports {
			// Resolve import to actual file
			targetFile := r.resolveImportToFile(importPath)
			if targetFile == "" {
				continue
			}

			// For now, if a file is imported at all, consider all its exports as used
			// A more sophisticated implementation would track specific symbol imports
			if importedSymbols[targetFile] == nil {
				importedSymbols[targetFile] = make(map[string]bool)
			}
			// Mark file as having imports (simple approach)
			importedSymbols[targetFile]["*"] = true

			_ = sourceFile // Avoid unused variable warning
		}
	}

	// Check each file's exports
	for filePath, exports := range r.Graph.Exports {
		// Skip if file has no exports
		if len(exports) == 0 {
			continue
		}

		// Check if this file is imported at all
		if symbols, exists := importedSymbols[filePath]; !exists || len(symbols) == 0 {
			// File has exports but is never imported
			var exportNames []string
			for _, export := range exports {
				exportNames = append(exportNames, export.Names...)
			}

			if len(exportNames) > 0 {
				violations = append(violations, Violation{
					Rule:    r.Name(),
					Path:    filePath,
					Message: fmt.Sprintf("exports %s but is never imported", formatNames(exportNames)),
				})
			}
		}
	}

	return violations
}

// Fix generates fixes for unused exports violations
func (r *UnusedExportsRule) Fix(files []walker.FileInfo, dirs map[string]*walker.DirInfo) []Fix {
	if r.Graph == nil {
		return []Fix{}
	}

	var fixes []Fix

	// Build a map of which files are imported
	importedFiles := make(map[string]bool)
	for _, imports := range r.Graph.Dependencies {
		for _, importPath := range imports {
			targetFile := r.resolveImportToFile(importPath)
			if targetFile != "" {
				importedFiles[targetFile] = true
			}
		}
	}

	// Check each file's exports
	for filePath, exports := range r.Graph.Exports {
		// Skip if file has no exports
		if len(exports) == 0 {
			continue
		}

		// Check if this file is never imported
		if !importedFiles[filePath] {
			// Find the actual file info
			var fileInfo *walker.FileInfo
			for i := range files {
				if files[i].Path == filePath {
					fileInfo = &files[i]
					break
				}
			}

			if fileInfo == nil {
				continue
			}

			// Generate a fix to remove export keywords (TypeScript/JavaScript only for now)
			if strings.HasSuffix(filePath, ".ts") || strings.HasSuffix(filePath, ".tsx") ||
				strings.HasSuffix(filePath, ".js") || strings.HasSuffix(filePath, ".jsx") {

				// Read the file content
				content, err := os.ReadFile(fileInfo.AbsPath)
				if err != nil {
					continue
				}

				// Remove export keywords
				modified := removeExportKeywords(string(content))

				if modified != string(content) {
					fixes = append(fixes, Fix{
						FilePath: filePath,
						Action:   "modify",
						OldValue: string(content),
						NewValue: modified,
					})
				}
			}
		}
	}

	return fixes
}

// removeExportKeywords removes export keywords from TypeScript/JavaScript code
func removeExportKeywords(content string) string {
	lines := strings.Split(content, "\n")
	var modified []string

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Handle different export patterns
		if strings.HasPrefix(trimmed, "export default ") {
			// Convert "export default X" to just "X" (or comment out)
			// For safety, let's comment it out instead of removing
			modified = append(modified, "// "+line+" // TODO: Unused export removed by structurelint --fix")
		} else if strings.HasPrefix(trimmed, "export {") {
			// Comment out export { ... }
			modified = append(modified, "// "+line+" // TODO: Unused export removed by structurelint --fix")
		} else if strings.HasPrefix(trimmed, "export const ") ||
			strings.HasPrefix(trimmed, "export let ") ||
			strings.HasPrefix(trimmed, "export var ") ||
			strings.HasPrefix(trimmed, "export function ") ||
			strings.HasPrefix(trimmed, "export class ") ||
			strings.HasPrefix(trimmed, "export interface ") ||
			strings.HasPrefix(trimmed, "export type ") ||
			strings.HasPrefix(trimmed, "export enum ") {
			// Remove the "export " keyword
			newLine := strings.Replace(line, "export ", "", 1)
			modified = append(modified, newLine)
		} else {
			modified = append(modified, line)
		}
	}

	return strings.Join(modified, "\n")
}

// resolveImportToFile attempts to resolve an import path to an actual file
func (r *UnusedExportsRule) resolveImportToFile(importPath string) string {
	// Try to find a matching file in AllFiles
	for _, file := range r.Graph.AllFiles {
		if strings.HasPrefix(file, importPath) || file == importPath {
			return file
		}

		// Try with common extensions
		for _, ext := range []string{".ts", ".tsx", ".js", ".jsx", ".go", ".py"} {
			if file == importPath+ext {
				return file
			}
		}
	}

	return ""
}

// formatNames formats a list of names for display
func formatNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	if len(names) == 1 {
		return fmt.Sprintf("'%s'", names[0])
	}
	if len(names) <= 3 {
		quoted := make([]string, len(names))
		for i, name := range names {
			quoted[i] = fmt.Sprintf("'%s'", name)
		}
		return strings.Join(quoted, ", ")
	}
	// Show first 3 and count
	return fmt.Sprintf("'%s', '%s', '%s' and %d more", names[0], names[1], names[2], len(names)-3)
}

// NewUnusedExportsRule creates a new UnusedExportsRule
func NewUnusedExportsRule(importGraph *graph.ImportGraph) *UnusedExportsRule {
	return &UnusedExportsRule{
		Graph: importGraph,
	}
}
