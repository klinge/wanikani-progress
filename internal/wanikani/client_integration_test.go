//go:build integration
// +build integration

package wanikani

import (
	"context"
	"testing"
	"time"

	"wanikani-api/internal/config"
)

// These tests require a valid WaniKani API token in .env file or WANIKANI_API_TOKEN environment variable.
// Run with: go test -tags=integration -v ./internal/wanikani

func getAPIToken(t *testing.T) string {
	// Load config which will read from .env file or environment variables
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("Failed to load config (is WANIKANI_API_TOKEN set?): %v", err)
	}
	return cfg.WaniKaniAPIToken
}

func TestIntegration_FetchSubjects(t *testing.T) {
	token := getAPIToken(t)
	client := NewClient()
	client.SetAPIToken(token)

	ctx := context.Background()

	t.Log("Fetching subjects from WaniKani API...")
	subjects, err := client.FetchSubjects(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to fetch subjects: %v", err)
	}

	t.Logf("Successfully fetched %d subjects", len(subjects))

	if len(subjects) == 0 {
		t.Log("Warning: No subjects returned. This might be expected for a new account.")
		return
	}

	// Verify structure of first subject
	subject := subjects[0]
	if subject.ID == 0 {
		t.Error("Subject ID should not be zero")
	}
	if subject.Object == "" {
		t.Error("Subject Object should not be empty")
	}
	if subject.Data.Level == 0 {
		t.Error("Subject Level should not be zero")
	}

	t.Logf("Sample subject: ID=%d, Object=%s, Level=%d, Characters=%s",
		subject.ID, subject.Object, subject.Data.Level, subject.Data.Characters)
}

func TestIntegration_FetchSubjects_WithUpdatedAfter(t *testing.T) {
	token := getAPIToken(t)
	client := NewClient()
	client.SetAPIToken(token)

	ctx := context.Background()

	// Fetch subjects updated in the last 30 days
	updatedAfter := time.Now().AddDate(0, 0, -30)
	t.Logf("Fetching subjects updated after %s...", updatedAfter.Format(time.RFC3339))

	subjects, err := client.FetchSubjects(ctx, &updatedAfter)
	if err != nil {
		t.Fatalf("Failed to fetch subjects with updated_after: %v", err)
	}

	t.Logf("Successfully fetched %d subjects updated in the last 30 days", len(subjects))

	// Verify all returned subjects were updated after the specified time
	for _, subject := range subjects {
		if subject.DataUpdatedAt.Before(updatedAfter) {
			t.Errorf("Subject %d was updated at %s, which is before the requested time %s",
				subject.ID, subject.DataUpdatedAt, updatedAfter)
		}
	}
}

func TestIntegration_FetchAssignments(t *testing.T) {
	token := getAPIToken(t)
	client := NewClient()
	client.SetAPIToken(token)

	ctx := context.Background()

	t.Log("Fetching assignments from WaniKani API...")
	assignments, err := client.FetchAssignments(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to fetch assignments: %v", err)
	}

	t.Logf("Successfully fetched %d assignments", len(assignments))

	if len(assignments) == 0 {
		t.Log("Warning: No assignments returned. This might be expected for a new account.")
		return
	}

	// Verify structure of first assignment
	assignment := assignments[0]
	if assignment.ID == 0 {
		t.Error("Assignment ID should not be zero")
	}
	if assignment.Data.SubjectID == 0 {
		t.Error("Assignment SubjectID should not be zero")
	}
	if assignment.Data.SubjectType == "" {
		t.Error("Assignment SubjectType should not be empty")
	}

	t.Logf("Sample assignment: ID=%d, SubjectID=%d, SubjectType=%s, SRSStage=%d",
		assignment.ID, assignment.Data.SubjectID, assignment.Data.SubjectType, assignment.Data.SRSStage)
}

