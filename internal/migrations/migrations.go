package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS

// Run executes all pending database migrations
func Run(db *sql.DB) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return fmt.Errorf("failed to set goose dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// Version returns the current migration version
func Version(db *sql.DB) (int64, error) {
	if err := goose.SetDialect("sqlite3"); err != nil {
		return 0, fmt.Errorf("failed to set goose dialect: %w", err)
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return 0, fmt.Errorf("failed to get migration version: %w", err)
	}

	return version, nil
}
