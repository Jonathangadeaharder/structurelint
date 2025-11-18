package rules

import (
	"regexp"
	"testing"

	"github.com/structurelint/structurelint/internal/walker"
)

func TestPathBasedLayerRule_ThreeLayerArchitecture(t *testing.T) {
	// Arrange - Classic 3-layer architecture
	layers := []PathLayer{
		{
			Name:     "presentation",
			Patterns: []string{"src/presentation/**"},
			CanDependOn: []string{"business"},
			ForbiddenPaths: []string{"**/data/**", "**/repositories/**"},
		},
		{
			Name:     "business",
			Patterns: []string{"src/business/**"},
			CanDependOn: []string{"data"},
			ForbiddenPaths: []string{},
		},
		{
			Name:     "data",
			Patterns: []string{"src/data/**"},
			CanDependOn: []string{},
			ForbiddenPaths: []string{"**/presentation/**", "**/business/**"},
		},
	}

	files := []walker.FileInfo{
		{Path: "src/presentation/controllers/user_controller.py", IsDir: false},
		{Path: "src/business/services/user_service.py", IsDir: false},
		{Path: "src/data/repositories/user_repository.py", IsDir: false},
		{Path: "src/presentation/data/cache.py", IsDir: false}, // VIOLATION: presentation has data path
	}

	rule := NewPathBasedLayerRule(layers)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}

	if len(violations) > 0 {
		if violations[0].Path != "src/presentation/data/cache.py" {
			t.Errorf("Expected violation for presentation/data/cache.py, got %s", violations[0].Path)
		}
	}
}

