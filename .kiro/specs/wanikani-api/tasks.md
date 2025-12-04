# Implementation Plan

- [x] 1. Set up project structure and core interfaces
  - Initialize Go module with `go mod init`
  - Create directory structure: `cmd/`, `internal/`, `pkg/`
  - Define core interfaces in `internal/domain/` for WaniKaniClient, DataStore, SyncService
  - Set up configuration loading from environment variables
  - _Requirements: All_

- [x] 2. Implement database layer with SQLite
  - Create database schema with migrations for subjects, assignments, reviews, statistics_snapshots, sync_metadata tables
  - Implement DataStore interface with SQLite backend
  - Add support for transactions and rollback
  - _Requirements: 2.3, 2.4, 3.2, 3.3, 3.5, 4.2, 4.4, 8.2, 8.3, 10.2_

- [ ]* 2.1 Write property test for upsert idempotence
  - **Property 5: Upsert idempotence**
  - **Validates: Requirements 2.4, 3.3**

- [ ]* 2.2 Write property test for data persistence round trip
  - **Property 4: Data persistence round trip**
  - **Validates: Requirements 2.3, 3.2, 4.2, 8.2**

- [ ]* 2.3 Write property test for transaction atomicity
  - **Property 19: Transaction atomicity**
  - **Validates: Requirements 10.2**

- [x] 3. Implement WaniKani API client
  - Create HTTP client with authentication header support
  - Implement token storage and management
  - Add methods for fetching subjects, assignments, reviews, and statistics
  - Implement automatic pagination handling
  - _Requirements: 1.1, 1.2, 1.4, 2.1, 2.2, 3.1, 4.1, 8.1_

- [x] 3.1 Write property test for token persistence and usage
  - **Property 1: Token persistence and usage**
  - **Validates: Requirements 1.1, 1.2**

- [ ]* 3.2 Write property test for token update
  - **Property 2: Token update replaces previous token**
  - **Validates: Requirements 1.4**

- [ ]* 3.3 Write property test for pagination completeness
  - **Property 3: Pagination completeness**
  - **Validates: Requirements 2.2**

- [x] 4. Implement rate limiting and retry logic
  - Add rate limit tracking from API response headers
  - Implement request throttling to respect rate limits
  - Add exponential backoff retry logic for transient errors
  - Expose rate limit status through GetRateLimitStatus method
  - _Requirements: 2.5, 7.1, 7.2, 7.3, 7.4_

- [ ]* 4.1 Write property test for rate limit compliance
  - **Property 13: Rate limit compliance**
  - **Validates: Requirements 7.2, 7.3**

- [ ]* 4.2 Write property test for rate limit status accuracy
  - **Property 14: Rate limit status accuracy**
  - **Validates: Requirements 7.4**

- [x] 5. Implement sync service with incremental updates
  - Create SyncService that orchestrates data synchronization
  - Implement sync methods for each data type (subjects, assignments, reviews, statistics)
  - Add logic to use last sync timestamp for incremental updates
  - Ensure sync ordering: subjects → assignments → reviews
  - Implement sync locking to prevent concurrent syncs
  - _Requirements: 6.1, 6.2, 6.3, 6.4, 6.5, 9.2, 9.5_

- [x] 5.1 Write property test for incremental sync timestamps
  - **Property 9: Incremental sync uses timestamps**
  - **Validates: Requirements 6.1, 3.4**

- [x] 5.2 Write property test for successful sync updates timestamp
  - **Property 10: Successful sync updates timestamp**
  - **Validates: Requirements 6.2**

- [x] 5.3 Write property test for failed sync preserves timestamp
  - **Property 11: Failed sync preserves timestamp**
  - **Validates: Requirements 6.3**

- [ ]* 5.4 Write property test for sync ordering
  - **Property 12: Sync ordering maintains referential integrity**
  - **Validates: Requirements 6.4, 9.2**

- [ ]* 5.5 Write property test for concurrent sync prevention
  - **Property 18: Concurrent sync prevention**
  - **Validates: Requirements 9.5**

