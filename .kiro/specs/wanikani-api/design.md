# Design Document

## Overview

The WaniKani Data API system is a local API service that fetches, stores, and serves data from the WaniKani language learning platform. The system consists of three main layers:

1. **API Client Layer**: Handles communication with the external WaniKani API, including authentication, pagination, rate limiting, and error handling
2. **Data Storage Layer**: Persists WaniKani data locally with support for incremental updates and historical tracking
3. **API Server Layer**: Provides a REST API for querying stored data

The system is designed for batch-style daily synchronization rather than real-time updates, using WaniKani's `updated_after` parameter for efficient incremental syncs.

## Architecture

### High-Level Architecture

```
┌─────────────────┐
│   API Server    │ ← REST API for querying local data
└────────┬────────┘
         │
┌────────▼────────┐
│   Data Store    │ ← Local database (SQLite/PostgreSQL)
└────────▲────────┘
         │
┌────────┴────────┐
│   API Client    │ ← Communicates with WaniKani API
└────────┬────────┘
         │
┌────────▼────────┐
│  WaniKani API   │ ← External service
└─────────────────┘
```

### Technology Stack

- **Runtime**: Go 1.21+
- **API Server**: net/http (standard library) with gorilla/mux for routing
- **Database**: SQLite with mattn/go-sqlite3 driver (easily upgradable to PostgreSQL)
- **HTTP Client**: net/http (standard library)
- **Scheduling**: robfig/cron for daily sync jobs
- **Testing**: testing (standard library) for unit tests, gopter for property-based testing
- **Database Migrations**: golang-migrate/migrate

## Components and Interfaces

### 1. WaniKani API Client

**Responsibilities:**
- Authenticate with WaniKani API using API token
- Fetch subjects, assignments, reviews, and statistics
- Handle pagination automatically
- Implement rate limiting and retry logic
- Parse and validate API responses

**Interface:**

```go
type WaniKaniClient interface {
    // Authentication
    SetAPIToken(token string)
    
    // Data fetching
    FetchSubjects(ctx context.Context, updatedAfter *time.Time) ([]Subject, error)
    FetchAssignments(ctx context.Context, updatedAfter *time.Time) ([]Assignment, error)
    FetchReviews(ctx context.Context, updatedAfter *time.Time) ([]Review, error)
    FetchStatistics(ctx context.Context) (*Statistics, error)
    
    // Rate limit info
    GetRateLimitStatus() RateLimitInfo
}

type RateLimitInfo struct {
    Remaining int
    ResetAt   time.Time
}
```
```

### 2. Data Store

**Responsibilities:**
- Persist subjects, assignments, reviews, and statistics
- Support upsert operations (insert or update)
- Maintain referential integrity
- Track sync timestamps
- Provide query interfaces

**Interface:**

```go
type DataStore interface {
    // Subjects
    UpsertSubjects(ctx context.Context, subjects []Subject) error
    GetSubjects(ctx context.Context, filters SubjectFilters) ([]Subject, error)
    
    // Assignments
    UpsertAssignments(ctx context.Context, assignments []Assignment) error
    GetAssignments(ctx context.Context, filters AssignmentFilters) ([]Assignment, error)
    
    // Reviews
    UpsertReviews(ctx context.Context, reviews []Review) error
    GetReviews(ctx context.Context, filters ReviewFilters) ([]Review, error)
    
    // Statistics
    InsertStatistics(ctx context.Context, stats Statistics, timestamp time.Time) error
    GetStatistics(ctx context.Context, dateRange *DateRange) ([]StatisticsSnapshot, error)
    GetLatestStatistics(ctx context.Context) (*StatisticsSnapshot, error)
    
    // Assignment Snapshots
    UpsertAssignmentSnapshot(ctx context.Context, snapshot AssignmentSnapshot) error
    GetAssignmentSnapshots(ctx context.Context, dateRange *DateRange) ([]AssignmentSnapshot, error)
    CalculateAssignmentSnapshot(ctx context.Context, date time.Time) ([]AssignmentSnapshot, error)
    
    // Sync tracking
    GetLastSyncTime(ctx context.Context, dataType DataType) (*time.Time, error)
    SetLastSyncTime(ctx context.Context, dataType DataType, timestamp time.Time) error
    
    // Transactions
    BeginTx(ctx context.Context) (*sql.Tx, error)
}
```
```

