package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Config holds all template paths and generator configuration
type Config struct {
	TemplatePaths TemplatePaths `yaml:"template_paths"`
}

// TemplatePaths defines all customizable template file paths
type TemplatePaths struct {
	// Entity templates
	Entity string `yaml:"entity"`

	// Resource templates
	Resource      string `yaml:"resource"`
	CreateRequest string `yaml:"resource_create"`
	UpdateRequest string `yaml:"resource_update"`

	// Handler templates
	HandlerCreator string `yaml:"handler_creator"`
	HandlerFinder  string `yaml:"handler_finder"`
	HandlerUpdater string `yaml:"handler_updater"`
	HandlerDeleter string `yaml:"handler_deleter"`

	// Service templates
	ServiceCreator string `yaml:"service_creator"`
	ServiceFinder  string `yaml:"service_finder"`
	ServiceUpdater string `yaml:"service_updater"`
	ServiceDeleter string `yaml:"service_deleter"`

	// Repository templates
	RepositoryCreator string `yaml:"repository_creator"`
	RepositoryFinder  string `yaml:"repository_finder"`
	RepositoryUpdater string `yaml:"repository_updater"`
	RepositoryDeleter string `yaml:"repository_deleter"`

	// Builder templates
	Builder string `yaml:"builder"`
	Routes  string `yaml:"routes"`
}

// Load loads configuration from file, with fallbacks
func Load(configPath, templateDir string) (*Config, error) {
	cfg := &Config{}

	// Try to load from config file
	if configPath != "" {
		data, err := os.ReadFile(configPath)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	}

	// Set default template paths if not configured
	setDefaultTemplatePaths(cfg, templateDir)

	return cfg, nil
}

func setDefaultTemplatePaths(cfg *Config, templateDir string) {
	baseDir := templateDir
	if baseDir == "" {
		baseDir = "./templates"
	}

	paths := &cfg.TemplatePaths

	// Entity templates
	if paths.Entity == "" {
		paths.Entity = filepath.Join(baseDir, "entity/entity.tmpl")
	}

	// Resource templates
	if paths.Resource == "" {
		paths.Resource = filepath.Join(baseDir, "resource/resource.tmpl")
	}
	if paths.CreateRequest == "" {
		paths.CreateRequest = filepath.Join(baseDir, "resource/create_request.tmpl")
	}
	if paths.UpdateRequest == "" {
		paths.UpdateRequest = filepath.Join(baseDir, "resource/update_request.tmpl")
	}

	// Handler templates
	if paths.HandlerCreator == "" {
		paths.HandlerCreator = filepath.Join(baseDir, "module/handler/creator.tmpl")
	}
	if paths.HandlerFinder == "" {
		paths.HandlerFinder = filepath.Join(baseDir, "module/handler/finder.tmpl")
	}
	if paths.HandlerUpdater == "" {
		paths.HandlerUpdater = filepath.Join(baseDir, "module/handler/updater.tmpl")
	}
	if paths.HandlerDeleter == "" {
		paths.HandlerDeleter = filepath.Join(baseDir, "module/handler/deleter.tmpl")
	}

	// Service templates
	if paths.ServiceCreator == "" {
		paths.ServiceCreator = filepath.Join(baseDir, "module/service/creator.tmpl")
	}
	if paths.ServiceFinder == "" {
		paths.ServiceFinder = filepath.Join(baseDir, "module/service/finder.tmpl")
	}
	if paths.ServiceUpdater == "" {
		paths.ServiceUpdater = filepath.Join(baseDir, "module/service/updater.tmpl")
	}
	if paths.ServiceDeleter == "" {
		paths.ServiceDeleter = filepath.Join(baseDir, "module/service/deleter.tmpl")
	}

	// Repository templates
	if paths.RepositoryCreator == "" {
		paths.RepositoryCreator = filepath.Join(baseDir, "module/repository/creator.tmpl")
	}
	if paths.RepositoryFinder == "" {
		paths.RepositoryFinder = filepath.Join(baseDir, "module/repository/finder.tmpl")
	}
	if paths.RepositoryUpdater == "" {
		paths.RepositoryUpdater = filepath.Join(baseDir, "module/repository/updater.tmpl")
	}
	if paths.RepositoryDeleter == "" {
		paths.RepositoryDeleter = filepath.Join(baseDir, "module/repository/deleter.tmpl")
	}

	// Builder templates
	if paths.Builder == "" {
		paths.Builder = filepath.Join(baseDir, "builder/builder.tmpl")
	}
	if paths.Routes == "" {
		paths.Routes = filepath.Join(baseDir, "routes/routes.tmpl")
	}
}
