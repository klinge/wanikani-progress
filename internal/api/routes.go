package api

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// setupRoutes configures all API routes
func setupRoutes(router *mux.Router, handler *Handler, token string, logger *logrus.Logger) {
	// Add CORS middleware to the main router
	router.Use(CORSMiddleware())
	
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

	// Data endpoints (OPTIONS bypass auth, GET/POST require auth)
	api.HandleFunc("/subjects", func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")
	authAPI.HandleFunc("/subjects", handler.HandleGetSubjects).Methods("GET")
	
	api.HandleFunc("/assignments", func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")
	authAPI.HandleFunc("/assignments", handler.HandleGetAssignments).Methods("GET")
	
	api.HandleFunc("/assignments/snapshots", func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")
	authAPI.HandleFunc("/assignments/snapshots", handler.HandleGetAssignmentSnapshots).Methods("GET")
	
	api.HandleFunc("/reviews", func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")
	authAPI.HandleFunc("/reviews", handler.HandleGetReviews).Methods("GET")
	
	api.HandleFunc("/statistics/latest", func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")
	authAPI.HandleFunc("/statistics/latest", handler.HandleGetLatestStatistics).Methods("GET")
	
	api.HandleFunc("/statistics", func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")
	authAPI.HandleFunc("/statistics", handler.HandleGetStatistics).Methods("GET")

	// Sync endpoints
	api.HandleFunc("/sync", func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")
	authAPI.HandleFunc("/sync", handler.HandleTriggerSync).Methods("POST")
	
	api.HandleFunc("/sync/status", func(w http.ResponseWriter, r *http.Request) {}).Methods("OPTIONS")
	authAPI.HandleFunc("/sync/status", handler.HandleGetSyncStatus).Methods("GET")
}
