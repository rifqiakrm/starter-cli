package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateServices generates service components
func (g *Generator) generateServices(schema, entity, version, action, outputDir string) error {
	actions := getActions(action)
	data := g.createTemplateData(schema, entity, version)

	dir := filepath.Join(outputDir, schema, version, "service")
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return err
	}

	for _, act := range actions {
		var templatePath string
		switch act {
		case "creator":
			templatePath = g.config.TemplatePaths.ServiceCreator
		case "finder":
			templatePath = g.config.TemplatePaths.ServiceFinder
		case "updater":
			templatePath = g.config.TemplatePaths.ServiceUpdater
		case "deleter":
			templatePath = g.config.TemplatePaths.ServiceDeleter
		default:
			return fmt.Errorf("unknown service action: %s", act)
		}

		code, err := g.generateFromTemplate("service_"+act, data, templatePath)
		if err != nil {
			return fmt.Errorf("service %s template error: %v", act, err)
		}

		filename := filepath.Join(dir, fmt.Sprintf("%s_%s.service.go", strings.ToLower(entity), act))
		if err := os.WriteFile(filename, []byte(code), 0644); err != nil {
			return fmt.Errorf("write service %s error: %v", act, err)
		}

		fmt.Printf("✅ Generated service.%s: %s\n", act, filename)
	}

	return nil
}
