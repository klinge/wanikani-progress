package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"wanikani-api/internal/domain"
)

// Handler handles HTTP requests
type Handler struct {
	service *Service
	logger  *logrus.Logger
}

// NewHandler creates a new HTTP handler
func NewHandler(service *Service, logger *logrus.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error information
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// writeError writes an error response
func (h *Handler) writeError(w http.ResponseWriter, code int, errorCode, message string, details map[string]string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)

	h.logger.WithFields(logrus.Fields{
		"status_code": code,
		"error_code":  errorCode,
		"message":     message,
		"details":     details,
	}).Warn("API error response")

	json.NewEncoder(w).Encode(ErrorResponse{
		Error: ErrorDetail{
			Code:    errorCode,
			Message: message,
			Details: details,
		},
	})
}

// handleServiceError handles errors from the service layer and writes appropriate HTTP responses
func (h *Handler) handleServiceError(w http.ResponseWriter, err error) {
	// Check for specific error types by examining the error message
	errMsg := err.Error()

	// Authentication errors
	if contains(errMsg, "Invalid API token") || contains(errMsg, "API token not set") {
		h.writeError(w, http.StatusUnauthorized, "AUTH_ERROR", "Authentication failed", map[string]string{
			"detail": "Invalid or missing API token",
		})
		return
	}

	// Network errors
	if contains(errMsg, "network error") || contains(errMsg, "connection") || contains(errMsg, "timeout") {
		h.writeError(w, http.StatusServiceUnavailable, "NETWORK_ERROR", "Unable to connect to WaniKani API", map[string]string{
			"detail": "Please check your network connection and try again",
		})
		return
	}

	// Rate limit errors
	if contains(errMsg, "rate limit") {
		h.writeError(w, http.StatusTooManyRequests, "RATE_LIMIT_ERROR", "Rate limit exceeded", map[string]string{
			"detail": "Too many requests to WaniKani API. Please try again later",
		})
		return
	}

	// Default to internal server error
	h.logger.WithError(err).Error("Unhandled service error")
	h.writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred", nil)
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && stringContains(s, substr)))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// writeJSON writes a JSON response
func writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	json.NewEncoder(w).Encode(data)
}

// HandleGetSubjects handles GET /api/subjects
func (h *Handler) HandleGetSubjects(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filters := domain.SubjectFilters{}

	h.logger.WithField("endpoint", "GET /api/subjects").Debug("Handling request")

	// Parse type filter
	if typeParam := r.URL.Query().Get("type"); typeParam != "" {
		// Validate subject type
		if typeParam != "radical" && typeParam != "kanji" && typeParam != "vocabulary" {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
				"type": "Must be one of: radical, kanji, vocabulary",
			})
			return
		}
		filters.Type = typeParam
	}

	// Parse level filter
	if levelParam := r.URL.Query().Get("level"); levelParam != "" {
		level, err := strconv.Atoi(levelParam)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
				"level": "Must be a valid integer",
			})
			return
		}
		if level < 1 || level > 60 {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
				"level": "Must be between 1 and 60",
			})
			return
		}
		filters.Level = &level
	}

	subjects, err := h.service.GetSubjects(ctx, filters)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"endpoint": "GET /api/subjects",
		"count":    len(subjects),
		"filters":  filters,
	}).Info("Request completed successfully")

	writeJSON(w, subjects)
}

// HandleGetAssignments handles GET /api/assignments
func (h *Handler) HandleGetAssignments(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filters := domain.AssignmentFilters{}

	h.logger.WithField("endpoint", "GET /api/assignments").Debug("Handling request")

	// Parse srs_stage filter
	if srsStageParam := r.URL.Query().Get("srs_stage"); srsStageParam != "" {
		srsStage, err := strconv.Atoi(srsStageParam)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
				"srs_stage": "Must be a valid integer",
			})
			return
		}
		// WaniKani SRS stages range from 0 (initiate) to 9 (burned)
		if srsStage < 0 || srsStage > 9 {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
				"srs_stage": "Must be between 0 and 9",
			})
			return
		}
		filters.SRSStage = &srsStage
	}

	assignments, err := h.service.GetAssignmentsWithSubjects(ctx, filters)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"endpoint": "GET /api/assignments",
		"count":    len(assignments),
		"filters":  filters,
	}).Info("Request completed successfully")

	writeJSON(w, assignments)
}

// HandleGetReviews handles GET /api/reviews
func (h *Handler) HandleGetReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	filters := domain.ReviewFilters{}

	h.logger.WithField("endpoint", "GET /api/reviews").Debug("Handling request")

	// Parse from date filter
	if fromParam := r.URL.Query().Get("from"); fromParam != "" {
		from, err := time.Parse("2006-01-02", fromParam)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
				"from": "Must be in YYYY-MM-DD format",
			})
			return
		}
		filters.From = &from
	}

	// Parse to date filter
	if toParam := r.URL.Query().Get("to"); toParam != "" {
		to, err := time.Parse("2006-01-02", toParam)
		if err != nil {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
				"to": "Must be in YYYY-MM-DD format",
			})
			return
		}
		filters.To = &to
	}

	// Validate date range
	if filters.From != nil && filters.To != nil && filters.From.After(*filters.To) {
		h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
			"from": "Must be before or equal to 'to' date",
		})
		return
	}

	reviews, err := h.service.GetReviewsWithDetails(ctx, filters)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"endpoint": "GET /api/reviews",
		"count":    len(reviews),
		"filters":  filters,
	}).Info("Request completed successfully")

	writeJSON(w, reviews)
}

