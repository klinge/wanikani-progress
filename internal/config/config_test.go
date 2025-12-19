package config

import (
	"os"
	"testing"
)

func TestLoad_WithEnvironmentVariables(t *testing.T) {
	// Set environment variables
	os.Setenv("WANIKANI_API_TOKEN", "test-token-123")
	os.Setenv("DATABASE_PATH", "/tmp/test.db")
	os.Setenv("API_PORT", "9090")
	defer func() {
		os.Unsetenv("WANIKANI_API_TOKEN")
		os.Unsetenv("DATABASE_PATH")
		os.Unsetenv("API_PORT")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if config.WaniKaniAPIToken != "test-token-123" {
		t.Errorf("expected API token 'test-token-123', got '%s'", config.WaniKaniAPIToken)
	}

	if config.DatabasePath != "/tmp/test.db" {
		t.Errorf("expected database path '/tmp/test.db', got '%s'", config.DatabasePath)
	}

	if config.APIPort != 9090 {
		t.Errorf("expected API port 9090, got %d", config.APIPort)
	}
}

func TestLoad_WithDefaults(t *testing.T) {
	// Clear all environment variables that might be set from .env file
	os.Unsetenv("DATABASE_PATH")
	os.Unsetenv("SYNC_SCHEDULE")
	os.Unsetenv("API_PORT")
	os.Unsetenv("LOG_LEVEL")

	// Set only required variable
	os.Setenv("WANIKANI_API_TOKEN", "test-token")
	defer func() {
		os.Unsetenv("WANIKANI_API_TOKEN")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	// Check defaults
	if config.DatabasePath != "./wanikani.db" {
		t.Errorf("expected default database path './wanikani.db', got '%s'", config.DatabasePath)
	}

	if config.SyncSchedule != "0 2 * * *" {
		t.Errorf("expected default sync schedule '0 2 * * *', got '%s'", config.SyncSchedule)
	}

	if config.APIPort != 8080 {
		t.Errorf("expected default API port 8080, got %d", config.APIPort)
	}

	if config.LogLevel != "info" {
		t.Errorf("expected default log level 'info', got '%s'", config.LogLevel)
	}
}

func TestLoad_MissingRequiredToken(t *testing.T) {
	// Ensure token is not set
	os.Unsetenv("WANIKANI_API_TOKEN")

	_, err := Load()
	if err == nil {
		t.Error("expected error when WANIKANI_API_TOKEN is missing, got nil")
	}
}
