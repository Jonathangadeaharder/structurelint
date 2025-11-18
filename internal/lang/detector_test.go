package lang

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetector_DetectGoProject(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte("module example.com/test\n\ngo 1.21\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	detector := NewDetector(tmpDir)

	// Act
	languages, err := detector.Detect()

	// Assert
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	if len(languages) != 1 {
		t.Fatalf("Expected 1 language, got %d", len(languages))
	}

	if languages[0].Language != Go {
		t.Errorf("Expected Go language, got %v", languages[0].Language)
	}

	if languages[0].RootDir != tmpDir {
		t.Errorf("Expected root dir %s, got %s", tmpDir, languages[0].RootDir)
	}
}

func TestDetector_DetectPythonProject(t *testing.T) {
	tests := []struct {
		name         string
		manifestFile string
		content      string
	}{
		{"pyproject.toml", "pyproject.toml", "[tool.poetry]\nname = \"test\"\n"},
		{"setup.py", "setup.py", "from setuptools import setup\nsetup(name='test')\n"},
		{"requirements.txt", "requirements.txt", "flask==2.0.0\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tmpDir := t.TempDir()
			manifestPath := filepath.Join(tmpDir, tt.manifestFile)
			if err := os.WriteFile(manifestPath, []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create %s: %v", tt.manifestFile, err)
			}

			detector := NewDetector(tmpDir)

			// Act
			languages, err := detector.Detect()

			// Assert
			if err != nil {
				t.Fatalf("Detect() failed: %v", err)
			}

			if len(languages) != 1 {
				t.Fatalf("Expected 1 language, got %d", len(languages))
			}

			if languages[0].Language != Python {
				t.Errorf("Expected Python language, got %v", languages[0].Language)
			}
		})
	}
}

