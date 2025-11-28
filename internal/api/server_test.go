package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"wanikani-api/internal/domain"
	"wanikani-api/internal/store/sqlite"
	"wanikani-api/internal/sync"
	"wanikani-api/internal/wanikani"
)

// setupTestServer creates a test server with an in-memory database
func setupTestServer(t *testing.T) (*Server, *sqlite.Store) {
	// Create in-memory SQLite database
	store, err := sqlite.New(":memory:")
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}

	logger := testLogger()

	// Create a mock client
	client := wanikani.NewClient(logger)
	client.SetAPIToken("test-token")

	// Create sync service
	syncService := sync.NewService(client, store, logger)

	// Create server without authentication for tests
	server := NewServer(store, syncService, 8080, "", logger)

	return server, store
}

// getRouter returns the router for testing
func (s *Server) getRouter() *mux.Router {
	return s.router
}

func TestGetSubjects(t *testing.T) {
	server, store := setupTestServer(t)
	defer store.Close()

	ctx := context.Background()

	// Insert test subjects
	testSubjects := []domain.Subject{
		{
			ID:            1,
			Object:        "radical",
			URL:           "https://api.wanikani.com/v2/subjects/1",
			DataUpdatedAt: time.Now(),
			Data: domain.SubjectData{
				Level:      1,
				Characters: "一",
				Meanings: []domain.Meaning{
					{Meaning: "one", Primary: true},
				},
			},
		},
		{
			ID:            2,
			Object:        "kanji",
			URL:           "https://api.wanikani.com/v2/subjects/2",
			DataUpdatedAt: time.Now(),
			Data: domain.SubjectData{
				Level:      2,
				Characters: "二",
				Meanings: []domain.Meaning{
					{Meaning: "two", Primary: true},
				},
			},
		},
	}

	err := store.UpsertSubjects(ctx, testSubjects)
	if err != nil {
		t.Fatalf("Failed to insert test subjects: %v", err)
	}

	// Test GET /api/subjects
	req := httptest.NewRequest("GET", "/api/subjects", nil)
	w := httptest.NewRecorder()
	server.getRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var subjects []domain.Subject
	if err := json.NewDecoder(w.Body).Decode(&subjects); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(subjects) != 2 {
		t.Errorf("Expected 2 subjects, got %d", len(subjects))
	}
}

func TestGetSubjectsWithFilters(t *testing.T) {
	server, store := setupTestServer(t)
	defer store.Close()

	ctx := context.Background()

	// Insert test subjects
	testSubjects := []domain.Subject{
		{
			ID:            1,
			Object:        "radical",
			URL:           "https://api.wanikani.com/v2/subjects/1",
			DataUpdatedAt: time.Now(),
			Data: domain.SubjectData{
				Level:      1,
				Characters: "一",
				Meanings: []domain.Meaning{
					{Meaning: "one", Primary: true},
				},
			},
		},
		{
			ID:            2,
			Object:        "kanji",
			URL:           "https://api.wanikani.com/v2/subjects/2",
			DataUpdatedAt: time.Now(),
			Data: domain.SubjectData{
				Level:      2,
				Characters: "二",
				Meanings: []domain.Meaning{
					{Meaning: "two", Primary: true},
				},
			},
		},
	}

	err := store.UpsertSubjects(ctx, testSubjects)
	if err != nil {
		t.Fatalf("Failed to insert test subjects: %v", err)
	}

	// Test with level filter
	req := httptest.NewRequest("GET", "/api/subjects?level=1", nil)
	w := httptest.NewRecorder()
	server.getRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var subjects []domain.Subject
	if err := json.NewDecoder(w.Body).Decode(&subjects); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(subjects) != 1 {
		t.Errorf("Expected 1 subject, got %d", len(subjects))
	}

	if subjects[0].Data.Level != 1 {
		t.Errorf("Expected level 1, got %d", subjects[0].Data.Level)
	}
}

