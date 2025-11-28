package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wanikani-api/internal/domain"
)

// TestAuthenticationErrorHandling tests that authentication errors are properly handled
func TestAuthenticationErrorHandling(t *testing.T) {
	store := &errorMockStore{authError: true}
	syncService := &mockSyncService{}
	service := NewService(store, syncService)
	handler := NewHandler(service, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subjects", nil)
	w := httptest.NewRecorder()

	handler.HandleGetSubjects(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected status 401, got %d", w.Code)
	}
}

// TestNetworkErrorHandling tests that network errors are properly handled
func TestNetworkErrorHandling(t *testing.T) {
	store := &errorMockStore{networkError: true}
	syncService := &mockSyncService{}
	service := NewService(store, syncService)
	handler := NewHandler(service, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subjects", nil)
	w := httptest.NewRecorder()

	handler.HandleGetSubjects(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected status 503, got %d", w.Code)
	}
}

// TestRateLimitErrorHandling tests that rate limit errors are properly handled
func TestRateLimitErrorHandling(t *testing.T) {
	store := &errorMockStore{rateLimitError: true}
	syncService := &mockSyncService{}
	service := NewService(store, syncService)
	handler := NewHandler(service, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subjects", nil)
	w := httptest.NewRecorder()

	handler.HandleGetSubjects(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

// TestInternalErrorHandling tests that generic errors are handled as internal errors
func TestInternalErrorHandling(t *testing.T) {
	store := &errorMockStore{genericError: true}
	syncService := &mockSyncService{}
	service := NewService(store, syncService)
	handler := NewHandler(service, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/subjects", nil)
	w := httptest.NewRecorder()

	handler.HandleGetSubjects(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", w.Code)
	}
}

// errorMockStore is a mock store that returns specific error types
type errorMockStore struct {
	authError      bool
	networkError   bool
	rateLimitError bool
	genericError   bool
}

func (m *errorMockStore) UpsertSubjects(ctx context.Context, subjects []domain.Subject) error {
	return m.getError()
}

func (m *errorMockStore) GetSubjects(ctx context.Context, filters domain.SubjectFilters) ([]domain.Subject, error) {
	return nil, m.getError()
}

func (m *errorMockStore) UpsertAssignments(ctx context.Context, assignments []domain.Assignment) error {
	return m.getError()
}

func (m *errorMockStore) GetAssignments(ctx context.Context, filters domain.AssignmentFilters) ([]domain.Assignment, error) {
	return nil, m.getError()
}

func (m *errorMockStore) UpsertReviews(ctx context.Context, reviews []domain.Review) error {
	return m.getError()
}

func (m *errorMockStore) GetReviews(ctx context.Context, filters domain.ReviewFilters) ([]domain.Review, error) {
	return nil, m.getError()
}

func (m *errorMockStore) InsertStatistics(ctx context.Context, stats domain.Statistics, timestamp time.Time) error {
	return m.getError()
}

func (m *errorMockStore) GetStatistics(ctx context.Context, dateRange *domain.DateRange) ([]domain.StatisticsSnapshot, error) {
	return nil, m.getError()
}

func (m *errorMockStore) GetLatestStatistics(ctx context.Context) (*domain.StatisticsSnapshot, error) {
	return nil, m.getError()
}

func (m *errorMockStore) GetLastSyncTime(ctx context.Context, dataType domain.DataType) (*time.Time, error) {
	return nil, m.getError()
}

func (m *errorMockStore) SetLastSyncTime(ctx context.Context, dataType domain.DataType, timestamp time.Time) error {
	return m.getError()
}

func (m *errorMockStore) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return nil, m.getError()
}

func (m *errorMockStore) UpsertAssignmentSnapshot(ctx context.Context, snapshot domain.AssignmentSnapshot) error {
	return m.getError()
}

func (m *errorMockStore) GetAssignmentSnapshots(ctx context.Context, dateRange *domain.DateRange) ([]domain.AssignmentSnapshot, error) {
	return nil, m.getError()
}

func (m *errorMockStore) CalculateAssignmentSnapshot(ctx context.Context, date time.Time) ([]domain.AssignmentSnapshot, error) {
	return nil, m.getError()
}

func (m *errorMockStore) getError() error {
	if m.authError {
		return errors.New("Invalid API token")
	}
	if m.networkError {
		return errors.New("network error: connection refused")
	}
	if m.rateLimitError {
		return errors.New("rate limit exceeded")
	}
	if m.genericError {
		return errors.New("database error")
	}
	return nil
}
