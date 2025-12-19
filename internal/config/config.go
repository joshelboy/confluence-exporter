package config

import (
	"encoding/json"
	"os"
)

// Config holds all the configuration for the application
type Config struct {
	Confluence ConfluenceConfig `json:"confluence"`
	Export     ExportConfig     `json:"export"`
	Logging    LoggingConfig    `json:"logging"`
}

// ConfluenceConfig holds Confluence API connection settings
type ConfluenceConfig struct {
	BaseURL  string `json:"baseUrl"`
	APIToken string `json:"apiToken"`
	Username string `json:"username"`
}

// ExportConfig holds settings for the export process
type ExportConfig struct {
	SpaceKey           string       `json:"spaceKey"`
	PageID             string       `json:"pageId"`
	OutputDir          string       `json:"outputDir"`
	OutputType         string       `json:"outputType"`
	Recursive          bool         `json:"recursive"`
	IncludeAttachments bool         `json:"includeAttachments"`
	ConcurrentRequests int          `json:"concurrentRequests"`
	Format             FormatConfig `json:"format"`
}

// FormatConfig holds settings for markdown formatting
type FormatConfig struct {
	IncludeFrontMatter bool `json:"includeFrontMatter"`
	PreserveLinks      bool `json:"preserveLinks"`
}

// LoggingConfig holds logging settings
type LoggingConfig struct {
	Level string `json:"level"`
	File  string `json:"file"`
}

// LoadConfig reads the config file from the specified path
func LoadConfig(path string) (*Config, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(file, &config); err != nil {
		return nil, err
	}

	// Set default values
	if config.Export.OutputType == "" {
		config.Export.OutputType = "file"
	}
	if config.Export.OutputDir == "" {
		config.Export.OutputDir = "./output"
	}

	return &config, nil
}
