package migrate

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// Run executes the database migrations
func Run(dbURL string, migrationsPath string) error {
	log.Printf("Running migrations from %s", migrationsPath)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", migrationsPath),
		dbURL,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Println("Database migrations: no change")
			return nil
		}
		return fmt.Errorf("failed to run migrate up: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}
