package sync

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/sirupsen/logrus"
	"wanikani-api/internal/domain"
)

// testLogger creates a logger for testing that discards output
func testLogger() *logrus.Logger {
	logger := logrus.New()
	logger.SetOutput(io.Discard)
	return logger
}

// Mock client for testing
type mockClient struct {
	subjects   []domain.Subject
	assignments []domain.Assignment
	reviews    []domain.Review
	statistics *domain.Statistics
	fetchError error
	delay      time.Duration
}

func (m *mockClient) SetAPIToken(token string) {}

func (m *mockClient) FetchSubjects(ctx context.Context, updatedAfter *time.Time) ([]domain.Subject, error) {
	if m.delay > 0 {
		time.Sleep(m.delay)
	}
	if m.fetchError != nil {
		return nil, m.fetchError
	}
	return m.subjects, nil
}

func (m *mockClient) FetchAssignments(ctx context.Context, updatedAfter *time.Time) ([]domain.Assignment, error) {
	if m.fetchError != nil {
		return nil, m.fetchError
	}
	return m.assignments, nil
}

func (m *mockClient) FetchReviews(ctx context.Context, updatedAfter *time.Time) ([]domain.Review, error) {
	if m.fetchError != nil {
		return nil, m.fetchError
	}
	return m.reviews, nil
}

func (m *mockClient) FetchStatistics(ctx context.Context) (*domain.Statistics, error) {
	if m.fetchError != nil {
		return nil, m.fetchError
	}
	return m.statistics, nil
}

func (m *mockClient) GetRateLimitStatus() domain.RateLimitInfo {
	return domain.RateLimitInfo{}
}

// Mock store for testing
type mockStore struct {
	lastSyncTimes       map[domain.DataType]*time.Time
	upsertError         error
	insertError         error
	syncTimeError       error
	snapshotUpsertError error
	snapshotCalcError   error
}

func newMockStore() *mockStore {
	return &mockStore{
		lastSyncTimes: make(map[domain.DataType]*time.Time),
	}
}

func (m *mockStore) UpsertSubjects(ctx context.Context, subjects []domain.Subject) error {
	return m.upsertError
}

func (m *mockStore) GetSubjects(ctx context.Context, filters domain.SubjectFilters) ([]domain.Subject, error) {
	return nil, nil
}

func (m *mockStore) UpsertAssignments(ctx context.Context, assignments []domain.Assignment) error {
	return m.upsertError
}

func (m *mockStore) GetAssignments(ctx context.Context, filters domain.AssignmentFilters) ([]domain.Assignment, error) {
	return nil, nil
}

func (m *mockStore) UpsertReviews(ctx context.Context, reviews []domain.Review) error {
	return m.upsertError
}

func (m *mockStore) GetReviews(ctx context.Context, filters domain.ReviewFilters) ([]domain.Review, error) {
	return nil, nil
}

func (m *mockStore) InsertStatistics(ctx context.Context, stats domain.Statistics, timestamp time.Time) error {
	return m.insertError
}

func (m *mockStore) GetStatistics(ctx context.Context, dateRange *domain.DateRange) ([]domain.StatisticsSnapshot, error) {
	return nil, nil
}

func (m *mockStore) GetLatestStatistics(ctx context.Context) (*domain.StatisticsSnapshot, error) {
	return nil, nil
}

func (m *mockStore) GetLastSyncTime(ctx context.Context, dataType domain.DataType) (*time.Time, error) {
	if m.syncTimeError != nil {
		return nil, m.syncTimeError
	}
	return m.lastSyncTimes[dataType], nil
}

func (m *mockStore) SetLastSyncTime(ctx context.Context, dataType domain.DataType, timestamp time.Time) error {
	if m.syncTimeError != nil {
		return m.syncTimeError
	}
	m.lastSyncTimes[dataType] = &timestamp
	return nil
}

func (m *mockStore) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return nil, nil
}

func (m *mockStore) UpsertAssignmentSnapshot(ctx context.Context, snapshot domain.AssignmentSnapshot) error {
	return m.snapshotUpsertError
}

func (m *mockStore) GetAssignmentSnapshots(ctx context.Context, dateRange *domain.DateRange) ([]domain.AssignmentSnapshot, error) {
	return nil, nil
}

