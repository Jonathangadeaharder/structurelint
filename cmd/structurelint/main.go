package main

import (
	"fmt"
	"os"

	"github.com/structurelint/structurelint/internal/linter"
)

// Version is set during build via ldflags
var Version = "dev"

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command line arguments
	if len(os.Args) > 1 {
		arg := os.Args[1]
		if arg == "--version" || arg == "-v" {
			fmt.Printf("structurelint version %s\n", Version)
			return nil
		}
		if arg == "--help" || arg == "-h" {
			printHelp()
			return nil
		}
	}

	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	// Create and run linter
	l := linter.New()
	violations, err := l.Lint(path)
	if err != nil {
		return err
	}

	// Report violations
	if len(violations) > 0 {
		for _, v := range violations {
			fmt.Printf("%s: %s\n", v.Path, v.Message)
		}
		return fmt.Errorf("found %d violation(s)", len(violations))
	}

	fmt.Println("No violations found")
	return nil
}

func printHelp() {
	fmt.Println(`structurelint - Project structure and architecture linter

Usage:
  structurelint [path]         Lint the project at path (default: current directory)
  structurelint --version      Show version information
  structurelint --help         Show this help message

Options:
  -v, --version               Show version information
  -h, --help                  Show help message

Configuration:
  structurelint looks for .structurelint.yml or .structurelint.yaml files
  in the current directory and parent directories.

Examples:
  structurelint                Lint current directory
  structurelint ./src          Lint src directory
  structurelint /path/to/project

Documentation:
  https://github.com/structurelint/structurelint`)
}