### 3. Sync Service

**Responsibilities:**
- Orchestrate incremental data synchronization
- Ensure correct order of operations (subjects → assignments → reviews)
- Handle sync failures and rollback
- Log sync results

**Interface:**

```go
type SyncService interface {
    // Perform full sync of all data types
    SyncAll(ctx context.Context) ([]SyncResult, error)
    
    // Sync specific data types
    SyncSubjects(ctx context.Context) (SyncResult, error)
    SyncAssignments(ctx context.Context) (SyncResult, error)
    SyncReviews(ctx context.Context) (SyncResult, error)
    SyncStatistics(ctx context.Context) (SyncResult, error)
    
    // Create assignment snapshot after successful sync
    CreateAssignmentSnapshot(ctx context.Context) error
    
    // Check if sync is in progress
    IsSyncing() bool
}

type SyncResult struct {
    DataType       DataType
    RecordsUpdated int
    Success        bool
    Error          string
    Timestamp      time.Time
}
```
```

### 4. API Server

**Responsibilities:**
- Authenticate incoming requests using API token
- Expose REST endpoints for querying data
- Validate request parameters
- Format responses as JSON
- Handle errors with appropriate status codes

**Authentication:**
- All API endpoints require authentication via Bearer token in Authorization header
- Token is configured via environment variable `LOCAL_API_TOKEN`
- If no token is configured, server operates without authentication (with warning)
- Authentication failures return 401 Unauthorized

**Endpoints:**

```
GET  /api/subjects?type=kanji&level=5
GET  /api/assignments?srs_stage=apprentice
GET  /api/assignments/snapshots?from=2024-01-01&to=2024-01-31
GET  /api/reviews?from=2024-01-01&to=2024-01-31
GET  /api/statistics/latest
GET  /api/statistics?from=2024-01-01&to=2024-01-31
POST /api/sync
GET  /api/sync/status
GET  /api/health (no authentication required)
```

### 5. Scheduler

**Responsibilities:**
- Execute daily sync at configured time
- Prevent concurrent syncs
- Log sync execution

**Interface:**

```go
type Scheduler interface {
    Start(cronExpression string) error
    Stop()
    IsRunning() bool
}
```

## Data Models

### Subject

```go
type Subject struct {
    ID            int       `json:"id"`
    Object        string    `json:"object"` // "radical", "kanji", or "vocabulary"
    URL           string    `json:"url"`
    DataUpdatedAt time.Time `json:"data_updated_at"`
    Data          SubjectData `json:"data"`
}

type SubjectData struct {
    Level      int       `json:"level"`
    Characters string    `json:"characters"`
    Meanings   []Meaning `json:"meanings"`
    Readings   []Reading `json:"readings,omitempty"`
    // ... additional fields
}
```

### Assignment

```go
type Assignment struct {
    ID            int       `json:"id"`
    Object        string    `json:"object"` // "assignment"
    URL           string    `json:"url"`
    DataUpdatedAt time.Time `json:"data_updated_at"`
    Data          AssignmentData `json:"data"`
}

type AssignmentData struct {
    SubjectID   int        `json:"subject_id"`
    SubjectType string     `json:"subject_type"` // "radical", "kanji", or "vocabulary"
    SRSStage    int        `json:"srs_stage"`
    UnlockedAt  *time.Time `json:"unlocked_at"`
    StartedAt   *time.Time `json:"started_at"`
    PassedAt    *time.Time `json:"passed_at"`
    // ... additional fields
}
```

### Review

```go
type Review struct {
    ID            int       `json:"id"`
    Object        string    `json:"object"` // "review"
    URL           string    `json:"url"`
    DataUpdatedAt time.Time `json:"data_updated_at"`
    Data          ReviewData `json:"data"`
}

