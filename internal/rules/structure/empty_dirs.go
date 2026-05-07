package structure

import (
	"github.com/Jonathangadeaharder/structurelint/internal/rules"
	"github.com/Jonathangadeaharder/structurelint/internal/walker"
)

// EmptyDirsRule flags directories that contain no files and no subdirectories.
// Empty directories committed to git rot quickly; they usually mean either
// dead scaffolding or files were deleted but the dir was kept by accident.
type EmptyDirsRule struct{}

func (r *EmptyDirsRule) Name() string { return "disallow-empty-dirs" }

func (r *EmptyDirsRule) Check(_ []walker.FileInfo, dirs map[string]*walker.DirInfo) []rules.Violation {
	var violations []rules.Violation
	for path, info := range dirs {
		if info.FileCount == 0 && info.SubdirCount == 0 {
			violations = append(violations, rules.Violation{
				Rule:    r.Name(),
				Path:    path,
				Message: "directory is empty",
				Suggestions: []string{
					"Delete the directory if it is no longer needed",
					"Add a `.gitkeep` only if the directory must exist before content is added",
				},
			})
		}
	}
	return violations
}

func NewEmptyDirsRule() *EmptyDirsRule { return &EmptyDirsRule{} }
