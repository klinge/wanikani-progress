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