type ReviewData struct {
    AssignmentID             int       `json:"assignment_id"`
    SubjectID                int       `json:"subject_id"`
    CreatedAt                time.Time `json:"created_at"`
    IncorrectMeaningAnswers  int       `json:"incorrect_meaning_answers"`
    IncorrectReadingAnswers  int       `json:"incorrect_reading_answers"`
}
```

### Statistics

```go
type Statistics struct {
    Object        string    `json:"object"` // "report"
    URL           string    `json:"url"`
    DataUpdatedAt time.Time `json:"data_updated_at"`
    Data          StatisticsData `json:"data"`
}

type StatisticsData struct {
    Lessons []LessonStatistics `json:"lessons"`
    Reviews []ReviewStatistics `json:"reviews"`
}

type StatisticsSnapshot struct {
    ID         int        `json:"id"`
    Timestamp  time.Time  `json:"timestamp"`
    Statistics Statistics `json:"statistics"`
}
```

### Assignment Snapshot

```go
type AssignmentSnapshot struct {
    Date        time.Time `json:"date"`
    SRSStage    int       `json:"srs_stage"`
    SubjectType string    `json:"subject_type"`
    Count       int       `json:"count"`
}

type AssignmentSnapshotSummary struct {
    Date string                           `json:"date"`
    Data map[string]map[string]int        `json:"data"` // SRS stage name -> subject type -> count
}

// SRS Stage mapping
const (
    SRSStageInitiate    = 0
    SRSStageApprentice1 = 1
    SRSStageApprentice2 = 2
    SRSStageApprentice3 = 3
    SRSStageApprentice4 = 4
    SRSStageGuru1       = 5
    SRSStageGuru2       = 6
    SRSStageMaster      = 7
    SRSStageEnlightened = 8
    SRSStageBurned      = 9
)

