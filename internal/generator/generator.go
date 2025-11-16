package generator

import (
	"github.com/rifqiakrm/starter-cli/internal/config"
)

// Generator holds the main generation logic
type Generator struct {
	config *config.Config
}

// NewGenerator creates a new generator instance
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config: cfg,
	}
}
