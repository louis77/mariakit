package schema

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// JSONMapping represents a custom type mapping for JSON columns
type JSONMapping struct {
	Type   string `yaml:"type"`
	Import string `yaml:"import,omitempty"`
}

// Config represents the configuration file structure
type Config struct {
	JSONMappings map[string]JSONMapping `yaml:"json_mappings"`
}

// LoadConfig loads configuration from a YAML file
func LoadConfig(configPath string) (*Config, error) {
	// Return empty config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &Config{JSONMappings: make(map[string]JSONMapping)}, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Initialize map if nil
	if config.JSONMappings == nil {
		config.JSONMappings = make(map[string]JSONMapping)
	}

	return &config, nil
}

// GetJSONMapping returns the custom JSON mapping for a table.column combination
func (c *Config) GetJSONMapping(tableName, columnName string) (JSONMapping, bool) {
	key := fmt.Sprintf("%s.%s", tableName, columnName)
	mapping, exists := c.JSONMappings[key]
	return mapping, exists
}

// GetRequiredImports returns all unique import paths needed for JSON mappings
func (c *Config) GetRequiredImports() []string {
	imports := make(map[string]bool)
	for _, mapping := range c.JSONMappings {
		if mapping.Import != "" {
			imports[mapping.Import] = true
		}
	}

	var result []string
	for imp := range imports {
		result = append(result, imp)
	}
	return result
}
