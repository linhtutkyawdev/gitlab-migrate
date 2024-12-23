package utils

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration loaded from YAML
type Config struct {
	SourceBaseURL          string `yaml:"source_base_url"`
	SourceAccessToken      string `yaml:"source_access_token"`
	DestinationBaseURL     string `yaml:"destination_base_url"`
	DestinationAccessToken string `yaml:"destination_access_token"`
	AuthUser               string `yaml:"auth_user"`
	AuthPassword           string `yaml:"auth_password"`
}

// Validate checks if all required fields are properly set and formatted
func (c *Config) Validate() error {
	if strings.TrimSpace(c.SourceBaseURL) == "" {
		return fmt.Errorf("source_base_url is required")
	}
	if strings.TrimSpace(c.SourceAccessToken) == "" {
		return fmt.Errorf("source_access_token is required")
	}
	if strings.TrimSpace(c.DestinationBaseURL) == "" {
		return fmt.Errorf("destination_base_url is required")
	}
	if strings.TrimSpace(c.DestinationAccessToken) == "" {
		return fmt.Errorf("destination_access_token is required")
	}

	// Validate URLs
	if err := validateURL(c.SourceBaseURL); err != nil {
		return fmt.Errorf("invalid source_base_url: %w", err)
	}
	if err := validateURL(c.DestinationBaseURL); err != nil {
		return fmt.Errorf("invalid destination_base_url: %w", err)
	}

	return nil
}

// validateURL checks if the provided URL is valid
func validateURL(urlStr string) error {
	u, err := url.Parse(urlStr)
	if err != nil {
		return err
	}
	if !u.IsAbs() {
		return fmt.Errorf("URL must be absolute")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL must use HTTP or HTTPS protocol")
	}
	return nil
}

// LoadConfig loads and validates configuration from the specified YAML file
func LoadConfig(filePath string) (*Config, error) {
	if strings.TrimSpace(filePath) == "" {
		return nil, fmt.Errorf("config file path cannot be empty")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %w", err)
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// GenerateOutputFileName generates a consistent file name based on command parameters
func GenerateOutputFileName(command string, groupID, projectID string, isDestination bool, isRecursive bool) string {
	prefix := "s"
	if isDestination {
		prefix = "d"
	}

	var identifier string
	switch command {
	case "groups":
		identifier = "groups"
	case "projects":
		if groupID != "" {
			identifier = fmt.Sprintf("projects_g-%s", groupID)
		} else {
			identifier = "projects"
		}
	case "variables":
		if groupID != "" {
			if isRecursive {
				identifier = fmt.Sprintf("variables_g-%s_recursive", groupID)
			} else {
				identifier = fmt.Sprintf("variables_g-%s", groupID)
			}
		} else if projectID != "" {
			identifier = fmt.Sprintf("variables_p-%s", projectID)
		} else {
			identifier = "variables"
		}
	}

	fileName := fmt.Sprintf("%s-gitlab_get_%s.json", prefix, identifier)
	return filepath.Join("data", fileName)
}

// EnsureDataDir ensures that the data directory exists
func EnsureDataDir() error {
	dataDir := "data"
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}
	return nil
}