func TestDetector_DetectTypeScriptProject(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()

	packageJSON := `{
		"name": "test-project",
		"dependencies": {
			"typescript": "^5.0.0"
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	detector := NewDetector(tmpDir)

	// Act
	languages, err := detector.Detect()

	// Assert
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	if len(languages) != 1 {
		t.Fatalf("Expected 1 language, got %d", len(languages))
	}

	if languages[0].Language != TypeScript {
		t.Errorf("Expected TypeScript language, got %v", languages[0].Language)
	}
}

func TestDetector_DetectReactProject(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()

	packageJSON := `{
		"name": "test-project",
		"dependencies": {
			"react": "^18.0.0",
			"typescript": "^5.0.0"
		}
	}`

	if err := os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	detector := NewDetector(tmpDir)

	// Act
	languages, err := detector.Detect()

	// Assert
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	if len(languages) != 1 {
		t.Fatalf("Expected 1 language, got %d", len(languages))
	}

	if languages[0].Language != TypeScript {
		t.Errorf("Expected TypeScript language, got %v", languages[0].Language)
	}

	if len(languages[0].SubLanguages) != 1 || languages[0].SubLanguages[0] != React {
		t.Errorf("Expected React in SubLanguages, got %v", languages[0].SubLanguages)
	}
}

func TestDetector_DetectMultipleLanguages(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()

	// Create Go project in root
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create Python project in subdirectory
	pythonDir := filepath.Join(tmpDir, "scripts")
	if err := os.MkdirAll(pythonDir, 0755); err != nil {
		t.Fatalf("Failed to create scripts dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pythonDir, "requirements.txt"), []byte("flask==2.0.0\n"), 0644); err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	// Create TypeScript project in subdirectory
	frontendDir := filepath.Join(tmpDir, "frontend")
	if err := os.MkdirAll(frontendDir, 0755); err != nil {
		t.Fatalf("Failed to create frontend dir: %v", err)
	}
	packageJSON := `{"name": "frontend", "devDependencies": {"typescript": "^5.0.0"}}`
	if err := os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte(packageJSON), 0644); err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	detector := NewDetector(tmpDir)

	// Act
	languages, err := detector.Detect()

	// Assert
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	if len(languages) != 3 {
		t.Fatalf("Expected 3 languages, got %d", len(languages))
	}

	// Check that we have all three languages
	foundGo := false
	foundPython := false
	foundTypeScript := false

	for _, lang := range languages {
		switch lang.Language {
		case Go:
			foundGo = true
		case Python:
			foundPython = true
		case TypeScript:
			foundTypeScript = true
		}
	}

	if !foundGo {
		t.Error("Go language not detected")
	}
	if !foundPython {
		t.Error("Python language not detected")
	}
	if !foundTypeScript {
		t.Error("TypeScript language not detected")
	}
}

func TestDetector_SkipsCommonDirectories(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()

	// Create main Go project
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644); err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Create node_modules with package.json (should be ignored)
	nodeModules := filepath.Join(tmpDir, "node_modules", "some-package")
	if err := os.MkdirAll(nodeModules, 0755); err != nil {
		t.Fatalf("Failed to create node_modules: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nodeModules, "package.json"), []byte(`{"name": "pkg"}`), 0644); err != nil {
		t.Fatalf("Failed to create package.json in node_modules: %v", err)
	}

	detector := NewDetector(tmpDir)

	// Act
	languages, err := detector.Detect()

	// Assert
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	// Should only detect Go, not the JavaScript in node_modules
	if len(languages) != 1 {
		t.Fatalf("Expected 1 language (Go only), got %d", len(languages))
	}

	if languages[0].Language != Go {
		t.Errorf("Expected only Go language, got %v", languages[0].Language)
	}
}

func TestLanguage_DefaultNamingConvention(t *testing.T) {
	tests := []struct {
		language   Language
		wantConvention string
	}{
		{Go, "snake_case"},
		{Python, "snake_case"},
		{TypeScript, "camelCase"},
		{JavaScript, "camelCase"},
		{React, "PascalCase"},
		{Rust, "snake_case"},
		{Java, "PascalCase"},
		{CSharp, "PascalCase"},
		{Ruby, "snake_case"},
	}

	for _, tt := range tests {
		t.Run(tt.language.String(), func(t *testing.T) {
			got := tt.language.DefaultNamingConvention()
			if got != tt.wantConvention {
				t.Errorf("DefaultNamingConvention() = %v, want %v", got, tt.wantConvention)
			}
		})
	}
}

func TestDetector_DetectRustProject(t *testing.T) {
	// Arrange
	tmpDir := t.TempDir()
	cargoToml := `[package]
name = "test-project"
version = "0.1.0"
`
	if err := os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(cargoToml), 0644); err != nil {
		t.Fatalf("Failed to create Cargo.toml: %v", err)
	}

	detector := NewDetector(tmpDir)

	// Act
	languages, err := detector.Detect()

	// Assert
	if err != nil {
		t.Fatalf("Detect() failed: %v", err)
	}

	if len(languages) != 1 {
		t.Fatalf("Expected 1 language, got %d", len(languages))
	}

	if languages[0].Language != Rust {
		t.Errorf("Expected Rust language, got %v", languages[0].Language)
	}
}

func TestDetector_DetectJavaProject(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{"Maven", "pom.xml", "<project></project>"},
		{"Gradle", "build.gradle", "plugins { id 'java' }"},
		{"Gradle Kotlin", "build.gradle.kts", "plugins { kotlin(\"jvm\") version \"1.8.0\" }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tmpDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(tmpDir, tt.filename), []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create %s: %v", tt.filename, err)
			}

			detector := NewDetector(tmpDir)

			// Act
			languages, err := detector.Detect()

			// Assert
			if err != nil {
				t.Fatalf("Detect() failed: %v", err)
			}

			if len(languages) != 1 {
				t.Fatalf("Expected 1 language, got %d", len(languages))
			}

			if languages[0].Language != Java {
				t.Errorf("Expected Java language, got %v", languages[0].Language)
			}
		})
	}
}

func TestDetector_DetectCSharpProject(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		content  string
	}{
		{"csproj", "MyApp.csproj", "<Project Sdk=\"Microsoft.NET.Sdk\"></Project>"},
		{"solution", "MyApp.sln", "Microsoft Visual Studio Solution File"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			tmpDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(tmpDir, tt.filename), []byte(tt.content), 0644); err != nil {
				t.Fatalf("Failed to create %s: %v", tt.filename, err)
			}

			detector := NewDetector(tmpDir)

			// Act
			languages, err := detector.Detect()

			// Assert
			if err != nil {
				t.Fatalf("Detect() failed: %v", err)
			}

			if len(languages) != 1 {
				t.Fatalf("Expected 1 language, got %d", len(languages))
			}

			if languages[0].Language != CSharp {
				t.Errorf("Expected C# language, got %v", languages[0].Language)
			}
		})
	}
}
