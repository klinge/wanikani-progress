# Testing Guide

This project uses both unit tests and property-based tests. Property-based tests are more thorough but take longer to run.

## Running Tests

### Run all tests (including property tests)
```bash
go test ./...
```

### Run only fast unit tests (exclude property tests)
```bash
go test -short ./...
```

### Run only property tests
```bash
go test -run Property ./...
```

### Run tests for a specific package
```bash
# Fast tests only
go test -short ./internal/api/

# All tests including property tests
go test ./internal/api/
```

## Test Types

### Unit Tests
- Fast, focused tests for specific functionality
- Run by default with `go test`
- Should complete in seconds

### Property-Based Tests
- Thorough tests that verify properties across many random inputs
- Named with `TestProperty_*` prefix
- Run 100+ iterations per property
- May take 30-90 seconds per test file
- Can be skipped with `-short` flag

## CI/CD Recommendations

For continuous integration:
- Run fast tests (`-short`) on every commit
- Run full test suite (including property tests) on pull requests and before merges
- Consider running property tests nightly for comprehensive coverage

## Writing Tests

When adding new property-based tests:
1. Name them with `TestProperty_` prefix
2. Add a check for `testing.Short()` at the start:
   ```go
   func TestProperty_MyFeature(t *testing.T) {
       if testing.Short() {
           t.Skip("Skipping property test in short mode")
       }
       // ... rest of test
   }
   ```
3. Document which property and requirements the test validates
