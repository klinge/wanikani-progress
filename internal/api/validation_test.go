package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wanikani-api/internal/domain"
)

// TestSubjectTypeValidation tests validation of subject type parameter
func TestSubjectTypeValidation(t *testing.T) {
	tests := []struct {
		name           string
		typeParam      string
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "valid radical type",
			typeParam:      "radical",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid kanji type",
			typeParam:      "kanji",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid vocabulary type",
			typeParam:      "vocabulary",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid type",
			typeParam:      "invalid",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty type",
			typeParam:      "",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			syncService := &mockSyncService{}
			service := NewService(store, syncService)
			handler := NewHandler(service, testLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/subjects?type="+tt.typeParam, nil)
			w := httptest.NewRecorder()

			handler.HandleGetSubjects(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectError {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if errResp.Error.Code != "VALIDATION_ERROR" {
					t.Errorf("expected VALIDATION_ERROR, got %s", errResp.Error.Code)
				}
				if errResp.Error.Details["type"] == "" {
					t.Error("expected type field in error details")
				}
			}
		})
	}
}

// TestLevelValidation tests validation of level parameter
func TestLevelValidation(t *testing.T) {
	tests := []struct {
		name           string
		levelParam     string
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "valid level 1",
			levelParam:     "1",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid level 60",
			levelParam:     "60",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid level 30",
			levelParam:     "30",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "level too low",
			levelParam:     "0",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "level too high",
			levelParam:     "61",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid level format",
			levelParam:     "abc",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			syncService := &mockSyncService{}
			service := NewService(store, syncService)
			handler := NewHandler(service, testLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/subjects?level="+tt.levelParam, nil)
			w := httptest.NewRecorder()

			handler.HandleGetSubjects(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectError {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if errResp.Error.Code != "VALIDATION_ERROR" {
					t.Errorf("expected VALIDATION_ERROR, got %s", errResp.Error.Code)
				}
			}
		})
	}
}

// TestSRSStageValidation tests validation of SRS stage parameter
func TestSRSStageValidation(t *testing.T) {
	tests := []struct {
		name           string
		srsStageParam  string
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "valid stage 0",
			srsStageParam:  "0",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid stage 9",
			srsStageParam:  "9",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "valid stage 5",
			srsStageParam:  "5",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "stage too low",
			srsStageParam:  "-1",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "stage too high",
			srsStageParam:  "10",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid stage format",
			srsStageParam:  "abc",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			syncService := &mockSyncService{}
			service := NewService(store, syncService)
			handler := NewHandler(service, testLogger())

			req := httptest.NewRequest(http.MethodGet, "/api/assignments?srs_stage="+tt.srsStageParam, nil)
			w := httptest.NewRecorder()

			handler.HandleGetAssignments(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectError {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if errResp.Error.Code != "VALIDATION_ERROR" {
					t.Errorf("expected VALIDATION_ERROR, got %s", errResp.Error.Code)
				}
			}
		})
	}
}