- [x] 6. Implement referential integrity checks
  - Add foreign key constraints in database schema
  - Implement validation to ensure assignments reference valid subjects
  - Implement validation to ensure reviews reference valid assignments
  - _Requirements: 3.5, 4.4_

- [ ]* 6.1 Write property test for referential integrity
  - **Property 6: Referential integrity preservation**
  - **Validates: Requirements 3.5, 4.4**

- [x] 7. Implement API server with REST endpoints
  - Create HTTP server using gorilla/mux
  - Implement GET /api/subjects endpoint with filtering
  - Implement GET /api/assignments endpoint with joins to subjects
  - Implement GET /api/reviews endpoint with joins to assignments and subjects
  - Implement GET /api/statistics/latest endpoint
  - Implement GET /api/statistics endpoint with date range filtering
  - Implement POST /api/sync endpoint to trigger manual sync
  - Implement GET /api/sync/status endpoint
  - _Requirements: 5.1, 5.2, 5.3, 5.5, 8.4, 8.5_

- [x] 7.1 Write property test for query filter correctness
  - **Property 7: Query filter correctness**
  - **Validates: Requirements 5.1, 4.3, 8.4**

- [ ]* 7.2 Write property test for join completeness
  - **Property 8: Join completeness**
  - **Validates: Requirements 5.2, 5.3**

- [ ]* 7.3 Write property test for latest statistics retrieval
  - **Property 16: Latest statistics retrieval**
  - **Validates: Requirements 8.5**

- [ ] 8. Implement input validation and error handling
  - Add validation for query parameters (levels, dates, SRS stages)
  - Implement standardized error response format
  - Add error handling for authentication failures
  - Add error handling for network errors
  - Return appropriate HTTP status codes (400, 404, 500)
  - _Requirements: 1.3, 5.4, 10.1, 10.3, 10.4_

- [ ]* 8.1 Write property test for validation error specificity
  - **Property 20: Validation error specificity**
  - **Validates: Requirements 10.4**

- [ ]* 9. Implement scheduler for daily syncs
  - Integrate robfig/cron library
  - Create Scheduler implementation that triggers SyncService
  - Add configuration for cron expression
  - Implement logging for sync results
  - _Requirements: 9.1, 9.3, 9.4_
  - _Note: Can use system cron to call POST /api/sync endpoint instead_

- [ ]* 9.1 Write property test for sync result logging
  - **Property 17: Sync result logging**
  - **Validates: Requirements 9.3**

- [x] 10. Implement statistics historical tracking
  - Ensure statistics snapshots are stored with timestamps
  - Implement query methods for date range filtering
  - Verify all historical snapshots are preserved
  - _Requirements: 8.3_

- [ ]* 10.1 Write property test for statistics historical preservation
  - **Property 15: Statistics historical preservation**
  - **Validates: Requirements 8.3**

- [x] 11. Create main application entry point
  - Implement cmd/wanikani-api/main.go
  - Wire up all components (client, store, sync service, scheduler, API server)
  - Add graceful shutdown handling
  - Add logging configuration
  - _Requirements: All_

- [x] 12. Add logging throughout the application
  - Use structured logging (e.g., logrus or zap)
  - Log sync operations with results
  - Log API requests and errors
  - Log rate limit events
  - _Requirements: 9.3, 9.4_

- [x] 13. Implement API authentication with token
  - Add LOCAL_API_TOKEN to configuration
  - Create authentication middleware to validate Bearer tokens
  - Apply middleware to all API endpoints except /health
  - Return 401 Unauthorized for missing or invalid tokens
  - Log authentication failures
  - _Requirements: 11.1, 11.2, 11.3, 11.4, 11.5_

- [x] 13.1 Write property test for API authentication enforcement
  - **Property 21: API authentication enforcement**
  - **Validates: Requirements 11.1, 11.2, 11.3**

- [x] 14. Create README with setup and usage instructions
  - Document environment variables (including LOCAL_API_TOKEN)
  - Provide example configuration
  - Document API endpoints and authentication
  - Include build and run instructions
  - _Requirements: All_

- [x] 15. Final checkpoint - Ensure all tests pass
  - Ensure all tests pass, ask the user if questions arise.