func TestPathBasedLayerRule_CleanArchitecture(t *testing.T) {
	// Arrange - Clean Architecture (Uncle Bob)
	layers := []PathLayer{
		{
			Name:     "entities",
			Patterns: []string{"src/domain/entities/**"},
			CanDependOn: []string{},
			ForbiddenPaths: []string{"**/usecases/**", "**/adapters/**", "**/frameworks/**"},
		},
		{
			Name:     "usecases",
			Patterns: []string{"src/domain/usecases/**"},
			CanDependOn: []string{"entities"},
			ForbiddenPaths: []string{"**/adapters/**", "**/frameworks/**"},
		},
		{
			Name:     "adapters",
			Patterns: []string{"src/adapters/**"},
			CanDependOn: []string{"usecases", "entities"},
			ForbiddenPaths: []string{"**/frameworks/**"},
		},
		{
			Name:     "frameworks",
			Patterns: []string{"src/frameworks/**"},
			CanDependOn: []string{"adapters", "usecases", "entities"},
			ForbiddenPaths: []string{},
		},
	}

	files := []walker.FileInfo{
		{Path: "src/domain/entities/user.py", IsDir: false},
		{Path: "src/domain/usecases/create_user.py", IsDir: false},
		{Path: "src/adapters/repositories/user_repository.py", IsDir: false},
		{Path: "src/frameworks/web/fastapi_app.py", IsDir: false},
		{Path: "src/domain/entities/frameworks/helper.py", IsDir: false}, // VIOLATION: entities has frameworks path
	}

	rule := NewPathBasedLayerRule(layers)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (entities with frameworks path), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestPathBasedLayerRule_HexagonalArchitecture(t *testing.T) {
	// Arrange - Hexagonal/Ports & Adapters
	layers := []PathLayer{
		{
			Name:     "core",
			Patterns: []string{"src/core/**"},
			CanDependOn: []string{},
			ForbiddenPaths: []string{"**/ports/**", "**/adapters/**"},
		},
		{
			Name:     "ports",
			Patterns: []string{"src/ports/**"},
			CanDependOn: []string{"core"},
			ForbiddenPaths: []string{"**/adapters/**"},
		},
		{
			Name:     "adapters",
			Patterns: []string{"src/adapters/**"},
			CanDependOn: []string{"ports", "core"},
			ForbiddenPaths: []string{},
		},
	}

	files := []walker.FileInfo{
		{Path: "src/core/domain/user.py", IsDir: false},
		{Path: "src/ports/repositories/user_port.py", IsDir: false},
		{Path: "src/adapters/postgres/user_adapter.py", IsDir: false},
		{Path: "src/core/adapters/helper.py", IsDir: false}, // VIOLATION: core has adapters path
	}

	rule := NewPathBasedLayerRule(layers)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (core with adapters path), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestPathBasedLayerRule_MVCArchitecture(t *testing.T) {
	// Arrange - MVC pattern
	layers := []PathLayer{
		{
			Name:     "views",
			Patterns: []string{"src/views/**", "templates/**"},
			CanDependOn: []string{"controllers"},
			ForbiddenPaths: []string{"**/models/**"},
		},
		{
			Name:     "controllers",
			Patterns: []string{"src/controllers/**"},
			CanDependOn: []string{"models"},
			ForbiddenPaths: []string{},
		},
		{
			Name:     "models",
			Patterns: []string{"src/models/**"},
			CanDependOn: []string{},
			ForbiddenPaths: []string{"**/views/**", "**/controllers/**"},
		},
	}

	files := []walker.FileInfo{
		{Path: "src/views/user_view.py", IsDir: false},
		{Path: "src/controllers/user_controller.py", IsDir: false},
		{Path: "src/models/user.py", IsDir: false},
		{Path: "src/models/controllers/db_controller.py", IsDir: false}, // VIOLATION: models has controllers path
	}

	rule := NewPathBasedLayerRule(layers)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (models with controllers path), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestPathBasedLayerRule_NoViolations(t *testing.T) {
	// Arrange
	layers := []PathLayer{
		{
			Name:     "presentation",
			Patterns: []string{"src/presentation/**"},
			CanDependOn: []string{"business"},
			ForbiddenPaths: []string{"**/data/**"},
		},
		{
			Name:     "business",
			Patterns: []string{"src/business/**"},
			CanDependOn: []string{"data"},
			ForbiddenPaths: []string{},
		},
		{
			Name:     "data",
			Patterns: []string{"src/data/**"},
			CanDependOn: []string{},
			ForbiddenPaths: []string{},
		},
	}

	// All files in correct layers
	files := []walker.FileInfo{
		{Path: "src/presentation/controllers/user_controller.py", IsDir: false},
		{Path: "src/business/services/user_service.py", IsDir: false},
		{Path: "src/data/repositories/user_repository.py", IsDir: false},
	}

	rule := NewPathBasedLayerRule(layers)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 0 {
		t.Errorf("Expected no violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Unexpected violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestPathBasedLayerRule_MultipleViolations(t *testing.T) {
	// Arrange
	layers := []PathLayer{
		{
			Name:     "presentation",
			Patterns: []string{"src/presentation/**"},
			CanDependOn: []string{"business"},
			ForbiddenPaths: []string{"**/data/**", "**/db/**"},
		},
	}

	files := []walker.FileInfo{
		{Path: "src/presentation/controllers/user_controller.py", IsDir: false},
		{Path: "src/presentation/data/cache.py", IsDir: false},     // VIOLATION 1
		{Path: "src/presentation/db/connection.py", IsDir: false},  // VIOLATION 2
	}

	rule := NewPathBasedLayerRule(layers)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 2 {
		t.Errorf("Expected 2 violations, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestGlobToRegex_Wildcards(t *testing.T) {
	tests := []struct {
		pattern string
		path    string
		matches bool
	}{
		{"src/**", "src/foo/bar.py", true},
		{"src/**", "src/foo.py", true},
		{"src/**", "other/foo.py", false},
		{"**/test/**", "src/test/foo.py", true},
		{"**/test/**", "test/foo.py", true},
		{"**/test/**", "src/tests/foo.py", false}, // "tests" != "test"
		{"*.py", "foo.py", true},
		{"*.py", "foo/bar.py", false}, // * doesn't match /
		{"src/*.py", "src/foo.py", true},
		{"src/*.py", "src/foo/bar.py", false},
	}

	for _, tt := range tests {
		regex := globToRegex(tt.pattern)
		matched, err := regexp.Compile(regex)
		if err != nil {
			t.Fatalf("Failed to compile regex for pattern %s: %v", tt.pattern, err)
		}

		result := matched.MatchString(tt.path)
		if result != tt.matches {
			t.Errorf("Pattern %s against %s: got %v, want %v (regex: %s)",
				tt.pattern, tt.path, result, tt.matches, regex)
		}
	}
}

func TestPathBasedLayerRule_RegexPatterns(t *testing.T) {
	// Arrange - Using regex patterns
	layers := []PathLayer{
		{
			Name:     "api",
			Patterns: []string{"src/api/**"},
			CanDependOn: []string{},
			ForbiddenPaths: []string{"**/internal/**"},
		},
	}

	files := []walker.FileInfo{
		{Path: "src/api/routes/users.py", IsDir: false},
		{Path: "src/api/internal/helper.py", IsDir: false}, // VIOLATION
	}

	rule := NewPathBasedLayerRule(layers)

	// Act
	violations := rule.Check(files, nil)

	// Assert
	if len(violations) != 1 {
		t.Errorf("Expected 1 violation (api with internal path), got %d", len(violations))
		for _, v := range violations {
			t.Logf("Violation: %s - %s", v.Path, v.Message)
		}
	}
}

func TestPathBasedLayerRule_FilesOutsideLayers(t *testing.T) {
	// Arrange
	layers := []PathLayer{
		{
			Name:     "core",
			Patterns: []string{"src/core/**"},
			CanDependOn: []string{},
			ForbiddenPaths: []string{},
		},
	}

	files := []walker.FileInfo{
		{Path: "src/core/domain.py", IsDir: false},
		{Path: "src/utils/helper.py", IsDir: false},      // Not in any layer
		{Path: "scripts/deploy.py", IsDir: false},         // Not in any layer
	}

	rule := NewPathBasedLayerRule(layers)

	// Act
	violations := rule.Check(files, nil)

	// Assert - Files outside layers should not trigger violations
	if len(violations) != 0 {
		t.Errorf("Expected no violations for files outside layers, got %d", len(violations))
		for _, v := range violations {
			t.Logf("Unexpected violation: %s - %s", v.Path, v.Message)
		}
	}
}
