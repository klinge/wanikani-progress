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