// GetSRSStageName returns the human-readable name for an SRS stage
func GetSRSStageName(stage int) string {
    switch {
    case stage >= 1 && stage <= 4:
        return "apprentice"
    case stage >= 5 && stage <= 6:
        return "guru"
    case stage == 7:
        return "master"
    case stage == 8:
        return "enlightened"
    case stage == 9:
        return "burned"
    default:
        return "unknown"
    }
}
```


## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Token persistence and usage
*For any* valid API token, after setting the token in the API Client, all subsequent requests should include that token in the authorization header.
**Validates: Requirements 1.1, 1.2**

### Property 2: Token update replaces previous token
*For any* two different API tokens, setting the first token then setting the second token should result in all subsequent requests using only the second token.
**Validates: Requirements 1.4**

### Property 3: Pagination completeness
*For any* paginated API response, the API Client should fetch all pages and return the complete dataset, not just the first page.
**Validates: Requirements 2.2**

### Property 4: Data persistence round trip
*For any* WaniKani data object (subject, assignment, review, or statistics), after persisting it to the Data Store and then retrieving it, the retrieved data should match the original data.
**Validates: Requirements 2.3, 3.2, 4.2, 8.2**

### Property 5: Upsert idempotence
*For any* data record with a unique ID, inserting it twice should result in exactly one record in the Data Store with the most recent data.
**Validates: Requirements 2.4, 3.3**

### Property 6: Referential integrity preservation
*For any* assignment or review, the referenced subject_id or assignment_id must exist in the Data Store, and deleting a referenced entity should fail or cascade appropriately.
**Validates: Requirements 3.5, 4.4**

### Property 7: Query filter correctness
*For any* filter criteria applied to subjects, assignments, or reviews, all returned results should satisfy the filter conditions, and no matching records should be excluded.
**Validates: Requirements 5.1, 4.3, 8.4**

### Property 8: Join completeness
*For any* assignment returned by the API Server, the response should include the complete associated subject data; for any review, the response should include both assignment and subject data.
**Validates: Requirements 5.2, 5.3**

### Property 9: Incremental sync uses timestamps
*For any* sync operation after the initial sync, the API Client should use the last successful sync timestamp as the updated_after parameter to fetch only modified data.
**Validates: Requirements 6.1, 3.4**

### Property 10: Successful sync updates timestamp
*For any* data type, after a successful sync operation, the stored last sync timestamp should be updated to the current time.
**Validates: Requirements 6.2**

### Property 11: Failed sync preserves timestamp
*For any* data type, if a sync operation fails, the last sync timestamp should remain unchanged from before the sync attempt.
**Validates: Requirements 6.3**

### Property 12: Sync ordering maintains referential integrity
*For any* full sync operation, subjects must be synced before assignments, and assignments must be synced before reviews, to ensure foreign key relationships are valid.
**Validates: Requirements 6.4, 9.2**

### Property 13: Rate limit compliance
*For any* sequence of API requests, the request rate should not exceed the rate limit specified by the WaniKani API headers, and requests should be throttled when approaching the limit.
**Validates: Requirements 7.2, 7.3**

### Property 14: Rate limit status accuracy
*For any* point in time, the exposed rate limit status should accurately reflect the remaining quota and reset time based on the most recent API response headers.
**Validates: Requirements 7.4**

### Property 15: Statistics historical preservation
*For any* sequence of statistics snapshots, all snapshots should be preserved in the Data Store with their timestamps, enabling retrieval of any historical snapshot.
**Validates: Requirements 8.3**

### Property 16: Latest statistics retrieval
*For any* set of statistics snapshots in the Data Store, requesting the latest statistics should return the snapshot with the most recent timestamp.
**Validates: Requirements 8.5**

### Property 17: Sync result logging
*For any* sync operation, the system should log a sync result containing the data type, number of records updated, success status, and timestamp.
**Validates: Requirements 9.3**

### Property 18: Concurrent sync prevention
*For any* sync operation in progress, attempting to start another sync operation should be prevented until the first sync completes.
**Validates: Requirements 9.5**

### Property 19: Transaction atomicity
*For any* Data Store write operation that fails, no partial changes should be persisted, and the Data Store should remain in its pre-operation state.
**Validates: Requirements 10.2**

### Property 20: Validation error specificity
*For any* invalid input data, the validation error message should identify which specific fields are invalid and why.
**Validates: Requirements 10.4**

### Property 21: API authentication enforcement
*For any* API request without a valid authorization token (when LOCAL_API_TOKEN is configured), the API Server should return a 401 Unauthorized response and reject the request.
**Validates: Requirements 11.1, 11.2, 11.3**

### Property 22: Snapshot creation after successful sync
*For any* successful sync operation, a daily snapshot should be created with assignment counts grouped by SRS stage and subject type for the current date.
**Validates: Requirements 12.1**

### Property 23: Snapshot excludes unstarted assignments
*For any* set of assignments including those with SRS stage 0, the calculated snapshot should exclude all assignments with SRS stage 0 from the counts.
**Validates: Requirements 12.2**

### Property 24: Snapshot data persistence round trip
*For any* assignment snapshot, after persisting it to the Data Store and then retrieving it, the retrieved snapshot should contain the same date, SRS stage, subject type, and count.
**Validates: Requirements 12.3**

### Property 25: Snapshot upsert idempotence
*For any* date, creating multiple snapshots on the same date should result in exactly one set of snapshot records for that date with the most recent counts.
**Validates: Requirements 12.4**

### Property 26: Snapshot API response format
*For any* set of snapshots retrieved from the API, the response should group data by date with SRS stage names (not numbers) and include all three subject types.
**Validates: Requirements 12.5**

### Property 27: Snapshot totals accuracy
*For any* snapshot date and SRS stage, the total count should equal the sum of counts across all subject types (radical, kanji, vocabulary) for that stage.
**Validates: Requirements 12.6**

### Property 28: Snapshot date range filtering
*For any* date range filter, all returned snapshots should have dates within the specified range, and no snapshots outside the range should be included.
**Validates: Requirements 12.7**

### Property 29: Snapshot date ordering
*For any* set of snapshots, the API response should order snapshots by date in ascending chronological order.
**Validates: Requirements 12.8**

## Error Handling

### API Client Error Handling

1. **Network Errors**: Wrap all HTTP requests in try-catch blocks. Return descriptive errors for network failures.
2. **Authentication Errors**: Detect 401 responses and return clear "Invalid API token" messages.
3. **Rate Limit Errors**: Detect 429 responses, extract retry-after header, and wait before retrying.
4. **Retry Logic**: Implement exponential backoff for transient errors (network issues, 5xx responses). Maximum 3 retries.
5. **Validation Errors**: Validate API responses against expected schemas. Return errors for unexpected formats.

### Data Store Error Handling

1. **Transaction Management**: Wrap multi-record operations in transactions. Rollback on any error.
2. **Constraint Violations**: Catch foreign key and unique constraint violations. Return specific error messages.
3. **Connection Errors**: Implement connection pooling with retry logic for database connection failures.

### API Server Error Handling

1. **Authentication Errors**: Validate Authorization header. Return 401 for missing or invalid tokens.
2. **Input Validation**: Validate all query parameters. Return 400 with specific field errors.
3. **Not Found**: Return 404 when requested resources don't exist.
4. **Internal Errors**: Catch all unhandled exceptions. Return 500 with generic message. Log full error details.
5. **Error Response Format**: Standardize error responses:
```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Invalid query parameters",
    "details": {
      "level": "Must be a number between 1 and 60"
    }
  }
}
```

**Authentication Error Example:**
```json
{
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Authentication required",
    "details": {
      "header": "Authorization header with Bearer token is required"
    }
  }
}
```

## Testing Strategy

### Unit Testing

The system will use Go's standard testing package for unit testing. Unit tests will cover:

1. **Specific Examples**: Test concrete scenarios like fetching a known subject, storing a specific assignment
2. **Edge Cases**: Test boundary conditions like empty responses, first sync with no timestamp, maximum pagination
3. **Error Conditions**: Test specific error scenarios like authentication failures, network timeouts, invalid input
4. **Integration Points**: Test interactions between components like API Client → Data Store

Example unit tests:
- Test that fetching subjects with no results returns an empty array
- Test that setting an invalid API token returns an authentication error
- Test that querying with an invalid date format returns a 400 error

### Property-Based Testing

The system will use gopter for property-based testing. Property-based tests verify universal properties across many randomly generated inputs.

**Configuration**:
- Each property test will run a minimum of 100 iterations using gopter's parameters
- Tests will use custom generators for WaniKani data types
- Each test will be tagged with a comment referencing the design document property

**Property Test Requirements**:
- Each correctness property listed above must be implemented by a single property-based test
- Tests must be tagged with: `// Feature: wanikani-api, Property N: [property text]`
- Tests should use realistic data generators that respect WaniKani's data constraints

