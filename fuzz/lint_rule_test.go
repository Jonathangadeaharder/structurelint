package fuzz

import (
	"encoding/json"
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/output"
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
)

func FuzzLintRule(f *testing.F) {
	for _, s := range []string{"max-depth", "naming-convention", "file-existence", "disallowed-patterns", "", "unknown-rule", "a"} {
		f.Add(s)
	}

	f.Fuzz(func(t *testing.T, ruleName string) {
		violations := []rules.Violation{{Rule: ruleName, Path: "test.txt", Message: "test"}}
		textF := &output.TextFormatter{}
		textF.Format(violations)
		jsonF := &output.JSONFormatter{Version: "fuzz"}
		result, err := jsonF.Format(violations)
		if err == nil {
			var parsed output.JSONOutput
			json.Unmarshal([]byte(result), &parsed)
		}
	})
}