func TestIntegration_FetchReviews(t *testing.T) {
	token := getAPIToken(t)
	client := NewClient()
	client.SetAPIToken(token)

	ctx := context.Background()

	// Fetch reviews from the last 7 days to limit the amount of data
	updatedAfter := time.Now().AddDate(0, 0, -7)
	t.Logf("Fetching reviews updated after %s...", updatedAfter.Format(time.RFC3339))

	reviews, err := client.FetchReviews(ctx, &updatedAfter)
	if err != nil {
		t.Fatalf("Failed to fetch reviews: %v", err)
	}

	t.Logf("Successfully fetched %d reviews from the last 7 days", len(reviews))

	if len(reviews) == 0 {
		t.Log("No reviews in the last 7 days. This is expected if you haven't done reviews recently.")
		return
	}

	// Verify structure of first review
	review := reviews[0]
	if review.ID == 0 {
		t.Error("Review ID should not be zero")
	}
	if review.Data.AssignmentID == 0 {
		t.Error("Review AssignmentID should not be zero")
	}
	if review.Data.SubjectID == 0 {
		t.Error("Review SubjectID should not be zero")
	}

	t.Logf("Sample review: ID=%d, AssignmentID=%d, SubjectID=%d, Created=%s",
		review.ID, review.Data.AssignmentID, review.Data.SubjectID, review.Data.CreatedAt)
}

func TestIntegration_FetchStatistics(t *testing.T) {
	token := getAPIToken(t)
	client := NewClient()
	client.SetAPIToken(token)

	ctx := context.Background()

	t.Log("Fetching statistics from WaniKani API...")
	stats, err := client.FetchStatistics(ctx)
	if err != nil {
		t.Fatalf("Failed to fetch statistics: %v", err)
	}

	if stats == nil {
		t.Fatal("Statistics should not be nil")
	}

	t.Logf("Successfully fetched statistics")
	t.Logf("Lessons available: %d", len(stats.Data.Lessons))
	t.Logf("Reviews available: %d", len(stats.Data.Reviews))

	if stats.Object != "report" {
		t.Errorf("Expected Object to be 'report', got '%s'", stats.Object)
	}
}

func TestIntegration_RateLimitTracking(t *testing.T) {
	token := getAPIToken(t)
	client := NewClient()
	client.SetAPIToken(token)

	ctx := context.Background()

	// Make a request to populate rate limit info
	t.Log("Making request to check rate limit tracking...")
	_, err := client.FetchStatistics(ctx)
	if err != nil {
		t.Fatalf("Failed to fetch statistics: %v", err)
	}

	// Check rate limit info
	rateLimitInfo := client.GetRateLimitStatus()
	t.Logf("Rate limit remaining: %d", rateLimitInfo.Remaining)
	t.Logf("Rate limit resets at: %s", rateLimitInfo.ResetAt)

	if rateLimitInfo.Remaining < 0 {
		t.Error("Rate limit remaining should not be negative")
	}

	if rateLimitInfo.ResetAt.IsZero() {
		t.Log("Warning: Rate limit reset time not set. API might not be returning rate limit headers.")
	}
}

func TestIntegration_Pagination(t *testing.T) {
	token := getAPIToken(t)
	client := NewClient()
	client.SetAPIToken(token)

	ctx := context.Background()

	t.Log("Testing pagination by fetching all subjects...")
	subjects, err := client.FetchSubjects(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to fetch subjects: %v", err)
	}

	t.Logf("Fetched %d total subjects across all pages", len(subjects))

	// WaniKani typically returns 1000 items per page for subjects
	// If we have more than 1000, pagination worked
	if len(subjects) > 1000 {
		t.Logf("✓ Pagination working correctly (fetched more than 1000 subjects)")
	} else if len(subjects) > 0 {
		t.Logf("Fetched %d subjects (less than one full page, or account has limited data)", len(subjects))
	}
}

func TestIntegration_AuthenticationError(t *testing.T) {
	client := NewClient()
	client.SetAPIToken("invalid-token-12345")

	ctx := context.Background()

	t.Log("Testing with invalid API token...")
	_, err := client.FetchSubjects(ctx, nil)
	if err == nil {
		t.Fatal("Expected authentication error with invalid token, got nil")
	}

	t.Logf("Got expected error: %v", err)

	// Check if it's an auth error
	if _, ok := err.(*authError); !ok {
		// The error might be wrapped, check the message
		if err.Error() == "" {
			t.Error("Expected non-empty error message")
		}
	}
}
