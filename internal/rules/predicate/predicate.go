// Package predicate provides a DSL for composing rule predicates
package predicate

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/structurelint/structurelint/internal/graph"
	"github.com/structurelint/structurelint/internal/walker"
)

// Predicate is a function that evaluates to true or false for a given file
type Predicate func(file walker.FileInfo, ctx *Context) bool

// Context provides additional information for predicate evaluation
type Context struct {
	Graph      *graph.ImportGraph
	AllFiles   []walker.FileInfo
	AllDirs    map[string]*walker.DirInfo
	RootPath   string
	CustomData map[string]interface{}
}

// Builder provides a fluent interface for building predicates
type Builder struct {
	predicate Predicate
}

// New creates a new predicate builder
func New() *Builder {
	return &Builder{
		predicate: func(file walker.FileInfo, ctx *Context) bool {
			return true // Default: always true
		},
	}
}

// WithPredicate sets the initial predicate
func WithPredicate(p Predicate) *Builder {
	return &Builder{predicate: p}
}

// Build returns the final predicate
func (b *Builder) Build() Predicate {
	return b.predicate
}

// And combines predicates with logical AND
func (b *Builder) And(other Predicate) *Builder {
	current := b.predicate
	b.predicate = func(file walker.FileInfo, ctx *Context) bool {
		return current(file, ctx) && other(file, ctx)
	}
	return b
}

// Or combines predicates with logical OR
func (b *Builder) Or(other Predicate) *Builder {
	current := b.predicate
	b.predicate = func(file walker.FileInfo, ctx *Context) bool {
		return current(file, ctx) || other(file, ctx)
	}
	return b
}

// Not negates the current predicate
func (b *Builder) Not() *Builder {
	current := b.predicate
	b.predicate = func(file walker.FileInfo, ctx *Context) bool {
		return !current(file, ctx)
	}
	return b
}

// --- Path-based predicates ---

// PathMatches checks if file path matches a glob pattern
func PathMatches(pattern string) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		matched, err := filepath.Match(pattern, file.Path)
		return err == nil && matched
	}
}

// PathContains checks if file path contains a substring
func PathContains(substr string) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return strings.Contains(file.Path, substr)
	}
}

// PathStartsWith checks if file path starts with a prefix
func PathStartsWith(prefix string) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return strings.HasPrefix(file.Path, prefix)
	}
}

// PathEndsWith checks if file path ends with a suffix
func PathEndsWith(suffix string) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return strings.HasSuffix(file.Path, suffix)
	}
}

// PathRegex checks if file path matches a regular expression
func PathRegex(pattern string) Predicate {
	re := regexp.MustCompile(pattern)
	return func(file walker.FileInfo, ctx *Context) bool {
		return re.MatchString(file.Path)
	}
}

// --- File type predicates ---

// IsFile checks if the entry is a file (not a directory)
func IsFile() Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return !file.IsDir
	}
}

// IsDirectory checks if the entry is a directory
func IsDirectory() Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return file.IsDir
	}
}

// HasExtension checks if file has a specific extension
func HasExtension(ext string) Predicate {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}
	return func(file walker.FileInfo, ctx *Context) bool {
		return filepath.Ext(file.Path) == ext
	}
}

// --- Layer predicates ---

// InLayer checks if file belongs to a specific layer
func InLayer(layerName string) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		if ctx.Graph == nil {
			return false
		}
		layer := ctx.Graph.GetLayerForFile(file.Path)
		return layer != nil && layer.Name == layerName
	}
}

// HasLayer checks if file belongs to any layer
func HasLayer() Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		if ctx.Graph == nil {
			return false
		}
		layer := ctx.Graph.GetLayerForFile(file.Path)
		return layer != nil
	}
}

// --- Dependency predicates ---

// DependsOn checks if file depends on another file/pattern
func DependsOn(pattern string) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		if ctx.Graph == nil {
			return false
		}
		deps := ctx.Graph.GetDependencies(file.Path)
		for _, dep := range deps {
			matched, err := filepath.Match(pattern, dep)
			if err == nil && matched {
				return true
			}
		}
		return false
	}
}

// HasDependencies checks if file has any dependencies
func HasDependencies() Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		if ctx.Graph == nil {
			return false
		}
		deps := ctx.Graph.GetDependencies(file.Path)
		return len(deps) > 0
	}
}

// HasIncomingRefs checks if file is imported by other files
func HasIncomingRefs() Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		if ctx.Graph == nil {
			return false
		}
		refs := ctx.Graph.IncomingRefs[file.Path]
		return refs > 0
	}
}

// IsOrphaned checks if file has no incoming or outgoing dependencies
func IsOrphaned() Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		if ctx.Graph == nil {
			return false
		}
		deps := ctx.Graph.GetDependencies(file.Path)
		refs := ctx.Graph.IncomingRefs[file.Path]
		return len(deps) == 0 && refs == 0
	}
}

// --- Size predicates ---

// SizeGreaterThan checks if file size exceeds a threshold
func SizeGreaterThan(bytes int64) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		// Get file size from filesystem
		info, err := os.Stat(file.AbsPath)
		if err != nil || info.IsDir() {
			return false
		}
		return info.Size() > bytes
	}
}

// SizeLessThan checks if file size is below a threshold
func SizeLessThan(bytes int64) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		// Get file size from filesystem
		info, err := os.Stat(file.AbsPath)
		if err != nil || info.IsDir() {
			return false
		}
		return info.Size() < bytes
	}
}

// --- Depth predicates ---

// DepthEquals checks if file is at a specific depth
func DepthEquals(depth int) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return file.Depth == depth
	}
}

// DepthGreaterThan checks if file depth exceeds a threshold
func DepthGreaterThan(depth int) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return file.Depth > depth
	}
}

// DepthLessThan checks if file depth is below a threshold
func DepthLessThan(depth int) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return file.Depth < depth
	}
}

// --- Naming predicates ---

// NameMatches checks if file name matches a pattern
func NameMatches(pattern string) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		matched, err := filepath.Match(pattern, filepath.Base(file.Path))
		return err == nil && matched
	}
}

// NameContains checks if file name contains a substring
func NameContains(substr string) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return strings.Contains(filepath.Base(file.Path), substr)
	}
}

// NameRegex checks if file name matches a regular expression
func NameRegex(pattern string) Predicate {
	re := regexp.MustCompile(pattern)
	return func(file walker.FileInfo, ctx *Context) bool {
		return re.MatchString(filepath.Base(file.Path))
	}
}

// --- Composite predicates ---

// All checks if all predicates are true
func All(predicates ...Predicate) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		for _, p := range predicates {
			if !p(file, ctx) {
				return false
			}
		}
		return true
	}
}

// Any checks if any predicate is true
func Any(predicates ...Predicate) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		for _, p := range predicates {
			if p(file, ctx) {
				return true
			}
		}
		return false
	}
}

// None checks if no predicate is true
func None(predicates ...Predicate) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		for _, p := range predicates {
			if p(file, ctx) {
				return false
			}
		}
		return true
	}
}

// Not creates a negated predicate
func Not(p Predicate) Predicate {
	return func(file walker.FileInfo, ctx *Context) bool {
		return !p(file, ctx)
	}
}

// --- Custom predicates ---

// Custom allows defining a custom predicate function
func Custom(fn func(file walker.FileInfo, ctx *Context) bool) Predicate {
	return fn
}
