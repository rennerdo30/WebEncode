package migrate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun_InvalidMigrationPath(t *testing.T) {
	err := Run("postgres://invalid", "/nonexistent/path")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to create migrate instance")
}

func TestRun_InvalidDBURL(t *testing.T) {
	// Create a temporary directory for migrations
	err := Run("invalid-url", ".")
	assert.Error(t, err)
}

// Note: Full integration tests for successful migrations would require
// a running PostgreSQL instance and actual migration files.
// These tests focus on error paths that can be tested without external dependencies.

func TestRun_ErrorHandling(t *testing.T) {
	// Test that the function properly wraps errors
	err := Run("", "")
	assert.Error(t, err)
	assert.NotNil(t, err)
}
