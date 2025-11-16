package types

// RoutesConfig holds configuration for routes generation
type RoutesConfig struct {
	Module        string
	Version       string
	Tables        []TableConfig
	RoutePrefix   string
	ImportPath    string
	HandlerPrefix string
	HandlerStruct string
}

// HandlerMethodConfig holds configuration for route methods
type HandlerMethodConfig struct {
	MethodName string // "Finder", "Creator", "Updater", "Deleter"
	HTTPMethod string // "GET", "POST", "PUT", "DELETE"
	Action     string // "find", "create", "update", "delete"
	RoutePath  string // "/users", "/roles", etc.
}
