package rules

import (
	"testing"
)

func TestInfrastructureProfile_GitHubActions(t *testing.T) {
	// Arrange
	profile := NewInfrastructureProfile(nil)
	tests := []struct {
		path     string
		expected bool
	}{
		{".github/workflows/ci.yml", true},
		{".github/workflows/deploy.yaml", true},
		{"src/main.go", false},
		{"pkg/utils/helper.go", false},
	}

	for _, tt := range tests {
		// Act
		result := profile.IsInfrastructure(tt.path)

		// Assert
		if result != tt.expected {
			t.Errorf("IsInfrastructure(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestInfrastructureProfile_Docker(t *testing.T) {
	// Arrange
	profile := NewInfrastructureProfile(nil)
	tests := []struct {
		path     string
		expected bool
	}{
		{"docker/Dockerfile", true},
		{"docker/app/Dockerfile", true},
		{"src/main.go", false},      // Regular source code
		{"pkg/docker_client.go", false}, // Has "docker" in name but not infrastructure
	}

	for _, tt := range tests {
		// Act
		result := profile.IsInfrastructure(tt.path)

		// Assert
		if result != tt.expected {
			t.Errorf("IsInfrastructure(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestInfrastructureProfile_Kubernetes(t *testing.T) {
	// Arrange
	profile := NewInfrastructureProfile(nil)
	tests := []struct {
		path     string
		expected bool
	}{
		{"k8s/deployment.yaml", true},
		{"k8s/service.yaml", true},
		{"kubernetes/ingress.yaml", true},
		{"src/k8s_client.go", false},
	}

	for _, tt := range tests {
		// Act
		result := profile.IsInfrastructure(tt.path)

		// Assert
		if result != tt.expected {
			t.Errorf("IsInfrastructure(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestInfrastructureProfile_Terraform(t *testing.T) {
	// Arrange
	profile := NewInfrastructureProfile(nil)
	tests := []struct {
		path     string
		expected bool
	}{
		{"terraform/main.tf", true},
		{"terraform/variables.tf", true},
		{"terraform/modules/vpc/main.tf", true},
		{"src/terraform_helper.go", false},
	}

	for _, tt := range tests {
		// Act
		result := profile.IsInfrastructure(tt.path)

		// Assert
		if result != tt.expected {
			t.Errorf("IsInfrastructure(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestInfrastructureProfile_Scripts(t *testing.T) {
	// Arrange
	profile := NewInfrastructureProfile(nil)
	tests := []struct {
		path     string
		expected bool
	}{
		{"scripts/deploy.sh", true},
		{"scripts/build.py", true},
		{"scripts/utils/setup.js", true},
		{"src/script_runner.go", false},
	}

	for _, tt := range tests {
		// Act
		result := profile.IsInfrastructure(tt.path)

		// Assert
		if result != tt.expected {
			t.Errorf("IsInfrastructure(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestInfrastructureProfile_UserPatterns(t *testing.T) {
	// Arrange
	userPatterns := []string{
		"deployment/**",
		"config/**",
	}
	profile := NewInfrastructureProfile(userPatterns)
	tests := []struct {
		path     string
		expected bool
	}{
		{"deployment/app.yaml", true},
		{"config/database.yml", true},
		{".github/workflows/ci.yml", true}, // Default pattern still works
		{"src/main.go", false},
	}

	for _, tt := range tests {
		// Act
		result := profile.IsInfrastructure(tt.path)

		// Assert
		if result != tt.expected {
			t.Errorf("IsInfrastructure(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestInfrastructureProfile_ShouldExemptFromRule(t *testing.T) {
	// Arrange
	profile := NewInfrastructureProfile(nil)
	tests := []struct {
		path     string
		rule     string
		expected bool
	}{
		{".github/workflows/ci.yml", "max-cognitive-complexity", true},
		{".github/workflows/ci.yml", "max-halstead-effort", true},
		{".github/workflows/ci.yml", "test-adjacency", true},
		{".github/workflows/ci.yml", "disallow-unused-exports", true},
		{".github/workflows/ci.yml", "max-depth", false}, // Not exempted
		{"src/main.go", "max-cognitive-complexity", false}, // Not infrastructure
	}

	for _, tt := range tests {
		// Act
		result := profile.ShouldExemptFromRule(tt.path, tt.rule)

		// Assert
		if result != tt.expected {
			t.Errorf("ShouldExemptFromRule(%q, %q) = %v, want %v", tt.path, tt.rule, result, tt.expected)
		}
	}
}

func TestInfrastructureProfile_GitLabCI(t *testing.T) {
	// Arrange
	profile := NewInfrastructureProfile(nil)
	tests := []struct {
		path     string
		expected bool
	}{
		{".gitlab/.gitlab-ci.yml", true},
		{".gitlab/ci/templates/build.yml", true},
		{"src/gitlab.go", false},
	}

	for _, tt := range tests {
		// Act
		result := profile.IsInfrastructure(tt.path)

		// Assert
		if result != tt.expected {
			t.Errorf("IsInfrastructure(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestInfrastructureProfile_Ansible(t *testing.T) {
	// Arrange
	profile := NewInfrastructureProfile(nil)
	tests := []struct {
		path     string
		expected bool
	}{
		{"ansible/playbook.yml", true},
		{"ansible/roles/webserver/tasks/main.yml", true},
		{"src/ansible_runner.go", false},
	}

	for _, tt := range tests {
		// Act
		result := profile.IsInfrastructure(tt.path)

		// Assert
		if result != tt.expected {
			t.Errorf("IsInfrastructure(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestInfrastructureProfile_Helm(t *testing.T) {
	// Arrange
	profile := NewInfrastructureProfile(nil)
	tests := []struct {
		path     string
		expected bool
	}{
		{"helm/charts/myapp/values.yaml", true},
		{"helm/templates/deployment.yaml", true},
		{"src/helm_client.go", false},
	}

	for _, tt := range tests {
		// Act
		result := profile.IsInfrastructure(tt.path)

		// Assert
		if result != tt.expected {
			t.Errorf("IsInfrastructure(%q) = %v, want %v", tt.path, result, tt.expected)
		}
	}
}

func TestMatchesGlobPattern_RecursiveWildcard(t *testing.T) {
	// Arrange
	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{".github/workflows/ci.yml", ".github/**", true},
		{"k8s/prod/deployment.yaml", "k8s/**", true},
		{"src/main.go", ".github/**", false},
		{"docker/app/Dockerfile", "docker/**", true},
	}

	for _, tt := range tests {
		// Act
		result := matchesGlobPattern(tt.path, tt.pattern)

		// Assert
		if result != tt.expected {
			t.Errorf("matchesGlobPattern(%q, %q) = %v, want %v", tt.path, tt.pattern, result, tt.expected)
		}
	}
}
