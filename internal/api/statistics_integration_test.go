package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	_ "github.com/mattn/go-sqlite3"
	"wanikani-api/internal/domain"
	"wanikani-api/internal/migrations"
	"wanikani-api/internal/store/sqlite"
)

// TestStatisticsHistoricalTrackingIntegration tests the full statistics flow
func TestStatisticsHistoricalTrackingIntegration(t *testing.T) {
	// Create temporary database
	dbPath := "test_statistics_integration.db"
	defer os.Remove(dbPath)

	// Open database and run migrations
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	if err := migrations.Run(db); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Fatalf("failed to close migration connection: %v", err)
	}

	// Create store
	store, err := sqlite.New(dbPath)
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer store.Close()

	// Create mock sync service
	mockSync := &mockSyncService{}

	// Create service and handler
	service := NewService(store, mockSync)
	logger := logrus.New()
	logger.SetOutput(os.Stderr)
	handler := NewHandler(service, logger)

	// Create router
	router := mux.NewRouter()
	setupRoutes(router, handler, "", logger)

	ctx := context.Background()

	// Insert multiple statistics snapshots
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
			},
		}

		timestamp := baseTime.Add(time.Duration(i) * 24 * time.Hour)
		err := store.InsertStatistics(ctx, stats, timestamp)
		if err != nil {
			t.Fatalf("failed to insert statistics: %v", err)
		}
	}

	t.Run("GET /api/statistics/latest returns most recent snapshot", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/statistics/latest", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var snapshot domain.StatisticsSnapshot
		err := json.NewDecoder(w.Body).Decode(&snapshot)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Should be the most recent (day 4)
		expectedTime := baseTime.Add(4 * 24 * time.Hour)
		if snapshot.Timestamp.Unix() != expectedTime.Unix() {
			t.Errorf("expected timestamp %v, got %v", expectedTime, snapshot.Timestamp)
		}
	})

	t.Run("GET /api/statistics returns all snapshots", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/statistics", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var snapshots []domain.StatisticsSnapshot
		err := json.NewDecoder(w.Body).Decode(&snapshots)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(snapshots) != 5 {
			t.Errorf("expected 5 snapshots, got %d", len(snapshots))
		}

		// Verify they're in descending order
		for i := 1; i < len(snapshots); i++ {
			if snapshots[i].Timestamp.After(snapshots[i-1].Timestamp) {
				t.Errorf("snapshots not in descending order")
			}
		}
	})

	t.Run("GET /api/statistics with date range filters correctly", func(t *testing.T) {
		// Query for days 1-3
		from := baseTime.Add(1 * 24 * time.Hour).Format("2006-01-02")
		to := baseTime.Add(3 * 24 * time.Hour).Format("2006-01-02")

		req := httptest.NewRequest("GET", "/api/statistics?from="+from+"&to="+to, nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var snapshots []domain.StatisticsSnapshot
		err := json.NewDecoder(w.Body).Decode(&snapshots)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if len(snapshots) != 3 {
			t.Errorf("expected 3 snapshots in date range, got %d", len(snapshots))
		}

		// Verify all are within range
		fromTime := baseTime.Add(1 * 24 * time.Hour)
		toTime := baseTime.Add(3 * 24 * time.Hour)
		for _, snapshot := range snapshots {
			if snapshot.Timestamp.Before(fromTime) || snapshot.Timestamp.After(toTime) {
				t.Errorf("snapshot timestamp %v outside range [%v, %v]",
					snapshot.Timestamp, fromTime, toTime)
			}
		}
	})

	t.Run("historical snapshots are preserved after new inserts", func(t *testing.T) {
		// Insert more snapshots
		for i := 5; i < 10; i++ {
			stats := domain.Statistics{
				Object:        "report",
				URL:           "https://api.wanikani.com/v2/summary",
				DataUpdatedAt: baseTime.Add(time.Duration(i) * 24 * time.Hour),
				Data: domain.StatisticsData{
					Lessons: []domain.LessonStatistics{
						{
							AvailableAt: baseTime.Add(time.Duration(i) * 24 * time.Hour),
							SubjectIDs:  []int{i + 1},
						},
					},
				},
			}

			timestamp := baseTime.Add(time.Duration(i) * 24 * time.Hour)
			err := store.InsertStatistics(ctx, stats, timestamp)
			if err != nil {
				t.Fatalf("failed to insert statistics: %v", err)
			}
		}

		// Query all snapshots
		req := httptest.NewRequest("GET", "/api/statistics", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var snapshots []domain.StatisticsSnapshot
		err := json.NewDecoder(w.Body).Decode(&snapshots)
		if err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		// Should have all 10 snapshots
		if len(snapshots) != 10 {
			t.Errorf("expected 10 total snapshots, got %d", len(snapshots))
		}
	})
}
