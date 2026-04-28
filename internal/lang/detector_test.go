package lang

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDetector_DetectGoProject(t *testing.T) {
	tmpDir := t.TempDir()
	goModPath := filepath.Join(tmpDir, "go.mod")
	require.NoError(t, os.WriteFile(goModPath, []byte("module example.com/test\n\ngo 1.21\n"), 0644))

	detector := NewDetector(tmpDir)

	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, Go, languages[0].Language)
	assert.Equal(t, tmpDir, languages[0].RootDir)
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
			tmpDir := t.TempDir()
			manifestPath := filepath.Join(tmpDir, tt.manifestFile)
			require.NoError(t, os.WriteFile(manifestPath, []byte(tt.content), 0644))

			detector := NewDetector(tmpDir)

			languages, err := detector.Detect()

			require.NoError(t, err)
			require.Len(t, languages, 1)
			assert.Equal(t, Python, languages[0].Language)
		})
	}
}

func TestDetector_DetectTypeScriptProject(t *testing.T) {
	tmpDir := t.TempDir()

	packageJSON := `{
		"name": "test-project",
		"dependencies": {
			"typescript": "^5.0.0"
		}
	}`

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644))

	detector := NewDetector(tmpDir)

	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, TypeScript, languages[0].Language)
}

func TestDetector_DetectReactProject(t *testing.T) {
	tmpDir := t.TempDir()

	packageJSON := `{
		"name": "test-project",
		"dependencies": {
			"react": "^18.0.0",
			"typescript": "^5.0.0"
		}
	}`

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "package.json"), []byte(packageJSON), 0644))

	detector := NewDetector(tmpDir)

	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, TypeScript, languages[0].Language)
	require.Len(t, languages[0].SubLanguages, 1)
	assert.Equal(t, React, languages[0].SubLanguages[0])
}

func TestDetector_DetectMultipleLanguages(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644))

	pythonDir := filepath.Join(tmpDir, "scripts")
	require.NoError(t, os.MkdirAll(pythonDir, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(pythonDir, "requirements.txt"), []byte("flask==2.0.0\n"), 0644))

	frontendDir := filepath.Join(tmpDir, "frontend")
	require.NoError(t, os.MkdirAll(frontendDir, 0755))
	packageJSON := `{"name": "frontend", "devDependencies": {"typescript": "^5.0.0"}}`
	require.NoError(t, os.WriteFile(filepath.Join(frontendDir, "package.json"), []byte(packageJSON), 0644))

	detector := NewDetector(tmpDir)

	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 3)

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

	assert.True(t, foundGo, "Go language not detected")
	assert.True(t, foundPython, "Python language not detected")
	assert.True(t, foundTypeScript, "TypeScript language not detected")
}

func TestDetector_SkipsCommonDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte("module test\n"), 0644))

	nodeModules := filepath.Join(tmpDir, "node_modules", "some-package")
	require.NoError(t, os.MkdirAll(nodeModules, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(nodeModules, "package.json"), []byte(`{"name": "pkg"}`), 0644))

	detector := NewDetector(tmpDir)

	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, Go, languages[0].Language)
}

func TestLanguage_DefaultNamingConvention(t *testing.T) {
	tests := []struct {
		language        Language
		wantConvention  string
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
			assert.Equal(t, tt.wantConvention, tt.language.DefaultNamingConvention())
		})
	}
}

func TestDetector_DetectRustProject(t *testing.T) {
	tmpDir := t.TempDir()
	cargoToml := `[package]
name = "test-project"
version = "0.1.0"
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Cargo.toml"), []byte(cargoToml), 0644))

	detector := NewDetector(tmpDir)

	languages, err := detector.Detect()

	require.NoError(t, err)
	require.Len(t, languages, 1)
	assert.Equal(t, Rust, languages[0].Language)
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
			tmpDir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, tt.filename), []byte(tt.content), 0644))

			detector := NewDetector(tmpDir)

			languages, err := detector.Detect()

			require.NoError(t, err)
			require.Len(t, languages, 1)
			assert.Equal(t, Java, languages[0].Language)
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
			tmpDir := t.TempDir()
			require.NoError(t, os.WriteFile(filepath.Join(tmpDir, tt.filename), []byte(tt.content), 0644))

			detector := NewDetector(tmpDir)

			languages, err := detector.Detect()

			require.NoError(t, err)
			require.Len(t, languages, 1)
			assert.Equal(t, CSharp, languages[0].Language)
		})
	}
}