// TestDateRangeValidation tests validation of date range parameters
func TestDateRangeValidation(t *testing.T) {
	tests := []struct {
		name           string
		fromParam      string
		toParam        string
		expectError    bool
		expectedStatus int
		errorField     string
	}{
		{
			name:           "valid date range",
			fromParam:      "2024-01-01",
			toParam:        "2024-01-31",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "same dates",
			fromParam:      "2024-01-01",
			toParam:        "2024-01-01",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "from after to",
			fromParam:      "2024-01-31",
			toParam:        "2024-01-01",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
			errorField:     "from",
		},
		{
			name:           "invalid from format",
			fromParam:      "2024/01/01",
			toParam:        "2024-01-31",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
			errorField:     "from",
		},
		{
			name:           "invalid to format",
			fromParam:      "2024-01-01",
			toParam:        "2024/01/31",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
			errorField:     "to",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			syncService := &mockSyncService{}
			service := NewService(store, syncService)
			handler := NewHandler(service, testLogger())

			url := "/api/reviews?"
			if tt.fromParam != "" {
				url += "from=" + tt.fromParam
			}
			if tt.toParam != "" {
				if tt.fromParam != "" {
					url += "&"
				}
				url += "to=" + tt.toParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.HandleGetReviews(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectError {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if errResp.Error.Code != "VALIDATION_ERROR" {
					t.Errorf("expected VALIDATION_ERROR, got %s", errResp.Error.Code)
				}
				if tt.errorField != "" && errResp.Error.Details[tt.errorField] == "" {
					t.Errorf("expected %s field in error details", tt.errorField)
				}
			}
		})
	}
}

// TestStatisticsDateRangeValidation tests validation of date range for statistics endpoint
func TestStatisticsDateRangeValidation(t *testing.T) {
	tests := []struct {
		name           string
		fromParam      string
		toParam        string
		expectError    bool
		expectedStatus int
	}{
		{
			name:           "valid date range",
			fromParam:      "2024-01-01",
			toParam:        "2024-01-31",
			expectError:    false,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "from after to",
			fromParam:      "2024-01-31",
			toParam:        "2024-01-01",
			expectError:    true,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStore{}
			syncService := &mockSyncService{}
			service := NewService(store, syncService)
			handler := NewHandler(service, testLogger())

			url := "/api/statistics?"
			if tt.fromParam != "" {
				url += "from=" + tt.fromParam
			}
			if tt.toParam != "" {
				if tt.fromParam != "" {
					url += "&"
				}
				url += "to=" + tt.toParam
			}

			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()

			handler.HandleGetStatistics(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectError {
				var errResp ErrorResponse
				if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
					t.Fatalf("failed to decode error response: %v", err)
				}
				if errResp.Error.Code != "VALIDATION_ERROR" {
					t.Errorf("expected VALIDATION_ERROR, got %s", errResp.Error.Code)
				}
			}
		})
	}
}

// TestErrorResponseFormat tests that error responses follow the standardized format
func TestErrorResponseFormat(t *testing.T) {
	store := &mockStore{}
	syncService := &mockSyncService{}
	service := NewService(store, syncService)
	handler := NewHandler(service, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subjects?level=invalid", nil)
	w := httptest.NewRecorder()

	handler.HandleGetSubjects(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}

	var errResp ErrorResponse
	if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	// Verify error response structure
	if errResp.Error.Code == "" {
		t.Error("error code should not be empty")
	}
	if errResp.Error.Message == "" {
		t.Error("error message should not be empty")
	}
	if errResp.Error.Details == nil {
		t.Error("error details should not be nil")
	}
}

// Mock implementations for testing
type mockStore struct{}

func (m *mockStore) UpsertSubjects(ctx context.Context, subjects []domain.Subject) error {
	return nil
}

func (m *mockStore) GetSubjects(ctx context.Context, filters domain.SubjectFilters) ([]domain.Subject, error) {
	return []domain.Subject{}, nil
}

func (m *mockStore) UpsertAssignments(ctx context.Context, assignments []domain.Assignment) error {
	return nil
}

func (m *mockStore) GetAssignments(ctx context.Context, filters domain.AssignmentFilters) ([]domain.Assignment, error) {
	return []domain.Assignment{}, nil
}

func (m *mockStore) UpsertReviews(ctx context.Context, reviews []domain.Review) error {
	return nil
}

func (m *mockStore) GetReviews(ctx context.Context, filters domain.ReviewFilters) ([]domain.Review, error) {
	return []domain.Review{}, nil
}

func (m *mockStore) InsertStatistics(ctx context.Context, stats domain.Statistics, timestamp time.Time) error {
	return nil
}

func (m *mockStore) GetStatistics(ctx context.Context, dateRange *domain.DateRange) ([]domain.StatisticsSnapshot, error) {
	return []domain.StatisticsSnapshot{}, nil
}

func (m *mockStore) GetLatestStatistics(ctx context.Context) (*domain.StatisticsSnapshot, error) {
	return &domain.StatisticsSnapshot{}, nil
}

func (m *mockStore) GetLastSyncTime(ctx context.Context, dataType domain.DataType) (*time.Time, error) {
	return nil, nil
}

func (m *mockStore) SetLastSyncTime(ctx context.Context, dataType domain.DataType, timestamp time.Time) error {
	return nil
}

func (m *mockStore) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return nil, nil
}

