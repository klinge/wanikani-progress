# Environment Configuration

This project uses environment variables for configuration. You can set these variables in two ways:

## Option 1: Using a .env file (Recommended for Development)

1. Copy the example file:
   ```bash
   cp .env.example .env
   ```

2. Edit `.env` and add your actual values:
   ```bash
   WANIKANI_API_TOKEN=your_actual_api_token_here
   DATABASE_PATH=./wanikani.db
   SYNC_SCHEDULE=0 2 * * *
   API_PORT=8080
   LOG_LEVEL=info
   ```

3. The application will automatically load these values when it starts.

**Note:** The `.env` file is in `.gitignore` to prevent accidentally committing secrets.

## Option 2: Using System Environment Variables

You can also set environment variables directly in your shell:

```bash
export WANIKANI_API_TOKEN=your_api_token_here
export DATABASE_PATH=./wanikani.db
export API_PORT=8080
```

Or pass them when running the application:

```bash
WANIKANI_API_TOKEN=your_token go run cmd/wanikani-api/main.go
```

## Configuration Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `WANIKANI_API_TOKEN` | Yes | - | Your WaniKani API token (get it from https://www.wanikani.com/settings/personal_access_tokens) |
| `DATABASE_PATH` | No | `./wanikani.db` | Path to the SQLite database file |
| `SYNC_SCHEDULE` | No | `0 2 * * *` | NOT USED: Cron expression for daily sync (default: 2 AM daily) |
| `API_PORT` | No | `8080` | Port for the API server |
| `LOG_LEVEL` | No | `info` | Logging level (debug, info, warn, error) |

## Priority

Environment variables take precedence in this order:
1. System environment variables (highest priority)
2. Variables from `.env` file
3. Default values (lowest priority)

This means you can override `.env` values by setting system environment variables.


## Building the Application

### Using Make (Recommended)

```bash
# Build the binary to bin/wanikani-api
make build

# Run tests
make test

# Build and run
make run

# Clean build artifacts
make clean

# See all available commands
make help
```

### Using Go directly

```bash
# Build to bin/ directory
go build -o bin/wanikani-api ./cmd/wanikani-api

# Run directly without building
go run ./cmd/wanikani-api
```

### Using the build script

```bash
./scripts/build.sh
```

### Build Output

All compiled binaries are placed in the `bin/` directory, which is ignored by git. This follows Go's standard convention for build artifacts.
