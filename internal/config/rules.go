package config

// RuleConfigs holds all rule configurations with type safety.
// Each field corresponds to a rule name in .structurelint.yml.
type RuleConfigs struct {
	MaxDepth               *MaxDepthConfig               `yaml:"max-depth,omitempty"`
	MaxFilesInDir          *MaxFilesInDirConfig          `yaml:"max-files-in-dir,omitempty"`
	MaxSubdirs             *MaxSubdirsConfig             `yaml:"max-subdirs,omitempty"`
	MaxCognitiveComplexity *MaxCognitiveComplexityConfig `yaml:"max-cognitive-complexity,omitempty"`
	MaxHalsteadEffort      *MaxHalsteadEffortConfig      `yaml:"max-halstead-effort,omitempty"`
	NamingConvention       NamingConventionConfig        `yaml:"naming-convention,omitempty"`
	FileExistence          FileExistenceConfig           `yaml:"file-existence,omitempty"`
	RegexMatch             RegexMatchConfig              `yaml:"regex-match,omitempty"`
	DisallowedPatterns     DisallowedPatternsConfig      `yaml:"disallowed-patterns,omitempty"`
	TestAdjacency          *TestAdjacencyConfig          `yaml:"test-adjacency,omitempty"`
	TestLocation           *TestLocationConfig           `yaml:"test-location,omitempty"`
	FileContent            *FileContentConfig            `yaml:"file-content,omitempty"`
	GitHubWorkflows        *GitHubWorkflowsConfig        `yaml:"github-workflows,omitempty"`
	LinterConfig           *LinterConfigConfig           `yaml:"linter-config,omitempty"`
	ApiSpec                *ApiSpecConfig                `yaml:"api-spec,omitempty"`
	ContractFramework      *ContractFrameworkConfig      `yaml:"contract-framework,omitempty"`
	SpecADREnforcement     *SpecADREnforcementConfig     `yaml:"spec-adr-enforcement,omitempty"`
	EnforceLayerBoundaries *EnforceLayerBoundariesConfig `yaml:"enforce-layer-boundaries,omitempty"`
	DisallowOrphanedFiles  *DisallowOrphanedFilesConfig  `yaml:"disallow-orphaned-files,omitempty"`
	DisallowUnusedExports  *DisallowUnusedExportsConfig  `yaml:"disallow-unused-exports,omitempty"`
	PropertyEnforcement    *PropertyEnforcementConfig    `yaml:"property-enforcement,omitempty"`
	PathBasedLayers        *PathBasedLayersConfig        `yaml:"path-based-layers,omitempty"`
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

// MaxCognitiveComplexityConfig configures the max-cognitive-complexity rule.
type MaxCognitiveComplexityConfig struct {
	Max          int      `yaml:"max"`
	TestMax      int      `yaml:"test-max,omitempty"`
	FilePatterns []string `yaml:"file-patterns,omitempty"`
}

// MaxHalsteadEffortConfig configures the max-halstead-effort rule.
type MaxHalsteadEffortConfig struct {
	Max          float64  `yaml:"max"`
	FilePatterns []string `yaml:"file-patterns,omitempty"`
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
	Pattern      string   `yaml:"pattern"`
	TestDir      string   `yaml:"test-dir,omitempty"`
	FilePatterns []string `yaml:"file-patterns,omitempty"`
	Exemptions   []string `yaml:"exemptions,omitempty"`
}

// TestLocationConfig configures the test-location rule.
type TestLocationConfig struct {
	IntegrationTestDir string   `yaml:"integration-test-dir,omitempty"`
	AllowAdjacent      bool     `yaml:"allow-adjacent,omitempty"`
	FilePatterns       []string `yaml:"file-patterns,omitempty"`
	Exemptions         []string `yaml:"exemptions,omitempty"`
}

// FileContentConfig configures the file-content rule.
type FileContentConfig struct {
	Templates   map[string]string `yaml:"templates,omitempty"`
	TemplateDir string            `yaml:"template-dir,omitempty"`
}

// GitHubWorkflowsConfig configures the github-workflows rule.
type GitHubWorkflowsConfig struct {
	RequireTests           bool     `yaml:"require-tests,omitempty"`
	RequireSecurity        bool     `yaml:"require-security,omitempty"`
	RequireQuality         bool     `yaml:"require-quality,omitempty"`
	RequireLogCommits      bool     `yaml:"require-log-commits,omitempty"`
	RequireRepomixArtifact bool     `yaml:"require-repomix-artifact,omitempty"`
	RequiredJobs           []string `yaml:"required-jobs,omitempty"`
	RequiredTriggers       []string `yaml:"required-triggers,omitempty"`
	AllowMissing           []string `yaml:"allow-missing,omitempty"`
}

// LinterConfigConfig configures the linter-config rule.
type LinterConfigConfig struct {
	RequirePython     bool     `yaml:"require-python,omitempty"`
	RequireTypeScript bool     `yaml:"require-typescript,omitempty"`
	RequireGo         bool     `yaml:"require-go,omitempty"`
	RequireHTML       bool     `yaml:"require-html,omitempty"`
	RequireCSS        bool     `yaml:"require-css,omitempty"`
	RequireSQL        bool     `yaml:"require-sql,omitempty"`
	RequireRust       bool     `yaml:"require-rust,omitempty"`
	RequireMarkdown   bool     `yaml:"require-markdown,omitempty"`
	RequireJava       bool     `yaml:"require-java,omitempty"`
	RequireCpp        bool     `yaml:"require-cpp,omitempty"`
	RequireCSharp     bool     `yaml:"require-csharp,omitempty"`
	CustomLinters     []string `yaml:"custom-linters,omitempty"`
}

// ApiSpecConfig configures the api-spec rule.
type ApiSpecConfig struct {
	RequireOpenAPI  bool     `yaml:"require-openapi,omitempty"`
	RequireAsyncAPI bool     `yaml:"require-asyncapi,omitempty"`
	CustomSpecs     []string `yaml:"custom-specs,omitempty"`
}

// ContractFrameworkConfig configures the contract-framework rule.
type ContractFrameworkConfig struct {
	RequirePython     bool     `yaml:"require-python,omitempty"`
	RequireRust       bool     `yaml:"require-rust,omitempty"`
	RequireTypeScript bool     `yaml:"require-typescript,omitempty"`
	RequireGo         bool     `yaml:"require-go,omitempty"`
	RequireJava       bool     `yaml:"require-java,omitempty"`
	RequireCSharp     bool     `yaml:"require-csharp,omitempty"`
	RequireCPlusPlus  bool     `yaml:"require-cplusplus,omitempty"`
	CustomFrameworks  []string `yaml:"custom-frameworks,omitempty"`
}

// SpecADREnforcementConfig configures the spec-adr-enforcement rule.
type SpecADREnforcementConfig struct {
	RequireSpecFolder   bool     `yaml:"require-spec-folder,omitempty"`
	RequireADRFolder    bool     `yaml:"require-adr-folder,omitempty"`
	EnforceSpecTemplate bool     `yaml:"enforce-spec-template,omitempty"`
	EnforceADRTemplate  bool     `yaml:"enforce-adr-template,omitempty"`
	SpecFolderPaths     []string `yaml:"spec-folder-paths,omitempty"`
	ADRFolderPaths      []string `yaml:"adr-folder-paths,omitempty"`
	SpecFilePatterns    []string `yaml:"spec-file-patterns,omitempty"`
	ADRFilePatterns     []string `yaml:"adr-file-patterns,omitempty"`
}

// EnforceLayerBoundariesConfig configures the enforce-layer-boundaries rule.
// An empty struct means the rule only needs to be enabled (no parameters).
type EnforceLayerBoundariesConfig struct{}

// DisallowOrphanedFilesConfig configures the disallow-orphaned-files rule.
type DisallowOrphanedFilesConfig struct {
	EntryPointPatterns []string `yaml:"entry-point-patterns,omitempty"`
}

// DisallowUnusedExportsConfig configures the disallow-unused-exports rule.
type DisallowUnusedExportsConfig struct{}

// PropertyEnforcementConfig configures the property-enforcement rule.
type PropertyEnforcementConfig struct {
	MaxDependenciesPerFile int      `yaml:"max_dependencies_per_file,omitempty"`
	MaxDependencyDepth     int      `yaml:"max_dependency_depth,omitempty"`
	DetectCycles           bool     `yaml:"detect_cycles,omitempty"`
	ForbiddenPatterns      []string `yaml:"forbidden_patterns,omitempty"`
}

// PathLayerConfig represents a single layer in the path-based-layers rule.
type PathLayerConfig struct {
	Name           string   `yaml:"name"`
	Patterns       []string `yaml:"patterns,omitempty"`
	CanDependOn    []string `yaml:"canDependOn,omitempty"`
	ForbiddenPaths []string `yaml:"forbiddenPaths,omitempty"`
}

// PathBasedLayersConfig configures the path-based-layers rule.
type PathBasedLayersConfig struct {
	Layers []PathLayerConfig `yaml:"layers,omitempty"`
}
