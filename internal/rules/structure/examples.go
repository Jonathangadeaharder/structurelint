package structure

import "github.com/Jonathangadeaharder/structurelint/internal/rules"

func ExampleTestingStrategyRule() rules.Rule {
	unitTestsExist := NewFileExistenceRule(map[string]string{
		"*_test.go": "Unit tests should exist",
	})

	integrationTestsExist := NewFileExistenceRule(map[string]string{
		"tests/integration/*_test.go": "Integration tests should exist",
	})

	return rules.AnyOf(
		"testing-strategy",
		"Project must have either unit tests or integration tests",
		unitTestsExist,
		integrationTestsExist,
	)
}

func ExampleDocumentationCompletenessRule() rules.Rule {
	readmeExists := NewFileExistenceRule(map[string]string{
		"README.md": "README must exist",
	})

	contributingExists := NewFileExistenceRule(map[string]string{
		"CONTRIBUTING.md": "CONTRIBUTING guide must exist",
	})

	licenseExists := NewFileExistenceRule(map[string]string{
		"LICENSE*": "LICENSE file must exist",
	})

	return rules.AllOf(
		"documentation-completeness",
		"Project must have complete documentation",
		readmeExists,
		contributingExists,
		licenseExists,
	)
}

func ExampleArchitectureConsistencyRule() rules.Rule {
	cleanArchRule := NewFileExistenceRule(map[string]string{
		"internal/domain/*":         "Clean architecture: domain layer",
		"internal/application/*":    "Clean architecture: application layer",
		"internal/infrastructure/*": "Clean architecture: infrastructure layer",
	})

	hexagonalRule := NewFileExistenceRule(map[string]string{
		"internal/core/*":    "Hexagonal: core",
		"internal/adapters/*": "Hexagonal: adapters",
		"internal/ports/*":   "Hexagonal: ports",
	})

	return rules.ExactlyOneOf(
		"architecture-consistency",
		"Project must follow exactly one architecture style",
		cleanArchRule,
		hexagonalRule,
	)
}
