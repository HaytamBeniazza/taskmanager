package config

import (
	"os"
	"strconv"
)

// Config holds the application configuration
type Config struct {
	Port       int
	Env        string
	CorsOrigin string
	DBPath     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	// Default values
	config := &Config{
		Port:       8080,
		Env:        "development",
		CorsOrigin: "*",
		DBPath:     "taskmanager.db",
	}

	// Override from environment variables if provided
	if port, err := strconv.Atoi(os.Getenv("PORT")); err == nil && port > 0 {
		config.Port = port
	}

	if env := os.Getenv("ENV"); env != "" {
		config.Env = env
	}

	if corsOrigin := os.Getenv("CORS_ORIGIN"); corsOrigin != "" {
		config.CorsOrigin = corsOrigin
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		config.DBPath = dbPath
	}

	return config
}
