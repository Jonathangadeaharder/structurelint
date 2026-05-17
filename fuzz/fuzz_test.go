package fuzz

import (
	"strings"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

func FuzzMatchesPattern(f *testing.F) {
	corpus := []struct{ path, pattern string }{
		{"src/main.go", "*.go"},
		{"src/main.go", "**/*.go"},
		{"node_modules/pkg/index.js", "node_modules/"},
		{"test_test.go", "*_test.go"},
		{"README.md", "*.md"},
		{"src/internal/config/config.go", "src/**/*.go"},
	}
	for _, c := range corpus {
		f.Add(c.path, c.pattern)
	}

	f.Fuzz(func(t *testing.T, path, pattern string) {
		result := walker.MatchesPattern(path, pattern)
		if pattern == "" && result {
			t.Errorf("MatchesPattern(%q, %q) = true, want false", path, pattern)
		}
	})
}

func FuzzMatchesPatternGlob(f *testing.F) {
	corpus := []struct{ path, pattern string }{
		{"foo.go", "*.go"},
		{"bar.ts", "*.ts"},
		{"baz_test.go", "*_test.go"},
		{"cmd/main.go", "cmd/*.go"},
	}
	for _, c := range corpus {
		f.Add(c.path, c.pattern)
	}

	f.Fuzz(func(t *testing.T, path, pattern string) {
		if strings.Contains(path, "\x00") || strings.Contains(pattern, "\x00") {
			t.Skip()
		}
		result := walker.MatchesPattern(path, pattern)
		if pattern == "" && result {
			t.Errorf("MatchesPattern(%q, %q) = true, want false", path, pattern)
		}
	})
}
