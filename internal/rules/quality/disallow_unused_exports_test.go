package quality

import (
	"testing"

	"github.com/Jonathangadeaharder/structurelint/internal/graph"
	"github.com/Jonathangadeaharder/structurelint/internal/parser"
	"github.com/stretchr/testify/assert"
)

func TestUnusedExportsRule_Name(t *testing.T) {
	r := NewUnusedExportsRule(&graph.ImportGraph{}, nil, nil)
	assert.Equal(t, "disallow-unused-exports", r.Name())
}

func TestUnusedExportsRule_Check_NilGraph(t *testing.T) {
	r := NewUnusedExportsRule(nil, nil, nil)
	violations := r.Check(nil, nil)
	assert.Empty(t, violations)
}

func TestUnusedExportsRule_Check_AllExportsUsed(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"src/utils.go", "src/main.go"},
		Dependencies: map[string][]string{
			"src/main.go": {"src/utils.go"},
		},
		Exports: map[string][]parser.Export{
			"src/utils.go": {{Names: []string{"Helper"}}},
		},
	}

	r := NewUnusedExportsRule(g, nil, nil)
	violations := r.Check(nil, nil)
	assert.Empty(t, violations, "utils.go is imported by main.go, so exports are used")
}

func TestUnusedExportsRule_Check_UnusedExport(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"src/unused.go", "src/main.go"},
		Dependencies: map[string][]string{
			"src/main.go": {},
		},
		Exports: map[string][]parser.Export{
			"src/unused.go": {{Names: []string{"UnusedHelper"}}},
		},
	}

	r := NewUnusedExportsRule(g, nil, nil)
	violations := r.Check(nil, nil)
	assert.NotEmpty(t, violations)
	assert.Equal(t, "src/unused.go", violations[0].Path)
}

func TestUnusedExportsRule_Check_WithExcludePattern(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"src/unused.go", "src/internal/private.go"},
		Exports: map[string][]parser.Export{
			"src/unused.go":        {{Names: []string{"Helper"}}},
			"src/internal/private.go": {{Names: []string{"PrivateHelper"}}},
		},
	}

	r := NewUnusedExportsRule(g, []string{"src/internal/**"}, nil)
	violations := r.Check(nil, nil)
	// Only src/unused.go should be reported; src/internal/private.go is excluded
	assert.NotEmpty(t, violations)
	for _, v := range violations {
		assert.NotContains(t, v.Path, "internal", "excluded files should not appear")
	}
}

func TestUnusedExportsRule_Check_WithEntryPointPatterns(t *testing.T) {
	indexExports := []parser.Export{{Names: []string{"ComponentA", "ComponentB"}}}
	g := &graph.ImportGraph{
		AllFiles: []string{"src/index.ts", "src/internal.ts"},
		Exports: map[string][]parser.Export{
			"src/index.ts":   indexExports,
			"src/internal.ts": {{Names: []string{"PrivateHelper"}}},
		},
	}

	// index.ts is an entry point — its exports are considered public API
	r := NewUnusedExportsRule(g, nil, []string{"src/index.ts"})
	violations := r.Check(nil, nil)
	assert.NotEmpty(t, violations)
	for _, v := range violations {
		assert.NotEqual(t, "src/index.ts", v.Path, "entry point exports should not be flagged")
	}
}

func TestUnusedExportsRule_Check_NoExports(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"src/main.go"},
		Exports:  map[string][]parser.Export{},
	}

	r := NewUnusedExportsRule(g, nil, nil)
	violations := r.Check(nil, nil)
	assert.Empty(t, violations)
}

func TestUnusedExportsRule_Check_ImportResolution(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"src/foo.go", "src/bar.go"},
		Dependencies: map[string][]string{
			"src/bar.go": {"src/foo"},
		},
		Exports: map[string][]parser.Export{
			"src/foo.go": {{Names: []string{"FooHelper"}}},
		},
	}

	r := NewUnusedExportsRule(g, nil, nil)
	violations := r.Check(nil, nil)
	assert.Empty(t, violations, "src/foo should resolve to src/foo.go via extension resolution")
}

func TestUnusedExportsRule_Check_MultipleExports(t *testing.T) {
	g := &graph.ImportGraph{
		AllFiles: []string{"src/unused.go", "src/main.go"},
		Exports: map[string][]parser.Export{
			"src/unused.go": {{Names: []string{"A", "B", "C"}}},
		},
	}

	r := NewUnusedExportsRule(g, nil, nil)
	violations := r.Check(nil, nil)
	assert.NotEmpty(t, violations)
	assert.Contains(t, violations[0].Message, "A")
}

func TestNewUnusedExportsRule(t *testing.T) {
	g := &graph.ImportGraph{}
	r := NewUnusedExportsRule(g, []string{"**/*.ts"}, []string{"src/index.ts"})
	assert.NotNil(t, r)
	assert.Equal(t, g, r.ImportGraph)
	assert.Len(t, r.ExcludePatterns, 1)
	assert.Len(t, r.EntryPointPatterns, 1)
}
