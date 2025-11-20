package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/structurelint/structurelint/internal/scaffold"
)

func runScaffold(args []string) error {
	fs := flag.NewFlagSet("scaffold", flag.ExitOnError)
	lang := fs.String("lang", "", "Target language (go, typescript, python, java)")
	includeTests := fs.Bool("tests", true, "Include test files")
	listFlag := fs.Bool("list", false, "List available templates")

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: structurelint scaffold [options] <type> <name>\n\n")
		fmt.Fprintf(os.Stderr, "Generate boilerplate code from templates.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nTypes:\n")
		fmt.Fprintf(os.Stderr, "  service      Business logic layer\n")
		fmt.Fprintf(os.Stderr, "  repository   Data access layer\n")
		fmt.Fprintf(os.Stderr, "  controller   REST API controller (TypeScript)\n")
		fmt.Fprintf(os.Stderr, "  handler      HTTP handler (Go)\n")
		fmt.Fprintf(os.Stderr, "  model        Domain model/entity\n")
		fmt.Fprintf(os.Stderr, "\nLanguages:\n")
		fmt.Fprintf(os.Stderr, "  go           Go\n")
		fmt.Fprintf(os.Stderr, "  typescript   TypeScript\n")
		fmt.Fprintf(os.Stderr, "  python       Python\n")
		fmt.Fprintf(os.Stderr, "  java         Java (coming soon)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  structurelint scaffold service UserService --lang go\n")
		fmt.Fprintf(os.Stderr, "  structurelint scaffold repository UserRepo --lang typescript\n")
		fmt.Fprintf(os.Stderr, "  structurelint scaffold model User --lang python --tests=false\n")
		fmt.Fprintf(os.Stderr, "  structurelint scaffold --list\n")
	}

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Get working directory
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	// Create generator
	gen := scaffold.NewGenerator(workDir)

	// Handle list flag
	if *listFlag {
		return listTemplates(gen)
	}

	// Validate arguments
	if fs.NArg() < 2 {
		fs.Usage()
		return fmt.Errorf("missing required arguments: <type> <name>")
	}

	typeName := fs.Arg(0)
	componentName := fs.Arg(1)

	// Detect language if not specified
	if *lang == "" {
		detected := detectLanguage(workDir)
		if detected == "" {
			return fmt.Errorf("could not detect language, please specify with --lang")
		}
		*lang = detected
		fmt.Printf("ℹ Detected language: %s\n", *lang)
	}

	// Validate language
	language, err := parseLanguage(*lang)
	if err != nil {
		return err
	}

	// Validate type
	templateType, err := parseTemplateType(typeName)
	if err != nil {
		return err
	}

	// Get template (validate it exists)
	if _, err := gen.GetTemplate(language, templateType); err != nil {
		return fmt.Errorf("template not found: %s-%s (use --list to see available templates)", *lang, typeName)
	}

	// Prepare variables
	vars := scaffold.Variables{
		IncludeTests: *includeTests,
	}

	// Generate code
	fmt.Printf("Generating %s %s...\n", language, templateType)
	templateKey := string(language) + "-" + string(templateType)
	if err := gen.Generate(templateKey, componentName, vars); err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	fmt.Printf("\n✓ Successfully generated %s\n", componentName)
	fmt.Println("\nNext steps:")
	fmt.Println("  1. Review the generated files")
	fmt.Println("  2. Implement the TODO sections")
	fmt.Println("  3. Run tests to verify functionality")

	return nil
}

func listTemplates(gen *scaffold.Generator) error {
	fmt.Println("Available Templates:")
	fmt.Println()

	// Group by language
	byLang := make(map[scaffold.Language][]*scaffold.Template)
	for _, tmpl := range gen.ListTemplates() {
		byLang[tmpl.Language] = append(byLang[tmpl.Language], tmpl)
	}

	// Print grouped templates
	for lang, templates := range byLang {
		fmt.Printf("%s:\n", strings.ToUpper(string(lang)))
		for _, tmpl := range templates {
			fmt.Printf("  %-15s %s\n", tmpl.Type, tmpl.Description)
		}
		fmt.Println()
	}

	fmt.Println("Usage: structurelint scaffold <type> <name> --lang <language>")

	return nil
}