Example property tests:
- Generate random subjects and verify persistence round trip (Property 4)
- Generate random filter criteria and verify all results match filters (Property 7)
- Generate random sync sequences and verify timestamp updates (Property 10)

**Complementary Approach**:
Unit tests and property tests work together:
- Unit tests catch specific bugs in concrete scenarios
- Property tests verify general correctness across all inputs
- Together they provide comprehensive coverage

### Test Data Generators

Custom generators will be created for:
- Subjects (radicals, kanji, vocabulary with valid levels and characters)
- Assignments (with valid SRS stages and timestamps)
- Reviews (with valid correctness counts)
- Statistics (with valid lesson and review data)
- API responses (with pagination, rate limit headers)

## Implementation Notes

### Database Schema

**subjects table**:
- id (PRIMARY KEY)
- object (TEXT)
- data_updated_at (TEXT)
- data (JSON)

**assignments table**:
- id (PRIMARY KEY)
- subject_id (FOREIGN KEY → subjects.id)
- data_updated_at (TEXT)
- data (JSON)

**reviews table**:
- id (PRIMARY KEY)
- assignment_id (FOREIGN KEY → assignments.id)
- subject_id (FOREIGN KEY → subjects.id)
- data_updated_at (TEXT)
- data (JSON)

**statistics_snapshots table**:
- id (PRIMARY KEY, AUTOINCREMENT)
- timestamp (TEXT)
- data (JSON)

