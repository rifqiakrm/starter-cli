package generator

import (
	"fmt"
	"os"
	"strings"
)

// updateRoutesIncremental adds new tables to existing routes
func (g *Generator) updateRoutesIncremental(module, version string, newTables []string) error {
	fmt.Printf("🛣️  Updating existing routes for %s with new tables: %v\n", module, newTables)

	// Analyze existing routes
	analysis, err := g.AnalyzeRoutes(module)
	if err != nil {
		return fmt.Errorf("analyze routes error: %v", err)
	}

	// Filter out tables that already exist in routes
	tablesToAdd := g.filterNewTables(newTables, analysis.ExistingTables)
	if len(tablesToAdd) == 0 {
		fmt.Printf("✅ All tables already exist in routes: %v\n", newTables)
		return nil
	}

	fmt.Printf("🔍 Adding %d new tables to routes: %v\n", len(tablesToAdd), tablesToAdd)

	// Update each of the 4 route methods
	updatedContent := analysis.Content
	for method := range analysis.MethodBlocks {
		updatedContent = g.updateRouteMethod(updatedContent, module, method, tablesToAdd)
	}

	// Update the handler struct fields
	updatedContent = g.updateHandlerStruct(updatedContent, module, tablesToAdd)

	// Update the handler constructor
	updatedContent = g.updateRouteHandlerConstructor(updatedContent, module, tablesToAdd)

	// Write updated routes file
	if err := os.WriteFile(analysis.FilePath, []byte(updatedContent), 0644); err != nil {
		return fmt.Errorf("write routes error: %v", err)
	}

	// Generate permission constants for new tables
	if err := g.generatePermissionConstants(module, tablesToAdd); err != nil {
		fmt.Printf("⚠️  Permission constants generation skipped: %v\n", err)
		// Don't fail the whole process if permission generation fails
	}

	fmt.Printf("✅ Updated routes with new tables: %v\n", tablesToAdd)
	return nil
}

// updateRouteMethod - simpler approach: insert before v1's closing brace
func (g *Generator) updateRouteMethod(content, module, method string, tables []string) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Find the specific method function
		if strings.Contains(line, method+"HTTPHandler() {") {
			// STEP 1: First add handler declarations for new tables
			contentWithDeclarations := g.addHandlerDeclarations(content, module, method, tables)
			lines = strings.Split(contentWithDeclarations, "\n")
			// Find the v1 group's CLOSING brace in THIS method
			for j := i; j < len(lines); j++ {
				// Look for v1's closing brace (proper indentation)
				if strings.TrimSpace(lines[j]) == "}" &&
					strings.HasPrefix(lines[j], "\t}") {
					// This is v1's closing brace - insert routes BEFORE it
					newRoutes := g.generateEntityRoutes(module, method, tables)

					updatedLines := make([]string, len(lines)+len(strings.Split(newRoutes, "\n")))
					copy(updatedLines[:j], lines[:j])
					updatedLines[j] = newRoutes + "\n" + lines[j] // Add before closing brace
					copy(updatedLines[j+1:], lines[j+1:])

					return strings.Join(updatedLines, "\n")
				}
			}
			break
		}
	}

	return content
}

// addHandlerDeclarations adds handler variable declarations to route methods
func (g *Generator) addHandlerDeclarations(content, module, method string, tables []string) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.Contains(line, method+"HTTPHandler() {") {
			// Find where to insert declarations (before v1 group)
			for j := i; j < len(lines); j++ {
				if strings.Contains(lines[j], "v1 :=") && strings.Contains(lines[j], "Group(") {
					// Insert handler declarations BEFORE the v1 group
					newDeclarations := g.generateHandlerDeclarations(module, method, tables)

					updatedLines := make([]string, len(lines)+len(strings.Split(newDeclarations, "\n")))
					copy(updatedLines[:j], lines[:j])
					updatedLines[j] = newDeclarations + lines[j]
					copy(updatedLines[j+1:], lines[j+1:])

					return strings.Join(updatedLines, "\n")
				}
			}
			break
		}
	}

	return content
}

// generateHandlerDeclarations generates handler variable declarations
func (g *Generator) generateHandlerDeclarations(module, method string, tables []string) string {
	var result strings.Builder

	for _, entity := range tables {
		displayName := toPascalCase(entity)
		result.WriteString(fmt.Sprintf("\t%sHnd := %shandlerv1.New%s%sHandler(",
			entity, module, displayName, method))

		switch method {
		case "Finder":
			result.WriteString(fmt.Sprintf("h.%sFinder)\n", entity))
		case "Creator":
			result.WriteString(fmt.Sprintf("h.%sCreator, h.cloudStorage)\n", entity))
		case "Updater":
			result.WriteString(fmt.Sprintf("h.%sUpdater, h.cloudStorage)\n", entity))
		case "Deleter":
			result.WriteString(fmt.Sprintf("h.%sDeleter)\n", entity))
		}
	}

	return result.String()
}

