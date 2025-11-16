package generator

import (
	"fmt"

	"github.com/rifqiakrm/starter-cli/internal/types"
)

// GenerateModule generates module components based on the specified parts
func (g *Generator) GenerateModule(schema, entity, version string, parts []types.ModulePart, outputDir string) error {
	fmt.Printf("🚀 Generating module components for %s...\n", entity)

	for _, part := range parts {
		switch part.Component {
		case "handler":
			if err := g.generateHandlers(schema, entity, version, part.Action, outputDir); err != nil {
				return fmt.Errorf("handler generation failed: %v", err)
			}
		case "service":
			if err := g.generateServices(schema, entity, version, part.Action, outputDir); err != nil {
				return fmt.Errorf("service generation failed: %v", err)
			}
		case "repository":
			if err := g.generateRepositories(schema, entity, version, part.Action, outputDir); err != nil {
				return fmt.Errorf("repository generation failed: %v", err)
			}
		default:
			return fmt.Errorf("unknown module component: %s", part.Component)
		}
	}

	return nil
}

// GenerateAll generates entity, resource, and modules in one command
func (g *Generator) GenerateAll(schema, table, entity, version, migrationsPath, entityOut, resourceOut, moduleOut string, parts []types.ModulePart) error {
	fmt.Printf("🎯 Generating complete stack for %s.%s...\n", schema, table)

	// Generate entity
	if err := g.GenerateEntity(schema, table, migrationsPath, entityOut); err != nil {
		return fmt.Errorf("entity generation failed: %v", err)
	}

	// Generate resource
	if err := g.GenerateResource(schema, table, migrationsPath, resourceOut); err != nil {
		return fmt.Errorf("resource generation failed: %v", err)
	}

	// Generate modules
	if len(parts) > 0 {
		if err := g.GenerateModule(schema, entity, version, parts, moduleOut); err != nil {
			return fmt.Errorf("module generation failed: %v", err)
		}
	}

	fmt.Println("✅ Complete stack generated successfully!")
	return nil
}
