# WaniKani Client Integration Tests

This directory contains integration tests that verify the WaniKani API client works correctly with the real WaniKani API.

## Setup

1. **Get your WaniKani API token:**
   - Log in to [WaniKani](https://www.wanikani.com)
   - Go to Settings → API Tokens
   - Generate a new token (or use an existing one)

2. **Configure your token:**
   
   The integration tests use the same config loading as the main application.
   
   **Option A: Use .env file (recommended)**
   ```bash
   cp .env.example .env
   # Edit .env and set WANIKANI_API_TOKEN=your-actual-token-here
   ```

   **Option B: Set environment variable**
   ```bash
   export WANIKANI_API_TOKEN="your-actual-token-here"
   ```

## Running Integration Tests

### Run all integration tests:
```bash
go test -tags=integration -v ./internal/wanikani
```

### Run a specific integration test:
```bash
go test -tags=integration -v ./internal/wanikani -run TestIntegration_FetchSubjects
```

### Run with temporary environment variable:
```bash
WANIKANI_API_TOKEN="your-token" go test -tags=integration -v ./internal/wanikani
```

**Note:** The tests use `config.Load()` which automatically reads from `.env` file if present, so you typically don't need to set environment variables manually.

## What the Tests Verify

- ✅ **Authentication**: Token is properly included in requests
- ✅ **Fetch Subjects**: Can retrieve subjects from the API
- ✅ **Fetch Assignments**: Can retrieve user assignments
- ✅ **Fetch Reviews**: Can retrieve review history
- ✅ **Fetch Statistics**: Can retrieve summary statistics
- ✅ **Pagination**: Automatically fetches all pages of results
- ✅ **Incremental Updates**: `updated_after` parameter works correctly
- ✅ **Rate Limit Tracking**: Rate limit headers are captured
- ✅ **Error Handling**: Invalid tokens return proper errors

## Notes

- Integration tests are **skipped** during normal test runs (`go test ./...`)
- They only run when explicitly requested with the `-tags=integration` flag
- Tests require a valid WaniKani account with some data
- Some tests may return no data if your account is new or hasn't been used recently
- Tests use real API calls and count against your rate limit (but minimally)

## Troubleshooting

**"Failed to load config (is WANIKANI_API_TOKEN set?)"**
- Make sure you've created a `.env` file with your token, or
- Export the `WANIKANI_API_TOKEN` environment variable
- The tests use the same config loading as the main application

**"Failed to fetch subjects: authentication error"**
- Your API token may be invalid or expired
- Generate a new token from WaniKani settings

**"No subjects/assignments/reviews returned"**
- This is normal for new accounts or accounts with limited activity
- The tests will log warnings but won't fail
