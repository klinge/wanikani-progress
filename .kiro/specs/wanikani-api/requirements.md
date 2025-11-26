# Requirements Document

## Introduction

This document specifies the requirements for a WaniKani Data API system. The system SHALL fetch data from the WaniKani language learning platform and store it locally, providing an interface for accessing and managing this data. WaniKani is a spaced repetition system for learning Japanese kanji and vocabulary.

## Glossary

- **WaniKani API**: The external REST API provided by WaniKani for accessing user data and learning resources
- **API Client**: The component that communicates with the WaniKani API
- **Data Store**: The local storage system for persisting fetched WaniKani data
- **API Server**: The local REST API that provides access to stored WaniKani data
- **Subject**: A learning item in WaniKani (radical, kanji, or vocabulary)
- **Review**: A user's answer to a quiz question for a specific subject
- **Assignment**: The relationship between a user and a subject, tracking progress
- **API Token**: The authentication credential for accessing the WaniKani API
- **Summary Statistics**: Aggregate data about user progress including lesson counts, review counts, and SRS stage distributions
- **Sync Operation**: The process of fetching updated data from WaniKani and storing it locally

## Requirements

### Requirement 1

**User Story:** As a developer, I want to authenticate with the WaniKani API, so that I can access user-specific data securely.

#### Acceptance Criteria

1. WHEN the API Client receives a valid API Token, THE API Client SHALL store the token securely for subsequent requests
2. WHEN the API Client makes a request to WaniKani, THE API Client SHALL include the API Token in the authorization header
3. IF the WaniKani API returns an authentication error, THEN THE API Client SHALL return a clear error message indicating invalid credentials
4. WHEN the API Token is updated, THE API Client SHALL use the new token for all subsequent requests

### Requirement 2

**User Story:** As a developer, I want to fetch subjects from WaniKani, so that I can access information about radicals, kanji, and vocabulary.

#### Acceptance Criteria

1. WHEN the API Client requests subjects, THE API Client SHALL retrieve all subject types (radicals, kanji, vocabulary) from the WaniKani API
2. WHEN the WaniKani API returns paginated results, THE API Client SHALL fetch all pages until the complete dataset is retrieved
3. WHEN subjects are fetched, THE Data Store SHALL persist each subject with its complete metadata
4. WHEN a subject already exists in the Data Store, THE Data Store SHALL update the existing record with new data
5. IF the WaniKani API returns an error during fetch, THEN THE API Client SHALL retry the request up to three times with exponential backoff

### Requirement 3

**User Story:** As a developer, I want to fetch user assignments from WaniKani, so that I can track learning progress for each subject.

#### Acceptance Criteria

1. WHEN the API Client requests assignments, THE API Client SHALL retrieve all assignment records from the WaniKani API
2. WHEN assignments are fetched, THE Data Store SHALL persist each assignment with its progress data
3. WHEN an assignment already exists in the Data Store, THE Data Store SHALL update the existing record preserving the relationship to its subject
4. WHEN the API Client fetches assignments, THE API Client SHALL include timestamps to retrieve only updated assignments since the last fetch
5. WHEN assignments reference subjects, THE Data Store SHALL maintain referential integrity between assignments and subjects

### Requirement 4

**User Story:** As a developer, I want to fetch user reviews from WaniKani, so that I can analyze learning history and performance.

#### Acceptance Criteria

1. WHEN the API Client requests reviews, THE API Client SHALL retrieve all review records from the WaniKani API
2. WHEN reviews are fetched, THE Data Store SHALL persist each review with its correctness data and timestamps
3. WHEN the API Client fetches reviews, THE API Client SHALL support filtering by date range to retrieve specific time periods
4. WHEN reviews reference assignments, THE Data Store SHALL maintain referential integrity between reviews and assignments

### Requirement 5

**User Story:** As a developer, I want to query stored WaniKani data through a local API, so that I can build applications without repeatedly calling the external WaniKani API.

