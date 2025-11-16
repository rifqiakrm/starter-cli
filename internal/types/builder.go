package types

// BuilderConfig holds configuration for builder generation
type BuilderConfig struct {
	Module        string
	Version       string
	Tables        []TableConfig
	RoutePrefix   string
	ImportPath    string
	HandlerPrefix string
	HasAuth       bool
	HasCron       bool
	CustomImports []string
}

// TableConfig holds entity-specific configuration
type TableConfig struct {
	Name        string
	DisplayName string
	Module      string
}
