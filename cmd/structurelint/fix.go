package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/structurelint/structurelint/internal/autofix"
	"github.com/structurelint/structurelint/internal/linter"
)

func runFix(args []string) error {
	fs := flag.NewFlagSet("fix", flag.ExitOnError)
	dryRun := fs.Bool("dry-run", false, "Show what would be fixed without applying changes")
	interactive := fs.Bool("interactive", false, "Prompt before applying each fix")
	autoFlag := fs.Bool("auto", false, "Automatically apply all safe fixes without prompting")
	ruleFilter := fs.String("rule", "", "Only fix violations from this rule")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: structurelint fix [options] [path]\n\n")
		fmt.Fprintf(os.Stderr, "Auto-fix violations detected by structurelint.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  structurelint fix --dry-run          # Preview fixes without applying\n")
		fmt.Fprintf(os.Stderr, "  structurelint fix --interactive      # Prompt for each fix\n")
		fmt.Fprintf(os.Stderr, "  structurelint fix --auto             # Apply all safe fixes\n")
		fmt.Fprintf(os.Stderr, "  structurelint fix --rule naming      # Only fix naming violations\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Get target path
	targetPath := "."
	if fs.NArg() > 0 {
		targetPath = fs.Arg(0)
	}

	// Use the linter to get violations
	fmt.Println("Checking for violations...")
	l := linter.New()
	allViolations, err := l.Lint(targetPath)
	if err != nil {
		return fmt.Errorf("failed to lint: %w", err)
	}

	// Filter to only violations with auto-fix
	fixableViolations := filterFixable(allViolations, *ruleFilter)

	if len(fixableViolations) == 0 {
		fmt.Println("✓ No fixable violations found")
		return nil
	}

	fmt.Printf("\nFound %d fixable violation(s)\n\n", len(fixableViolations))

	// Create auto-fix engine
	engine := autofix.NewEngine(*dryRun)

	// Convert violations to fixes (pass nil for files since we use built-in AutoFix)
	fixes, err := engine.GenerateFixes(fixableViolations, nil)
	if err != nil {
		return fmt.Errorf("failed to generate fixes: %w", err)
	}

	// Apply fixes
	result, err := applyFixes(engine, fixes, *interactive, *autoFlag, *dryRun)
	if err != nil {
		return fmt.Errorf("failed to apply fixes: %w", err)
	}

	// Print summary
	fmt.Println("\n" + result.String())

	if !*dryRun && result.Applied > 0 {
		fmt.Println("\n✓ Fixes applied successfully!")
		fmt.Println("  Run 'git diff' to review changes")
		fmt.Println("  Run 'structurelint' to verify fixes")
	}

	return nil
}

// filterFixable filters violations to only those with auto-fix capability
func filterFixable(violations []linter.Violation, ruleFilter string) []linter.Violation {
	var fixable []linter.Violation

	for _, v := range violations {
		// Skip if no auto-fix available
		if v.AutoFix == nil {
			continue
		}

		// Apply rule filter if specified
		if ruleFilter != "" && !strings.Contains(v.Rule, ruleFilter) {
			continue
		}

		fixable = append(fixable, v)
	}

	return fixable
}

// applyFixes applies fixes with optional interactive prompting
func applyFixes(
	engine *autofix.Engine,
	fixes []*autofix.Fix,
	interactive bool,
	auto bool,
	dryRun bool,
) (*autofix.FixResult, error) {
	result := &autofix.FixResult{
		DryRun: dryRun,
		Fixes:  fixes,
	}

	reader := bufio.NewReader(os.Stdin)

	for _, fix := range fixes {
		// Show fix details
		fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
		fmt.Printf("Rule:        %s\n", fix.Violation.Rule)
		fmt.Printf("File:        %s\n", fix.Violation.Path)
		fmt.Printf("Description: %s\n", fix.Description)
		fmt.Printf("Confidence:  %.0f%%\n", fix.Confidence*100)

		if len(fix.Actions) > 0 {
			fmt.Printf("\nActions:\n")
			for i, action := range fix.Actions {
				fmt.Printf("  %d. %s\n", i+1, action.Describe())
			}
		}

		// Decide whether to apply
		shouldApply := false

		if dryRun {
			shouldApply = true // Just show, don't actually apply
		} else if auto {
			shouldApply = fix.Safe // Auto-apply only safe fixes
			if !fix.Safe {
				fmt.Printf("\n⚠ Skipping unsafe fix (use --interactive to review)\n")
			}
		} else if interactive {
			// Prompt user
			fmt.Printf("\nApply this fix? [y/n/q] ")
			response, err := reader.ReadString('\n')
			if err != nil {
				return result, fmt.Errorf("failed to read input: %w", err)
			}

			response = strings.ToLower(strings.TrimSpace(response))

			if response == "q" {
				fmt.Println("\nQuitting...")
				return result, nil
			}

			shouldApply = (response == "y" || response == "yes")
		} else {
			// Default: apply safe fixes, prompt for unsafe
			if fix.Safe {
				shouldApply = true
			} else {
				fmt.Printf("\n⚠ Unsafe fix requires confirmation. Apply? [y/n] ")
				response, err := reader.ReadString('\n')
				if err != nil {
					return result, fmt.Errorf("failed to read input: %w", err)
				}
				shouldApply = strings.ToLower(strings.TrimSpace(response)) == "y"
			}
		}

		// Apply or skip
		if shouldApply {
			if dryRun {
				fmt.Printf("\n[DRY RUN] Would apply fix\n")
				result.Applied++
			} else {
				// Apply the fix
				applied, err := engine.ApplyFixes([]*autofix.Fix{fix})
				if err != nil {
					fmt.Printf("\n✗ Failed to apply fix: %v\n", err)
					result.Failed++
					result.Errors = append(result.Errors, err)
				} else if applied > 0 {
					fmt.Printf("\n✓ Fix applied\n")
					result.Applied++
				}
			}
		} else {
			fmt.Printf("\n⊘ Skipped\n")
			result.Skipped++
		}
	}

	return result, nil
}

func printFixHelp() {
	fmt.Println(`structurelint fix - Auto-fix detected violations

Usage:
  structurelint fix [options] [path]

Description:
  Automatically fix violations detected by structurelint. The fix command
  supports dry-run mode for previewing changes, interactive mode for
  reviewing each fix, and automatic mode for applying safe fixes without
  prompting.

Options:
  --config <path>      Path to configuration file (default: .structurelint.yml)
  --dry-run            Show what would be fixed without applying changes
  --interactive        Prompt before applying each fix
  --auto               Automatically apply all safe fixes without prompting
  --rule <name>        Only fix violations from this rule

Modes:
  Default:      Apply safe fixes automatically, prompt for unsafe fixes
  --dry-run:    Preview all fixes without making changes
  --interactive: Prompt for every fix (safe or unsafe)
  --auto:       Apply safe fixes only, skip unsafe fixes

Examples:
  structurelint fix --dry-run
    Preview all available fixes without applying them

  structurelint fix --interactive
    Review and approve each fix one by one

  structurelint fix --auto
    Automatically apply all safe fixes

  structurelint fix --rule github-workflows
    Only fix violations from the github-workflows rule

  structurelint fix --auto --rule naming
    Automatically fix all naming convention violations

Safety:
  - Each fix has a confidence level (0-100%)
  - Safe fixes can be applied automatically with --auto
  - Unsafe fixes require explicit confirmation
  - In dry-run mode, no changes are made to files
  - Always review changes with 'git diff' after applying fixes

Workflow:
  1. Run 'structurelint' to detect violations
  2. Run 'structurelint fix --dry-run' to preview fixes
  3. Run 'structurelint fix' to apply fixes
  4. Run 'git diff' to review changes
  5. Run 'structurelint' to verify fixes resolved violations

Documentation:
  https://github.com/structurelint/structurelint`)
}