#### Acceptance Criteria

1. WHEN the API Server receives a request for subjects, THE API Server SHALL return subjects filtered by the requested criteria
2. WHEN the API Server receives a request for assignments, THE API Server SHALL return assignments with their associated subject data
3. WHEN the API Server receives a request for reviews, THE API Server SHALL return reviews with their associated assignment and subject data
4. WHEN the API Server receives invalid query parameters, THE API Server SHALL return a clear error message with status code 400
5. WHEN the API Server processes a request, THE API Server SHALL return results in JSON format

### Requirement 6

**User Story:** As a developer, I want to sync data incrementally from WaniKani, so that I can keep my local data up-to-date efficiently without fetching unchanged data.

#### Acceptance Criteria

1. WHEN the API Client performs a sync operation, THE API Client SHALL use the updated_after parameter to fetch only data modified since the last successful sync
2. WHEN a sync operation completes successfully, THE Data Store SHALL record the sync timestamp for each data type (subjects, assignments, reviews, statistics)
3. WHEN a sync operation fails, THE Data Store SHALL preserve the previous sync timestamp to allow retry from the correct point
4. WHEN the API Client syncs data, THE API Client SHALL process subjects before assignments and assignments before reviews to maintain referential integrity
5. WHEN no previous sync timestamp exists, THE API Client SHALL fetch all historical data for the initial sync

### Requirement 7

**User Story:** As a developer, I want the system to handle rate limits from the WaniKani API, so that my application remains compliant with API usage policies.

#### Acceptance Criteria

1. WHEN the WaniKani API returns a rate limit error, THE API Client SHALL wait for the specified duration before retrying
2. WHEN the API Client makes requests, THE API Client SHALL respect the rate limit headers provided by the WaniKani API
3. WHEN multiple requests are queued, THE API Client SHALL throttle requests to stay within rate limits
4. WHEN rate limit information is available, THE API Client SHALL expose the remaining quota to calling code

### Requirement 8

**User Story:** As a language learner, I want to track my WaniKani statistics over time, so that I can monitor my progress and identify trends.

#### Acceptance Criteria

1. WHEN the API Client fetches summary statistics, THE API Client SHALL retrieve the current snapshot from the WaniKani API
2. WHEN statistics are fetched, THE Data Store SHALL persist the statistics with a timestamp indicating when the snapshot was taken
3. WHEN statistics are stored, THE Data Store SHALL preserve all historical snapshots to enable time-series analysis
4. WHEN the API Server receives a request for statistics, THE API Server SHALL return statistics filtered by date range
5. WHEN the API Server receives a request for the latest statistics, THE API Server SHALL return the most recent snapshot

### Requirement 9

**User Story:** As a developer, I want to schedule automatic daily syncs, so that my local data stays current without manual intervention.

#### Acceptance Criteria

1. WHEN a scheduled sync is configured, THE system SHALL execute a sync operation at the specified daily time
2. WHEN a scheduled sync runs, THE system SHALL fetch updated subjects, assignments, reviews, and statistics in sequence
3. WHEN a scheduled sync completes, THE system SHALL log the sync result including the number of records updated
4. IF a scheduled sync fails, THEN THE system SHALL log the error and retry at the next scheduled time
5. WHEN a scheduled sync is running, THE system SHALL prevent concurrent sync operations from starting

### Requirement 10

**User Story:** As a developer, I want to handle errors gracefully, so that the system remains stable when issues occur.

#### Acceptance Criteria

1. WHEN the WaniKani API is unreachable, THE API Client SHALL return a clear error indicating network connectivity issues
2. WHEN the Data Store encounters a write error, THE Data Store SHALL rollback partial changes to maintain data consistency
3. WHEN the API Server encounters an internal error, THE API Server SHALL return status code 500 with a generic error message without exposing internal details
4. WHEN validation fails on input data, THE system SHALL return specific error messages indicating which fields are invalid
