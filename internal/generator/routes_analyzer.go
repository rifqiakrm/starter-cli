package generator

import (
	"fmt"
	"os"
	"strings"
)

// RoutesAnalysis holds analysis results of existing routes file
type RoutesAnalysis struct {
	FilePath       string
	Content        string
	ExistingTables []string
	MethodBlocks   map[string]MethodBlock // "Finder", "Creator", "Updater", "Deleter"
}

// MethodBlock represents a route method block
type MethodBlock struct {
	StartLine int
	EndLine   int
	Content   string
}

// AnalyzeRoutes reads and analyzes existing routes file
func (g *Generator) AnalyzeRoutes(module string) (*RoutesAnalysis, error) {
	routesPath := fmt.Sprintf("app/%s_routes.go", module)

	// Check if routes file exists
	if _, err := os.Stat(routesPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("routes file not found: %s", routesPath)
	}

	// Read file content
	content, err := os.ReadFile(routesPath)
	if err != nil {
		return nil, fmt.Errorf("read routes error: %v", err)
	}

	contentStr := string(content)
	analysis := &RoutesAnalysis{
		FilePath: routesPath,
		Content:  contentStr,
	}

	// Extract existing tables and method blocks
	analysis.ExistingTables = g.extractExistingRoutesTables(contentStr)
	analysis.MethodBlocks = g.extractMethodBlocks(contentStr)

	return analysis, nil
}

// extractExistingRoutesTables finds all tables in the routes file
func (g *Generator) extractExistingRoutesTables(content string) []string {
	var tables []string
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for handler assignments: "userHnd :="
		if strings.Contains(line, "Hnd :=") && strings.Contains(line, "handlerv1.New") {
			entity := g.extractEntityFromHandlerLine(line)
			if entity != "" && !contains(tables, entity) {
				tables = append(tables, entity)
			}
		}
	}

	return tables
}

// extractEntityFromHandlerLine extracts entity name from handler line
func (g *Generator) extractEntityFromHandlerLine(line string) string {
	// Example: "userHnd := authhandlerv1.NewUserFinderHandler"
	if strings.Contains(line, "handlerv1.New") {
		parts := strings.Split(line, "handlerv1.New")
		if len(parts) > 1 {
			handlerPart := parts[1]
			// Remove method suffix to get entity name
			suffixes := []string{"FinderHandler", "CreatorHandler", "UpdaterHandler", "DeleterHandler", "("}
			for _, suffix := range suffixes {
				handlerPart = strings.Split(handlerPart, suffix)[0]
			}
			return strings.ToLower(handlerPart)
		}
	}
	return ""
}

// extractMethodBlocks extracts the four main route method blocks
func (g *Generator) extractMethodBlocks(content string) map[string]MethodBlock {
	blocks := make(map[string]MethodBlock)
	lines := strings.Split(content, "\n")

	currentMethod := ""
	startLine := -1

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Look for method function definitions
		if strings.HasPrefix(trimmed, "func (h *") && strings.Contains(trimmed, "HTTPHandler() {") {
			if strings.Contains(trimmed, "FinderHTTPHandler") {
				currentMethod = "Finder"
			} else if strings.Contains(trimmed, "CreatorHTTPHandler") {
				currentMethod = "Creator"
			} else if strings.Contains(trimmed, "UpdaterHTTPHandler") {
				currentMethod = "Updater"
			} else if strings.Contains(trimmed, "DeleterHTTPHandler") {
				currentMethod = "Deleter"
			}

			if currentMethod != "" {
				startLine = i
			}
		}

		// Look for closing brace of method
		if startLine != -1 && trimmed == "}" {
			// Check if this is the end of our method (next line is empty or next method)
			if i+1 >= len(lines) ||
				strings.TrimSpace(lines[i+1]) == "" ||
				strings.Contains(lines[i+1], "func (h *") {

				blocks[currentMethod] = MethodBlock{
					StartLine: startLine,
					EndLine:   i,
					Content:   strings.Join(lines[startLine:i+1], "\n"),
				}
				currentMethod = ""
				startLine = -1
			}
		}
	}

	return blocks
}
