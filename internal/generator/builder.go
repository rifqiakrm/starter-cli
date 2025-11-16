package generator

import "fmt"

// GenerateBuilder generates builder files and routes
func (g *Generator) GenerateBuilder(module, version string, tables []string, newModule bool) error {
	fmt.Printf("🚀 Generating builder for module '%s' with tables: %v\n", module, tables)

	if newModule {
		return g.generateNewModule(module, version, tables)
	} else {
		return g.generateIncremental(module, version, tables)
	}
}

// generateNewModule creates complete new module
func (g *Generator) generateNewModule(module, version string, tables []string) error {
	fmt.Printf("🆕 Creating new module '%s'\n", module)

	// Generate complete builder
	if err := g.generateCompleteBuilder(module, version, tables); err != nil {
		return err
	}

	// Generate complete routes
	if err := g.generateCompleteRoutes(module, version, tables); err != nil {
		return err
	}

	//// Generate module-specific cache keys
	//if err := g.generateCacheKeys(module, version, tables); err != nil {
	//	return err
	//}

	return nil
}

// generateIncremental adds to existing module
func (g *Generator) generateIncremental(module, version string, tables []string) error {
	fmt.Printf("📈 Adding to existing module '%s'\n", module)

	// Update builder incrementally
	if err := g.updateBuilderIncremental(module, version, tables); err != nil {
		return fmt.Errorf("update builder error: %v", err)
	}

	// Update routes incrementally
	if err := g.updateRoutesIncremental(module, version, tables); err != nil {
		return fmt.Errorf("update routes error: %v", err)
	}

	return nil
}