- [x] 16. Implement assignment snapshots data model and database schema
  - Add AssignmentSnapshot and AssignmentSnapshotSummary types to internal/domain/types.go
  - Add GetSRSStageName() helper function for SRS stage name mapping
  - Create database migration for assignment_snapshots table with composite primary key
  - _Requirements: 12.1, 12.2, 12.3_

- [x] 17. Implement assignment snapshot storage methods
  - Add UpsertAssignmentSnapshot method to DataStore interface
  - Add GetAssignmentSnapshots method with date range filtering
  - Add CalculateAssignmentSnapshot method to compute snapshot from current assignments
  - Implement methods in SQLite store with proper transaction handling
  - _Requirements: 12.1, 12.2, 12.3, 12.4_

- [ ]* 17.1 Write property test for snapshot data persistence
  - **Property 24: Snapshot data persistence round trip**
  - **Validates: Requirements 12.3**

- [ ]* 17.2 Write property test for snapshot upsert idempotence
  - **Property 25: Snapshot upsert idempotence**
  - **Validates: Requirements 12.4**

- [ ]* 17.3 Write property test for snapshot excludes unstarted assignments
  - **Property 23: Snapshot excludes unstarted assignments**
  - **Validates: Requirements 12.2**

- [x] 18. Integrate snapshot creation into sync service
  - Add CreateAssignmentSnapshot method to SyncService interface
  - Implement snapshot creation logic that calls CalculateAssignmentSnapshot
  - Call CreateAssignmentSnapshot after successful SyncAll operation
  - Ensure snapshot creation doesn't fail the entire sync if it encounters errors
  - _Requirements: 12.1, 12.2_

- [ ]* 18.1 Write property test for snapshot creation after sync
  - **Property 22: Snapshot creation after successful sync**
  - **Validates: Requirements 12.1**

- [x] 19. Implement assignment snapshots API endpoint
  - Add GET /api/assignments/snapshots endpoint to routes
  - Implement handler to retrieve snapshots with date range filtering
  - Transform flat snapshot records into nested structure grouped by date and SRS stage name
  - Calculate and include totals for each SRS stage
  - Order results by date in ascending order
  - Add input validation for date range parameters
  - _Requirements: 12.5, 12.6, 12.7, 12.8_

- [ ]* 19.1 Write property test for snapshot API response format
  - **Property 26: Snapshot API response format**
  - **Validates: Requirements 12.5**

- [ ]* 19.2 Write property test for snapshot totals accuracy
  - **Property 27: Snapshot totals accuracy**
  - **Validates: Requirements 12.6**

- [ ]* 19.3 Write property test for snapshot date range filtering
  - **Property 28: Snapshot date range filtering**
  - **Validates: Requirements 12.7**

- [ ]* 19.4 Write property test for snapshot date ordering
  - **Property 29: Snapshot date ordering**
  - **Validates: Requirements 12.8**

- [x] 20. Update documentation for assignment snapshots
  - Add assignment snapshots endpoint to README
  - Document query parameters and response format
  - Provide example API calls and responses
  - Explain SRS stage name mapping
  - _Requirements: 12.5, 12.6, 12.7, 12.8_

- [ ] 21. Final checkpoint - Ensure all tests pass including new snapshot tests
  - Ensure all tests pass, ask the user if questions arise.

- [-] 22. Implement proper database migrations using goose
  - Install goose library: `go get github.com/pressly/goose/v3`
  - Create migrations directory structure at project root
  - Extract current schema from store.go into initial migration file (00001_initial_schema.sql)
  - Include both up and down migrations using goose directives (-- +goose Up/Down)
  - Create migration for assignment_snapshots table (00002_add_assignment_snapshots.sql)
  - Update store.go to remove inline migrate() method and use goose instead
  - Add embed directive to embed migration files in binary
  - Implement runMigrations() function that calls goose.Up() on startup
  - Add migration execution to application startup in main.go
  - Add logging for migration status (version applied, errors, etc.)
  - Test both up and down migrations in development
  - Document migration workflow in README (including optional goose CLI usage)
  - _Design: Database Migration Strategy section_
  - _Note: This addresses the technical debt of having schema creation mixed with application code_
