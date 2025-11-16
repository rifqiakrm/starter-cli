package generator

import (
	"fmt"
	"os"
	"strings"
)

// generatePermissionConstants generates permission constants for tables
func (g *Generator) generatePermissionConstants(module string, tables []string) error {
	fmt.Printf("🔐 Generating permission constants for %s tables: %v\n", module, tables)

	permissionFilePath := "common/constant/permission.go"

	// Check if permission file exists
	if _, err := os.Stat(permissionFilePath); os.IsNotExist(err) {
		return fmt.Errorf("permission file not found: %s", permissionFilePath)
	}

	// Read existing permission file
	content, err := os.ReadFile(permissionFilePath)
	if err != nil {
		return fmt.Errorf("read permission file error: %v", err)
	}

	contentStr := string(content)

	// Generate new permission constants
	newPermissions := g.generatePermissionConstantsForModule(module, tables)

	// Find insertion point and update content
	updatedContent := g.insertPermissionConstants(contentStr, module, newPermissions)

	// Write updated content
	if err := os.WriteFile(permissionFilePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("write permission file error: %v", err)
	}

	fmt.Printf("✅ Updated permission constants for %s in: %s\n", module, permissionFilePath)
	return nil
}

// generatePermissionConstantsForModule generates permission constants for a module's tables
func (g *Generator) generatePermissionConstantsForModule(module string, tables []string) string {
	var result strings.Builder

	// Start the module permission block
	result.WriteString(fmt.Sprintf("// %s permissions\n", toPascalCase(tables[0])))
	result.WriteString("const (\n")

	for _, entity := range tables {
		displayName := toPascalCase(entity)

		// Generate permissions for CRUD operations
		permissions := []struct {
			suffix string
			action string
			desc   string
		}{
			{"View", "view", fmt.Sprintf("allows viewing %s", entity)},
			{"Create", "create", fmt.Sprintf("allows creating %s", entity)},
			{"Update", "update", fmt.Sprintf("allows updating %s", entity)},
			{"Delete", "delete", fmt.Sprintf("allows deleting %s", entity)},
			{"List", "list", fmt.Sprintf("allows listing %s", entity)},
			{"Manage", "manage", fmt.Sprintf("allows managing %s", entity)},
		}

		for _, perm := range permissions {
			constName := fmt.Sprintf("Perm%s%s", displayName, perm.suffix)
			permissionKey := fmt.Sprintf("%s:%s", entity, perm.action)

			result.WriteString(fmt.Sprintf("\t// %s %s\n", constName, perm.desc))
			result.WriteString(fmt.Sprintf("\t%s = \"%s\"\n\n", constName, permissionKey))
		}
	}

	// Close the const block
	result.WriteString(")\n\n")

	return result.String()
}

// insertPermissionConstants inserts new permission constants in the appropriate section
func (g *Generator) insertPermissionConstants(content, module, newPermissions string) string {
	lines := strings.Split(content, "\n")

	// Find the position to insert - after the last existing module section
	insertPosition := -1
	currentModule := ""

	for i, line := range lines {
		// Look for module comment sections like "// Course permissions"
		if strings.HasPrefix(line, "// ") && strings.Contains(line, "permissions") {
			currentModule = strings.TrimPrefix(line, "// ")
			currentModule = strings.TrimSuffix(currentModule, " permissions")
			currentModule = strings.TrimSpace(currentModule)
		}

		// Look for the System permissions section (usually last) to insert before it
		if currentModule == "System" {
			insertPosition = i
			break
		}
	}

	// If System permissions not found, insert before the last closing brace or at the end
	if insertPosition == -1 {
		// Find the last closing brace of const blocks
		for i := len(lines) - 1; i >= 0; i-- {
			if strings.TrimSpace(lines[i]) == ")" {
				insertPosition = i + 1 // Insert after the closing brace
				break
			}
		}

		// If still not found, append at the end
		if insertPosition == -1 {
			insertPosition = len(lines)
		}
	}

	// Insert the new permissions
	updatedLines := make([]string, 0, len(lines)+strings.Count(newPermissions, "\n"))
	updatedLines = append(updatedLines, lines[:insertPosition]...)
	updatedLines = append(updatedLines, strings.Split(newPermissions, "\n")...)
	updatedLines = append(updatedLines, lines[insertPosition:]...)

	return strings.Join(updatedLines, "\n")
}
