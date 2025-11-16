package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rifqiakrm/starter-cli/internal/parser"
)

// GenerateResource generates resource code from SQL migrations
func (g *Generator) GenerateResource(schema, table, migrationsPath, outputDir string) error {
	fmt.Printf("🔍 Finding migration for %s.%s...\n", schema, table)

	// Find SQL migration file
	sqlFile, err := parser.FindMigration(migrationsPath, schema, table)
	if err != nil {
		return fmt.Errorf("migration not found: %v", err)
	}

	// Parse SQL to get table structure
	fmt.Printf("📝 Parsing SQL file: %s\n", sqlFile)
	tbl, err := parser.ParseSQL(sqlFile)
	if err != nil {
		return fmt.Errorf("parse error: %v", err)
	}

	// Format output directory (supports %s placeholder for schema)
	formattedOutputDir := strings.Replace(outputDir, "%s", schema, 1)

	// Generate main resource with resource-specific functions
	fmt.Printf("⚡ Generating resource for %s...\n", tbl.Name)
	resourceCode, err := g.generateResourceTemplate("resource", tbl, g.config.TemplatePaths.Resource)
	if err != nil {
		return fmt.Errorf("resource template error: %v", err)
	}

	// Generate create request
	createRequestCode, err := g.generateResourceTemplate("create_request", tbl, g.config.TemplatePaths.CreateRequest)
	if err != nil {
		return fmt.Errorf("create request template error: %v", err)
	}

	// Generate update request
	updateRequestCode, err := g.generateResourceTemplate("update_request", tbl, g.config.TemplatePaths.UpdateRequest)
	if err != nil {
		return fmt.Errorf("update request template error: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll(formattedOutputDir, 0755); err != nil {
		return fmt.Errorf("mkdir error: %v", err)
	}

	// Write resource file
	resourceFile := filepath.Join(formattedOutputDir, fmt.Sprintf("%s.resource.go", tbl.NameLower))
	combinedCode := resourceCode + "\n" + createRequestCode + "\n" + updateRequestCode
	if err := os.WriteFile(resourceFile, []byte(combinedCode), 0644); err != nil {
		return fmt.Errorf("write resource error: %v", err)
	}

	fmt.Printf("✅ Generated resource: %s\n", resourceFile)
	return nil
}

// generateResourceTemplate is specifically for resource templates with resource functions
func (g *Generator) generateResourceTemplate(templateName string, data interface{}, templatePath string) (string, error) {
	tmplContent, err := g.loadTemplate(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %v", templatePath, err)
	}

	// Resource-specific template functions
	funcMap := template.FuncMap{
		"ToCamel":         toCamelCase,
		"ToPascalCase":    toPascalCase,
		"IsSensitive":     isSensitive,
		"IncludeInCreate": includeInCreate,
		"IncludeInUpdate": includeInUpdate,
		"GoResourceType":  goResourceType,
		"GoRequestType":   goRequestType,
		"MapFromEntity":   mapFromEntity,
	}

	tmpl, err := template.New(templateName).Funcs(funcMap).Parse(tmplContent)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
