# WaniKani Data API

A local API service that fetches, stores, and serves data from the [WaniKani](https://www.wanikani.com) language learning platform. This system provides efficient incremental synchronization and a REST API for querying your WaniKani learning data locally.

## Features

- 🔄 **Incremental Sync**: Efficiently fetch only updated data using timestamps
- 💾 **Local Storage**: SQLite database for fast local queries
- 🔐 **Dual Authentication**: Secure access to both WaniKani API and local API
- 📊 **Historical Tracking**: Preserve statistics snapshots over time
- ⚡ **Rate Limiting**: Automatic compliance with WaniKani API rate limits
- 🔁 **Retry Logic**: Exponential backoff for transient errors
- 📅 **Scheduled Syncs**: Optional daily synchronization (via cron)
- 🧪 **Property-Based Testing**: Comprehensive test coverage with formal correctness properties

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [API Endpoints](#api-endpoints)
- [Authentication](#authentication)
- [Building](#building)
- [Testing](#testing)
- [Architecture](#architecture)
- [Contributing](#contributing)

## Quick Start

```bash
# 1. Clone the repository
git clone <repository-url>
cd wanikani-api

# 2. Set up configuration
cp .env.example .env
# Edit .env and add your WaniKani API token

# 3. Build and run
make build
./bin/wanikani-api

# Or run directly without building
go run ./cmd/wanikani-api
```

The API server will start on `http://localhost:8080` (or the port specified in your configuration).

## Installation

### Prerequisites

- Go 1.21 or higher
- SQLite3 (usually pre-installed on most systems)
- A WaniKani account with an API token

### Getting Your WaniKani API Token

1. Log in to [WaniKani](https://www.wanikani.com)
2. Go to [Settings → Personal Access Tokens](https://www.wanikani.com/settings/personal_access_tokens)
3. Generate a new token with appropriate permissions
4. Copy the token for use in configuration

### Install Dependencies

```bash
# Download Go module dependencies
go mod download

# Or use Make
make install
```

## Configuration

### Environment Variables

The application is configured using environment variables. You can set these in a `.env` file or as system environment variables.

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `WANIKANI_API_TOKEN` | **Yes** | - | Your WaniKani API token for accessing the external API |
| `LOCAL_API_TOKEN` | No | - | Token for authenticating requests to your local API (recommended) |
| `DATABASE_PATH` | No | `./data/wanikani.db` | Path to the SQLite database file |
| `SYNC_SCHEDULE` | No | `0 2 * * *` | Cron expression for scheduled syncs (default: 2 AM daily) |
| `API_PORT` | No | `8080` | Port for the API server to listen on |
| `LOG_LEVEL` | No | `info` | Logging verbosity: `debug`, `info`, `warn`, `error` |

### Configuration Setup

#### Option 1: Using .env file (Recommended)

```bash
# Copy the example configuration
cp .env.example .env

# Edit .env with your values
nano .env  # or use your preferred editor
```

Example `.env` file:
```bash
WANIKANI_API_TOKEN=your_wanikani_api_token_here
LOCAL_API_TOKEN=your_secure_local_token_here
DATABASE_PATH=./data/wanikani.db
SYNC_SCHEDULE=0 2 * * *
API_PORT=8080
LOG_LEVEL=info
```

#### Option 2: System Environment Variables

```bash
export WANIKANI_API_TOKEN=your_token_here
export LOCAL_API_TOKEN=your_local_token_here
export API_PORT=8080
```

**Note**: The `.env` file is in `.gitignore` to prevent accidentally committing secrets.

## Usage

### Starting the Server

```bash
# Using the compiled binary
./bin/wanikani-api

# Or run directly with Go
go run ./cmd/wanikani-api

# Or use Make
make run
```

The server will:
1. Load configuration from environment variables
2. Initialize the SQLite database
3. Start the API server
4. Begin listening for requests

### Initial Sync

On first run, trigger a full sync to fetch all your WaniKani data:

```bash
curl -X POST http://localhost:8080/api/sync \
  -H "Authorization: Bearer your_local_token_here"
```

This will fetch:
- All subjects (radicals, kanji, vocabulary)
- All assignments (your progress on each subject)
- All reviews (your quiz history)
- Current statistics snapshot

### Scheduled Syncs

For automatic daily syncs, you can:

**Option 1: Use system cron**
```bash
# Add to your crontab (crontab -e)
0 2 * * * curl -X POST http://localhost:8080/api/sync -H "Authorization: Bearer your_token" >> /var/log/wanikani-sync.log 2>&1
```

**Option 2: Keep the application running**
The application includes built-in scheduling support (currently optional in implementation).

### Querying Data

Once synced, query your data through the local API:

```bash
# Get all kanji from level 5
curl http://localhost:8080/api/subjects?type=kanji&level=5 \
  -H "Authorization: Bearer your_local_token_here"

# Get assignments in apprentice stage
curl http://localhost:8080/api/assignments?srs_stage=apprentice \
  -H "Authorization: Bearer your_local_token_here"

# Get reviews from January 2024
curl "http://localhost:8080/api/reviews?from=2024-01-01&to=2024-01-31" \
  -H "Authorization: Bearer your_local_token_here"

# Get latest statistics
curl http://localhost:8080/api/statistics/latest \
  -H "Authorization: Bearer your_local_token_here"
```

## API Endpoints

All endpoints except `/health` require authentication when `LOCAL_API_TOKEN` is configured.

### Health Check

```
GET /health
```

Returns server health status. No authentication required.

**Response:**
```json
{
  "status": "ok"
}
```

### Subjects

```
GET /api/subjects
```

Retrieve subjects (radicals, kanji, vocabulary) with optional filtering.

**Query Parameters:**
- `type` - Filter by subject type: `radical`, `kanji`, or `vocabulary`
- `level` - Filter by WaniKani level (1-60)

**Example:**
```bash
curl http://localhost:8080/api/subjects?type=kanji&level=5 \
  -H "Authorization: Bearer your_token"
```

### Assignments

```
GET /api/assignments
```

Retrieve user assignments with progress data. Includes associated subject information.

**Query Parameters:**
- `srs_stage` - Filter by SRS stage (0-9)
- `level` - Filter by subject level (1-60)

**Example:**
```bash
curl http://localhost:8080/api/assignments?srs_stage=4 \
  -H "Authorization: Bearer your_token"
```

### Reviews

```
GET /api/reviews
```

Retrieve review history with associated assignment and subject data.

**Query Parameters:**
- `from` - Start date (ISO 8601 format: `YYYY-MM-DD`)
- `to` - End date (ISO 8601 format: `YYYY-MM-DD`)

**Example:**
```bash
curl "http://localhost:8080/api/reviews?from=2024-01-01&to=2024-01-31" \
  -H "Authorization: Bearer your_token"
```

### Statistics (Latest)

```
GET /api/statistics/latest
```

Retrieve the most recent statistics snapshot.

**Example:**
```bash
curl http://localhost:8080/api/statistics/latest \
  -H "Authorization: Bearer your_token"
```

### Statistics (Historical)

```
GET /api/statistics
```

Retrieve statistics snapshots within a date range.

**Query Parameters:**
- `from` - Start date (ISO 8601 format: `YYYY-MM-DD`)
- `to` - End date (ISO 8601 format: `YYYY-MM-DD`)

**Example:**
```bash
curl "http://localhost:8080/api/statistics?from=2024-01-01&to=2024-01-31" \
  -H "Authorization: Bearer your_token"
```

### Trigger Sync

```
POST /api/sync
```

Manually trigger a data synchronization with WaniKani.

**Example:**
```bash
curl -X POST http://localhost:8080/api/sync \
  -H "Authorization: Bearer your_token"
```

**Response:**
```json
{
  "message": "Sync completed successfully",
  "results": [
    {
      "data_type": "subjects",
      "records_updated": 42,
      "success": true,
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ]
}
```

### Sync Status

```
GET /api/sync/status
```

Check if a sync operation is currently in progress.

**Example:**
```bash
curl http://localhost:8080/api/sync/status \
  -H "Authorization: Bearer your_token"
```

**Response:**
```json
{
  "syncing": false,
  "last_sync": {
    "subjects": "2024-01-15T10:30:00Z",
    "assignments": "2024-01-15T10:30:15Z",
    "reviews": "2024-01-15T10:30:30Z",
    "statistics": "2024-01-15T10:30:45Z"
  }
}
```

## Authentication

### Local API Authentication

When `LOCAL_API_TOKEN` is configured, all API endpoints (except `/health`) require authentication using a Bearer token.

**Request Header:**
```
Authorization: Bearer your_local_token_here
```

**Example:**
```bash
curl http://localhost:8080/api/subjects \
  -H "Authorization: Bearer your_local_token_here"
```

**Error Response (401 Unauthorized):**
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

### Security Recommendations

1. **Always set LOCAL_API_TOKEN** in production environments
2. **Use strong, random tokens** (e.g., generated with `openssl rand -hex 32`)
3. **Never commit tokens** to version control (`.env` is in `.gitignore`)
4. **Use HTTPS** if exposing the API over a network
5. **Restrict network access** using firewall rules if needed

## Building

### Using Make (Recommended)

```bash
# Build the binary
make build

# Run tests
make test

# Run tests with coverage
make test-coverage

# Build and run
make run

# Clean build artifacts
make clean

# Format code
make fmt

# See all available commands
make help
```

### Using Go Directly

```bash
# Build to bin/ directory
go build -o bin/wanikani-api ./cmd/wanikani-api

# Run without building
go run ./cmd/wanikani-api

# Run tests
go test ./...

# Run only fast tests (skip property tests)
go test -short ./...
```

### Build Output

All compiled binaries are placed in the `bin/` directory, which is ignored by git.

### Cross-Platform Builds

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/wanikani-api-linux-amd64 ./cmd/wanikani-api

# macOS
GOOS=darwin GOARCH=amd64 go build -o bin/wanikani-api-darwin-amd64 ./cmd/wanikani-api

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/wanikani-api-windows-amd64.exe ./cmd/wanikani-api
```

## Testing

The project includes both unit tests and property-based tests for comprehensive coverage.

### Running Tests

```bash
# Run all tests (including property tests)
go test ./...

# Run only fast unit tests (skip property tests)
go test -short ./...

# Run tests for a specific package
go test ./internal/api/

# Run with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Test Types

**Unit Tests:**
- Fast, focused tests for specific functionality
- Test concrete scenarios and edge cases
- Complete in seconds

**Property-Based Tests:**
- Verify correctness properties across many random inputs
- Named with `TestProperty_*` prefix
- Run 100+ iterations per property
- May take 30-90 seconds per test file
- Can be skipped with `-short` flag

### CI/CD Recommendations

- Run fast tests (`-short`) on every commit
- Run full test suite on pull requests
- Consider running property tests nightly for comprehensive coverage

For more details, see [README_TESTING.md](README_TESTING.md).

## Architecture

### High-Level Overview

```
┌─────────────────┐
│   API Server    │ ← REST API for querying local data
└────────┬────────┘
         │
┌────────▼────────┐
│   Data Store    │ ← SQLite database
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

### Components

1. **API Client Layer**: Handles communication with WaniKani API
   - Authentication and token management
   - Automatic pagination handling
   - Rate limiting and retry logic
   - Error handling

2. **Data Storage Layer**: Local SQLite database
   - Subjects, assignments, reviews, statistics
   - Incremental sync tracking
   - Referential integrity
   - Transaction support

3. **API Server Layer**: REST API for local queries
   - Authentication middleware
   - Query filtering and validation
   - JSON response formatting
   - Error handling

4. **Sync Service**: Orchestrates data synchronization
   - Incremental updates using timestamps
   - Correct ordering (subjects → assignments → reviews)
   - Sync locking to prevent concurrent operations
   - Result logging

### Database Schema

- `subjects` - Learning items (radicals, kanji, vocabulary)
- `assignments` - User progress on subjects
- `reviews` - Quiz history
- `statistics_snapshots` - Historical statistics with timestamps
- `sync_metadata` - Last sync timestamps for incremental updates

For more details, see [docs/ARCHITECTURE.md](docs/ARCHITECTURE.md).

## Project Structure

```
wanikani-api/
├── cmd/
│   └── wanikani-api/      # Application entry point
├── internal/
│   ├── api/               # API server and handlers
│   ├── config/            # Configuration management
│   ├── domain/            # Domain types and interfaces
│   ├── store/             # Data storage implementations
│   │   └── sqlite/        # SQLite implementation
│   ├── sync/              # Sync service
│   ├── utils/             # Utilities (logging, etc.)
│   └── wanikani/          # WaniKani API client
├── data/                  # Database files (gitignored)
├── docs/                  # Additional documentation
├── scripts/               # Build and utility scripts
├── .env.example           # Example configuration
├── Makefile               # Build automation
└── README.md              # This file
```

## Troubleshooting

### Common Issues

**"Failed to load configuration"**
- Ensure `.env` file exists or environment variables are set
- Check that `WANIKANI_API_TOKEN` is set

**"Authentication required" (401)**
- Include `Authorization: Bearer <token>` header in requests
- Verify `LOCAL_API_TOKEN` matches between server and client

**"Invalid API token" from WaniKani**
- Verify your WaniKani API token is correct
- Check token hasn't been revoked at WaniKani settings

**Database locked errors**
- Ensure only one instance of the application is running
- Check file permissions on the database file

**Sync fails with rate limit errors**
- The client automatically handles rate limits
- If persistent, check WaniKani API status

### Logging

Increase log verbosity for debugging:

```bash
LOG_LEVEL=debug ./bin/wanikani-api
```

Log levels: `debug`, `info`, `warn`, `error`

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes with tests
4. Run the full test suite: `go test ./...`
5. Submit a pull request

### Development Guidelines

- Follow Go best practices and idioms
- Add tests for new functionality
- Update documentation as needed
- Run `go fmt` before committing
- Ensure all tests pass

## License

[Add your license here]

## Acknowledgments

- [WaniKani](https://www.wanikani.com) for providing the API
- Built with Go and SQLite
- Property-based testing with [gopter](https://github.com/leanovate/gopter)

## Support

For issues and questions:
- Check existing [documentation](docs/)
- Review [WaniKani API documentation](https://docs.api.wanikani.com/)
- Open an issue on GitHub

---

**Happy learning! 頑張って！(Ganbatte!)**
