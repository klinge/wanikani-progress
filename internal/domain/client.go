package domain

import (
	"context"
	"time"
)

// WaniKaniClient defines the interface for communicating with the WaniKani API
type WaniKaniClient interface {
	// SetAPIToken sets the API token for authentication
	SetAPIToken(token string)

	// FetchSubjects retrieves subjects from the WaniKani API
	// If updatedAfter is provided, only subjects modified after that time are returned
	FetchSubjects(ctx context.Context, updatedAfter *time.Time) ([]Subject, error)

	// FetchAssignments retrieves assignments from the WaniKani API
	// If updatedAfter is provided, only assignments modified after that time are returned
	FetchAssignments(ctx context.Context, updatedAfter *time.Time) ([]Assignment, error)

	// FetchReviews retrieves reviews from the WaniKani API
	// If updatedAfter is provided, only reviews modified after that time are returned
	FetchReviews(ctx context.Context, updatedAfter *time.Time) ([]Review, error)

	// FetchStatistics retrieves the current statistics snapshot from the WaniKani API
	FetchStatistics(ctx context.Context) (*Statistics, error)

	// GetRateLimitStatus returns the current rate limit information
	GetRateLimitStatus() RateLimitInfo
}
