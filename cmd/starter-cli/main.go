package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rifqiakrm/starter-cli/internal/config"
	"github.com/rifqiakrm/starter-cli/internal/generator"
	"github.com/rifqiakrm/starter-cli/internal/types"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "all", "entity", "resource", "module":
		runGenerator(command)
	case "builder": // NEW COMMAND
		runBuilder()
	case "init":
		initTemplates()
	case "help", "-h", "--help": // ADD HELP COMMAND HANDLING
		printUsage()
	case "version":
		fmt.Println("starter-cli v1.0.0")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runGenerator(command string) {
	fs := flag.NewFlagSet(command, flag.ExitOnError)

	// Common flags
	schema := fs.String("schema", "public", "Schema name")
	table := fs.String("table", "", "Table name (required for entity/resource/all)")
	version := fs.String("version", "v1", "API version")
	templateDir := fs.String("template-dir", "", "Custom template directory")
	configFile := fs.String("config", "", "Config file path")
	migrations := fs.String("migrations", "./db/migrations", "Migrations path")

	// Output directories
	entityOut := fs.String("entity-out", "./modules/%s/entity", "Entity output directory")
	resourceOut := fs.String("resource-out", "./modules/%s/resource", "Resource output directory")
	moduleOut := fs.String("module-out", "./modules", "Module output directory")

	// Enhanced module parts
	moduleParts := fs.String("parts", "handler,service,repository", "Module parts to generate")

	err := fs.Parse(os.Args[2:])
	if err != nil {
		return
	}

	// Load configuration
	cfg, err := config.Load(*configFile, *templateDir)
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Validate inputs
	if command != "module" && *table == "" {
		log.Fatal("Missing required flag: --table")
	}
	if command == "module" && *table == "" {
		log.Fatal("Missing required flag: --table")
	}

	// Parse module parts for module/all commands
	var parts []types.ModulePart
	if command == "module" || command == "all" {
		parts = parseModuleParts(*moduleParts)
	}

	// Run the appropriate generator
	gen := generator.NewGenerator(cfg)

	switch command {
	case "init":
		initTemplates()
	case "help":
		printUsage()
	case "all":
		// Use table name directly as entity name
		entityName := strings.ToLower(*table)
		if err := gen.GenerateAll(*schema, *table, entityName, *version, *migrations, *entityOut, *resourceOut, *moduleOut, parts); err != nil {
			log.Fatalf("Generate all error: %v", err)
		}
	case "entity":
		if err := gen.GenerateEntity(*schema, *table, *migrations, *entityOut); err != nil {
			log.Fatalf("Generate entity error: %v", err)
		}
	case "resource":
		if err := gen.GenerateResource(*schema, *table, *migrations, *resourceOut); err != nil {
			log.Fatalf("Generate resource error: %v", err)
		}
	case "module":
		// Use table name directly as entity name
		entityName := strings.ToLower(*table)
		if err := gen.GenerateModule(*schema, entityName, *version, parts, *moduleOut); err != nil {
			log.Fatalf("Generate module error: %v", err)
		}
	case "version":
		fmt.Println("starter-cli v1.0.0")
	}
}

func runBuilder() {
	fs := flag.NewFlagSet("builder", flag.ExitOnError)

	// Builder-specific flags
	module := fs.String("module", "", "Module name (e.g., auth, inventory, listings)")
	tables := fs.String("tables", "", "Comma-separated table names (e.g., users,roles,permissions)") // CHANGED: tables -> tables
	version := fs.String("version", "v1", "API version")
	newModule := fs.Bool("new-module", false, "Generate complete new module")
	dryRun := fs.Bool("dry-run", false, "Show what will be generated without writing files")

	_ = fs.Parse(os.Args[2:])

	// Validate required flags
	if *module == "" || *tables == "" {
		log.Fatal("Missing required flags: --module and --tables")
	}

	// Parse tables
	tableList := strings.Split(*tables, ",")
	for i, table := range tableList {
		tableList[i] = strings.TrimSpace(table)
	}

	// Load configuration
	cfg, err := config.Load("", "")
	if err != nil {
		log.Fatalf("Config error: %v", err)
	}

	// Run builder generator
	gen := generator.NewGenerator(cfg)

	if *dryRun {
		fmt.Printf("🧪 DRY RUN: Would generate builder for module=%s, tables=%v, version=%s, new-module=%t\n",
			*module, tableList, *version, *newModule)
		return
	}

	if err := gen.GenerateBuilder(*module, *version, tableList, *newModule); err != nil {
		log.Fatalf("Builder generation error: %v", err)
	}
}