func (m *mockStore) CalculateAssignmentSnapshot(ctx context.Context, date time.Time) ([]domain.AssignmentSnapshot, error) {
	if m.snapshotCalcError != nil {
		return nil, m.snapshotCalcError
	}
	// Return a simple snapshot for testing
	return []domain.AssignmentSnapshot{
		{
			Date:        date,
			SRSStage:    1,
			SubjectType: "kanji",
			Count:       5,
		},
	}, nil
}

// mockClientWithTimestampCapture captures the updatedAfter parameter
type mockClientWithTimestampCapture struct {
	capturedUpdatedAfter **time.Time
	subjects             []domain.Subject
	assignments          []domain.Assignment
	reviews              []domain.Review
	statistics           *domain.Statistics
}

func (m *mockClientWithTimestampCapture) SetAPIToken(token string) {}

func (m *mockClientWithTimestampCapture) FetchSubjects(ctx context.Context, updatedAfter *time.Time) ([]domain.Subject, error) {
	*m.capturedUpdatedAfter = updatedAfter
	return m.subjects, nil
}

func (m *mockClientWithTimestampCapture) FetchAssignments(ctx context.Context, updatedAfter *time.Time) ([]domain.Assignment, error) {
	*m.capturedUpdatedAfter = updatedAfter
	return m.assignments, nil
}

func (m *mockClientWithTimestampCapture) FetchReviews(ctx context.Context, updatedAfter *time.Time) ([]domain.Review, error) {
	*m.capturedUpdatedAfter = updatedAfter
	return m.reviews, nil
}

func (m *mockClientWithTimestampCapture) FetchStatistics(ctx context.Context) (*domain.Statistics, error) {
	return m.statistics, nil
}

func (m *mockClientWithTimestampCapture) GetRateLimitStatus() domain.RateLimitInfo {
	return domain.RateLimitInfo{}
}

// Generators for property-based testing

// genDataType generates random DataType values
func genDataType() gopter.Gen {
	return gen.OneConstOf(
		domain.DataTypeSubjects,
		domain.DataTypeAssignments,
		domain.DataTypeReviews,
		domain.DataTypeStatistics,
	)
}

// genPastTimestamp generates random timestamps in the past
func genPastTimestamp() gopter.Gen {
	return gen.Int64Range(1, 365*24*60*60).Map(func(secondsAgo int64) time.Time {
		return time.Now().Add(-time.Duration(secondsAgo) * time.Second)
	})
}