**sync_metadata table**:
- data_type (PRIMARY KEY)
- last_sync_time (TEXT)

**assignment_snapshots table**:
- date (TEXT, part of composite PRIMARY KEY)
- srs_stage (INTEGER, part of composite PRIMARY KEY)
- subject_type (TEXT, part of composite PRIMARY KEY)
- count (INTEGER)
- PRIMARY KEY (date, srs_stage, subject_type)

### Rate Limiting Strategy

Implement a token bucket algorithm:
1. Track remaining requests from API response headers
2. Track reset time from API response headers
3. Before each request, check if quota is available
4. If quota exhausted, wait until reset time
5. Implement request queue to serialize requests

### Pagination Strategy

WaniKani API uses cursor-based pagination:
1. Make initial request
2. Check for `pages.next_url` in response
3. If present, fetch next page
4. Repeat until `next_url` is null
5. Aggregate all results before returning

### Incremental Sync Strategy

1. Query sync_metadata for last sync time of data type
2. If no timestamp exists, perform full sync (no updated_after parameter)
3. If timestamp exists, pass it as updated_after parameter
4. WaniKani returns only records modified since that time
5. Upsert returned records
6. Update sync_metadata with current timestamp

### Assignment Snapshot Strategy

Assignment snapshots provide a daily view of the distribution of assignments across SRS stages and subject types. This enables tracking progress over time.

**Snapshot Creation Process:**
1. After a successful sync operation, calculate the current state of all assignments
2. Group assignments by SRS stage (1-9, excluding 0) and subject type (radical, kanji, vocabulary)
3. Count assignments in each group
4. Store counts in assignment_snapshots table with today's date
5. Use UPSERT logic so multiple syncs on the same day update the existing snapshot

**Snapshot Data Structure:**
Each snapshot record contains:
- `date`: The date of the snapshot (YYYY-MM-DD format)
- `srs_stage`: Numeric SRS stage (1-9)
- `subject_type`: Type of subject (radical, kanji, vocabulary)
- `count`: Number of assignments in this category

**API Response Format:**
The API transforms the flat snapshot records into a nested structure:
```json
{
  "2024-01-15": {
    "apprentice": {
      "radical": 6,
      "kanji": 15,
      "vocabulary": 20,
      "total": 41
    },
    "guru": {
      "radical": 10,
      "kanji": 25,
      "vocabulary": 30,
      "total": 65
    }
  }
}
```

**SRS Stage Mapping:**
- Stages 1-4 → "apprentice"
- Stages 5-6 → "guru"
- Stage 7 → "master"
- Stage 8 → "enlightened"
- Stage 9 → "burned"

**Date Range Filtering:**
The API supports filtering snapshots by date range using `from` and `to` query parameters in YYYY-MM-DD format.

### Configuration

Configuration will be stored in environment variables or a config file:
- `WANIKANI_API_TOKEN`: API token for authentication with external WaniKani API (required)
- `LOCAL_API_TOKEN`: API token for authentication with local API server (optional, recommended)
- `DATABASE_PATH`: Path to SQLite database file (default: "./wanikani.db")
- `SYNC_SCHEDULE`: Cron expression for daily sync (default: "0 2 * * *" for 2 AM daily)
- `API_PORT`: Port for API server (default: 8080)
- `LOG_LEVEL`: Logging verbosity (default: "info")

Configuration can be loaded using a simple struct:

```go
type Config struct {
    WaniKaniAPIToken string
    LocalAPIToken    string
    DatabasePath     string
    SyncSchedule     string
    APIPort          int
    LogLevel         string
}
```
