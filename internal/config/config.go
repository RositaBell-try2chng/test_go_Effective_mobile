package config

import (
	"fmt"
	"os"
	"regexp"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Host string `yaml:"host"`
	} `yaml:"server"`

	Database struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		DBName   string `yaml:"dbname"`
		SSLMode  string `yaml:"sslmode"`
	} `yaml:"database"`

	Logging struct {
		Level  string `yaml:"level"`
		Format string `yaml:"format"`
	} `yaml:"logging"`
}

func Load() (*Config, error) {
	config := &Config{}

	// Try to load from environment variable first
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Read file content
	content, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Replace environment variables
	contentStr := replaceEnvVars(string(content))

	// Parse YAML
	if err := yaml.Unmarshal([]byte(contentStr), config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	return config, nil
}

// replaceEnvVars replaces ${VAR:-default} patterns with environment variables
func replaceEnvVars(content string) string {
	// Pattern to match ${VAR:-default} or ${VAR}
	re := regexp.MustCompile(`\$\{([^:}]+)(?::([^}]*))?\}`)
	return re.ReplaceAllStringFunc(content, func(match string) string {
		groups := re.FindStringSubmatch(match)
		if len(groups) < 2 {
			return match
		}

		varName := groups[1]
		defaultValue := ""
		if len(groups) > 2 {
			defaultValue = groups[2]
		}

		// Check if environment variable exists
		if envValue := os.Getenv(varName); envValue != "" {
			return envValue
		}

		// Return default value if provided
		if defaultValue != "" {
			return defaultValue
		}

		// Return the original match if no value found
		return match
	})
}

func (c *Config) GetDatabaseURL() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host, c.Database.Port, c.Database.User, c.Database.Password,
		c.Database.DBName, c.Database.SSLMode)
}
