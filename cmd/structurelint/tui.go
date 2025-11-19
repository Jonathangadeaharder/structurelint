package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/structurelint/structurelint/internal/linter"
	"github.com/structurelint/structurelint/internal/tui"
)

func runTUI(args []string) error {
	fs := flag.NewFlagSet("tui", flag.ExitOnError)
	fixableOnly := fs.Bool("fixable-only", false, "Show only fixable violations")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: structurelint tui [options] [path]\n\n")
		fmt.Fprintf(os.Stderr, "Interactive terminal UI for navigating and fixing violations.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nKeyboard Controls:\n")
		fmt.Fprintf(os.Stderr, "  ↑/↓ or j/k    Navigate violations\n")
		fmt.Fprintf(os.Stderr, "  Enter/Space   View violation details\n")
		fmt.Fprintf(os.Stderr, "  f             Preview and apply fix\n")
		fmt.Fprintf(os.Stderr, "  g             View dependency graph\n")
		fmt.Fprintf(os.Stderr, "  Esc           Go back\n")
		fmt.Fprintf(os.Stderr, "  q             Quit\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  structurelint tui              # Launch interactive mode\n")
		fmt.Fprintf(os.Stderr, "  structurelint tui --fixable-only  # Show only auto-fixable violations\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Get target path
	targetPath := "."
	if fs.NArg() > 0 {
		targetPath = fs.Arg(0)
	}

	// Run linter to get violations
	fmt.Println("Checking for violations...")
	l := linter.New()
	violations, err := l.Lint(targetPath)
	if err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}

	// Filter to fixable only if requested
	if *fixableOnly {
		var fixable []linter.Violation
		for _, v := range violations {
			if v.AutoFix != nil {
				fixable = append(fixable, v)
			}
		}
		violations = fixable
	}

	if len(violations) == 0 {
		fmt.Println("✓ No violations found")
		return nil
	}

	// Create and run TUI
	model := tui.NewModel(violations, *fixableOnly)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("TUI error: %w", err)
	}

	return nil
}

func printTUIHelp() {
	fmt.Println(`structurelint tui - Interactive terminal interface

Usage:
  structurelint tui [options] [path]

Description:
  Launch an interactive terminal UI for navigating and fixing violations.
  The TUI provides a rich, keyboard-driven interface for reviewing violations,
  viewing details, and applying fixes interactively.

Options:
  --fixable-only   Show only violations that can be auto-fixed

Keyboard Controls:
  Navigation:
    ↑/↓ or j/k     Move up/down in the violation list
    Enter/Space    View detailed information about the selected violation

  Actions:
    f              Preview and apply an auto-fix for the selected violation
    g              View dependency graph for the selected file (coming soon)

  General:
    Esc            Go back to the previous view
    q              Quit the interactive mode

Views:
  List View:
    - Shows all violations with file paths and rule names
    - 🔧 icon indicates auto-fixable violations
    - Use arrow keys to navigate

  Detail View:
    - Full violation information
    - Expected vs actual values
    - Suggestions for fixes
    - Context information

  Fix Preview:
    - Shows what the fix will do
    - Confidence level (0-100%)
    - Safety indicator
    - Actions that will be performed
    - Confirm before applying

Examples:
  structurelint tui
    Launch interactive mode for all violations

  structurelint tui --fixable-only
    Show only violations that can be automatically fixed

  structurelint tui ./src
    Check and navigate violations in the src/ directory

Workflow:
  1. Launch TUI: structurelint tui
  2. Navigate violations with arrow keys
  3. Press Enter to view details
  4. Press 'f' to preview available fixes
  5. Press 'y' to apply fix or 'n' to cancel
  6. Repeat until all violations are resolved
  7. Press 'q' to quit

Tips:
  - Use 'f' to quickly apply fixes without leaving the TUI
  - The violation list updates automatically after fixes are applied
  - Use --fixable-only to focus on violations you can fix immediately
  - Unsafe fixes will show a warning before application

Documentation:
  https://github.com/structurelint/structurelint`)
}
