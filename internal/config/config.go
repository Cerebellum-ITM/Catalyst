package config

import (
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// Config holds the application's configuration.
type Config struct {
	RuneCraftHost string `toml:"runecraft_host"`
}

// Load loads the configuration from the user's config directory.
// If the config file doesn't exist, it creates a default one.
func Load() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configDir := filepath.Join(homeDir, ".config")

	catalystConfigDir := filepath.Join(configDir, "Catalyst")
	if err := os.MkdirAll(catalystConfigDir, 0755); err != nil {
		return nil, err
	}

	configFile := filepath.Join(catalystConfigDir, "config.toml")

	// Create a default config if it doesn't exist
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		if err := os.WriteFile(configFile, []byte(defaultConfigContent), 0644); err != nil {
			return nil, err
		}
	}

	// Read the config file
	var cfg Config
	if _, err := toml.DecodeFile(configFile, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
