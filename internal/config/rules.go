package config

// RuleConfigs holds all rule configurations with type safety.
// Each field corresponds to a rule name in .structurelint.yml.
type RuleConfigs struct {
	MaxDepth               *MaxDepthConfig               `yaml:"max-depth,omitempty"`
	MaxFilesInDir          *MaxFilesInDirConfig          `yaml:"max-files-in-dir,omitempty"`
	MaxSubdirs             *MaxSubdirsConfig             `yaml:"max-subdirs,omitempty"`
	NamingConvention       NamingConventionConfig        `yaml:"naming-convention,omitempty"`
	FileExistence          FileExistenceConfig           `yaml:"file-existence,omitempty"`
	RegexMatch             RegexMatchConfig              `yaml:"regex-match,omitempty"`
	DisallowedPatterns     DisallowedPatternsConfig      `yaml:"disallowed-patterns,omitempty"`
	TestAdjacency          *TestAdjacencyConfig          `yaml:"test-adjacency,omitempty"`
	TestLocation           *TestLocationConfig           `yaml:"test-location,omitempty"`
	EnforceLayerBoundaries *EnforceLayerBoundariesConfig `yaml:"enforce-layer-boundaries,omitempty"`
	DisallowOrphanedFiles  *DisallowOrphanedFilesConfig  `yaml:"disallow-orphaned-files,omitempty"`
	DisallowImportCycles   *DisallowImportCyclesConfig   `yaml:"disallow-import-cycles,omitempty"`
	PathBasedLayers        *PathBasedLayersConfig        `yaml:"path-based-layers,omitempty"`
	SpecADR                *SpecADRConfig                `yaml:"spec-adr,omitempty"`
	SpecADREnforcement     *SpecADRConfig                `yaml:"spec-adr-enforcement,omitempty"`
}

// MaxDepthConfig configures the max-depth rule.
type MaxDepthConfig struct {
	Max int `yaml:"max"`
}

// MaxFilesInDirConfig configures the max-files-in-dir rule.
type MaxFilesInDirConfig struct {
	Max int `yaml:"max"`
}

// MaxSubdirsConfig configures the max-subdirs rule.
type MaxSubdirsConfig struct {
	Max int `yaml:"max"`
}

// NamingConventionConfig configures the naming-convention rule.
// Maps file glob patterns to naming conventions (e.g., "*.ts" -> "camelCase").
type NamingConventionConfig map[string]string

// FileExistenceConfig configures the file-existence rule.
// Maps file patterns to existence requirements (e.g., "index.ts" -> "exists:1").
type FileExistenceConfig map[string]string

// RegexMatchConfig configures the regex-match rule.
// Maps file patterns to regex requirements.
type RegexMatchConfig map[string]string

// DisallowedPatternsConfig configures the disallowed-patterns rule.
// A list of glob patterns that are not allowed.
type DisallowedPatternsConfig []string

// TestAdjacencyConfig configures the test-adjacency rule.
type TestAdjacencyConfig struct {
	Pattern      string   `yaml:"pattern" json:"pattern"`
	TestDir      string   `yaml:"test-dir,omitempty" json:"test-dir,omitempty"`
	FilePatterns []string `yaml:"file-patterns,omitempty" json:"file-patterns,omitempty"`
	Exemptions   []string `yaml:"exemptions,omitempty" json:"exemptions,omitempty"`
}

// TestLocationConfig configures the test-location rule.
type TestLocationConfig struct {
	IntegrationTestDir string   `yaml:"integration-test-dir,omitempty" json:"integration-test-dir,omitempty"`
	AllowAdjacent      bool     `yaml:"allow-adjacent,omitempty" json:"allow-adjacent,omitempty"`
	FilePatterns       []string `yaml:"file-patterns,omitempty" json:"file-patterns,omitempty"`
	Exemptions         []string `yaml:"exemptions,omitempty" json:"exemptions,omitempty"`
}

// EnforceLayerBoundariesConfig configures the enforce-layer-boundaries rule.
// An empty struct means the rule only needs to be enabled (no parameters).
type EnforceLayerBoundariesConfig struct{}

// DisallowOrphanedFilesConfig configures the disallow-orphaned-files rule.
type DisallowOrphanedFilesConfig struct {
	EntryPointPatterns []string `yaml:"entry-point-patterns,omitempty" json:"entry-point-patterns,omitempty"`
}

// DisallowImportCyclesConfig configures the disallow-import-cycles rule.
// An empty struct means the rule only needs to be enabled (no parameters).
type DisallowImportCyclesConfig struct{}

// PathLayerConfig represents a single layer in the path-based-layers rule.
type PathLayerConfig struct {
	Name           string   `yaml:"name" json:"name"`
	Patterns       []string `yaml:"patterns,omitempty" json:"patterns,omitempty"`
	CanDependOn    []string `yaml:"canDependOn,omitempty" json:"canDependOn,omitempty"`
	ForbiddenPaths []string `yaml:"forbiddenPaths,omitempty" json:"forbiddenPaths,omitempty"`
}

// PathBasedLayersConfig configures the path-based-layers rule.
type PathBasedLayersConfig struct {
	Layers []PathLayerConfig `yaml:"layers,omitempty" json:"layers,omitempty"`
}

// SpecADRConfig configures the spec-adr rule.
type SpecADRConfig struct {
	RequireSpecFolder    *bool    `yaml:"require-spec-folder,omitempty" json:"require-spec-folder,omitempty"`
	RequireADRFolder     *bool    `yaml:"require-adr-folder,omitempty" json:"require-adr-folder,omitempty"`
	EnforceSpecTemplate  *bool    `yaml:"enforce-spec-template,omitempty" json:"enforce-spec-template,omitempty"`
	EnforceADRTemplate   *bool    `yaml:"enforce-adr-template,omitempty" json:"enforce-adr-template,omitempty"`
	SpecFolderPaths      []string `yaml:"spec-folder-paths,omitempty" json:"spec-folder-paths,omitempty"`
	ADRFolderPaths       []string `yaml:"adr-folder-paths,omitempty" json:"adr-folder-paths,omitempty"`
	SpecFilePatterns     []string `yaml:"spec-file-patterns,omitempty" json:"spec-file-patterns,omitempty"`
	ADRFilePatterns      []string `yaml:"adr-file-patterns,omitempty" json:"adr-file-patterns,omitempty"`
	SpecRequiredHeadings []string `yaml:"spec-required-headings,omitempty" json:"spec-required-headings,omitempty"`
	ADRRequiredHeadings  []string `yaml:"adr-required-headings,omitempty" json:"adr-required-headings,omitempty"`
	ADRRequiredMetadata  []string `yaml:"adr-required-metadata,omitempty" json:"adr-required-metadata,omitempty"`
}
