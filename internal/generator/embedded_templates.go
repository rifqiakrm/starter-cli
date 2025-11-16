package generator

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

//go:embed templates/**/*
var templateFS embed.FS

// loadTemplate loads template from embedded filesystem or filesystem
func (g *Generator) loadTemplate(templatePath string) (string, error) {
	// First try to load from filesystem (for development/custom templates)
	if _, err := os.Stat(templatePath); err == nil {
		content, err := os.ReadFile(templatePath)
		if err != nil {
			return "", err
		}
		return string(content), nil
	}

	// Fallback to embedded templates
	// Convert path like "./templates/entity/entity.tmpl" to "templates/entity/entity.tmpl"
	embeddedPath := strings.TrimPrefix(templatePath, "./")
	content, err := templateFS.ReadFile(embeddedPath)
	if err != nil {
		return "", fmt.Errorf("embedded template not found: %s", embeddedPath)
	}

	return string(content), nil
}

// CopyEmbeddedTemplates copies all embedded templates to the filesystem
func CopyEmbeddedTemplates(targetDir string) error {
	// Walk through the embedded template filesystem
	err := fs.WalkDir(templateFS, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip the root directory
		if path == "templates" {
			return nil
		}

		targetPath := filepath.Join(targetDir, strings.TrimPrefix(path, "templates/"))

		if d.IsDir() {
			// Create directory
			return os.MkdirAll(targetPath, 0755)
		} else {
			// Copy file
			content, err := templateFS.ReadFile(path)
			if err != nil {
				return fmt.Errorf("read embedded file %s: %v", path, err)
			}

			if err := os.WriteFile(targetPath, content, 0644); err != nil {
				return fmt.Errorf("write template file %s: %v", targetPath, err)
			}

			fmt.Printf("📄 Created: %s\n", targetPath)
		}

		return nil
	})

	return err
}
