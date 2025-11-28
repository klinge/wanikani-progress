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

// TestStore_StatisticsHistoricalTracking tests comprehensive historical tracking of statistics
func TestStore_StatisticsHistoricalTracking(t *testing.T) {
	dbPath := "test_statistics_historical.db"
	defer os.Remove(dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	t.Run("snapshots are stored with timestamps", func(t *testing.T) {
		// Create multiple statistics snapshots with different timestamps
		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		
		for i := 0; i < 5; i++ {
			stats := domain.Statistics{
				Object:        "report",
				URL:           "https://api.wanikani.com/v2/summary",
				DataUpdatedAt: baseTime.Add(time.Duration(i) * 24 * time.Hour),
				Data: domain.StatisticsData{
					Lessons: []domain.LessonStatistics{
						{
							AvailableAt: baseTime.Add(time.Duration(i) * 24 * time.Hour),
							SubjectIDs:  []int{i + 1, i + 2, i + 3},
						},
					},
					Reviews: []domain.ReviewStatistics{
						{
							AvailableAt: baseTime.Add(time.Duration(i) * 24 * time.Hour),
							SubjectIDs:  []int{i * 10, i*10 + 1},
						},
					},
				},
			}
			
			timestamp := baseTime.Add(time.Duration(i) * 24 * time.Hour)
			err := store.InsertStatistics(ctx, stats, timestamp)
			if err != nil {
				t.Fatalf("failed to insert statistics snapshot %d: %v", i, err)
			}
		}

		// Verify all snapshots were stored
		allSnapshots, err := store.GetStatistics(ctx, nil)
		if err != nil {
			t.Fatalf("failed to get all statistics: %v", err)
		}

		if len(allSnapshots) != 5 {
			t.Errorf("expected 5 snapshots, got %d", len(allSnapshots))
		}

		// Verify each snapshot has the correct timestamp
		for i, snapshot := range allSnapshots {
			expectedTime := baseTime.Add(time.Duration(4-i) * 24 * time.Hour) // Reversed order (DESC)
			if snapshot.Timestamp.Unix() != expectedTime.Unix() {
				t.Errorf("snapshot %d: expected timestamp %v, got %v", i, expectedTime, snapshot.Timestamp)
			}
		}
	})

	t.Run("date range filtering works correctly", func(t *testing.T) {
		// Query with date range
		baseTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		dateRange := &domain.DateRange{
			From: baseTime.Add(1 * 24 * time.Hour),
			To:   baseTime.Add(3 * 24 * time.Hour),
		}

		filtered, err := store.GetStatistics(ctx, dateRange)
		if err != nil {
			t.Fatalf("failed to get filtered statistics: %v", err)
		}

		// Should return snapshots from day 1, 2, and 3 (3 snapshots)
		if len(filtered) != 3 {
			t.Errorf("expected 3 snapshots in date range, got %d", len(filtered))
		}

		// Verify all returned snapshots are within the date range
		for _, snapshot := range filtered {
			if snapshot.Timestamp.Before(dateRange.From) || snapshot.Timestamp.After(dateRange.To) {
				t.Errorf("snapshot timestamp %v is outside date range [%v, %v]", 
					snapshot.Timestamp, dateRange.From, dateRange.To)
			}
		}
	})

	t.Run("all historical snapshots are preserved", func(t *testing.T) {
		// Insert more snapshots to verify preservation
		baseTime := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		
		for i := 0; i < 10; i++ {
			stats := domain.Statistics{
				Object:        "report",
				URL:           "https://api.wanikani.com/v2/summary",
				DataUpdatedAt: baseTime.Add(time.Duration(i) * time.Hour),
				Data: domain.StatisticsData{
					Lessons: []domain.LessonStatistics{
						{
							AvailableAt: baseTime.Add(time.Duration(i) * time.Hour),
							SubjectIDs:  []int{100 + i},
						},
					},
				},
			}
			
			timestamp := baseTime.Add(time.Duration(i) * time.Hour)
			err := store.InsertStatistics(ctx, stats, timestamp)
			if err != nil {
				t.Fatalf("failed to insert statistics snapshot: %v", err)
			}
		}

		// Get all snapshots (should include previous 5 + new 10 = 15 total)
		allSnapshots, err := store.GetStatistics(ctx, nil)
		if err != nil {
			t.Fatalf("failed to get all statistics: %v", err)
		}

		if len(allSnapshots) != 15 {
			t.Errorf("expected 15 total snapshots, got %d", len(allSnapshots))
		}

		// Verify snapshots are ordered by timestamp descending
		for i := 1; i < len(allSnapshots); i++ {
			if allSnapshots[i].Timestamp.After(allSnapshots[i-1].Timestamp) {
				t.Errorf("snapshots not ordered correctly: snapshot %d (%v) is after snapshot %d (%v)",
					i, allSnapshots[i].Timestamp, i-1, allSnapshots[i-1].Timestamp)
			}
		}
	})

	t.Run("latest statistics returns most recent snapshot", func(t *testing.T) {
		latest, err := store.GetLatestStatistics(ctx)
		if err != nil {
			t.Fatalf("failed to get latest statistics: %v", err)
		}

		if latest == nil {
			t.Fatal("expected latest statistics, got nil")
		}

		// Get all snapshots to verify latest is actually the most recent
		allSnapshots, err := store.GetStatistics(ctx, nil)
		if err != nil {
			t.Fatalf("failed to get all statistics: %v", err)
		}

		// The latest should match the first in the list (DESC order)
		if latest.ID != allSnapshots[0].ID {
			t.Errorf("latest statistics ID %d doesn't match most recent snapshot ID %d", 
				latest.ID, allSnapshots[0].ID)
		}

		if latest.Timestamp.Unix() != allSnapshots[0].Timestamp.Unix() {
			t.Errorf("latest statistics timestamp %v doesn't match most recent snapshot timestamp %v",
				latest.Timestamp, allSnapshots[0].Timestamp)
		}
	})

	t.Run("empty date range returns all snapshots", func(t *testing.T) {
		allSnapshots, err := store.GetStatistics(ctx, nil)
		if err != nil {
			t.Fatalf("failed to get statistics with nil date range: %v", err)
		}

		if len(allSnapshots) == 0 {
			t.Error("expected snapshots with nil date range, got 0")
		}
	})

	t.Run("statistics data integrity is preserved", func(t *testing.T) {
		// Insert a snapshot with complex data
		baseTime := time.Date(2024, 3, 1, 12, 0, 0, 0, time.UTC)
		stats := domain.Statistics{
			Object:        "report",
			URL:           "https://api.wanikani.com/v2/summary",
			DataUpdatedAt: baseTime,
			Data: domain.StatisticsData{
				Lessons: []domain.LessonStatistics{
					{
						AvailableAt: baseTime,
						SubjectIDs:  []int{1, 2, 3, 4, 5},
					},
					{
						AvailableAt: baseTime.Add(1 * time.Hour),
						SubjectIDs:  []int{6, 7, 8},
					},
				},
				Reviews: []domain.ReviewStatistics{
					{
						AvailableAt: baseTime,
						SubjectIDs:  []int{10, 20, 30},
					},
				},
			},
		}

		err := store.InsertStatistics(ctx, stats, baseTime)
		if err != nil {
			t.Fatalf("failed to insert complex statistics: %v", err)
		}

		// Retrieve and verify data integrity
		retrieved, err := store.GetStatistics(ctx, &domain.DateRange{
			From: baseTime.Add(-1 * time.Minute),
			To:   baseTime.Add(1 * time.Minute),
		})
		if err != nil {
			t.Fatalf("failed to retrieve statistics: %v", err)
		}

		if len(retrieved) != 1 {
			t.Fatalf("expected 1 snapshot, got %d", len(retrieved))
		}

		snapshot := retrieved[0]
		
		// Verify lessons data
		if len(snapshot.Statistics.Data.Lessons) != 2 {
			t.Errorf("expected 2 lesson statistics, got %d", len(snapshot.Statistics.Data.Lessons))
		}

		if len(snapshot.Statistics.Data.Lessons[0].SubjectIDs) != 5 {
			t.Errorf("expected 5 subject IDs in first lesson, got %d", 
				len(snapshot.Statistics.Data.Lessons[0].SubjectIDs))
		}

		// Verify reviews data
		if len(snapshot.Statistics.Data.Reviews) != 1 {
			t.Errorf("expected 1 review statistics, got %d", len(snapshot.Statistics.Data.Reviews))
		}

		if len(snapshot.Statistics.Data.Reviews[0].SubjectIDs) != 3 {
			t.Errorf("expected 3 subject IDs in review, got %d",
				len(snapshot.Statistics.Data.Reviews[0].SubjectIDs))
		}
	})
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

	t.Run("assignment with non-existent subject fails", func(t *testing.T) {
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
	})

	t.Run("assignment with valid subject succeeds", func(t *testing.T) {
		// First create a subject
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

		// Now insert assignment with valid subject
		assignments := []domain.Assignment{
			{
				ID:            100,
				Object:        "assignment",
				URL:           "https://api.wanikani.com/v2/assignments/100",
				DataUpdatedAt: time.Now(),
				Data: domain.AssignmentData{
					SubjectID:   1,
					SubjectType: "kanji",
					SRSStage:    3,
				},
			},
		}

		err = store.UpsertAssignments(ctx, assignments)
		if err != nil {
			t.Errorf("expected no error when inserting assignment with valid subject, got: %v", err)
		}
	})

	t.Run("review with non-existent assignment fails", func(t *testing.T) {
		// Try to insert a review without an assignment (should fail)
		reviews := []domain.Review{
			{
				ID:            200,
				Object:        "review",
				URL:           "https://api.wanikani.com/v2/reviews/200",
				DataUpdatedAt: time.Now(),
				Data: domain.ReviewData{
					AssignmentID: 999, // Non-existent assignment
					SubjectID:    1,
					CreatedAt:    time.Now(),
				},
			},
		}

		err = store.UpsertReviews(ctx, reviews)
		if err == nil {
			t.Error("expected error when inserting review with non-existent assignment, got nil")
		}
	})

	t.Run("review with non-existent subject fails", func(t *testing.T) {
		// Try to insert a review with non-existent subject (should fail)
		reviews := []domain.Review{
			{
				ID:            201,
				Object:        "review",
				URL:           "https://api.wanikani.com/v2/reviews/201",
				DataUpdatedAt: time.Now(),
				Data: domain.ReviewData{
					AssignmentID: 100, // Valid assignment
					SubjectID:    999, // Non-existent subject
					CreatedAt:    time.Now(),
				},
			},
		}

		err = store.UpsertReviews(ctx, reviews)
		if err == nil {
			t.Error("expected error when inserting review with non-existent subject, got nil")
		}
	})

	t.Run("review with valid assignment and subject succeeds", func(t *testing.T) {
		// Insert a review with valid references
		reviews := []domain.Review{
			{
				ID:            202,
				Object:        "review",
				URL:           "https://api.wanikani.com/v2/reviews/202",
				DataUpdatedAt: time.Now(),
				Data: domain.ReviewData{
					AssignmentID:            100,
					SubjectID:               1,
					CreatedAt:               time.Now(),
					IncorrectMeaningAnswers: 0,
					IncorrectReadingAnswers: 1,
				},
			},
		}

		err = store.UpsertReviews(ctx, reviews)
		if err != nil {
			t.Errorf("expected no error when inserting review with valid references, got: %v", err)
		}
	})
}

func TestStore_AssignmentSnapshots(t *testing.T) {
	dbPath := "test_assignment_snapshots.db"
	defer os.Remove(dbPath)

	store, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	ctx := context.Background()

	t.Run("upsert and get assignment snapshots", func(t *testing.T) {
		// Create test snapshots
		date1 := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
		snapshots := []domain.AssignmentSnapshot{
			{
				Date:        date1,
				SRSStage:    1,
				SubjectType: "kanji",
				Count:       10,
			},
			{
				Date:        date1,
				SRSStage:    1,
				SubjectType: "vocabulary",
				Count:       15,
			},
			{
				Date:        date1,
				SRSStage:    5,
				SubjectType: "kanji",
				Count:       20,
			},
		}

		// Upsert snapshots
		for _, snapshot := range snapshots {
			err := store.UpsertAssignmentSnapshot(ctx, snapshot)
			if err != nil {
				t.Fatalf("failed to upsert snapshot: %v", err)
			}
		}

		// Get all snapshots
		retrieved, err := store.GetAssignmentSnapshots(ctx, nil)
		if err != nil {
			t.Fatalf("failed to get snapshots: %v", err)
		}

		if len(retrieved) != 3 {
			t.Errorf("expected 3 snapshots, got %d", len(retrieved))
		}

		// Verify data
		if retrieved[0].Count != 10 {
			t.Errorf("expected count 10, got %d", retrieved[0].Count)
		}
	})

	t.Run("upsert idempotence", func(t *testing.T) {
		// Upsert the same snapshot twice with different counts
		date := time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)
		snapshot := domain.AssignmentSnapshot{
			Date:        date,
			SRSStage:    2,
			SubjectType: "radical",
			Count:       5,
		}

		err := store.UpsertAssignmentSnapshot(ctx, snapshot)
		if err != nil {
			t.Fatalf("failed to upsert snapshot: %v", err)
		}

		// Update with new count
		snapshot.Count = 8
		err = store.UpsertAssignmentSnapshot(ctx, snapshot)
		if err != nil {
			t.Fatalf("failed to update snapshot: %v", err)
		}

		// Verify only one record exists with updated count
		dateRange := &domain.DateRange{
			From: date,
			To:   date,
		}
		retrieved, err := store.GetAssignmentSnapshots(ctx, dateRange)
		if err != nil {
			t.Fatalf("failed to get snapshots: %v", err)
		}

		count := 0
		for _, s := range retrieved {
			if s.SRSStage == 2 && s.SubjectType == "radical" {
				count++
				if s.Count != 8 {
					t.Errorf("expected count 8, got %d", s.Count)
				}
			}
		}

		if count != 1 {
			t.Errorf("expected 1 snapshot with SRS stage 2 and type radical, got %d", count)
		}
	})

	t.Run("date range filtering", func(t *testing.T) {
		// Create snapshots for multiple dates
		date1 := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
		date2 := time.Date(2024, 2, 2, 0, 0, 0, 0, time.UTC)
		date3 := time.Date(2024, 2, 3, 0, 0, 0, 0, time.UTC)

		snapshots := []domain.AssignmentSnapshot{
			{Date: date1, SRSStage: 1, SubjectType: "kanji", Count: 10},
			{Date: date2, SRSStage: 1, SubjectType: "kanji", Count: 12},
			{Date: date3, SRSStage: 1, SubjectType: "kanji", Count: 15},
		}

		for _, snapshot := range snapshots {
			err := store.UpsertAssignmentSnapshot(ctx, snapshot)
			if err != nil {
				t.Fatalf("failed to upsert snapshot: %v", err)
			}
		}

		// Query with date range
		dateRange := &domain.DateRange{
			From: date1,
			To:   date2,
		}

		filtered, err := store.GetAssignmentSnapshots(ctx, dateRange)
		if err != nil {
			t.Fatalf("failed to get filtered snapshots: %v", err)
		}

		// Count snapshots within the date range
		count := 0
		for _, s := range filtered {
			if !s.Date.Before(date1) && !s.Date.After(date2) {
				count++
			}
		}

		if count < 2 {
			t.Errorf("expected at least 2 snapshots in date range, got %d", count)
		}
	})

	t.Run("calculate assignment snapshot", func(t *testing.T) {
		// First create subjects
		subjects := []domain.Subject{
			{
				ID:            1,
				Object:        "kanji",
				URL:           "https://api.wanikani.com/v2/subjects/1",
				DataUpdatedAt: time.Now(),
				Data:          domain.SubjectData{Level: 5, Characters: "一"},
			},
			{
				ID:            2,
				Object:        "vocabulary",
				URL:           "https://api.wanikani.com/v2/subjects/2",
				DataUpdatedAt: time.Now(),
				Data:          domain.SubjectData{Level: 5, Characters: "一つ"},
			},
			{
				ID:            3,
				Object:        "radical",
				URL:           "https://api.wanikani.com/v2/subjects/3",
				DataUpdatedAt: time.Now(),
				Data:          domain.SubjectData{Level: 1, Characters: "丨"},
			},
		}
		err := store.UpsertSubjects(ctx, subjects)
		if err != nil {
			t.Fatalf("failed to upsert subjects: %v", err)
		}

		// Create assignments with various SRS stages
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
					SRSStage:    1, // Apprentice
					StartedAt:   &now,
				},
			},
			{
				ID:            101,
				Object:        "assignment",
				URL:           "https://api.wanikani.com/v2/assignments/101",
				DataUpdatedAt: now,
				Data: domain.AssignmentData{
					SubjectID:   2,
					SubjectType: "vocabulary",
					SRSStage:    5, // Guru
					StartedAt:   &now,
				},
			},
			{
				ID:            102,
				Object:        "assignment",
				URL:           "https://api.wanikani.com/v2/assignments/102",
				DataUpdatedAt: now,
				Data: domain.AssignmentData{
					SubjectID:   3,
					SubjectType: "radical",
					SRSStage:    0, // Unstarted - should be excluded
					StartedAt:   nil,
				},
			},
		}

		err = store.UpsertAssignments(ctx, assignments)
		if err != nil {
			t.Fatalf("failed to upsert assignments: %v", err)
		}

		// Calculate snapshot
		date := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
		calculated, err := store.CalculateAssignmentSnapshot(ctx, date)
		if err != nil {
			t.Fatalf("failed to calculate snapshot: %v", err)
		}

		// Verify results
		if len(calculated) == 0 {
			t.Fatal("expected calculated snapshots, got none")
		}

		// Verify SRS stage 0 is excluded
		for _, snapshot := range calculated {
			if snapshot.SRSStage == 0 {
				t.Error("SRS stage 0 should be excluded from snapshot")
			}
		}

		// Verify we have snapshots for SRS stages 1 and 5
		foundStage1 := false
		foundStage5 := false
		for _, snapshot := range calculated {
			if snapshot.SRSStage == 1 && snapshot.SubjectType == "kanji" {
				foundStage1 = true
				if snapshot.Count != 1 {
					t.Errorf("expected count 1 for stage 1 kanji, got %d", snapshot.Count)
				}
			}
			if snapshot.SRSStage == 5 && snapshot.SubjectType == "vocabulary" {
				foundStage5 = true
				if snapshot.Count != 1 {
					t.Errorf("expected count 1 for stage 5 vocabulary, got %d", snapshot.Count)
				}
			}
		}

		if !foundStage1 {
			t.Error("expected snapshot for SRS stage 1 kanji")
		}
		if !foundStage5 {
			t.Error("expected snapshot for SRS stage 5 vocabulary")
		}
	})
}
