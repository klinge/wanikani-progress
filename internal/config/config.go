package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds the application configuration
type Config struct {
	WaniKaniAPIToken string
	LocalAPIToken    string
	DatabasePath     string
	SyncSchedule     string
	APIPort          int
	LogLevel         string
}

// Load loads configuration from .env file and environment variables with defaults
func Load() (*Config, error) {
	// Load .env file if it exists (silently ignore if not found)
	_ = godotenv.Load()

	config := &Config{
		WaniKaniAPIToken: getEnv("WANIKANI_API_TOKEN", ""),
		LocalAPIToken:    getEnv("LOCAL_API_TOKEN", ""),
		DatabasePath:     getEnv("DATABASE_PATH", "./wanikani.db"),
		SyncSchedule:     getEnv("SYNC_SCHEDULE", "0 2 * * *"),
		APIPort:          getEnvAsInt("API_PORT", 8080),
		LogLevel:         getEnv("LOG_LEVEL", "info"),
	}

	// Validate required configuration
	if config.WaniKaniAPIToken == "" {
		return nil, fmt.Errorf("WANIKANI_API_TOKEN environment variable is required")
	}

	return config, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt retrieves an environment variable as an integer or returns a default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}