func TestSyncSubjects_Success(t *testing.T) {
	client := &mockClient{
		subjects: []domain.Subject{
			{ID: 1, Object: "kanji"},
			{ID: 2, Object: "vocabulary"},
		},
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	result := service.SyncSubjects(context.Background())

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.RecordsUpdated != 2 {
		t.Errorf("expected 2 records updated, got %d", result.RecordsUpdated)
	}
	if result.DataType != domain.DataTypeSubjects {
		t.Errorf("expected DataTypeSubjects, got %s", result.DataType)
	}
}

func TestSyncAssignments_Success(t *testing.T) {
	client := &mockClient{
		assignments: []domain.Assignment{
			{ID: 1, Object: "assignment"},
		},
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	result := service.SyncAssignments(context.Background())

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.RecordsUpdated != 1 {
		t.Errorf("expected 1 record updated, got %d", result.RecordsUpdated)
	}
}

func TestSyncReviews_Success(t *testing.T) {
	client := &mockClient{
		reviews: []domain.Review{
			{ID: 1, Object: "review"},
			{ID: 2, Object: "review"},
			{ID: 3, Object: "review"},
		},
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	result := service.SyncReviews(context.Background())

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.RecordsUpdated != 3 {
		t.Errorf("expected 3 records updated, got %d", result.RecordsUpdated)
	}
}

func TestSyncStatistics_Success(t *testing.T) {
	client := &mockClient{
		statistics: &domain.Statistics{
			Object: "report",
		},
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	result := service.SyncStatistics(context.Background())

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	if result.RecordsUpdated != 1 {
		t.Errorf("expected 1 record updated, got %d", result.RecordsUpdated)
	}
}

func TestSyncSubjects_FetchError(t *testing.T) {
	client := &mockClient{
		fetchError: errors.New("network error"),
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	result := service.SyncSubjects(context.Background())

	if result.Success {
		t.Error("expected failure, got success")
	}
	if result.Error == "" {
		t.Error("expected error message")
	}
}

func TestSyncSubjects_StoreError(t *testing.T) {
	client := &mockClient{
		subjects: []domain.Subject{{ID: 1}},
	}
	store := newMockStore()
	store.upsertError = errors.New("database error")
	service := NewService(client, store, testLogger())

	result := service.SyncSubjects(context.Background())

	if result.Success {
		t.Error("expected failure, got success")
	}
}

func TestSyncAll_Success(t *testing.T) {
	client := &mockClient{
		subjects:    []domain.Subject{{ID: 1}},
		assignments: []domain.Assignment{{ID: 1}},
		reviews:     []domain.Review{{ID: 1}},
		statistics:  &domain.Statistics{Object: "report"},
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	results, err := service.SyncAll(context.Background())

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}
	for _, result := range results {
		if !result.Success {
			t.Errorf("expected all syncs to succeed, got error for %s: %s", result.DataType, result.Error)
		}
	}
}

func TestSyncAll_StopsOnFirstFailure(t *testing.T) {
	client := &mockClient{
		fetchError: errors.New("api error"),
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	results, err := service.SyncAll(context.Background())

	if err == nil {
		t.Error("expected error, got nil")
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result (failed subjects sync), got %d", len(results))
	}
}

func TestIsSyncing_ConcurrentSyncPrevention(t *testing.T) {
	client := &mockClient{
		subjects:    []domain.Subject{{ID: 1}},
		assignments: []domain.Assignment{{ID: 1}},
		reviews:     []domain.Review{{ID: 1}},
		statistics:  &domain.Statistics{Object: "report"},
		delay:       50 * time.Millisecond, // Add delay to ensure sync is in progress
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	// Start first sync in goroutine
	done := make(chan bool)
	go func() {
		service.SyncAll(context.Background())
		done <- true
	}()

	// Give first sync time to start and set the syncing flag
	time.Sleep(20 * time.Millisecond)

	// Try to start second sync
	_, err := service.SyncAll(context.Background())

	if err == nil {
		t.Error("expected error for concurrent sync, got nil")
	}
	if err != nil && err.Error() != "sync already in progress" {
		t.Errorf("expected 'sync already in progress' error, got: %v", err)
	}

	<-done
}

func TestIsSyncing_ReturnsFalseWhenNotSyncing(t *testing.T) {
	service := NewService(&mockClient{}, newMockStore(), testLogger())

	if service.IsSyncing() {
		t.Error("expected IsSyncing to return false initially")
	}
}

func TestSyncSubjects_UsesLastSyncTime(t *testing.T) {
	lastSync := time.Now().Add(-24 * time.Hour)
	client := &mockClient{
		subjects: []domain.Subject{{ID: 1}},
	}
	store := newMockStore()
	store.lastSyncTimes[domain.DataTypeSubjects] = &lastSync
	service := NewService(client, store, testLogger())

	result := service.SyncSubjects(context.Background())

	if !result.Success {
		t.Errorf("expected success, got error: %s", result.Error)
	}
	// Verify that the last sync time was updated
	newSyncTime := store.lastSyncTimes[domain.DataTypeSubjects]
	if newSyncTime == nil {
		t.Error("expected sync time to be updated")
	}
	if !newSyncTime.After(lastSync) {
		t.Error("expected new sync time to be after old sync time")
	}
}

func TestSyncSubjects_EmptyResults(t *testing.T) {
	client := &mockClient{
		subjects: []domain.Subject{},
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	result := service.SyncSubjects(context.Background())

	if !result.Success {
		t.Errorf("expected success with empty results, got error: %s", result.Error)
	}
	if result.RecordsUpdated != 0 {
		t.Errorf("expected 0 records updated, got %d", result.RecordsUpdated)
	}
}

func TestCreateAssignmentSnapshot_Success(t *testing.T) {
	client := &mockClient{}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	err := service.CreateAssignmentSnapshot(context.Background())

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestCreateAssignmentSnapshot_CalculateError(t *testing.T) {
	client := &mockClient{}
	store := newMockStore()
	store.snapshotCalcError = errors.New("calculation error")
	service := NewService(client, store, testLogger())

	err := service.CreateAssignmentSnapshot(context.Background())

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestCreateAssignmentSnapshot_UpsertError(t *testing.T) {
	client := &mockClient{}
	store := newMockStore()
	store.snapshotUpsertError = errors.New("upsert error")
	service := NewService(client, store, testLogger())

	err := service.CreateAssignmentSnapshot(context.Background())

	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestSyncAll_SnapshotErrorDoesNotFailSync(t *testing.T) {
	client := &mockClient{
		subjects:    []domain.Subject{{ID: 1}},
		assignments: []domain.Assignment{{ID: 1}},
		reviews:     []domain.Review{{ID: 1}},
		statistics:  &domain.Statistics{Object: "report"},
	}
	store := newMockStore()
	service := NewService(client, store, testLogger())

	// First sync should succeed
	results, err := service.SyncAll(context.Background())

	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
	if len(results) != 4 {
		t.Errorf("expected 4 results, got %d", len(results))
	}

	// Now test with snapshot error - sync should still succeed
	store.snapshotCalcError = errors.New("snapshot calculation error")
	results2, err2 := service.SyncAll(context.Background())

	if err2 != nil {
		t.Errorf("expected no error even with snapshot failure, got: %v", err2)
	}
	if len(results2) != 4 {
		t.Errorf("expected 4 results, got %d", len(results2))
	}
	// All sync results should still be successful
	for _, result := range results2 {
		if !result.Success {
			t.Errorf("expected all syncs to succeed, got error for %s: %s", result.DataType, result.Error)
		}
	}
}

// Feature: wanikani-api, Property 9: Incremental sync uses timestamps
// Validates: Requirements 6.1, 3.4
func TestProperty_IncrementalSyncUsesTimestamps(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("sync operations after initial sync use last sync timestamp", prop.ForAll(
		func(dataType domain.DataType, lastSyncTime time.Time) bool {
			// Create a mock client that tracks the updatedAfter parameter
			var capturedUpdatedAfter *time.Time
			client := &mockClientWithTimestampCapture{
				capturedUpdatedAfter: &capturedUpdatedAfter,
				subjects:             []domain.Subject{{ID: 1}},
				assignments:          []domain.Assignment{{ID: 1}},
				reviews:              []domain.Review{{ID: 1}},
			}

			// Create a store with a previous sync timestamp
			store := newMockStore()
			store.lastSyncTimes[dataType] = &lastSyncTime

			service := NewService(client, store, testLogger())
			ctx := context.Background()

			// Perform sync based on data type
			var result domain.SyncResult
			switch dataType {
			case domain.DataTypeSubjects:
				result = service.SyncSubjects(ctx)
			case domain.DataTypeAssignments:
				result = service.SyncAssignments(ctx)
			case domain.DataTypeReviews:
				result = service.SyncReviews(ctx)
			case domain.DataTypeStatistics:
				// Statistics don't use incremental sync, skip
				return true
			default:
				return true
			}

			// Verify the sync was successful
			if !result.Success {
				return false
			}

			// Verify that the client received the last sync timestamp
			if capturedUpdatedAfter == nil {
				return false
			}

			// The captured timestamp should match the last sync time
			return capturedUpdatedAfter.Equal(lastSyncTime)
		},
		genDataType(),
		genPastTimestamp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: wanikani-api, Property 10: Successful sync updates timestamp
// Validates: Requirements 6.2
func TestProperty_SuccessfulSyncUpdatesTimestamp(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("successful sync updates the last sync timestamp", prop.ForAll(
		func(dataType domain.DataType, initialSyncTime *time.Time) bool {
			// Create a mock client with data to sync
			client := &mockClient{
				subjects:    []domain.Subject{{ID: 1, Object: "kanji"}},
				assignments: []domain.Assignment{{ID: 1, Object: "assignment"}},
				reviews:     []domain.Review{{ID: 1, Object: "review"}},
				statistics:  &domain.Statistics{Object: "report"},
			}

			// Create a store with an optional initial sync timestamp
			store := newMockStore()
			if initialSyncTime != nil {
				store.lastSyncTimes[dataType] = initialSyncTime
			}

			service := NewService(client, store, testLogger())
			ctx := context.Background()

			// Record the time before sync
			beforeSync := time.Now()

			// Perform sync based on data type
			var result domain.SyncResult
			switch dataType {
			case domain.DataTypeSubjects:
				result = service.SyncSubjects(ctx)
			case domain.DataTypeAssignments:
				result = service.SyncAssignments(ctx)
			case domain.DataTypeReviews:
				result = service.SyncReviews(ctx)
			case domain.DataTypeStatistics:
				result = service.SyncStatistics(ctx)
			default:
				return true
			}

			// Verify the sync was successful
			if !result.Success {
				return false
			}

			// Get the updated sync timestamp from the store
			updatedSyncTime, err := store.GetLastSyncTime(ctx, dataType)
			if err != nil {
				return false
			}

			// Verify that the timestamp was updated
			if updatedSyncTime == nil {
				return false
			}

			// The updated timestamp should be after or equal to the time before sync
			if updatedSyncTime.Before(beforeSync) {
				return false
			}

			// If there was an initial sync time, the new timestamp should be after it
			if initialSyncTime != nil {
				if !updatedSyncTime.After(*initialSyncTime) && !updatedSyncTime.Equal(*initialSyncTime) {
					return false
				}
			}

			return true
		},
		genDataType(),
		gen.PtrOf(genPastTimestamp()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: wanikani-api, Property 11: Failed sync preserves timestamp
// Validates: Requirements 6.3
func TestProperty_FailedSyncPreservesTimestamp(t *testing.T) {
	properties := gopter.NewProperties(nil)

	properties.Property("failed sync operations preserve the last sync timestamp", prop.ForAll(
		func(dataType domain.DataType, initialSyncTime time.Time) bool {
			// Create a store with an initial sync timestamp
			store := newMockStore()
			store.lastSyncTimes[dataType] = &initialSyncTime

			ctx := context.Background()

			// Test with fetch error
			clientWithFetchError := &mockClient{
				fetchError: errors.New("api error"),
			}
			serviceFetchError := NewService(clientWithFetchError, store, testLogger())

			// Perform sync based on data type
			var result domain.SyncResult
			switch dataType {
			case domain.DataTypeSubjects:
				result = serviceFetchError.SyncSubjects(ctx)
			case domain.DataTypeAssignments:
				result = serviceFetchError.SyncAssignments(ctx)
			case domain.DataTypeReviews:
				result = serviceFetchError.SyncReviews(ctx)
			case domain.DataTypeStatistics:
				result = serviceFetchError.SyncStatistics(ctx)
			default:
				return true
			}

			// Verify the sync failed
			if result.Success {
				return false
			}

			// Verify the timestamp was preserved
			preservedTime, err := store.GetLastSyncTime(ctx, dataType)
			if err != nil {
				return false
			}
			if preservedTime == nil {
				return false
			}
			if !preservedTime.Equal(initialSyncTime) {
				return false
			}

			// Test with store error (for non-statistics types)
			if dataType != domain.DataTypeStatistics {
				store2 := newMockStore()
				store2.lastSyncTimes[dataType] = &initialSyncTime
				store2.upsertError = errors.New("database error")

				clientWithData := &mockClient{
					subjects:    []domain.Subject{{ID: 1}},
					assignments: []domain.Assignment{{ID: 1}},
					reviews:     []domain.Review{{ID: 1}},
				}
				serviceStoreError := NewService(clientWithData, store2, testLogger())

				switch dataType {
				case domain.DataTypeSubjects:
					result = serviceStoreError.SyncSubjects(ctx)
				case domain.DataTypeAssignments:
					result = serviceStoreError.SyncAssignments(ctx)
				case domain.DataTypeReviews:
					result = serviceStoreError.SyncReviews(ctx)
				}

				// Verify the sync failed
				if result.Success {
					return false
				}

				// Verify the timestamp was preserved
				preservedTime2, err := store2.GetLastSyncTime(ctx, dataType)
				if err != nil {
					return false
				}
				if preservedTime2 == nil {
					return false
				}
				if !preservedTime2.Equal(initialSyncTime) {
					return false
				}
			} else {
				// For statistics, test with insert error
				store2 := newMockStore()
				store2.lastSyncTimes[dataType] = &initialSyncTime
				store2.insertError = errors.New("database error")

				clientWithData := &mockClient{
					statistics: &domain.Statistics{Object: "report"},
				}
				serviceStoreError := NewService(clientWithData, store2, testLogger())

				result = serviceStoreError.SyncStatistics(ctx)

				// Verify the sync failed
				if result.Success {
					return false
				}

				// Verify the timestamp was preserved
				preservedTime2, err := store2.GetLastSyncTime(ctx, dataType)
				if err != nil {
					return false
				}
				if preservedTime2 == nil {
					return false
				}
				if !preservedTime2.Equal(initialSyncTime) {
					return false
				}
			}

			return true
		},
		genDataType(),
		genPastTimestamp(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
