package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rifqiakrm/starter-cli/internal/types"
)

// generateCompleteBuilder generates a complete new builder file
func (g *Generator) generateCompleteBuilder(module, version string, tables []string) error {
	fmt.Printf("🏗️  Generating complete builder for %s\n", module)

	// Prepare builder configuration
	config := g.buildBuilderConfig(module, version, tables)

	// Generate builder code
	builderCode, err := g.generateBuilderCode(config)
	if err != nil {
		return fmt.Errorf("builder template error: %v", err)
	}

	// Create module directory
	moduleDir := filepath.Join("modules", module)
	if err := os.MkdirAll(moduleDir, 0755); err != nil {
		return fmt.Errorf("mkdir error: %v", err)
	}

	// Write builder file
	builderFile := filepath.Join(moduleDir, "builder.go")
	if err := os.WriteFile(builderFile, []byte(builderCode), 0644); err != nil {
		return fmt.Errorf("write builder error: %v", err)
	}

	fmt.Printf("✅ Generated builder: %s\n", builderFile)
	return nil
}

// buildBuilderConfig creates dynamic configuration for any module
func (g *Generator) buildBuilderConfig(module, version string, tables []string) types.BuilderConfig {
	tableConfigs := make([]types.TableConfig, 0, len(tables))

	for _, table := range tables {
		tableConfigs = append(tableConfigs, types.TableConfig{
			Name:        table,
			DisplayName: toPascalCase(table),
			Module:      module,
		})
	}

	return types.BuilderConfig{
		Module:        module,
		Version:       version,
		Tables:        tableConfigs,
		RoutePrefix:   fmt.Sprintf("/%s/%s", module, version),
		ImportPath:    fmt.Sprintf("gin-starter/modules/%s/%s", module, version),
		HandlerPrefix: toPascalCase(module),
		HasAuth:       module == "auth",
		HasCron:       module == "auth", // Only auth has cron in your example
		CustomImports: g.getCustomImports(module, version),
	}
}

// getCustomImports returns module-specific imports
func (g *Generator) getCustomImports(module, version string) []string {
	imports := []string{}

	if module == "auth" {
		imports = append(imports, `cronjobHandler "gin-starter/modules/auth/v1/cronjob/handler"`)
	}

	return imports
}

// generateBuilderCode generates the actual builder code from template
func (g *Generator) generateBuilderCode(config types.BuilderConfig) (string, error) {
	templatePath := g.config.TemplatePaths.Builder
	if templatePath == "" {
		templatePath = "templates/builder/builder.tmpl"
	}

	// Load template
	tmplContent, err := g.loadTemplate(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to load builder template: %v", err)
	}

	// Create template with functions
	funcMap := template.FuncMap{
		"ToCamel":  toCamelCase,
		"ToPascal": toPascalCase,
		"ToLower":  strings.ToLower,
	}

	tmpl, err := template.New("builder").Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("parse builder template error: %v", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, config); err != nil {
		return "", fmt.Errorf("execute builder template error: %v", err)
	}

	return buf.String(), nil
}