// generateEntityRoutes generates route code for new tables in a method
func (g *Generator) generateEntityRoutes(module, method string, tables []string) string {
	var result strings.Builder

	for _, entity := range tables {
		displayName := toPascalCase(entity)
		routePath := g.getRoutePath(entity)
		groupName := g.getGroupName(entity)
		pluralDisplayName := g.getPluralDisplayName(entity) // NEW: Get plural for method names

		result.WriteString("\n\t\t")
		result.WriteString(groupName)
		result.WriteString(" := v1.Group(\"/")
		result.WriteString(routePath)
		result.WriteString("\", middleware.RequirePermission(\n\t\t\tconstant.Perm")
		result.WriteString(displayName)
		result.WriteString(g.getPermissionAction(method))
		result.WriteString(",\n\t\t\tconstant.PermSystemManage,\n\t\t))\n\t\t{\n")

		switch method {
		case "Finder":
			result.WriteString("\t\t\t")
			result.WriteString(groupName)
			result.WriteString(".GET(\"\", ")
			result.WriteString(entity)
			result.WriteString("Hnd.GetAll")
			result.WriteString(pluralDisplayName) // Use plural for method name
			result.WriteString(")\n\t\t\t")
			result.WriteString(groupName)
			result.WriteString(".GET(\"/:id\", ")
			result.WriteString(entity)
			result.WriteString("Hnd.Get")
			result.WriteString(displayName)
			result.WriteString("ByID)\n")
		case "Creator":
			result.WriteString("\t\t\t")
			result.WriteString(groupName)
			result.WriteString(".POST(\"\", ")
			result.WriteString(entity)
			result.WriteString("Hnd.Create")
			result.WriteString(displayName)
			result.WriteString(")\n")
		case "Updater":
			result.WriteString("\t\t\t")
			result.WriteString(groupName)
			result.WriteString(".PUT(\"/:id\", ")
			result.WriteString(entity)
			result.WriteString("Hnd.Update")
			result.WriteString(displayName)
			result.WriteString(")\n")
		case "Deleter":
			result.WriteString("\t\t\t")
			result.WriteString(groupName)
			result.WriteString(".DELETE(\"/:id\", ")
			result.WriteString(entity)
			result.WriteString("Hnd.Delete")
			result.WriteString(displayName)
			result.WriteString("ByID)\n")
		}

		result.WriteString("\t\t}\n")
	}

	return result.String()
}

// getPluralDisplayName returns the plural form for method names
func (g *Generator) getPluralDisplayName(entity string) string {
	// Handle special cases for method names
	switch entity {
	case "category":
		return "Categories" // "GetAllCategories" not "GetAllCategorys"
	case "product":
		return "Products"
	case "user":
		return "Users"
	case "role":
		return "Roles"
	case "permission":
		return "Permissions"
	case "organization":
		return "Organizations"
	case "person":
		return "People"
	case "child":
		return "Children"
	case "man":
		return "Men"
	case "woman":
		return "Women"
	default:
		// Default pluralization rules
		if strings.HasSuffix(entity, "y") {
			return strings.TrimSuffix(toPascalCase(entity), "y") + "ies"
		} else if strings.HasSuffix(entity, "s") ||
			strings.HasSuffix(entity, "x") ||
			strings.HasSuffix(entity, "z") ||
			strings.HasSuffix(entity, "ch") ||
			strings.HasSuffix(entity, "sh") {
			return toPascalCase(entity) + "es"
		} else {
			return toPascalCase(entity) + "s"
		}
	}
}

// getPermissionAction returns the permission action for a method
func (g *Generator) getPermissionAction(method string) string {
	switch method {
	case "Finder":
		return "View"
	case "Creator":
		return "Create"
	case "Updater":
		return "Update"
	case "Deleter":
		return "Delete"
	default:
		return "View"
	}
}

// updateHandlerStruct adds new entity fields to handler struct
func (g *Generator) updateHandlerStruct(content, module string, tables []string) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.Contains(line, "type") && strings.Contains(line, "HTTPHandler struct {") {
			// Find where to insert new fields (before cloudStorage)
			for j := i; j < len(lines); j++ {
				if strings.Contains(lines[j], "cloudStorage") || strings.Contains(lines[j], "cache") {
					newFields := g.generateHandlerFields(module, tables)

					updatedLines := make([]string, len(lines)+len(strings.Split(newFields, "\n")))
					copy(updatedLines[:j], lines[:j])
					updatedLines[j] = newFields + lines[j]
					copy(updatedLines[j+1:], lines[j+1:])

					return strings.Join(updatedLines, "\n")
				}
			}
			break
		}
	}

	return content
}

