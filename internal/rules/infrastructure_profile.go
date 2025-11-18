package rules

import (
	"github.com/structurelint/structurelint/internal/walker"
)

// InfrastructureProfile identifies infrastructure code directories
// These directories typically contain declarative configs that need different validation
type InfrastructureProfile struct {
	// Default infrastructure patterns
	InfraPatterns []string

	// User-defined infrastructure patterns
	UserPatterns []string
}

// DefaultInfrastructurePatterns returns common infrastructure directory patterns
var DefaultInfrastructurePatterns = []string{
	".github/**",     // GitHub Actions workflows
	".gitlab/**",     // GitLab CI configs
	".circleci/**",   // CircleCI configs
	"docker/**",      // Docker configs
	"k8s/**",         // Kubernetes manifests
	"kubernetes/**",  // Kubernetes manifests
	"terraform/**",   // Terraform configs
	"ansible/**",     // Ansible playbooks
	"helm/**",        // Helm charts
	"scripts/**",     // Build/deploy scripts
	"ci/**",          // CI configs
	"cd/**",          // CD configs
	"infrastructure/**", // General infrastructure
}

// NewInfrastructureProfile creates a new infrastructure profile
func NewInfrastructureProfile(userPatterns []string) *InfrastructureProfile {
	return &InfrastructureProfile{
		InfraPatterns: DefaultInfrastructurePatterns,
		UserPatterns:  userPatterns,
	}
}

// IsInfrastructure checks if a file path is in an infrastructure directory
func (p *InfrastructureProfile) IsInfrastructure(path string) bool {
	// Check default patterns
	for _, pattern := range p.InfraPatterns {
		if matchesGlobPattern(path, pattern) {
			return true
		}
	}

	// Check user patterns
	for _, pattern := range p.UserPatterns {
		if matchesGlobPattern(path, pattern) {
			return true
		}
	}

	return false
}

// FilterInfrastructureFiles removes infrastructure files from a list
func (p *InfrastructureProfile) FilterInfrastructureFiles(files []interface{}) []interface{} {
	var filtered []interface{}
	for _, file := range files {
		// Support both FileInfo and string paths
		var path string
		switch f := file.(type) {
		case *walker.FileInfo:
			path = f.Path
		case string:
			path = f
		default:
			// Log a warning or handle unexpected types if necessary
			continue
		}

		if !p.IsInfrastructure(path) {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// ShouldExemptFromRule checks if a rule should be exempted for infrastructure files
// Some rules don't make sense for declarative configs (YAML, Dockerfiles, etc.)
func (p *InfrastructureProfile) ShouldExemptFromRule(path, ruleName string) bool {
	if !p.IsInfrastructure(path) {
		return false
	}

	// Rules that should be exempted for infrastructure code
	exemptedRules := map[string]bool{
		"max-cognitive-complexity": true, // Declarative configs don't have complexity
		"max-halstead-effort":      true, // Declarative configs don't have Halstead metrics
		"test-adjacency":           true, // Infrastructure doesn't need test adjacency
		"disallow-unused-exports":  true, // Config files don't have exports
	}

	return exemptedRules[ruleName]
}