func (m *mockStore) UpsertAssignmentSnapshot(ctx context.Context, snapshot domain.AssignmentSnapshot) error {
	return nil
}

func (m *mockStore) GetAssignmentSnapshots(ctx context.Context, dateRange *domain.DateRange) ([]domain.AssignmentSnapshot, error) {
	return []domain.AssignmentSnapshot{}, nil
}

func (m *mockStore) CalculateAssignmentSnapshot(ctx context.Context, date time.Time) ([]domain.AssignmentSnapshot, error) {
	return []domain.AssignmentSnapshot{}, nil
}

type mockSyncService struct{}

func (m *mockSyncService) SyncAll(ctx context.Context) ([]domain.SyncResult, error) {
	return []domain.SyncResult{}, nil
}

func (m *mockSyncService) SyncSubjects(ctx context.Context) domain.SyncResult {
	return domain.SyncResult{}
}

func (m *mockSyncService) SyncAssignments(ctx context.Context) domain.SyncResult {
	return domain.SyncResult{}
}

func (m *mockSyncService) SyncReviews(ctx context.Context) domain.SyncResult {
	return domain.SyncResult{}
}

func (m *mockSyncService) SyncStatistics(ctx context.Context) domain.SyncResult {
	return domain.SyncResult{}
}

func (m *mockSyncService) IsSyncing() bool {
	return false
}

func (m *mockSyncService) CreateAssignmentSnapshot(ctx context.Context) error {
	return nil
}

// TestAssignmentSnapshotsEndpoint tests the assignment snapshots endpoint
func TestAssignmentSnapshotsEndpoint(t *testing.T) {
	store := &mockStore{}
	syncService := &mockSyncService{}
	service := NewService(store, syncService)
	handler := NewHandler(service, testLogger())

	t.Run("valid request without date range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/assignments/snapshots", nil)
		w := httptest.NewRecorder()

		handler.HandleGetAssignmentSnapshots(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		// Verify response is valid JSON
		var result map[string]map[string]map[string]int
		if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
	})

	t.Run("valid request with date range", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/assignments/snapshots?from=2024-01-01&to=2024-01-31", nil)
		w := httptest.NewRecorder()

		handler.HandleGetAssignmentSnapshots(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
	})

	t.Run("invalid from date format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/assignments/snapshots?from=invalid", nil)
		w := httptest.NewRecorder()

		handler.HandleGetAssignmentSnapshots(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var errResp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		if errResp.Error.Code != "VALIDATION_ERROR" {
			t.Errorf("expected VALIDATION_ERROR, got %s", errResp.Error.Code)
		}
	})

	t.Run("invalid to date format", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/assignments/snapshots?to=invalid", nil)
		w := httptest.NewRecorder()

		handler.HandleGetAssignmentSnapshots(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var errResp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		if errResp.Error.Code != "VALIDATION_ERROR" {
			t.Errorf("expected VALIDATION_ERROR, got %s", errResp.Error.Code)
		}
	})

	t.Run("invalid date range - from after to", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/assignments/snapshots?from=2024-01-31&to=2024-01-01", nil)
		w := httptest.NewRecorder()

		handler.HandleGetAssignmentSnapshots(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}

		var errResp ErrorResponse
		if err := json.NewDecoder(w.Body).Decode(&errResp); err != nil {
			t.Fatalf("failed to decode error response: %v", err)
		}

		if errResp.Error.Code != "VALIDATION_ERROR" {
			t.Errorf("expected VALIDATION_ERROR, got %s", errResp.Error.Code)
		}
	})
}

