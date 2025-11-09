package main

import (
	"fmt"
	"os"

	"github.com/structurelint/structurelint/internal/linter"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	// Parse command line arguments
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