func TestGetSubjectsInvalidLevel(t *testing.T) {
	server, store := setupTestServer(t)
	defer store.Close()

	// Test with invalid level (out of range)
	req := httptest.NewRequest("GET", "/api/subjects?level=100", nil)
	w := httptest.NewRecorder()
	server.getRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("Expected error code VALIDATION_ERROR, got %s", errResp.Error.Code)
	}
}

func TestGetAssignments(t *testing.T) {
	server, store := setupTestServer(t)
	defer store.Close()

	ctx := context.Background()

	// Insert test subject first
	testSubjects := []domain.Subject{
		{
			ID:            1,
			Object:        "kanji",
			URL:           "https://api.wanikani.com/v2/subjects/1",
			DataUpdatedAt: time.Now(),
			Data: domain.SubjectData{
				Level:      1,
				Characters: "一",
				Meanings: []domain.Meaning{
					{Meaning: "one", Primary: true},
				},
			},
		},
	}

	err := store.UpsertSubjects(ctx, testSubjects)
	if err != nil {
		t.Fatalf("Failed to insert test subjects: %v", err)
	}

	// Insert test assignment
	testAssignments := []domain.Assignment{
		{
			ID:            1,
			Object:        "assignment",
			URL:           "https://api.wanikani.com/v2/assignments/1",
			DataUpdatedAt: time.Now(),
			Data: domain.AssignmentData{
				SubjectID:   1,
				SubjectType: "kanji",
				SRSStage:    1,
			},
		},
	}

	err = store.UpsertAssignments(ctx, testAssignments)
	if err != nil {
		t.Fatalf("Failed to insert test assignments: %v", err)
	}

	// Test GET /api/assignments
	req := httptest.NewRequest("GET", "/api/assignments", nil)
	w := httptest.NewRecorder()
	server.getRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var assignments []AssignmentWithSubject
	if err := json.NewDecoder(w.Body).Decode(&assignments); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(assignments) != 1 {
		t.Errorf("Expected 1 assignment, got %d", len(assignments))
	}

	if assignments[0].Subject == nil {
		t.Error("Expected subject to be joined, got nil")
	}

	if assignments[0].Subject.ID != 1 {
		t.Errorf("Expected subject ID 1, got %d", assignments[0].Subject.ID)
	}
}

func TestGetLatestStatistics(t *testing.T) {
	server, store := setupTestServer(t)
	defer store.Close()

	ctx := context.Background()

	// Insert test statistics
	testStats := domain.Statistics{
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

	err := store.InsertStatistics(ctx, testStats, time.Now())
	if err != nil {
		t.Fatalf("Failed to insert test statistics: %v", err)
	}

	// Test GET /api/statistics/latest
	req := httptest.NewRequest("GET", "/api/statistics/latest", nil)
	w := httptest.NewRecorder()
	server.getRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var snapshot domain.StatisticsSnapshot
	if err := json.NewDecoder(w.Body).Decode(&snapshot); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(snapshot.Statistics.Data.Lessons) != 1 {
		t.Errorf("Expected 1 lesson statistic, got %d", len(snapshot.Statistics.Data.Lessons))
	}
}

func TestGetSyncStatus(t *testing.T) {
	server, store := setupTestServer(t)
	defer store.Close()

	// Test GET /api/sync/status
	req := httptest.NewRequest("GET", "/api/sync/status", nil)
	w := httptest.NewRecorder()
	server.getRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var status SyncStatusResponse
	if err := json.NewDecoder(w.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if status.Syncing {
		t.Error("Expected syncing to be false initially")
	}
}

func TestInvalidDateFormat(t *testing.T) {
	server, store := setupTestServer(t)
	defer store.Close()

	// Test with invalid date format
	req := httptest.NewRequest("GET", "/api/reviews?from=invalid-date", nil)
	w := httptest.NewRecorder()
	server.getRouter().ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	if errResp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("Expected error code VALIDATION_ERROR, got %s", errResp.Error.Code)
	}
}