func parseModuleParts(partsStr string) []types.ModulePart {
	var parts []types.ModulePart

	for _, part := range strings.Split(partsStr, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		if strings.Contains(part, ".") {
			components := strings.Split(part, ".")
			if len(components) == 2 {
				parts = append(parts, types.ModulePart{
					Component: components[0],
					Action:    components[1],
				})
			}
		} else {
			parts = append(parts, types.ModulePart{
				Component: part,
				Action:    "", // empty means all actions
			})
		}
	}

	return parts
}

func printUsage() {
	fmt.Println(`Usage: starter-cli <command> [flags]

Commands:
  init      Initialize template directory for customization
  all       Generate entity, resource, and module components
  entity    Generate entity only from database table
  resource  Generate resource (DTO) only from database table
  module    Generate module components (handler, service, repository)
  builder   Generate builder and routes for modules
  help      See usage information
  version   Show version information

Common Flags:
  --schema         Database schema name (default: public)
  --table          Table name (required for entity/resource/module/all)
  --tables         Comma-separated table names (for builder command)
  --version        API version (default: v1)
  --parts          Module parts to generate: handler,service,repository (default: all)
  --template-dir   Custom template directory (overrides embedded templates)
  --migrations     Path to database migrations (default: ./db/migrations)

Template Customization:
  # Initialize template directory for customization
  starter-cli init
  
  # Use custom templates (overrides embedded templates)
  starter-cli all --schema=auth --table=users --version=v1 --template-dir=./templates

Examples:
  # Generate complete stack for new table
  starter-cli all --schema=auth --table=users --version=v1
  # if it's existing module then
  starter-cli builder --module=auth --tables=users --version=v1
  # if it's new module then
  starter-cli builder --module=auth --tables=users --version=v1 --new-module

  # Generate only entity from existing table
  starter-cli entity --schema=inventory --table=products

  # Generate resource DTOs from table
  starter-cli resource --schema=inventory --table=products

  # Generate specific module components
  starter-cli module --schema=inventory --table=categories --version=v1 --parts=handler,service

  # Create complete new module with multiple tables
  starter-cli builder --module=auth --tables=users,roles,permissions --version=v1 --new-module

  # Add tables to existing module
  starter-cli builder --module=auth --tables=organizations --version=v1

  # Preview changes without writing files
  starter-cli builder --module=auth --tables=organizations --dry-run

Module Parts Syntax:
  --parts=handler                    # All handler actions
  --parts=handler.creator            # Only creator handler
  --parts=handler.finder,service     # Finder handler + all services
  --parts=repository.creator,repository.updater  # Specific repository actions

Output:
  • Entities:   ./modules/{schema}/entity/
  • Resources:  ./modules/{schema}/resource/
  • Modules:    ./modules/{schema}/{version}/
  • Builder:    ./modules/{module/schema}/builder.go
  • Routes:     ./app/{module/schema}_routes.go`)
}

func initTemplates() {
	fmt.Println("🚀 Initializing template directory...")

	templateDir := "./templates"

	// Create the main template directory
	if err := os.MkdirAll(templateDir, 0755); err != nil {
		log.Fatalf("Failed to create template directory: %v", err)
	}

	// Copy all embedded templates to filesystem
	err := generator.CopyEmbeddedTemplates(templateDir)
	if err != nil {
		log.Fatalf("Failed to copy templates: %v", err)
	}

	fmt.Printf("✅ Template directory initialized at: %s\n", templateDir)
	fmt.Println("📝 You can now customize the templates and use --template-dir flag")
}