// generateHandlerFields generates handler struct fields for new tables
func (g *Generator) generateHandlerFields(module string, tables []string) string {
	var result strings.Builder

	for _, entity := range tables {
		displayName := toPascalCase(entity)
		result.WriteString(fmt.Sprintf("\t%sCreator %sservicev1.%sCreatorUseCase\n",
			entity, module, displayName))
		result.WriteString(fmt.Sprintf("\t%sFinder %sservicev1.%sFinderUseCase\n",
			entity, module, displayName))
		result.WriteString(fmt.Sprintf("\t%sUpdater %sservicev1.%sUpdaterUseCase\n",
			entity, module, displayName))
		result.WriteString(fmt.Sprintf("\t%sDeleter %sservicev1.%sDeleterUseCase\n",
			entity, module, displayName))
	}

	return result.String()
}

// updateRouteHandlerConstructor adds new entity parameters to handler constructor in routes
func (g *Generator) updateRouteHandlerConstructor(content, module string, tables []string) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Find the handler constructor function
		if strings.Contains(line, "func New") && strings.Contains(line, "HTTPHandler(") {
			// Find where to insert new parameters (before cloudStorage or cache)
			for j := i; j < len(lines); j++ {
				if strings.Contains(lines[j], "cloudStorage") || strings.Contains(lines[j], "cache") {
					// Generate new parameters for the constructor
					newParams := g.generateHandlerConstructorParams(module, tables)

					// Insert before cloudStorage/cache
					updatedLines := make([]string, len(lines)+len(strings.Split(newParams, "\n")))
					copy(updatedLines[:j], lines[:j])
					updatedLines[j] = newParams + lines[j]
					copy(updatedLines[j+1:], lines[j+1:])

					content = strings.Join(updatedLines, "\n")

					// Now update the constructor body (field assignments)
					return g.updateHandlerConstructorBody(content, module, tables)
				}
			}
			break
		}
	}

	return content
}

// generateHandlerConstructorParams generates constructor parameters for new tables
func (g *Generator) generateHandlerConstructorParams(module string, tables []string) string {
	var result strings.Builder

	for _, entity := range tables {
		displayName := toPascalCase(entity)
		result.WriteString(fmt.Sprintf("\t%sCreator %sservicev1.%sCreatorUseCase,\n",
			entity, module, displayName))
		result.WriteString(fmt.Sprintf("\t%sFinder %sservicev1.%sFinderUseCase,\n",
			entity, module, displayName))
		result.WriteString(fmt.Sprintf("\t%sUpdater %sservicev1.%sUpdaterUseCase,\n",
			entity, module, displayName))
		result.WriteString(fmt.Sprintf("\t%sDeleter %sservicev1.%sDeleterUseCase,\n",
			entity, module, displayName))
	}

	return result.String()
}

// updateHandlerConstructorBody updates the constructor body with new field assignments
func (g *Generator) updateHandlerConstructorBody(content, module string, tables []string) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		// Find the return statement in constructor
		if strings.Contains(line, "return &") && strings.Contains(line, "HTTPHandler{") {
			// Find where to insert new field assignments (before cloudStorage or cache)
			for j := i; j < len(lines); j++ {
				if strings.Contains(lines[j], "cloudStorage:") || strings.Contains(lines[j], "cache:") {
					// Generate new field assignments
					newAssignments := g.generateHandlerFieldAssignments(tables)

					// Insert before cloudStorage/cache
					updatedLines := make([]string, len(lines)+len(strings.Split(newAssignments, "\n")))
					copy(updatedLines[:j], lines[:j])
					updatedLines[j] = newAssignments + lines[j]
					copy(updatedLines[j+1:], lines[j+1:])

					return strings.Join(updatedLines, "\n")
				}
			}
			break
		}
	}

	return content
}

// generateHandlerFieldAssignments generates field assignments for new tables
func (g *Generator) generateHandlerFieldAssignments(tables []string) string {
	var result strings.Builder

	for _, entity := range tables {
		result.WriteString(fmt.Sprintf("\t\t%sCreator: %sCreator,\n", entity, entity))
		result.WriteString(fmt.Sprintf("\t\t%sFinder: %sFinder,\n", entity, entity))
		result.WriteString(fmt.Sprintf("\t\t%sUpdater: %sUpdater,\n", entity, entity))
		result.WriteString(fmt.Sprintf("\t\t%sDeleter: %sDeleter,\n", entity, entity))
	}

	return result.String()
}

// getGroupName returns the proper pluralized group name
func (g *Generator) getGroupName(entity string) string {
	// Handle special cases
	switch entity {
	case "category":
		return "categories"
	case "product":
		return "products"
	case "user":
		return "users"
	case "role":
		return "roles"
	case "permission":
		return "permissions"
	case "organization":
		return "organizations"
	case "person":
		return "people"
	case "child":
		return "children"
	case "man":
		return "men"
	case "woman":
		return "women"
	default:
		// Default pluralization rules
		if strings.HasSuffix(entity, "y") {
			return strings.TrimSuffix(entity, "y") + "ies"
		} else if strings.HasSuffix(entity, "s") ||
			strings.HasSuffix(entity, "x") ||
			strings.HasSuffix(entity, "z") ||
			strings.HasSuffix(entity, "ch") ||
			strings.HasSuffix(entity, "sh") {
			return entity + "es"
		} else {
			return entity + "s"
		}
	}
}
