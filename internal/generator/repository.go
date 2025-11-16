package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateRepositories generates repository components
func (g *Generator) generateRepositories(schema, entity, version, action, outputDir string) error {
	actions := getActions(action)
	data := g.createTemplateData(schema, entity, version)

	dir := filepath.Join(outputDir, schema, version, "repository")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	for _, act := range actions {
		var templatePath string
		switch act {
		case "creator":
			templatePath = g.config.TemplatePaths.RepositoryCreator
		case "finder":
			templatePath = g.config.TemplatePaths.RepositoryFinder
		case "updater":
			templatePath = g.config.TemplatePaths.RepositoryUpdater
		case "deleter":
			templatePath = g.config.TemplatePaths.RepositoryDeleter
		default:
			return fmt.Errorf("unknown repository action: %s", act)
		}

		code, err := g.generateFromTemplate("repository_"+act, data, templatePath)
		if err != nil {
			return fmt.Errorf("repository %s template error: %v", act, err)
		}

		filename := filepath.Join(dir, fmt.Sprintf("%s_%s.repository.go", strings.ToLower(entity), act))
		if err := os.WriteFile(filename, []byte(code), 0644); err != nil {
			return fmt.Errorf("write repository %s error: %v", act, err)
		}

		fmt.Printf("✅ Generated repository.%s: %s\n", act, filename)
	}

	// Generate cache keys when creating finder repositories
	if contains(actions, "finder") {
		if err := g.generateCacheKeysForEntity(schema, entity); err != nil {
			fmt.Printf("⚠️  Cache keys generation skipped: %v\n", err)
			// Don't fail the whole process if cache generation fails
		}
	}

	return nil
}
