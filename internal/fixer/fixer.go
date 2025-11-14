// @structurelint:ignore test-adjacency Fixer is tested through integration tests
package fixer

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/structurelint/structurelint/internal/rules"
)

// Fixer applies automated fixes to the codebase
type Fixer struct {
	dryRun  bool
	verbose bool
}

// New creates a new Fixer
func New(dryRun, verbose bool) *Fixer {
	return &Fixer{
		dryRun:  dryRun,
		verbose: verbose,
	}
}

// Apply applies a list of fixes to the filesystem
func (f *Fixer) Apply(fixes []rules.Fix) error {
	if len(fixes) == 0 {
		return nil
	}

	if f.verbose || f.dryRun {
		fmt.Printf("Applying %d fix(es)...\n", len(fixes))
	}

	for _, fix := range fixes {
		if err := f.applyFix(fix); err != nil {
			return fmt.Errorf("failed to apply fix for %s: %w", fix.FilePath, err)
		}
	}

	if f.dryRun {
		fmt.Printf("Dry run complete. No changes were made.\n")
	} else if f.verbose {
		fmt.Printf("Successfully applied %d fix(es).\n", len(fixes))
	}

	return nil
}

// applyFix applies a single fix
func (f *Fixer) applyFix(fix rules.Fix) error {
	switch fix.Action {
	case "rename":
		return f.applyRename(fix)
	case "delete":
		return f.applyDelete(fix)
	case "modify":
		return f.applyModify(fix)
	default:
		return fmt.Errorf("unknown fix action: %s", fix.Action)
	}
}

// applyRename renames a file or directory
func (f *Fixer) applyRename(fix rules.Fix) error {
	if f.verbose || f.dryRun {
		fmt.Printf("  [%s] Rename: %s -> %s\n", fix.Action, fix.OldValue, fix.NewValue)
	}

	if f.dryRun {
		return nil
	}

	// Ensure the directory for the new file exists
	newDir := filepath.Dir(fix.NewValue)
	if err := os.MkdirAll(newDir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", newDir, err)
	}

	return os.Rename(fix.OldValue, fix.NewValue)
}

// applyDelete deletes a file
func (f *Fixer) applyDelete(fix rules.Fix) error {
	if f.verbose || f.dryRun {
		fmt.Printf("  [%s] Delete: %s\n", fix.Action, fix.FilePath)
	}

	if f.dryRun {
		return nil
	}

	return os.Remove(fix.FilePath)
}

// applyModify modifies file content
func (f *Fixer) applyModify(fix rules.Fix) error {
	if f.verbose || f.dryRun {
		fmt.Printf("  [%s] Modify: %s\n", fix.Action, fix.FilePath)
	}

	if f.dryRun {
		return nil
	}

	// Read current file content
	content, err := os.ReadFile(fix.FilePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// For now, just write the new value directly
	// TODO: Implement more sophisticated content modification (e.g., line-based edits)
	_ = content // Will be used for more sophisticated fixes

	return os.WriteFile(fix.FilePath, []byte(fix.NewValue), 0644)
}
