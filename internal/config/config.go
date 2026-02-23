package config

import (
	"os"
	"path/filepath"
)

const DefaultTimezone = "Local"

type Config struct {
	HomeDir    string
	DataDir    string
	DBPath     string
	Timezone   string
	ConfigToml string
}

func Default() Config {
	homeDir := os.Getenv("TD_HOME")
	if homeDir == "" {
		userHome, err := os.UserHomeDir()
		if err == nil {
			homeDir = filepath.Join(userHome, ".td")
		}
	}
	dataDir := filepath.Join(homeDir, "data")
	return Config{
		HomeDir:    homeDir,
		DataDir:    dataDir,
		DBPath:     filepath.Join(dataDir, "td.db"),
		Timezone:   DefaultTimezone,
		ConfigToml: filepath.Join(homeDir, "config.toml"),
	}
}
