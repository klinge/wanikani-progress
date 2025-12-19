package api

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"

	"wanikani-api/internal/domain"
	"wanikani-api/internal/store/sqlite"
)

// Feature: wanikani-api, Property 7: Query filter correctness
// Validates: Requirements 5.1, 4.3, 8.4
func TestProperty_QueryFilterCorrectness(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	// Test subject filters
	properties.Property("all returned subjects match the filter criteria", prop.ForAll(
		func(subjects []domain.Subject, filterType string, filterLevel *int) bool {
			// Create temporary database
			dbPath := "test_filter_subjects_" + randomString(8) + ".db"
			defer os.Remove(dbPath)

			store, err := sqlite.New(dbPath)
			if err != nil {
				t.Logf("failed to create store: %v", err)
				return false
			}
			defer store.Close()

			ctx := context.Background()

			// Insert all subjects
			if err := store.UpsertSubjects(ctx, subjects); err != nil {
				t.Logf("failed to upsert subjects: %v", err)
				return false
			}

			// Apply filters
			filters := domain.SubjectFilters{
				Type:  filterType,
				Level: filterLevel,
			}

			// Query with filters
			results, err := store.GetSubjects(ctx, filters)
			if err != nil {
				t.Logf("failed to get subjects: %v", err)
				return false
			}

			// Verify all results match the filter
			for _, result := range results {
				if filterType != "" && result.Object != filterType {
					t.Logf("result object %q does not match filter type %q", result.Object, filterType)
					return false
				}
				if filterLevel != nil && result.Data.Level != *filterLevel {
					t.Logf("result level %d does not match filter level %d", result.Data.Level, *filterLevel)
					return false
				}
			}

			// Verify no matching records were excluded
			// Count how many subjects in the original set should match
			expectedCount := 0
			for _, subject := range subjects {
				matches := true
				if filterType != "" && subject.Object != filterType {
					matches = false
				}
				if filterLevel != nil && subject.Data.Level != *filterLevel {
					matches = false
				}
				if matches {
					expectedCount++
				}
			}

			if len(results) != expectedCount {
				t.Logf("expected %d results, got %d", expectedCount, len(results))
				return false
			}

			return true
		},
		genSubjects(),
		genSubjectType(),
		genOptionalLevel(),
	))

	// Test assignment filters
	properties.Property("all returned assignments match the filter criteria", prop.ForAll(
		func(testData assignmentTestData, filterSRSStage *int) bool {
			// Create temporary database
			dbPath := "test_filter_assignments_" + randomString(8) + ".db"
			defer os.Remove(dbPath)

			store, err := sqlite.New(dbPath)
			if err != nil {
				t.Logf("failed to create store: %v", err)
				return false
			}
			defer store.Close()

			ctx := context.Background()

			// Insert subjects first (for referential integrity)
			if err := store.UpsertSubjects(ctx, testData.Subjects); err != nil {
				t.Logf("failed to upsert subjects: %v", err)
				return false
			}

			// Insert assignments
			if err := store.UpsertAssignments(ctx, testData.Assignments); err != nil {
				t.Logf("failed to upsert assignments: %v", err)
				return false
			}

			// Apply filters
			filters := domain.AssignmentFilters{
				SRSStage: filterSRSStage,
			}

			// Query with filters
			results, err := store.GetAssignments(ctx, filters)
			if err != nil {
				t.Logf("failed to get assignments: %v", err)
				return false
			}

			// Verify all results match the filter
			for _, result := range results {
				if filterSRSStage != nil && result.Data.SRSStage != *filterSRSStage {
					t.Logf("result SRS stage %d does not match filter %d", result.Data.SRSStage, *filterSRSStage)
					return false
				}
			}

			// Verify no matching records were excluded
			expectedCount := 0
			for _, assignment := range testData.Assignments {
				matches := true
				if filterSRSStage != nil && assignment.Data.SRSStage != *filterSRSStage {
					matches = false
				}
				if matches {
					expectedCount++
				}
			}

			if len(results) != expectedCount {
				t.Logf("expected %d results, got %d", expectedCount, len(results))
				return false
			}

			return true
		},
		genAssignmentTestData(),
		genOptionalSRSStage(),
	))

	// Test review filters
	properties.Property("all returned reviews match the filter criteria", prop.ForAll(
		func(testData reviewTestData, filterFrom *time.Time, filterTo *time.Time) bool {
			// Create temporary database
			dbPath := "test_filter_reviews_" + randomString(8) + ".db"
			defer os.Remove(dbPath)

			store, err := sqlite.New(dbPath)
			if err != nil {
				t.Logf("failed to create store: %v", err)
				return false
			}
			defer store.Close()

			ctx := context.Background()

			// Insert subjects first
			if err := store.UpsertSubjects(ctx, testData.Subjects); err != nil {
				t.Logf("failed to upsert subjects: %v", err)
				return false
			}

			// Insert assignments
			if err := store.UpsertAssignments(ctx, testData.Assignments); err != nil {
				t.Logf("failed to upsert assignments: %v", err)
				return false
			}

			// Insert reviews
			if err := store.UpsertReviews(ctx, testData.Reviews); err != nil {
				t.Logf("failed to upsert reviews: %v", err)
				return false
			}

			// Apply filters
			filters := domain.ReviewFilters{
				From: filterFrom,
				To:   filterTo,
			}

			// Query with filters
			results, err := store.GetReviews(ctx, filters)
			if err != nil {
				t.Logf("failed to get reviews: %v", err)
				return false
			}

			// Verify all results match the filter
			// Use Unix timestamps for comparison to avoid precision issues with RFC3339 string comparison
			for _, result := range results {
				resultUnix := result.Data.CreatedAt.Unix()
				if filterFrom != nil {
					fromUnix := filterFrom.Unix()
					if resultUnix < fromUnix {
						t.Logf("result created_at %v (unix: %d) is before filter from %v (unix: %d)",
							result.Data.CreatedAt, resultUnix, *filterFrom, fromUnix)
						return false
					}
				}
				if filterTo != nil {
					toUnix := filterTo.Unix()
					if resultUnix > toUnix {
						t.Logf("result created_at %v (unix: %d) is after filter to %v (unix: %d)",
							result.Data.CreatedAt, resultUnix, *filterTo, toUnix)
						return false
					}
				}
			}

			// Verify no matching records were excluded
			expectedCount := 0
			for _, review := range testData.Reviews {
				matches := true
				reviewUnix := review.Data.CreatedAt.Unix()

				if filterFrom != nil && reviewUnix < filterFrom.Unix() {
					matches = false
				}
				if filterTo != nil && reviewUnix > filterTo.Unix() {
					matches = false
				}
				if matches {
					expectedCount++
				}
			}

			if len(results) != expectedCount {
				t.Logf("expected %d results, got %d", expectedCount, len(results))
				return false
			}

			return true
		},
		genReviewTestData(),
		genOptionalTime(),
		genOptionalTime(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Test data structures

type assignmentTestData struct {
	Subjects    []domain.Subject
	Assignments []domain.Assignment
}

type reviewTestData struct {
	Subjects    []domain.Subject
	Assignments []domain.Assignment
	Reviews     []domain.Review
}

// Generators

func genSubjects() gopter.Gen {
	return gen.SliceOfN(10, genSubject())
}

func genSubject() gopter.Gen {
	return gopter.CombineGens(
		gen.IntRange(1, 10000),
		genSubjectType(),
		gen.IntRange(1, 60),
		gen.AlphaString(),
	).Map(func(values []interface{}) domain.Subject {
		id := values[0].(int)
		objType := values[1].(string)
		level := values[2].(int)
		chars := values[3].(string)
		if chars == "" {
			chars = "字"
		}

		return domain.Subject{
			ID:            id,
			Object:        objType,
			URL:           "https://api.wanikani.com/v2/subjects/" + string(rune(id)),
			DataUpdatedAt: time.Now(),
			Data: domain.SubjectData{
				Level:      level,
				Characters: chars,
				Meanings: []domain.Meaning{
					{Meaning: "test", Primary: true},
				},
			},
		}
	})
}

func genSubjectType() gopter.Gen {
	return gen.OneConstOf("radical", "kanji", "vocabulary", "")
}

func genOptionalLevel() gopter.Gen {
	return gen.OneGenOf(
		gen.Const((*int)(nil)),
		gen.IntRange(1, 60).Map(func(v int) *int { return &v }),
	)
}

func genAssignmentTestData() gopter.Gen {
	return genSubjects().FlatMap(func(subjects interface{}) gopter.Gen {
		subjectList := subjects.([]domain.Subject)

		if len(subjectList) == 0 {
			return gen.Const(assignmentTestData{
				Subjects:    []domain.Subject{},
				Assignments: []domain.Assignment{},
			})
		}

		// Generate SRS stages for all assignments
		return gen.SliceOfN(len(subjectList), gen.IntRange(0, 9)).Map(func(srsStages []int) assignmentTestData {
			assignments := make([]domain.Assignment, len(subjectList))
			for i, subject := range subjectList {
				assignments[i] = domain.Assignment{
					ID:            100 + i,
					Object:        "assignment",
					URL:           "https://api.wanikani.com/v2/assignments/" + string(rune(100+i)),
					DataUpdatedAt: time.Now(),
					Data: domain.AssignmentData{
						SubjectID:   subject.ID,
						SubjectType: subject.Object,
						SRSStage:    srsStages[i],
					},
				}
			}
			return assignmentTestData{
				Subjects:    subjectList,
				Assignments: assignments,
			}
		})
	}, reflect.TypeOf(assignmentTestData{}))
}

func genOptionalSRSStage() gopter.Gen {
	return gen.OneGenOf(
		gen.Const((*int)(nil)),
		gen.IntRange(0, 9).Map(func(v int) *int { return &v }),
	)
}

func genReviewTestData() gopter.Gen {
	return genAssignmentTestData().FlatMap(func(assignmentData interface{}) gopter.Gen {
		data := assignmentData.(assignmentTestData)

		if len(data.Assignments) == 0 {
			return gen.Const(reviewTestData{
				Subjects:    data.Subjects,
				Assignments: data.Assignments,
				Reviews:     []domain.Review{},
			})
		}

		// Truncate base time to second precision to match RFC3339 storage
		baseTime := time.Now().Add(-365 * 24 * time.Hour).Truncate(time.Second)

		// Generate review data for all assignments
		return gopter.CombineGens(
			gen.SliceOfN(len(data.Assignments), gen.IntRange(0, 365)),
			gen.SliceOfN(len(data.Assignments), gen.IntRange(0, 5)),
			gen.SliceOfN(len(data.Assignments), gen.IntRange(0, 5)),
		).Map(func(values []interface{}) reviewTestData {
			daysOffsets := values[0].([]int)
			incorrectMeanings := values[1].([]int)
			incorrectReadings := values[2].([]int)

			reviews := make([]domain.Review, len(data.Assignments))
			for i, assignment := range data.Assignments {
				// Truncate to second precision to match RFC3339 storage
				createdAt := baseTime.Add(time.Duration(daysOffsets[i]) * 24 * time.Hour).Truncate(time.Second)
				reviews[i] = domain.Review{
					ID:            200 + i,
					Object:        "review",
					URL:           "https://api.wanikani.com/v2/reviews/" + string(rune(200+i)),
					DataUpdatedAt: time.Now(),
					Data: domain.ReviewData{
						AssignmentID:            assignment.ID,
						SubjectID:               assignment.Data.SubjectID,
						CreatedAt:               createdAt,
						IncorrectMeaningAnswers: incorrectMeanings[i],
						IncorrectReadingAnswers: incorrectReadings[i],
					},
				}
			}

			return reviewTestData{
				Subjects:    data.Subjects,
				Assignments: data.Assignments,
				Reviews:     reviews,
			}
		})
	}, reflect.TypeOf(reviewTestData{}))
}

func genOptionalTime() gopter.Gen {
	return gen.OneGenOf(
		gen.Const((*time.Time)(nil)),
		gen.Int64Range(0, 365).Map(func(daysAgo int64) *time.Time {
			// Truncate to second precision to match RFC3339 storage
			t := time.Now().Add(-time.Duration(daysAgo) * 24 * time.Hour).Truncate(time.Second)
			return &t
		}),
	)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// Feature: wanikani-api, Property 21: API authentication enforcement
// Validates: Requirements 11.1, 11.2, 11.3
func TestProperty_APIAuthenticationEnforcement(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping property test in short mode")
	}

	properties := gopter.NewProperties(gopter.DefaultTestParameters())

	// Property: For any API request without a valid authorization token (when LOCAL_API_TOKEN is configured),
	// the API Server should return a 401 Unauthorized response and reject the request
	properties.Property("requests without valid token are rejected with 401", prop.ForAll(
		func(validToken string, providedToken string, endpoint string) bool {

			// Create temporary database
			dbPath := "test_auth_" + randomString(8) + ".db"
			defer os.Remove(dbPath)

			store, err := sqlite.New(dbPath)
			if err != nil {
				t.Logf("failed to create store: %v", err)
				return false
			}
			defer store.Close()

			logger := testLogger()
			syncService := &mockSyncService{}

			// Create server with authentication enabled
			server := NewServer(store, syncService, 8080, validToken, logger)

			// Test the endpoint - use POST for /api/sync, GET for others
			method := "GET"
			if endpoint == "/api/sync" {
				method = "POST"
			}
			req := createTestRequest(method, endpoint, nil)

			// Set authorization header based on providedToken
			if providedToken != "" {
				req.Header.Set("Authorization", "Bearer "+providedToken)
			}

			w := executeTestRequest(server, req)

			// If the provided token matches the valid token, request should succeed (or fail for other reasons)
			// If the provided token doesn't match, should get 401
			if providedToken == validToken {
				// Valid token - should NOT get 401 (might get other errors like 404, 400, etc.)
				if w.Code == 401 {
					t.Logf("valid token was rejected with 401")
					return false
				}
			} else {
				// Invalid or missing token - should get 401
				if w.Code != 401 {
					t.Logf("invalid/missing token did not return 401, got %d", w.Code)
					return false
				}

				// Verify error response format
				var errResp ErrorResponse
				if err := decodeJSON(w.Body, &errResp); err != nil {
					t.Logf("failed to decode error response: %v", err)
					return false
				}

				if errResp.Error.Code != "UNAUTHORIZED" {
					t.Logf("expected error code UNAUTHORIZED, got %s", errResp.Error.Code)
					return false
				}
			}

			return true
		},
		genToken(),
		genOptionalToken(),
		genAPIEndpoint(),
	))

	// Property: Health check endpoint should not require authentication
	properties.Property("health endpoint does not require authentication", prop.ForAll(
		func(validToken string) bool {

			// Create temporary database
			dbPath := "test_health_" + randomString(8) + ".db"
			defer os.Remove(dbPath)

			store, err := sqlite.New(dbPath)
			if err != nil {
				t.Logf("failed to create store: %v", err)
				return false
			}
			defer store.Close()

			logger := testLogger()
			syncService := &mockSyncService{}

			// Create server with authentication enabled
			server := NewServer(store, syncService, 8080, validToken, logger)

			// Test health endpoint without authentication
			req := createTestRequest("GET", "/api/health", nil)
			w := executeTestRequest(server, req)

			// Health endpoint should return 200 without authentication
			if w.Code != 200 {
				t.Logf("health endpoint returned %d without auth, expected 200", w.Code)
				return false
			}

			return true
		},
		genToken(),
	))

	// Property: When no token is configured, all endpoints should work without authentication
	properties.Property("endpoints work without authentication when token not configured", prop.ForAll(
		func(endpoint string) bool {
			// Create temporary database
			dbPath := "test_no_auth_" + randomString(8) + ".db"
			defer os.Remove(dbPath)

			store, err := sqlite.New(dbPath)
			if err != nil {
				t.Logf("failed to create store: %v", err)
				return false
			}
			defer store.Close()

			logger := testLogger()
			syncService := &mockSyncService{}

			// Create server WITHOUT authentication (empty token)
			server := NewServer(store, syncService, 8080, "", logger)

			// Test endpoint without authorization header
			req := createTestRequest("GET", endpoint, nil)
			w := executeTestRequest(server, req)

			// Should NOT get 401 (might get 404, 400, or 200, but not 401)
			if w.Code == 401 {
				t.Logf("endpoint returned 401 when no auth configured")
				return false
			}

			return true
		},
		genAPIEndpoint(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Generators for authentication tests

func genToken() gopter.Gen {
	return gen.Identifier().SuchThat(func(s string) bool {
		return len(s) >= 10 && len(s) <= 50
	})
}

func genOptionalToken() gopter.Gen {
	return gen.OneGenOf(
		gen.Const(""),             // No token
		genToken(),                // Valid format token (but wrong value)
		gen.Const("invalid"),      // Invalid format
		gen.Const("Bearer token"), // Token with Bearer prefix (should not include it)
	)
}

func genAPIEndpoint() gopter.Gen {
	return gen.OneConstOf(
		"/api/subjects",
		"/api/assignments",
		"/api/reviews",
		"/api/statistics/latest",
		"/api/statistics",
		"/api/sync",
		"/api/sync/status",
	)
}
