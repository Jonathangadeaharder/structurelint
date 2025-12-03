package plugin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// RulePlugin defines the interface for external rule plugins
type RulePlugin interface {
	// Name returns the plugin name
	Name() string
	// Check executes the rule logic via the plugin
	Check(ctx context.Context, files []walker.FileInfo, config map[string]interface{}) ([]rules.Violation, error)
}

// ProcessPlugin implements RulePlugin using an external process
type ProcessPlugin struct {
	name       string
	executable string
	args       []string
}

// NewProcessPlugin creates a new process-based plugin
func NewProcessPlugin(name, executable string, args ...string) *ProcessPlugin {
	return &ProcessPlugin{
		name:       name,
		executable: executable,
		args:       args,
	}
}

func (p *ProcessPlugin) Name() string {
	return p.name
}

// PluginInput represents the input passed to the plugin process
type PluginInput struct {
	Config map[string]interface{} `json:"config"`
	Files  []string               `json:"files"`
}

// Check runs the external process and parses its output
func (p *ProcessPlugin) Check(ctx context.Context, files []walker.FileInfo, config map[string]interface{}) ([]rules.Violation, error) {
	// Prepare input
	filePaths := make([]string, len(files))
	for i, f := range files {
		filePaths[i] = f.AbsPath
	}

	input := PluginInput{
		Config: config,
		Files:  filePaths,
	}

	inputBytes, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plugin input: %w", err)
	}

	// Prepare command
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, p.executable, p.args...)
	cmd.Stdin = bytes.NewReader(inputBytes)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("plugin execution failed: %w (stderr: %s)", err, stderr.String())
	}

	// Parse output
	var violations []rules.Violation
	if err := json.Unmarshal(stdout.Bytes(), &violations); err != nil {
		return nil, fmt.Errorf("failed to parse plugin output: %w", err)
	}

	// Ensure rule name is set
	for i := range violations {
		if violations[i].Rule == "" {
			violations[i].Rule = p.name
		}
	}

	return violations, nil
}