func detectLanguage(dir string) string {
	// Check for language-specific files
	if _, err := os.Stat(dir + "/go.mod"); err == nil {
		return "go"
	}
	if _, err := os.Stat(dir + "/package.json"); err == nil {
		// Check if it's TypeScript
		if _, err := os.Stat(dir + "/tsconfig.json"); err == nil {
			return "typescript"
		}
		return "typescript" // Default to TypeScript for Node projects
	}
	if _, err := os.Stat(dir + "/requirements.txt"); err == nil {
		return "python"
	}
	if _, err := os.Stat(dir + "/setup.py"); err == nil {
		return "python"
	}
	if _, err := os.Stat(dir + "/pom.xml"); err == nil {
		return "java"
	}
	if _, err := os.Stat(dir + "/build.gradle"); err == nil {
		return "java"
	}

	return ""
}

func parseLanguage(lang string) (scaffold.Language, error) {
	switch strings.ToLower(lang) {
	case "go", "golang":
		return scaffold.LangGo, nil
	case "typescript", "ts":
		return scaffold.LangTypeScript, nil
	case "python", "py":
		return scaffold.LangPython, nil
	case "java":
		return scaffold.LangJava, nil
	default:
		return "", fmt.Errorf("unsupported language: %s", lang)
	}
}

func parseTemplateType(typ string) (scaffold.TemplateType, error) {
	switch strings.ToLower(typ) {
	case "service":
		return scaffold.TypeService, nil
	case "repository", "repo":
		return scaffold.TypeRepository, nil
	case "controller":
		return scaffold.TypeController, nil
	case "handler":
		return scaffold.TypeHandler, nil
	case "model", "entity":
		return scaffold.TypeModel, nil
	case "middleware":
		return scaffold.TypeMiddleware, nil
	case "test":
		return scaffold.TypeTest, nil
	default:
		return "", fmt.Errorf("unsupported template type: %s", typ)
	}
}

func printScaffoldHelp() {
	fmt.Println(`structurelint scaffold - Code generation from templates

Usage:
  structurelint scaffold [options] <type> <name>

Description:
  Generate boilerplate code from templates. Scaffold creates code that follows
  your project's architectural patterns and conventions automatically.

Arguments:
  type         Type of component to generate (service, repository, controller, etc.)
  name         Name of the component (e.g., UserService, OrderRepository)

Options:
  --lang <language>    Target language (go, typescript, python, java)
  --tests              Include test files (default: true)
  --list               List all available templates

Supported Languages:
  go           Go (services, repositories, handlers, models)
  typescript   TypeScript (services, controllers, models)
  python       Python (services, repositories, models)
  java         Java (coming soon)

Template Types:
  service      Business logic layer with CRUD operations
  repository   Data access layer with database operations
  controller   REST API controller (TypeScript/Java)
  handler      HTTP handler (Go)
  model        Domain model/entity with validation

Examples:
  structurelint scaffold service UserService --lang go
    Generate a Go service in internal/services/user_service.go

  structurelint scaffold repository OrderRepo --lang typescript
    Generate a TypeScript repository in src/repositories/order-repo.ts

  structurelint scaffold controller ProductController --lang typescript
    Generate a TypeScript controller in src/controllers/product-controller.ts

  structurelint scaffold model User --lang python --tests=false
    Generate a Python model in models/user.py without tests

  structurelint scaffold --list
    List all available templates

Features:
  - Automatic language detection from project files
  - Smart naming conventions (PascalCase, camelCase, snake_case, kebab-case)
  - Package/module detection from project configuration
  - Test file generation
  - TODO comments for implementation guidance
  - Follows best practices for each language

File Placement:
  Go:         internal/<type>/<name_snake>.go
  TypeScript: src/<type>s/<name-kebab>.<type>.ts
  Python:     <type>s/<name_snake>.py

Generated Code Structure:
  - Service:     CRUD operations, business logic layer
  - Repository:  Data access interface and implementation
  - Controller:  REST API endpoints with HTTP handling
  - Handler:     HTTP request handlers
  - Model:       Domain entities with validation

Tips:
  - Use PascalCase for component names (UserService, not userService)
  - Generated code includes TODO comments for implementation
  - Review and customize generated code before committing
  - Run tests after implementing TODOs
  - Use --tests=false to skip test file generation

Integration with Linter:
  Generated code follows structurelint rules automatically:
  - Correct file placement based on detected patterns
  - Proper naming conventions for the language
  - Architectural layer separation
  - Test file co-location

Documentation:
  https://github.com/structurelint/structurelint/docs/scaffold`)
}
