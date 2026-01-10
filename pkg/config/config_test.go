package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Defaults(t *testing.T) {
	// Clear any environment variables that might interfere
	originalNatsURL := os.Getenv("NATS_URL")
	originalPluginDir := os.Getenv("PLUGIN_DIR")
	originalPort := os.Getenv("PORT")
	originalDBURL := os.Getenv("DATABASE_URL")

	defer func() {
		// Restore original values
		if originalNatsURL != "" {
			os.Setenv("NATS_URL", originalNatsURL)
		}
		if originalPluginDir != "" {
			os.Setenv("PLUGIN_DIR", originalPluginDir)
		}
		if originalPort != "" {
			os.Setenv("PORT", originalPort)
		}
		if originalDBURL != "" {
			os.Setenv("DATABASE_URL", originalDBURL)
		}
	}()

	// Unset to use defaults
	os.Unsetenv("NATS_URL")
	os.Unsetenv("PLUGIN_DIR")
	os.Unsetenv("PORT")
	os.Unsetenv("DATABASE_URL")

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "nats://localhost:4222", cfg.NatsURL)
	assert.Equal(t, "./plugins", cfg.PluginDir)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "postgres://webencode:webencode@localhost:5432/webencode?sslmode=disable", cfg.DatabaseURL)
}

func TestLoad_FromEnv(t *testing.T) {
	// Set custom values
	os.Setenv("NATS_URL", "nats://custom:4222")
	os.Setenv("PLUGIN_DIR", "/custom/plugins")
	os.Setenv("PORT", "9090")
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost:5432/test")

	defer func() {
		os.Unsetenv("NATS_URL")
		os.Unsetenv("PLUGIN_DIR")
		os.Unsetenv("PORT")
		os.Unsetenv("DATABASE_URL")
	}()

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "nats://custom:4222", cfg.NatsURL)
	assert.Equal(t, "/custom/plugins", cfg.PluginDir)
	assert.Equal(t, "9090", cfg.Port)
	assert.Equal(t, "postgres://test:test@localhost:5432/test", cfg.DatabaseURL)
}

func TestLoad_PartialEnv(t *testing.T) {
	// Set only some values
	os.Setenv("PORT", "3000")
	os.Unsetenv("NATS_URL")
	os.Unsetenv("PLUGIN_DIR")
	os.Unsetenv("DATABASE_URL")

	defer func() {
		os.Unsetenv("PORT")
	}()

	cfg, err := Load()

	assert.NoError(t, err)
	assert.NotNil(t, cfg)
	assert.Equal(t, "3000", cfg.Port)
	// Other values should be defaults
	assert.Equal(t, "nats://localhost:4222", cfg.NatsURL)
	assert.Equal(t, "./plugins", cfg.PluginDir)
}

func TestConfig_Struct(t *testing.T) {
	cfg := Config{
		DatabaseURL: "postgres://localhost",
		NatsURL:     "nats://localhost",
		PluginDir:   "./plugins",
		Port:        "8080",
	}

	assert.Equal(t, "postgres://localhost", cfg.DatabaseURL)
	assert.Equal(t, "nats://localhost", cfg.NatsURL)
	assert.Equal(t, "./plugins", cfg.PluginDir)
	assert.Equal(t, "8080", cfg.Port)
}
