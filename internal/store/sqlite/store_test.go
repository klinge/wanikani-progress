package sqlite

import (
	"context"
	"os"
	"testing"
	"time"

	"wanikani-api/internal/domain"
)

func TestStore_UpsertAndGetSubjects(t *testing.T) {
	// Create temporary database
	dbPath := "test_subjects.db"
	defer os.Remove(dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create test subjects
	subjects := []domain.Subject{
		{
			ID:            1,
			Object:        "kanji",
			URL:           "https://api.wanikani.com/v2/subjects/1",
			DataUpdatedAt: time.Now(),
			Data: domain.SubjectData{
				Level:      5,
				Characters: "一",
				Meanings: []domain.Meaning{
					{Meaning: "one", Primary: true},
				},
			},
		},
		{
			ID:            2,
			Object:        "radical",
			URL:           "https://api.wanikani.com/v2/subjects/2",
			DataUpdatedAt: time.Now(),
			Data: domain.SubjectData{
				Level:      1,
				Characters: "丨",
				Meanings: []domain.Meaning{
					{Meaning: "stick", Primary: true},
				},
			},
		},
	}

	// Test upsert
	err = store.UpsertSubjects(ctx, subjects)
	if err != nil {
		t.Fatalf("failed to upsert subjects: %v", err)
	}

	// Test get all subjects
	retrieved, err := store.GetSubjects(ctx, domain.SubjectFilters{})
	if err != nil {
		t.Fatalf("failed to get subjects: %v", err)
	}

	if len(retrieved) != 2 {
		t.Errorf("expected 2 subjects, got %d", len(retrieved))
	}

	// Test filter by level
	level := 5
	filtered, err := store.GetSubjects(ctx, domain.SubjectFilters{Level: &level})
	if err != nil {
		t.Fatalf("failed to get filtered subjects: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("expected 1 subject with level 5, got %d", len(filtered))
	}

	if filtered[0].Data.Level != 5 {
		t.Errorf("expected level 5, got %d", filtered[0].Data.Level)
	}

	// Test upsert idempotence - update existing subject
	subjects[0].Data.Characters = "二"
	err = store.UpsertSubjects(ctx, subjects[:1])
	if err != nil {
		t.Fatalf("failed to update subject: %v", err)
	}

	updated, err := store.GetSubjects(ctx, domain.SubjectFilters{})
	if err != nil {
		t.Fatalf("failed to get updated subjects: %v", err)
	}

	if len(updated) != 2 {
		t.Errorf("expected 2 subjects after update, got %d", len(updated))
	}
}

func TestStore_UpsertAndGetAssignments(t *testing.T) {
	dbPath := "test_assignments.db"
	defer os.Remove(dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// First create a subject (for foreign key constraint)
	subjects := []domain.Subject{
		{
			ID:            1,
			Object:        "kanji",
			URL:           "https://api.wanikani.com/v2/subjects/1",
			DataUpdatedAt: time.Now(),
			Data: domain.SubjectData{
				Level:      5,
				Characters: "一",
			},
		},
	}
	err = store.UpsertSubjects(ctx, subjects)
	if err != nil {
		t.Fatalf("failed to upsert subjects: %v", err)
	}

	// Create test assignments
	now := time.Now()
	assignments := []domain.Assignment{
		{
			ID:            100,
			Object:        "assignment",
			URL:           "https://api.wanikani.com/v2/assignments/100",
			DataUpdatedAt: now,
			Data: domain.AssignmentData{
				SubjectID:   1,
				SubjectType: "kanji",
				SRSStage:    3,
				UnlockedAt:  &now,
			},
		},
	}

	err = store.UpsertAssignments(ctx, assignments)
	if err != nil {
		t.Fatalf("failed to upsert assignments: %v", err)
	}

	// Test get assignments
	retrieved, err := store.GetAssignments(ctx, domain.AssignmentFilters{})
	if err != nil {
		t.Fatalf("failed to get assignments: %v", err)
	}

	if len(retrieved) != 1 {
		t.Errorf("expected 1 assignment, got %d", len(retrieved))
	}

	// Test filter by SRS stage
	srsStage := 3
	filtered, err := store.GetAssignments(ctx, domain.AssignmentFilters{SRSStage: &srsStage})
	if err != nil {
		t.Fatalf("failed to get filtered assignments: %v", err)
	}

	if len(filtered) != 1 {
		t.Errorf("expected 1 assignment with SRS stage 3, got %d", len(filtered))
	}
}

func TestStore_TransactionRollback(t *testing.T) {
	dbPath := "test_transaction.db"
	defer os.Remove(dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Start a transaction
	tx, err := store.BeginTx(ctx)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}

	// Insert a subject within the transaction
	_, err = tx.ExecContext(ctx, `
		INSERT INTO subjects (id, object, url, data_updated_at, data)
		VALUES (?, ?, ?, ?, ?)
	`, 1, "kanji", "https://test.com", time.Now().Format(time.RFC3339), `{"level": 1}`)
	if err != nil {
		t.Fatalf("failed to insert in transaction: %v", err)
	}

	// Rollback the transaction
	err = tx.Rollback()
	if err != nil {
		t.Fatalf("failed to rollback transaction: %v", err)
	}

	// Verify the subject was not persisted
	subjects, err := store.GetSubjects(ctx, domain.SubjectFilters{})
	if err != nil {
		t.Fatalf("failed to get subjects: %v", err)
	}

	if len(subjects) != 0 {
		t.Errorf("expected 0 subjects after rollback, got %d", len(subjects))
	}
}

func TestStore_SyncMetadata(t *testing.T) {
	dbPath := "test_sync.db"
	defer os.Remove(dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Test getting sync time when none exists
	syncTime, err := store.GetLastSyncTime(ctx, domain.DataTypeSubjects)
	if err != nil {
		t.Fatalf("failed to get last sync time: %v", err)
	}

	if syncTime != nil {
		t.Errorf("expected nil sync time, got %v", syncTime)
	}

	// Set sync time
	now := time.Now()
	err = store.SetLastSyncTime(ctx, domain.DataTypeSubjects, now)
	if err != nil {
		t.Fatalf("failed to set last sync time: %v", err)
	}

	// Get sync time
	syncTime, err = store.GetLastSyncTime(ctx, domain.DataTypeSubjects)
	if err != nil {
		t.Fatalf("failed to get last sync time: %v", err)
	}

	if syncTime == nil {
		t.Fatal("expected sync time, got nil")
	}

	// Compare times (allowing for small differences due to formatting)
	if syncTime.Unix() != now.Unix() {
		t.Errorf("expected sync time %v, got %v", now, syncTime)
	}

	// Update sync time
	later := now.Add(1 * time.Hour)
	err = store.SetLastSyncTime(ctx, domain.DataTypeSubjects, later)
	if err != nil {
		t.Fatalf("failed to update last sync time: %v", err)
	}

	// Verify update
	syncTime, err = store.GetLastSyncTime(ctx, domain.DataTypeSubjects)
	if err != nil {
		t.Fatalf("failed to get updated sync time: %v", err)
	}

	if syncTime.Unix() != later.Unix() {
		t.Errorf("expected updated sync time %v, got %v", later, syncTime)
	}
}

func TestStore_Statistics(t *testing.T) {
	dbPath := "test_statistics.db"
	defer os.Remove(dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Create test statistics
	stats := domain.Statistics{
		Object:        "report",
		URL:           "https://api.wanikani.com/v2/summary",
		DataUpdatedAt: time.Now(),
		Data: domain.StatisticsData{
			Lessons: []domain.LessonStatistics{
				{
					AvailableAt: time.Now(),
					SubjectIDs:  []int{1, 2, 3},
				},
			},
		},
	}

	// Insert first snapshot
	timestamp1 := time.Now().Add(-2 * time.Hour)
	err = store.InsertStatistics(ctx, stats, timestamp1)
	if err != nil {
		t.Fatalf("failed to insert statistics: %v", err)
	}

	// Insert second snapshot
	timestamp2 := time.Now().Add(-1 * time.Hour)
	err = store.InsertStatistics(ctx, stats, timestamp2)
	if err != nil {
		t.Fatalf("failed to insert second statistics: %v", err)
	}

	// Get latest statistics
	latest, err := store.GetLatestStatistics(ctx)
	if err != nil {
		t.Fatalf("failed to get latest statistics: %v", err)
	}

	if latest == nil {
		t.Fatal("expected latest statistics, got nil")
	}

	// Verify it's the most recent one
	if latest.Timestamp.Unix() != timestamp2.Unix() {
		t.Errorf("expected timestamp %v, got %v", timestamp2, latest.Timestamp)
	}

	// Get all statistics
	allStats, err := store.GetStatistics(ctx, nil)
	if err != nil {
		t.Fatalf("failed to get all statistics: %v", err)
	}

	if len(allStats) != 2 {
		t.Errorf("expected 2 statistics snapshots, got %d", len(allStats))
	}
}

func TestStore_ReferentialIntegrity(t *testing.T) {
	dbPath := "test_referential.db"
	defer os.Remove(dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	// Try to insert an assignment without a subject (should fail)
	assignments := []domain.Assignment{
		{
			ID:            100,
			Object:        "assignment",
			URL:           "https://api.wanikani.com/v2/assignments/100",
			DataUpdatedAt: time.Now(),
			Data: domain.AssignmentData{
				SubjectID:   999, // Non-existent subject
				SubjectType: "kanji",
				SRSStage:    3,
			},
		},
	}

	err = store.UpsertAssignments(ctx, assignments)
	if err == nil {
		t.Error("expected error when inserting assignment with non-existent subject, got nil")
	}
}
