# Database Migration Strategy

## Problem

The current implementation mixes database schema creation with application code in `internal/store/sqlite/store.go`. This approach has several issues:

1. **No version tracking** - Can't determine what schema version the database is at
2. **CREATE TABLE IF NOT EXISTS is insufficient** - Won't update existing tables when schema changes
3. **No rollback capability** - Can't undo migrations if something goes wrong
4. **Mixed concerns** - Production code shouldn't handle migrations
5. **No migration history** - Can't audit what changes were applied and when
6. **Risk to production data** - Schema changes happen automatically on startup without control

## Solution

Implement proper database migrations using `goose` (github.com/pressly/goose/v3) as specified in the design document.

## Why Goose?

Goose was chosen over golang-migrate/migrate for several reasons:

1. **Embeddable**: Can embed migrations in the application binary (no separate CLI needed)
2. **Flexibility**: Supports both SQL and Go migrations
3. **Better DX**: Cleaner API with better error messages
4. **Active Development**: Well-maintained with regular updates
5. **Simplicity**: Not overengineered for our use case

## Implementation Plan

See task 22 in `tasks.md` for the detailed implementation plan.

## Migration Files

Migrations will be stored in `migrations/` directory with the following structure:

```
migrations/
  00001_initial_schema.sql           # Creates initial tables (up and down)
  00002_add_assignment_snapshots.sql # Adds assignment snapshots table
```

Each file contains both up and down migrations using goose directives:

```sql
-- +goose Up
CREATE TABLE subjects (...);

-- +goose Down
DROP TABLE subjects;
```

## Benefits

1. **Safety**: Migrations can be reviewed and tested before deployment
2. **Embedded**: Migrations are embedded in binary for easy deployment
3. **Auditability**: Clear history of all schema changes in `goose_db_version` table
4. **Rollback**: Can revert changes if needed
5. **Transactions**: Automatic transaction wrapping with rollback on failure
6. **Best Practice**: Follows industry standard for database management

## Migration Workflow

### Development (Embedded in Application)
Migrations run automatically when the application starts:

```go
// In main.go or store initialization
if err := runMigrations(db); err != nil {
    log.Fatalf("Failed to run migrations: %v", err)
}
```

### Development (Manual CLI - Optional)
```bash
# Install goose CLI
go install github.com/pressly/goose/v3/cmd/goose@latest

# Create new migration
goose -dir migrations create add_new_feature sql

# Apply migrations
goose -dir migrations sqlite3 ./wanikani.db up

# Rollback last migration
goose -dir migrations sqlite3 ./wanikani.db down

# Check status
goose -dir migrations sqlite3 ./wanikani.db status
```

### Production
```bash
# Migrations run automatically on application startup
./wanikani-api

# Or run manually before starting application
goose -dir migrations sqlite3 ./wanikani.db up
./wanikani-api
```

## Example Migration File

```sql
-- +goose Up
-- +goose StatementBegin
CREATE TABLE assignment_snapshots (
    date TEXT NOT NULL,
    srs_stage INTEGER NOT NULL,
    subject_type TEXT NOT NULL,
    count INTEGER NOT NULL,
    PRIMARY KEY (date, srs_stage, subject_type)
);

CREATE INDEX idx_assignment_snapshots_date ON assignment_snapshots(date);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_assignment_snapshots_date;
DROP TABLE IF EXISTS assignment_snapshots;
-- +goose StatementEnd
```

## Current Status

- ✅ Design document updated with goose migration strategy
- ✅ Task 22 added to implement migrations with goose
- ⏳ Implementation pending

## References

- Design Document: `.kiro/specs/wanikani-api/design.md` (Database Migration Strategy section)
- Task: `.kiro/specs/wanikani-api/tasks.md` (Task 22)
- Library: https://github.com/pressly/goose
- Documentation: https://pressly.github.io/goose/
