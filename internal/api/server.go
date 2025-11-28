package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"wanikani-api/internal/domain"
)

// Server represents the API server
type Server struct {
	router  *mux.Router
	server  *http.Server
	handler *Handler
	logger  *logrus.Logger
}

// NewServer creates a new API server
func NewServer(store domain.DataStore, syncService domain.SyncService, port int, token string, logger *logrus.Logger) *Server {
	// Create service layer
	service := NewService(store, syncService)

	// Create handler layer
	handler := NewHandler(service, logger)

	// Create router
	router := mux.NewRouter()

	// Setup routes with authentication
	setupRoutes(router, handler, token, logger)

	// Create HTTP server
	s := &Server{
		router:  router,
		handler: handler,
		logger:  logger,
		server: &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		},
	}

	return s
}

// Start starts the API server
func (s *Server) Start() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
