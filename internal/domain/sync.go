package domain

import "context"

// SyncService defines the interface for orchestrating data synchronization
type SyncService interface {
	// SyncAll performs a full sync of all data types
	SyncAll(ctx context.Context) ([]SyncResult, error)

	// SyncSubjects syncs only subjects
	SyncSubjects(ctx context.Context) (SyncResult, error)

	// SyncAssignments syncs only assignments
	SyncAssignments(ctx context.Context) (SyncResult, error)

	// SyncReviews syncs only reviews
	SyncReviews(ctx context.Context) (SyncResult, error)

	// SyncStatistics syncs only statistics
	SyncStatistics(ctx context.Context) (SyncResult, error)

	// IsSyncing returns true if a sync operation is currently in progress
	IsSyncing() bool
}
