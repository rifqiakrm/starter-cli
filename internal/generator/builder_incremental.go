package generator

import (
	"fmt"
	"os"
	"strings"
)

// updateBuilderIncremental adds new tables to existing builder
func (g *Generator) updateBuilderIncremental(module, version string, newTables []string) error {
	fmt.Printf("📈 Updating existing builder for %s with new tables: %v\n", module, newTables)

	// Analyze existing builder
	analysis, err := g.AnalyzeBuilder(module)
	if err != nil {
		return fmt.Errorf("analyze builder error: %v", err)
	}

	// Filter out tables that already exist
	tablesToAdd := g.filterNewTables(newTables, analysis.ExistingTables)
	if len(tablesToAdd) == 0 {
		fmt.Printf("✅ All tables already exist in builder: %v\n", newTables)
		return nil
	}

	fmt.Printf("🔍 Adding %d new tables: %v\n", len(tablesToAdd), tablesToAdd)

	// Generate code for new tables only
	newCode, err := g.generateIncrementalBuilderCode(module, version, tablesToAdd)
	if err != nil {
		return fmt.Errorf("generate incremental code error: %v", err)
	}

	// Update the builder file
	if err := g.updateBuilderFile(analysis, newCode, tablesToAdd); err != nil {
		return fmt.Errorf("update builder file error: %v", err)
	}

	fmt.Printf("✅ Updated builder with new tables: %v\n", tablesToAdd)
	return nil
}

// filterNewTables returns only tables that don't exist yet
func (g *Generator) filterNewTables(requested, existing []string) []string {
	var newTables []string
	for _, entity := range requested {
		if !contains(existing, entity) {
			newTables = append(newTables, entity)
		}
	}
	return newTables
}

// generateIncrementalBuilderCode generates code only for new tables
func (g *Generator) generateIncrementalBuilderCode(module, version string, tables []string) (string, error) {
	config := g.buildBuilderConfig(module, version, tables)

	var result strings.Builder
	result.WriteString("\n")

	for _, entity := range config.Tables {
		result.WriteString(fmt.Sprintf("\t// %s Repository\n", entity.DisplayName))
		result.WriteString(fmt.Sprintf("\t%sFinderRepo := repository.New%sFinderRepository(db, cache)\n",
			entity.Name, entity.DisplayName))
		result.WriteString(fmt.Sprintf("\t%sCreatorRepo := repository.New%sCreatorRepository(db, cache)\n",
			entity.Name, entity.DisplayName))
		result.WriteString(fmt.Sprintf("\t%sUpdaterRepo := repository.New%sUpdaterRepository(db, cache)\n",
			entity.Name, entity.DisplayName))
		result.WriteString(fmt.Sprintf("\t%sDeleterRepo := repository.New%sDeleterRepository(db, cache)\n",
			entity.Name, entity.DisplayName))

		result.WriteString(fmt.Sprintf("\n\t// %s Service\n", entity.DisplayName))
		result.WriteString(fmt.Sprintf("\t%sCreatorSvc := service.New%sCreator(cfg, %sCreatorRepo, %sFinderRepo, %sUpdaterRepo, cloudStorage)\n",
			entity.Name, entity.DisplayName, entity.Name, entity.Name, entity.Name))
		result.WriteString(fmt.Sprintf("\t%sFinderSvc := service.New%sFinder(cfg, %sFinderRepo, cloudStorage)\n",
			entity.Name, entity.DisplayName, entity.Name))
		result.WriteString(fmt.Sprintf("\t%sUpdaterSvc := service.New%sUpdater(cfg, %sFinderRepo, %sUpdaterRepo, cloudStorage)\n",
			entity.Name, entity.DisplayName, entity.Name, entity.Name))
		result.WriteString(fmt.Sprintf("\t%sDeleterSvc := service.New%sDeleter(cfg, %sDeleterRepo, cloudStorage)\n\n",
			entity.Name, entity.DisplayName, entity.Name))
	}

	return result.String(), nil
}

// updateBuilderFile updates the existing builder file with new tables
func (g *Generator) updateBuilderFile(analysis *BuilderAnalysis, newCode string, tables []string) error {
	lines := strings.Split(analysis.Content, "\n")

	// Step 1: Find safe insertion point for entity wiring (after last complete entity)
	entityInsertLine := g.findEntityWiringInsertionLine(lines)
	if entityInsertLine == -1 {
		return fmt.Errorf("could not find entity wiring insertion point")
	}

	// Step 2: Insert entity wiring
	updatedLines := make([]string, 0, len(lines)+20)
	updatedLines = append(updatedLines, lines[:entityInsertLine]...)
	updatedLines = append(updatedLines, strings.Split(newCode, "\n")...)
	updatedLines = append(updatedLines, lines[entityInsertLine:]...)

	// Step 3: Update handler constructor call
	finalContent := g.updateHandlerConstructor(strings.Join(updatedLines, "\n"), tables)

	// Write updated content
	return os.WriteFile(analysis.FilePath, []byte(finalContent), 0644)
}

// findEntityWiringInsertionLine finds safe place to insert new entity wiring
func (g *Generator) findEntityWiringInsertionLine(lines []string) int {
	// Look for handler creation line
	for i := 0; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if strings.Contains(line, "handler :=") || strings.Contains(line, "app.New") {
			// Go backwards to find a safe insertion point (after last complete entity)
			for j := i - 1; j >= 0; j-- {
				prevLine := strings.TrimSpace(lines[j])
				if prevLine == "" || strings.Contains(prevLine, "//") {
					continue
				}
				// Found a non-empty, non-comment line - insert after this
				return j + 1
			}
			return i
		}
	}
	return len(lines) - 2
}

// updateHandlerConstructor adds new tables to handler constructor parameters
func (g *Generator) updateHandlerConstructor(content string, tables []string) string {
	lines := strings.Split(content, "\n")

	for i, line := range lines {
		if strings.Contains(line, "app.New") && strings.Contains(line, "HTTPHandler(") {
			// Find where to insert new parameters (before cloudStorage or cache)
			for j := i; j < len(lines); j++ {
				if strings.Contains(lines[j], "// Cloud Storage") || strings.Contains(lines[j], "cache,") {
					// Insert new tables before cloudStorage/cache
					newParams := ""
					for _, entity := range tables {
						displayName := toPascalCase(entity)
						newParams += fmt.Sprintf("\t\t// %s\n\t\t%sCreatorSvc, %sFinderSvc, %sUpdaterSvc, %sDeleterSvc,\n",
							displayName, entity, entity, entity, entity)
					}

					// Create updated lines
					updatedLines := make([]string, len(lines)+len(strings.Split(newParams, "\n")))
					copy(updatedLines[:j], lines[:j])
					updatedLines[j] = newParams + lines[j]
					copy(updatedLines[j+1:], lines[j+1:])

					return strings.Join(updatedLines, "\n")
				}
			}
			break
		}
	}

	return content
}
