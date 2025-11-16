package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/rifqiakrm/starter-cli/internal/parser"
)

// GenerateEntity generates entity code from SQL migrations
// GenerateEntity generates entity code from SQL migrations
func (g *Generator) GenerateEntity(schema, table, migrationsPath, outputDir string) error {
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

	// Load and execute template
	fmt.Printf("⚡ Generating entity for %s...\n", tbl.Name)
	code, err := g.generateFromTemplate("entity", tbl, g.config.TemplatePaths.Entity)
	if err != nil {
		return fmt.Errorf("template error: %v", err)
	}

	// Create output directory
	if err := os.MkdirAll(formattedOutputDir, 0755); err != nil {
		return fmt.Errorf("mkdir error: %v", err)
	}

	// Write generated file
	filename := fmt.Sprintf("%s.entity.go", tbl.NameLower)
	outputPath := filepath.Join(formattedOutputDir, filename)
	if err := os.WriteFile(outputPath, []byte(code), 0644); err != nil {
		return fmt.Errorf("write error: %v", err)
	}

	fmt.Printf("✅ Generated entity: %s\n", outputPath)
	return nil
}

// generateFromTemplate is a helper to load and execute templates
func (g *Generator) generateFromTemplate(templateName string, data interface{}, templatePath string) (string, error) {
	// Try to load custom template first, fallback to embedded defaults
	tmplContent, err := g.loadTemplate(templatePath)
	if err != nil {
		return "", fmt.Errorf("failed to load template %s: %v", templatePath, err)
	}

	// Create template with helper functions
	funcMap := template.FuncMap{
		"ToCamel":      toCamelCase,
		"ToPascalCase": toPascalCase,
		"ToLowerCamel": toLowerCamelCase,
		"GoType":       goType,
		"IsAuditable":  isAuditable,
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