// HandleGetLatestStatistics handles GET /api/statistics/latest
func (h *Handler) HandleGetLatestStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.WithField("endpoint", "GET /api/statistics/latest").Debug("Handling request")

	snapshot, err := h.service.GetLatestStatistics(ctx)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	if snapshot == nil {
		h.writeError(w, http.StatusNotFound, "NOT_FOUND", "No statistics found", nil)
		return
	}

	h.logger.WithField("endpoint", "GET /api/statistics/latest").Info("Request completed successfully")
	writeJSON(w, snapshot)
}

// HandleGetStatistics handles GET /api/statistics
func (h *Handler) HandleGetStatistics(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var dateRange *domain.DateRange

	h.logger.WithField("endpoint", "GET /api/statistics").Debug("Handling request")

	// Parse date range filters
	fromParam := r.URL.Query().Get("from")
	toParam := r.URL.Query().Get("to")

	if fromParam != "" || toParam != "" {
		dateRange = &domain.DateRange{}

		if fromParam != "" {
			from, err := time.Parse("2006-01-02", fromParam)
			if err != nil {
				h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
					"from": "Must be in YYYY-MM-DD format",
				})
				return
			}
			dateRange.From = from
		}

		if toParam != "" {
			to, err := time.Parse("2006-01-02", toParam)
			if err != nil {
				h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
					"to": "Must be in YYYY-MM-DD format",
				})
				return
			}
			dateRange.To = to
		}

		// Validate date range
		if fromParam != "" && toParam != "" && dateRange.From.After(dateRange.To) {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
				"from": "Must be before or equal to 'to' date",
			})
			return
		}
	}

	snapshots, err := h.service.GetStatistics(ctx, dateRange)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"endpoint":   "GET /api/statistics",
		"count":      len(snapshots),
		"date_range": dateRange,
	}).Info("Request completed successfully")

	writeJSON(w, snapshots)
}

// SyncResponse represents the response from a sync operation
type SyncResponse struct {
	Message string              `json:"message"`
	Results []domain.SyncResult `json:"results"`
}

// HandleTriggerSync handles POST /api/sync
func (h *Handler) HandleTriggerSync(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	h.logger.WithField("endpoint", "POST /api/sync").Info("Manual sync triggered")

	results, err := h.service.TriggerSync(ctx)
	if err != nil {
		if err.Error() == "sync already in progress" {
			h.writeError(w, http.StatusConflict, "SYNC_IN_PROGRESS", "A sync operation is already in progress", nil)
			return
		}
		// Use the standard error handler for other errors
		h.handleServiceError(w, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"endpoint":      "POST /api/sync",
		"results_count": len(results),
	}).Info("Manual sync completed successfully")

	writeJSON(w, SyncResponse{
		Message: "Sync completed successfully",
		Results: results,
	})
}

// SyncStatusResponse represents the sync status
type SyncStatusResponse struct {
	Syncing bool `json:"syncing"`
}

// HandleGetSyncStatus handles GET /api/sync/status
func (h *Handler) HandleGetSyncStatus(w http.ResponseWriter, r *http.Request) {
	h.logger.WithField("endpoint", "GET /api/sync/status").Debug("Handling request")

	syncing := h.service.GetSyncStatus()

	h.logger.WithFields(logrus.Fields{
		"endpoint": "GET /api/sync/status",
		"syncing":  syncing,
	}).Debug("Request completed successfully")

	writeJSON(w, SyncStatusResponse{
		Syncing: syncing,
	})
}

// HandleGetAssignmentSnapshots handles GET /api/assignments/snapshots
func (h *Handler) HandleGetAssignmentSnapshots(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var dateRange *domain.DateRange

	h.logger.WithField("endpoint", "GET /api/assignments/snapshots").Debug("Handling request")

	// Parse date range filters
	fromParam := r.URL.Query().Get("from")
	toParam := r.URL.Query().Get("to")

	if fromParam != "" || toParam != "" {
		dateRange = &domain.DateRange{}

		if fromParam != "" {
			from, err := time.Parse("2006-01-02", fromParam)
			if err != nil {
				h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
					"from": "Must be in YYYY-MM-DD format",
				})
				return
			}
			dateRange.From = from
		}

		if toParam != "" {
			to, err := time.Parse("2006-01-02", toParam)
			if err != nil {
				h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
					"to": "Must be in YYYY-MM-DD format",
				})
				return
			}
			dateRange.To = to
		}

		// Validate date range
		if fromParam != "" && toParam != "" && dateRange.From.After(dateRange.To) {
			h.writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid query parameters", map[string]string{
				"from": "Must be before or equal to 'to' date",
			})
			return
		}
	}

	snapshots, err := h.service.GetAssignmentSnapshots(ctx, dateRange)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.logger.WithFields(logrus.Fields{
		"endpoint":   "GET /api/assignments/snapshots",
		"date_range": dateRange,
	}).Info("Request completed successfully")

	writeJSON(w, snapshots)
}
