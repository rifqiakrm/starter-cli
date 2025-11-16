package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateHandlers generates handler components
func (g *Generator) generateHandlers(schema, entity, version, action, outputDir string) error {
	actions := getActions(action)
	data := g.createTemplateData(schema, entity, version)

	dir := filepath.Join(outputDir, schema, version, "handler")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	for _, act := range actions {
		var templatePath string
		switch act {
		case "creator":
			templatePath = g.config.TemplatePaths.HandlerCreator
		case "finder":
			templatePath = g.config.TemplatePaths.HandlerFinder
		case "updater":
			templatePath = g.config.TemplatePaths.HandlerUpdater
		case "deleter":
			templatePath = g.config.TemplatePaths.HandlerDeleter
		default:
			return fmt.Errorf("unknown handler action: %s", act)
		}

		code, err := g.generateFromTemplate("handler_"+act, data, templatePath)
		if err != nil {
			return fmt.Errorf("handler %s template error: %v", act, err)
		}

		filename := filepath.Join(dir, fmt.Sprintf("%s_%s.handler.go", strings.ToLower(entity), act))
		if err := os.WriteFile(filename, []byte(code), 0644); err != nil {
			return fmt.Errorf("write handler %s error: %v", act, err)
		}

		fmt.Printf("✅ Generated handler.%s: %s\n", act, filename)
	}

	return nil
}
