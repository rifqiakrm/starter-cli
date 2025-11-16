package generator

import (
	"fmt"
	"os"
	"strings"
)

// BuilderAnalysis holds analysis results of existing builder
type BuilderAnalysis struct {
	FilePath         string
	Content          string
	ExistingTables   []string
	InsertionPoint   int // Position to insert new tables
	HandlerCallPoint int // Position to find handler calls
}

// AnalyzeBuilder reads and analyzes existing builder file
func (g *Generator) AnalyzeBuilder(module string) (*BuilderAnalysis, error) {
	builderPath := fmt.Sprintf("modules/%s/builder.go", module)

	// Check if builder exists
	if _, err := os.Stat(builderPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("builder file not found: %s", builderPath)
	}

	// Read file content
	content, err := os.ReadFile(builderPath)
	if err != nil {
		return nil, fmt.Errorf("read builder error: %v", err)
	}

	contentStr := string(content)
	analysis := &BuilderAnalysis{
		FilePath: builderPath,
		Content:  contentStr,
	}

	// Extract existing tables
	analysis.ExistingTables = g.extractExistingTables(contentStr)

	// Find insertion points
	analysis.InsertionPoint = g.findEntityInsertionPoint(contentStr)
	analysis.HandlerCallPoint = g.findHandlerCallPoint(contentStr)

	return analysis, nil
}

// extractExistingTables finds all tables in the builder
func (g *Generator) extractExistingTables(content string) []string {
	var tables []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for repository patterns: "userFinderRepo :="
		if strings.Contains(line, "Repo :=") && strings.Contains(line, "repository.New") {
			entity := g.extractEntityFromRepoLine(line)
			if entity != "" && !contains(tables, entity) {
				tables = append(tables, entity)
			}
		}
	}

	return tables
}

// extractEntityFromRepoLine extracts entity name from repository line
func (g *Generator) extractEntityFromRepoLine(line string) string {
	// Example: "userFinderRepo := repository.NewUserFinderRepository(db, cache)"
	if strings.Contains(line, "repository.New") {
		parts := strings.Split(line, "repository.New")
		if len(parts) > 1 {
			repoPart := parts[1]
			// Remove common suffixes to get entity name
			suffixes := []string{"FinderRepository", "CreatorRepository", "UpdaterRepository", "DeleterRepository", "("}
			for _, suffix := range suffixes {
				repoPart = strings.Split(repoPart, suffix)[0]
			}
			return strings.ToLower(repoPart)
		}
	}
	return ""
}

// findEntityInsertionPoint finds where to insert new tables INSIDE the function
func (g *Generator) findEntityInsertionPoint(content string) int {
	// Find the function body start
	funcStart := strings.Index(content, "func Build")
	if funcStart == -1 {
		return -1
	}

	// Find the opening brace of the function
	braceStart := strings.Index(content[funcStart:], "{")
	if braceStart == -1 {
		return -1
	}

	functionBodyStart := funcStart + braceStart + 1

	// Now find the last entity block before handler creation
	lines := strings.Split(content[functionBodyStart:], "\n")

	currentPos := functionBodyStart
	lastEntityEnd := currentPos

	for _, line := range lines {
		line = strings.TrimSpace(line)
		currentPos += len(line) + 1 // +1 for newline

		// Stop when we find handler creation
		if strings.Contains(line, "// Handler") ||
			strings.Contains(line, "handler :=") ||
			strings.Contains(line, "app.New") {
			// Return position at the end of previous line
			return lastEntityEnd
		}

		// If this is an entity line, update lastEntityEnd
		if strings.Contains(line, "Repo :=") ||
			strings.Contains(line, "Svc :=") ||
			strings.Contains(line, "//") && (strings.Contains(line, "Repository") || strings.Contains(line, "Service")) {
			lastEntityEnd = currentPos
		}
	}

	// Fallback: before handler calls at the end
	handlerCallsStart := strings.Index(content, "handler.")
	if handlerCallsStart != -1 {
		return handlerCallsStart
	}

	// Final fallback: before the last closing brace
	lastBrace := strings.LastIndex(content, "}")
	if lastBrace != -1 {
		return lastBrace
	}

	return functionBodyStart
}

// findHandlerCallPoint finds where handler calls are made
func (g *Generator) findHandlerCallPoint(content string) int {
	// Look for handler method calls
	return strings.Index(content, "handler.")
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
