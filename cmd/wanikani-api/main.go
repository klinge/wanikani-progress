package main

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"wanikani-api/internal/api"
	"wanikani-api/internal/config"
	"wanikani-api/internal/migrations"
	"wanikani-api/internal/store/sqlite"
	"wanikani-api/internal/sync"
	"wanikani-api/internal/utils"
	"wanikani-api/internal/wanikani"
)

func main() {
	// Load configuration first to get log level
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize structured logging
	log := logger.Init(cfg.LogLevel)
	log.Info("Starting WaniKani API application...")

	log.WithFields(map[string]interface{}{
		"api_port":      cfg.APIPort,
		"database_path": cfg.DatabasePath,
		"sync_schedule": cfg.SyncSchedule,
		"log_level":     cfg.LogLevel,
	}).Info("Configuration loaded")

	// Run database migrations
	log.Info("Running database migrations...")
	db, err := sql.Open("sqlite3", cfg.DatabasePath)
	if err != nil {
		log.WithError(err).Fatal("Failed to open database for migrations")
	}

	if err := migrations.Run(db); err != nil {
		db.Close()
		log.WithError(err).Fatal("Failed to run database migrations")
	}

	version, err := migrations.Version(db)
	if err != nil {
		log.WithError(err).Warn("Failed to get migration version")
	} else {
		log.WithField("version", version).Info("Database migrations completed successfully")
	}

	// Close the migration connection
	if err := db.Close(); err != nil {
		log.WithError(err).Warn("Failed to close migration database connection")
	}

	// Initialize database store
	store, err := sqlite.New(cfg.DatabasePath)
	if err != nil {
		log.WithError(err).Fatal("Failed to initialize database")
	}
	defer func() {
		if err := store.Close(); err != nil {
			log.WithError(err).Error("Error closing database")
		}
	}()
	log.Info("Database store initialized successfully")

	// Initialize WaniKani API client
	client := wanikani.NewClient(log)
	client.SetAPIToken(cfg.WaniKaniAPIToken)
	log.Info("WaniKani API client initialized")

	// Initialize sync service
	syncService := sync.NewService(client, store, log)
	log.Info("Sync service initialized")

	// Initialize API server
	server := api.NewServer(store, syncService, cfg.APIPort, cfg.LocalAPIToken, log)
	log.WithField("port", cfg.APIPort).Info("API server initialized")

	// Start API server in a goroutine
	serverErrors := make(chan error, 1)
	go func() {
		log.WithField("port", cfg.APIPort).Info("API server listening")
		if err := server.Start(); err != nil {
			serverErrors <- fmt.Errorf("API server error: %w", err)
		}
	}()

	// Set up graceful shutdown
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Wait for shutdown signal or server error
	select {
	case err := <-serverErrors:
		log.WithError(err).Fatal("Server error")
	case sig := <-shutdown:
		log.WithField("signal", sig).Info("Received shutdown signal")

		// Create shutdown context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Gracefully shutdown the server
		log.Info("Shutting down API server...")
		if err := server.Shutdown(ctx); err != nil {
			log.WithError(err).Error("Error during server shutdown")
		}

		log.Info("Application shutdown complete")
	}
}
