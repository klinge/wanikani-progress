package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// setupRoutes configures all API routes
func setupRoutes(router *mux.Router, handler *Handler, token string, logger *logrus.Logger) {
	// API routes
	api := router.PathPrefix("/api").Subrouter()

	// Health check endpoint (no authentication required)
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}).Methods("GET")

	// Create authenticated subrouter for protected endpoints
	authAPI := api.NewRoute().Subrouter()

	// Apply authentication middleware if token is configured
	if token != "" {
		authAPI.Use(AuthMiddleware(token, logger))
		logger.Info("API authentication enabled")
	} else {
		logger.Warn("LOCAL_API_TOKEN not configured - API running without authentication")
	}

	// Data endpoints
	authAPI.HandleFunc("/subjects", handler.HandleGetSubjects).Methods("GET")
	authAPI.HandleFunc("/assignments", handler.HandleGetAssignments).Methods("GET")
	authAPI.HandleFunc("/assignments/snapshots", handler.HandleGetAssignmentSnapshots).Methods("GET")
	authAPI.HandleFunc("/reviews", handler.HandleGetReviews).Methods("GET")
	authAPI.HandleFunc("/statistics/latest", handler.HandleGetLatestStatistics).Methods("GET")
	authAPI.HandleFunc("/statistics", handler.HandleGetStatistics).Methods("GET")

	// Sync endpoints
	authAPI.HandleFunc("/sync", handler.HandleTriggerSync).Methods("POST")
	authAPI.HandleFunc("/sync/status", handler.HandleGetSyncStatus).Methods("GET")
}
