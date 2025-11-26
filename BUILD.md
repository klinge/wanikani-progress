# Build Guide

## Directory Structure

```
wanikani-api/
├── bin/              # Compiled binaries (gitignored)
├── cmd/              # Application entry points
│   └── wanikani-api/ # Main application
├── internal/         # Private application code
│   ├── config/       # Configuration management
│   ├── domain/       # Domain types and interfaces
│   └── store/        # Data storage implementations
├── pkg/              # Public libraries (if any)
└── scripts/          # Build and utility scripts
```

## Build Output

### Standard Go Conventions

Go projects typically use these directories for build artifacts:

- **`bin/`** - Compiled binaries (most common, used by this project)
- **`dist/`** - Distribution packages (alternative)
- **`build/`** - Build artifacts (less common)

This project uses `bin/` as it's the most widely adopted convention in the Go community.

## Building

### Quick Start

```bash
# Using Make (easiest)
make build

# Using Go directly
go build -o bin/wanikani-api ./cmd/wanikani-api

# Using the build script
./scripts/build.sh
```

### Build Options

#### Development Build
```bash
make build
# or
go build -o bin/wanikani-api ./cmd/wanikani-api
```

#### Production Build (with optimizations)
```bash
go build -ldflags="-s -w" -o bin/wanikani-api ./cmd/wanikani-api
```
- `-s` removes symbol table
- `-w` removes DWARF debugging info
- Results in smaller binary size

#### Cross-Platform Builds
```bash
# Build for all platforms
make build-all

# Or manually for specific platforms
GOOS=linux GOARCH=amd64 go build -o bin/wanikani-api-linux-amd64 ./cmd/wanikani-api
GOOS=darwin GOARCH=amd64 go build -o bin/wanikani-api-darwin-amd64 ./cmd/wanikani-api
GOOS=windows GOARCH=amd64 go build -o bin/wanikani-api-windows-amd64.exe ./cmd/wanikani-api
```

## Running

### After Building
```bash
./bin/wanikani-api
```

### Without Building (Development)
```bash
go run ./cmd/wanikani-api
```

### With Make
```bash
make run
```

## Testing

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./internal/store/sqlite/...
```

## Cleaning

```bash
# Remove build artifacts
make clean

# Or manually
rm -rf bin/
```

## Installing Dependencies

```bash
make install
# or
go mod download
go mod tidy
```

## Code Quality

```bash
# Format code
make fmt
# or
go fmt ./...

# Run linter (requires golangci-lint)
make lint
```

## Tips

1. **Always use `bin/` for binaries** - It's gitignored and follows Go conventions
2. **Use Make for convenience** - Simplifies common tasks
3. **Cross-compile easily** - Go makes it trivial to build for other platforms
4. **Keep binaries out of git** - They're in `.gitignore` for a reason
5. **Use `go run` for quick testing** - No need to build during development