// TestAssignmentSnapshotsDataTransformation tests the data transformation logic
func TestAssignmentSnapshotsDataTransformation(t *testing.T) {
	// Create a custom mock store with test data
	date1, _ := time.Parse("2006-01-02", "2024-01-15")
	date2, _ := time.Parse("2006-01-02", "2024-01-16")

	testSnapshots := []domain.AssignmentSnapshot{
		{Date: date1, SRSStage: 1, SubjectType: "radical", Count: 5},
		{Date: date1, SRSStage: 1, SubjectType: "kanji", Count: 10},
		{Date: date1, SRSStage: 2, SubjectType: "vocabulary", Count: 8},
		{Date: date1, SRSStage: 5, SubjectType: "radical", Count: 12},
		{Date: date1, SRSStage: 5, SubjectType: "kanji", Count: 15},
		{Date: date1, SRSStage: 7, SubjectType: "vocabulary", Count: 20},
		{Date: date2, SRSStage: 1, SubjectType: "kanji", Count: 7},
		{Date: date2, SRSStage: 9, SubjectType: "radical", Count: 30},
	}

	customStore := &customMockStore{snapshots: testSnapshots}
	syncService := &mockSyncService{}
	service := NewService(customStore, syncService)
	handler := NewHandler(service, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/assignments/snapshots", nil)
	w := httptest.NewRecorder()

	handler.HandleGetAssignmentSnapshots(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var result map[string]map[string]map[string]int
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	// Verify structure for date 2024-01-15
	date1Str := "2024-01-15"
	if _, ok := result[date1Str]; !ok {
		t.Fatalf("expected date %s in result", date1Str)
	}

	// Check apprentice stage (stages 1-4) - should combine stages 1 and 2
	if apprentice, ok := result[date1Str]["apprentice"]; ok {
		if apprentice["radical"] != 5 {
			t.Errorf("expected 5 radicals in apprentice, got %d", apprentice["radical"])
		}
		if apprentice["kanji"] != 10 {
			t.Errorf("expected 10 kanji in apprentice, got %d", apprentice["kanji"])
		}
		if apprentice["vocabulary"] != 8 {
			t.Errorf("expected 8 vocabulary in apprentice, got %d", apprentice["vocabulary"])
		}
		expectedTotal := 5 + 10 + 8
		if apprentice["total"] != expectedTotal {
			t.Errorf("expected total %d in apprentice, got %d", expectedTotal, apprentice["total"])
		}
	} else {
		t.Error("expected apprentice stage in result")
	}

	// Check guru stage (stages 5-6)
	if guru, ok := result[date1Str]["guru"]; ok {
		if guru["radical"] != 12 {
			t.Errorf("expected 12 radicals in guru, got %d", guru["radical"])
		}
		if guru["kanji"] != 15 {
			t.Errorf("expected 15 kanji in guru, got %d", guru["kanji"])
		}
		expectedTotal := 12 + 15
		if guru["total"] != expectedTotal {
			t.Errorf("expected total %d in guru, got %d", expectedTotal, guru["total"])
		}
	} else {
		t.Error("expected guru stage in result")
	}

	// Check master stage (stage 7)
	if master, ok := result[date1Str]["master"]; ok {
		if master["vocabulary"] != 20 {
			t.Errorf("expected 20 vocabulary in master, got %d", master["vocabulary"])
		}
		if master["total"] != 20 {
			t.Errorf("expected total 20 in master, got %d", master["total"])
		}
	} else {
		t.Error("expected master stage in result")
	}

	// Verify structure for date 2024-01-16
	date2Str := "2024-01-16"
	if _, ok := result[date2Str]; !ok {
		t.Fatalf("expected date %s in result", date2Str)
	}

	// Check apprentice stage for date 2
	if apprentice, ok := result[date2Str]["apprentice"]; ok {
		if apprentice["kanji"] != 7 {
			t.Errorf("expected 7 kanji in apprentice, got %d", apprentice["kanji"])
		}
		if apprentice["total"] != 7 {
			t.Errorf("expected total 7 in apprentice, got %d", apprentice["total"])
		}
	} else {
		t.Error("expected apprentice stage in result for date 2")
	}

	// Check burned stage (stage 9)
	if burned, ok := result[date2Str]["burned"]; ok {
		if burned["radical"] != 30 {
			t.Errorf("expected 30 radicals in burned, got %d", burned["radical"])
		}
		if burned["total"] != 30 {
			t.Errorf("expected total 30 in burned, got %d", burned["total"])
		}
	} else {
		t.Error("expected burned stage in result for date 2")
	}
}

// customMockStore is a mock store that returns custom snapshot data
type customMockStore struct {
	mockStore
	snapshots []domain.AssignmentSnapshot
}

func (m *customMockStore) GetAssignmentSnapshots(ctx context.Context, dateRange *domain.DateRange) ([]domain.AssignmentSnapshot, error) {
	return m.snapshots, nil
}
