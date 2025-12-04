package migrations

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestMigrations(t *testing.T) {
	// Create a temporary database file
	tmpDB := "test_migrations.db"
	defer os.Remove(tmpDB)

	// Open database connection
	db, err := sql.Open("sqlite3", tmpDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := Run(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Verify migration version
	version, err := Version(db)
	if err != nil {
		t.Fatalf("Failed to get migration version: %v", err)
	}

	if version != 2 {
		t.Errorf("Expected migration version 2, got %d", version)
	}

	// Verify tables exist
	tables := []string{
		"subjects",
		"assignments",
		"reviews",
		"statistics_snapshots",
		"sync_metadata",
		"assignment_snapshots",
	}

	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err == sql.ErrNoRows {
			t.Errorf("Table %s does not exist", table)
		} else if err != nil {
			t.Errorf("Error checking table %s: %v", table, err)
		}
	}

	// Verify indexes exist
	indexes := []string{
		"idx_subjects_data_updated_at",
		"idx_assignments_subject_id",
		"idx_assignments_data_updated_at",
		"idx_reviews_assignment_id",
		"idx_reviews_subject_id",
		"idx_reviews_data_updated_at",
		"idx_statistics_snapshots_timestamp",
		"idx_assignment_snapshots_date",
	}

	for _, index := range indexes {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='index' AND name=?", index).Scan(&name)
		if err == sql.ErrNoRows {
			t.Errorf("Index %s does not exist", index)
		} else if err != nil {
			t.Errorf("Error checking index %s: %v", index, err)
		}
	}
}

func TestMigrationsIdempotent(t *testing.T) {
	// Create a temporary database file
	tmpDB := "test_migrations_idempotent.db"
	defer os.Remove(tmpDB)

	// Open database connection
	db, err := sql.Open("sqlite3", tmpDB)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer db.Close()

	// Run migrations first time
	if err := Run(db); err != nil {
		t.Fatalf("Failed to run migrations first time: %v", err)
	}

	version1, err := Version(db)
	if err != nil {
		t.Fatalf("Failed to get migration version after first run: %v", err)
	}

	// Run migrations second time (should be idempotent)
	if err := Run(db); err != nil {
		t.Fatalf("Failed to run migrations second time: %v", err)
	}

	version2, err := Version(db)
	if err != nil {
		t.Fatalf("Failed to get migration version after second run: %v", err)
	}

	if version1 != version2 {
		t.Errorf("Migration version changed on second run: %d -> %d", version1, version2)
	}

	if version2 != 2 {
		t.Errorf("Expected migration version 2, got %d", version2)
	}
}
