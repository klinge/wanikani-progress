# Integration Test Setup - Improved

## What Changed

The integration tests now use the same `config.Load()` function as the main application, making them much cleaner and more maintainable.

### Before
```go
func getAPIToken(t *testing.T) string {
    token := os.Getenv("WANIKANI_API_TOKEN")
    if token == "" {
        t.Skip("WANIKANI_API_TOKEN environment variable not set")
    }
    return token
}
```

### After
```go
func getAPIToken(t *testing.T) string {
    cfg, err := config.Load()
    if err != nil {
        t.Skipf("Failed to load config: %v", err)
    }
    return cfg.WaniKaniAPIToken
}
```

## Benefits

1. **Consistent Configuration**: Tests use the exact same config loading as the application
2. **Automatic .env Loading**: No need to manually export environment variables
3. **Better Error Messages**: Clear feedback if config is missing
4. **Maintainable**: Changes to config loading automatically apply to tests

## Usage

### Quick Start
```bash
# 1. Set up your .env file (if not already done)
cp .env.example .env
# Edit .env and set WANIKANI_API_TOKEN=your-token

# 2. Run integration tests
make test-int

# Or directly with go
go test -tags=integration -v ./internal/wanikani
```

### Available Make Targets
```bash
make test              # Run unit tests only
make test-integration  # Run integration tests (requires .env)
make test-int          # Alias for test-integration
make test-all          # Run both unit and integration tests
```

## How It Works

1. Integration tests call `config.Load()`
2. `config.Load()` uses `godotenv.Load()` to read `.env` file
3. Environment variables override `.env` values
4. Tests get the token from the loaded config
5. If config loading fails, tests are skipped with a helpful message

## No Manual Environment Variables Needed

The tests automatically pick up your `.env` file, so you don't need to:
- Export environment variables before running tests
- Pass variables inline with the test command
- Maintain separate test configuration

Just make sure your `.env` file exists and has `WANIKANI_API_TOKEN` set!
