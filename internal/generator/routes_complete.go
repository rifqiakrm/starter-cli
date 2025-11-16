package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rifqiakrm/starter-cli/internal/types"
)

// generateCompleteRoutes generates complete routes file for new module
func (g *Generator) generateCompleteRoutes(module, version string, tables []string) error {
	fmt.Printf("🛣️  Generating complete routes for %s\n", module)

	// Prepare routes configuration
	config := g.buildRoutesConfig(module, version, tables)

	// Generate routes code
	routesCode, err := g.generateRoutesCode(config)
	if err != nil {
		return fmt.Errorf("routes template error: %v", err)
	}

	// Create app directory if needed
	appDir := "app"
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("mkdir error: %v", err)
	}

	// Write routes file
	routesFile := filepath.Join(appDir, fmt.Sprintf("%s_routes.go", module))
	if err := os.WriteFile(routesFile, []byte(routesCode), 0644); err != nil {
		return fmt.Errorf("write routes error: %v", err)
	}

	fmt.Printf("✅ Generated routes: %s\n", routesFile)
	return nil
}

// buildRoutesConfig creates configuration for routes
func (g *Generator) buildRoutesConfig(module, version string, tables []string) types.RoutesConfig {
	tableConfigs := make([]types.TableConfig, 0, len(tables))

	for _, table := range tables {
		tableConfigs = append(tableConfigs, types.TableConfig{
			Name:        table,
			DisplayName: toPascalCase(table),
			Module:      module,
		})
	}

	return types.RoutesConfig{
		Module:        module,
		Version:       version,
		Tables:        tableConfigs,
		RoutePrefix:   fmt.Sprintf("/%s/%s", module, version),
		ImportPath:    fmt.Sprintf("gin-starter/modules/%s/%s", module, version),
		HandlerPrefix: toPascalCase(module),
		HandlerStruct: fmt.Sprintf("%sHTTPHandler", toPascalCase(module)),
	}
}

// generateRoutesCode generates the actual routes code from template
func (g *Generator) generateRoutesCode(config types.RoutesConfig) (string, error) {
	templatePath := g.config.TemplatePaths.Routes
	if templatePath == "" {
		templatePath = "templates/routes/routes.tmpl"
	}

	// Load template
	tmplContent, err := g.loadTemplate(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to load routes template: %v", err)
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"ToCamel":      toCamelCase,
		"ToPascal":     toPascalCase,
		"ToLower":      strings.ToLower,
		"GetRoutePath": g.getRoutePath,
	}

	tmpl, err := template.New("routes").Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("parse routes template error: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, config); err != nil {
		return "", fmt.Errorf("execute routes template error: %v", err)
	}

	return buf.String(), nil
}

// getRoutePath returns the route path for an table
func (g *Generator) getRoutePath(tableName string) string {
	// Convert "user" -> "users", "category" -> "categories"
	if strings.HasSuffix(tableName, "y") {
		return strings.TrimSuffix(tableName, "y") + "ies"
	}
	return tableName + "s"
}
