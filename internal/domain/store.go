package domain

import (
	"context"
	"database/sql"
	"time"
)

// DataStore defines the interface for persisting and querying WaniKani data
type DataStore interface {
	// UpsertSubjects inserts or updates subjects in the data store
	UpsertSubjects(ctx context.Context, subjects []Subject) error

	// GetSubjects retrieves subjects matching the provided filters
	GetSubjects(ctx context.Context, filters SubjectFilters) ([]Subject, error)

	// UpsertAssignments inserts or updates assignments in the data store
	UpsertAssignments(ctx context.Context, assignments []Assignment) error

	// GetAssignments retrieves assignments matching the provided filters
	GetAssignments(ctx context.Context, filters AssignmentFilters) ([]Assignment, error)

	// UpsertReviews inserts or updates reviews in the data store
	UpsertReviews(ctx context.Context, reviews []Review) error

	// GetReviews retrieves reviews matching the provided filters
	GetReviews(ctx context.Context, filters ReviewFilters) ([]Review, error)

	// InsertStatistics inserts a new statistics snapshot
	InsertStatistics(ctx context.Context, stats Statistics, timestamp time.Time) error

	// GetStatistics retrieves statistics snapshots within the provided date range
	GetStatistics(ctx context.Context, dateRange *DateRange) ([]StatisticsSnapshot, error)

	// GetLatestStatistics retrieves the most recent statistics snapshot
	GetLatestStatistics(ctx context.Context) (*StatisticsSnapshot, error)

	// UpsertAssignmentSnapshot inserts or updates an assignment snapshot
	UpsertAssignmentSnapshot(ctx context.Context, snapshot AssignmentSnapshot) error

	// GetAssignmentSnapshots retrieves assignment snapshots within the provided date range
	GetAssignmentSnapshots(ctx context.Context, dateRange *DateRange) ([]AssignmentSnapshot, error)

	// CalculateAssignmentSnapshot computes a snapshot from current assignments for a given date
	CalculateAssignmentSnapshot(ctx context.Context, date time.Time) ([]AssignmentSnapshot, error)

	// GetLastSyncTime retrieves the last successful sync timestamp for a data type
	GetLastSyncTime(ctx context.Context, dataType DataType) (*time.Time, error)

	// SetLastSyncTime updates the last successful sync timestamp for a data type
	SetLastSyncTime(ctx context.Context, dataType DataType, timestamp time.Time) error

	// BeginTx starts a new database transaction
	BeginTx(ctx context.Context) (*sql.Tx, error)
}
