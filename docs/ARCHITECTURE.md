# Go Project Architecture

## Current Structure

```
wanikani-api/
├── cmd/
│   └── wanikani-api/        # Application entry point (main.go)
├── internal/                # Private application code
│   ├── config/              # Configuration management
│   ├── domain/              # Domain types and interfaces
│   └── store/               # Data storage implementations
│       └── sqlite/
├── pkg/                     # Public libraries (currently empty)
├── bin/                     # Build output (gitignored)
└── scripts/                 # Build and utility scripts
```

## Is This Structure Common?

**Yes!** This follows the [Standard Go Project Layout](https://github.com/golang-standards/project-layout), which is widely adopted in the Go community.

### Key Directories Explained

#### `internal/` ✅ **Very Common**
- Code that is **private** to this application
- Cannot be imported by other projects (enforced by Go compiler)
- Perfect for application-specific logic
- **When to use**: Almost always for application code

**Common subdirectories in `internal/`:**
- `internal/domain/` - Domain models, interfaces, business logic
- `internal/config/` - Configuration management
- `internal/store/` or `internal/repository/` - Data access layer
- `internal/service/` - Business logic services
- `internal/handler/` or `internal/api/` - HTTP handlers
- `internal/middleware/` - HTTP middleware

#### `cmd/` ✅ **Very Common**
- Contains application entry points (main packages)
- Each subdirectory is a separate executable
- Example: `cmd/server/`, `cmd/cli/`, `cmd/worker/`

#### `pkg/` ⚠️ **Use Sparingly**
- Code that **can be imported** by other projects
- Should be well-documented and stable
- **When to use**: Only for reusable libraries you want to share
- **Current status**: Empty (which is fine!)

## Alternative Structures

### 1. Flat Structure (Small Projects)
```
wanikani-api/
├── main.go              # For very small projects
├── config.go
├── store.go
└── handlers.go
```
**When to use**: Tiny projects with < 5 files

### 2. Feature-Based Structure
```
internal/
├── subjects/            # Everything related to subjects
│   ├── handler.go
│   ├── service.go
│   ├── repository.go
│   └── types.go
├── assignments/         # Everything related to assignments
│   ├── handler.go
│   ├── service.go
│   └── repository.go
└── sync/                # Sync feature
    ├── service.go
    └── scheduler.go
```
**When to use**: Domain-driven design, microservices

### 3. Layer-Based Structure (What You Have)
```
internal/
├── domain/              # Domain models and interfaces
├── service/             # Business logic
├── repository/          # Data access (or "store")
├── handler/             # HTTP handlers (or "api")
└── config/              # Configuration
```
**When to use**: Traditional layered architecture (very common!)

### 4. Hexagonal/Clean Architecture
```
internal/
├── core/                # Business logic (no dependencies)
│   ├── domain/
│   └── ports/           # Interfaces
├── adapters/            # External implementations
│   ├── http/
│   ├── sqlite/
│   └── wanikani/
└── config/
```
**When to use**: Complex applications, strict separation of concerns

## Your Current Structure: Analysis

### ✅ What's Good

1. **`internal/domain/`** - Great for interfaces and types
2. **`internal/config/`** - Standard location for configuration
3. **`internal/store/`** - Clear data access layer
4. **`cmd/`** - Proper entry point location
5. **Separation of concerns** - Each package has a clear purpose

### 🤔 Potential Additions (As You Grow)

```
internal/
├── config/              # ✅ Already have
├── domain/              # ✅ Already have
├── store/               # ✅ Already have
├── client/              # WaniKani API client (from task 3)
│   └── wanikani/
├── service/             # Sync service (from task 5)
│   └── sync/
├── api/                 # REST API handlers (from task 7)
│   └── handlers/
├── scheduler/           # Cron scheduler (from task 9)
└── middleware/          # HTTP middleware (if needed)
```

## Common Patterns in Go Projects

### Pattern 1: Repository Pattern (What You're Using)
```
internal/
├── domain/              # Interfaces + types
└── store/               # Implementations
    ├── sqlite/
    └── postgres/        # Easy to add alternatives
```

### Pattern 2: Service Layer
```
internal/
├── domain/              # Types
├── repository/          # Data access
└── service/             # Business logic
    ├── sync.go
    └── statistics.go
```

### Pattern 3: Handler-Service-Repository
```
internal/
├── handler/             # HTTP layer
├── service/             # Business logic
└── repository/          # Data access
```

## Recommendations for Your Project

Your current structure is **excellent** for this project! Here's what I'd suggest:

### Keep As-Is ✅
- `internal/domain/` - Interfaces and types
- `internal/config/` - Configuration
- `internal/store/` - Data access

### Add As You Implement Tasks
```
internal/
├── client/              # Task 3: WaniKani API client
│   └── wanikani/
├── service/             # Task 5: Sync service
│   └── sync/
├── api/                 # Task 7: REST API
│   ├── handlers/
│   └── middleware/
└── scheduler/           # Task 9: Cron scheduler
```

### Alternative: Keep It Simpler
If you prefer fewer directories:
```
internal/
├── domain/              # Types and interfaces
├── config/              # Configuration
├── store/               # Data access
├── wanikani/            # WaniKani client
├── sync/                # Sync service
├── api/                 # REST handlers
└── scheduler/           # Cron jobs
```

## Real-World Examples

Popular Go projects using similar structures:

1. **Kubernetes** - Uses `pkg/` and `cmd/` heavily
2. **Docker** - Uses `internal/` for private code
3. **Prometheus** - Layer-based with `storage/`, `web/`, `config/`
4. **Grafana** - Feature-based structure
5. **Hugo** - Mix of `internal/` and `pkg/`

## Best Practices

1. ✅ **Use `internal/`** - Prevents accidental imports
2. ✅ **Use `cmd/`** - Clear entry points
3. ⚠️ **Use `pkg/` sparingly** - Only for truly reusable code
4. ✅ **Group by layer or feature** - Pick one and be consistent
5. ✅ **Keep packages focused** - Single responsibility
6. ✅ **Avoid circular dependencies** - Domain should not import store

## Conclusion

Your structure is **very common and well-organized**! It follows Go best practices and will scale well as your project grows. The `internal/` directory with subdirectories for different concerns is exactly what most Go developers would expect to see.

**TL;DR**: Yes, your structure is idiomatic Go! 🎉
