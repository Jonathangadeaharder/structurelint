package rules

func init() {
	// Structure rules are registered in internal/rules/structure/init.go
	// Graph-dependent rules are registered in internal/rules/graph/init.go
	// CI rules are registered in internal/rules/ci/init.go
	// These registrations are triggered by blank imports in internal/linter/factory.go
}
