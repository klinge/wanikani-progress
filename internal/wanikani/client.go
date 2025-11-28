package wanikani

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"wanikani-api/internal/domain"
)

const (
	baseURL        = "https://api.wanikani.com/v2"
	maxRetries     = 3
	initialBackoff = 1 * time.Second
)

// Client implements the WaniKaniClient interface
type Client struct {
	httpClient *http.Client
	apiToken   string
	logger     *logrus.Logger
	mu         sync.RWMutex // protects apiToken and rateLimitInfo
	rateLimit  domain.RateLimitInfo
}

// NewClient creates a new WaniKani API client
func NewClient(logger *logrus.Logger) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// SetAPIToken sets the API token for authentication
func (c *Client) SetAPIToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiToken = token
	c.logger.Debug("API token set successfully")
}

// GetRateLimitStatus returns the current rate limit information
func (c *Client) GetRateLimitStatus() domain.RateLimitInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.rateLimit
}

// FetchSubjects retrieves subjects from the WaniKani API
func (c *Client) FetchSubjects(ctx context.Context, updatedAfter *time.Time) ([]domain.Subject, error) {
	params := url.Values{}
	if updatedAfter != nil {
		params.Set("updated_after", updatedAfter.Format(time.RFC3339))
		c.logger.WithField("updated_after", updatedAfter.Format(time.RFC3339)).Debug("Fetching subjects with incremental update")
	} else {
		c.logger.Debug("Fetching all subjects")
	}

	var allSubjects []domain.Subject
	nextURL := fmt.Sprintf("%s/subjects?%s", baseURL, params.Encode())
	pageCount := 0

	for nextURL != "" {
		var response paginatedResponse
		var subjects []domain.Subject

		err := c.fetchWithRetry(ctx, nextURL, &response, &subjects)
		if err != nil {
			c.logger.WithError(err).Error("Failed to fetch subjects page")
			return nil, fmt.Errorf("failed to fetch subjects: %w", err)
		}

		pageCount++
		allSubjects = append(allSubjects, subjects...)
		nextURL = response.Pages.NextURL
	}

	c.logger.WithFields(logrus.Fields{
		"total_subjects": len(allSubjects),
		"pages_fetched":  pageCount,
	}).Info("Successfully fetched subjects from API")

	return allSubjects, nil
}

// FetchAssignments retrieves assignments from the WaniKani API
func (c *Client) FetchAssignments(ctx context.Context, updatedAfter *time.Time) ([]domain.Assignment, error) {
	params := url.Values{}
	if updatedAfter != nil {
		params.Set("updated_after", updatedAfter.Format(time.RFC3339))
		c.logger.WithField("updated_after", updatedAfter.Format(time.RFC3339)).Debug("Fetching assignments with incremental update")
	} else {
		c.logger.Debug("Fetching all assignments")
	}

	var allAssignments []domain.Assignment
	nextURL := fmt.Sprintf("%s/assignments?%s", baseURL, params.Encode())
	pageCount := 0

	for nextURL != "" {
		var response paginatedResponse
		var assignments []domain.Assignment

		err := c.fetchWithRetry(ctx, nextURL, &response, &assignments)
		if err != nil {
			c.logger.WithError(err).Error("Failed to fetch assignments page")
			return nil, fmt.Errorf("failed to fetch assignments: %w", err)
		}

		pageCount++
		allAssignments = append(allAssignments, assignments...)
		nextURL = response.Pages.NextURL
	}

	c.logger.WithFields(logrus.Fields{
		"total_assignments": len(allAssignments),
		"pages_fetched":     pageCount,
	}).Info("Successfully fetched assignments from API")

	return allAssignments, nil
}

// FetchReviews retrieves reviews from the WaniKani API
func (c *Client) FetchReviews(ctx context.Context, updatedAfter *time.Time) ([]domain.Review, error) {
	params := url.Values{}
	if updatedAfter != nil {
		params.Set("updated_after", updatedAfter.Format(time.RFC3339))
		c.logger.WithField("updated_after", updatedAfter.Format(time.RFC3339)).Debug("Fetching reviews with incremental update")
	} else {
		c.logger.Debug("Fetching all reviews")
	}

	var allReviews []domain.Review
	nextURL := fmt.Sprintf("%s/reviews?%s", baseURL, params.Encode())
	pageCount := 0

	for nextURL != "" {
		var response paginatedResponse
		var reviews []domain.Review

		err := c.fetchWithRetry(ctx, nextURL, &response, &reviews)
		if err != nil {
			c.logger.WithError(err).Error("Failed to fetch reviews page")
			return nil, fmt.Errorf("failed to fetch reviews: %w", err)
		}

		pageCount++
		allReviews = append(allReviews, reviews...)
		nextURL = response.Pages.NextURL
	}

	c.logger.WithFields(logrus.Fields{
		"total_reviews": len(allReviews),
		"pages_fetched": pageCount,
	}).Info("Successfully fetched reviews from API")

	return allReviews, nil
}

// FetchStatistics retrieves the current statistics snapshot from the WaniKani API
func (c *Client) FetchStatistics(ctx context.Context) (*domain.Statistics, error) {
	c.logger.Debug("Fetching statistics summary from API")
	endpoint := fmt.Sprintf("%s/summary", baseURL)

	// Summary endpoint returns data directly, not in a collection wrapper
	var stats domain.Statistics
	err := c.fetchWithRetry(ctx, endpoint, nil, &stats)
	if err != nil {
		c.logger.WithError(err).Error("Failed to fetch statistics")
		return nil, fmt.Errorf("failed to fetch statistics: %w", err)
	}

	c.logger.Info("Successfully fetched statistics from API")
	return &stats, nil
}

// fetchWithRetry performs an HTTP request with retry logic and exponential backoff
func (c *Client) fetchWithRetry(ctx context.Context, url string, paginationInfo *paginatedResponse, data interface{}) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Calculate wait duration based on error type
			waitDuration := backoff
			if rateLimitErr, ok := lastErr.(*rateLimitError); ok {
				// For rate limit errors, wait for the specified retry-after duration
				waitDuration = rateLimitErr.retryAfter
				c.logger.WithFields(logrus.Fields{
					"retry_after": waitDuration,
					"attempt":     attempt,
				}).Warn("Rate limit exceeded, waiting before retry")
			} else {
				c.logger.WithFields(logrus.Fields{
					"backoff": waitDuration,
					"attempt": attempt,
					"error":   lastErr,
				}).Warn("Request failed, retrying with exponential backoff")
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(waitDuration):
				backoff *= 2
			}
		}

		err := c.doRequest(ctx, url, paginationInfo, data)
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !isRetryableError(err) {
			c.logger.WithError(err).Error("Non-retryable error encountered")
			return err
		}
	}

	c.logger.WithError(lastErr).Error("Max retries exceeded")
	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doRequest performs a single HTTP request
func (c *Client) doRequest(ctx context.Context, url string, paginationInfo *paginatedResponse, data interface{}) error {
	// Check and wait for rate limit if necessary
	if err := c.waitForRateLimit(ctx); err != nil {
		return err
	}

	c.mu.RLock()
	token := c.apiToken
	c.mu.RUnlock()

	if token == "" {
		c.logger.Error("API token not set")
		return fmt.Errorf("API token not set")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Wanikani-Revision", "20170710")

	c.logger.WithField("url", url).Debug("Making API request")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.WithError(err).Error("Network error during API request")
		return &networkError{err: err}
	}
	defer resp.Body.Close()

	// Update rate limit information
	c.updateRateLimitInfo(resp)

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusUnauthorized {
			c.logger.Error("Authentication failed: Invalid API token")
			return &authError{message: "Invalid API token"}
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			retryAfter := parseRetryAfter(resp)
			c.logger.WithField("retry_after", retryAfter).Warn("Rate limit exceeded")
			return &rateLimitError{retryAfter: retryAfter}
		}
		if resp.StatusCode >= 500 {
			c.logger.WithFields(logrus.Fields{
				"status_code": resp.StatusCode,
				"body":        string(body),
			}).Error("Server error from WaniKani API")
			return &serverError{statusCode: resp.StatusCode, body: string(body)}
		}
		c.logger.WithFields(logrus.Fields{
			"status_code": resp.StatusCode,
			"body":        string(body),
		}).Error("Unexpected status code from API")
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// If we need pagination info, parse the full response
	if paginationInfo != nil {
		var fullResponse struct {
			Data  json.RawMessage `json:"data"`
			Pages struct {
				NextURL string `json:"next_url"`
			} `json:"pages"`
		}

		if err := json.Unmarshal(body, &fullResponse); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		paginationInfo.Pages.NextURL = fullResponse.Pages.NextURL

		// Parse the data array
		if err := json.Unmarshal(fullResponse.Data, data); err != nil {
			return fmt.Errorf("failed to parse data: %w", err)
		}
	} else {
		// For non-paginated responses (like statistics), parse the entire response directly
		if err := json.Unmarshal(body, data); err != nil {
			return fmt.Errorf("failed to parse data: %w", err)
		}
	}

	c.logger.WithField("url", url).Debug("API request completed successfully")
	return nil
}

// waitForRateLimit checks if we need to wait for rate limit reset and waits if necessary
func (c *Client) waitForRateLimit(ctx context.Context) error {
	c.mu.RLock()
	remaining := c.rateLimit.Remaining
	resetAt := c.rateLimit.ResetAt
	c.mu.RUnlock()

	// If we have remaining quota or no rate limit info yet, proceed
	if remaining > 0 || resetAt.IsZero() {
		return nil
	}

	// Calculate wait time until rate limit resets
	waitDuration := time.Until(resetAt)
	if waitDuration <= 0 {
		// Rate limit has already reset, proceed
		return nil
	}

	c.logger.WithFields(logrus.Fields{
		"wait_duration": waitDuration,
		"reset_at":      resetAt,
	}).Info("Rate limit quota exhausted, waiting for reset")

	// Wait until rate limit resets or context is cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitDuration):
		c.logger.Info("Rate limit reset, resuming requests")
		return nil
	}
}

// updateRateLimitInfo updates the rate limit information from response headers
func (c *Client) updateRateLimitInfo(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	oldRemaining := c.rateLimit.Remaining

	// Try different possible header names for rate limiting
	// WaniKani uses X-RateLimit-* headers
	if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
		fmt.Sscanf(remaining, "%d", &c.rateLimit.Remaining)
	} else if remaining := resp.Header.Get("RateLimit-Remaining"); remaining != "" {
		fmt.Sscanf(remaining, "%d", &c.rateLimit.Remaining)
	}

	if reset := resp.Header.Get("X-RateLimit-Reset"); reset != "" {
		var timestamp int64
		fmt.Sscanf(reset, "%d", &timestamp)
		c.rateLimit.ResetAt = time.Unix(timestamp, 0)
	} else if reset := resp.Header.Get("RateLimit-Reset"); reset != "" {
		var timestamp int64
		fmt.Sscanf(reset, "%d", &timestamp)
		c.rateLimit.ResetAt = time.Unix(timestamp, 0)
	}

	// Log rate limit updates if they changed significantly
	if oldRemaining != c.rateLimit.Remaining {
		c.logger.WithFields(logrus.Fields{
			"remaining": c.rateLimit.Remaining,
			"reset_at":  c.rateLimit.ResetAt,
		}).Debug("Rate limit status updated")
	}
}

// parseRetryAfter parses the Retry-After header
func parseRetryAfter(resp *http.Response) time.Duration {
	retryAfter := resp.Header.Get("Retry-After")
	if retryAfter == "" {
		return 60 * time.Second // Default to 60 seconds
	}

	// Try parsing as seconds
	var seconds int
	if _, err := fmt.Sscanf(retryAfter, "%d", &seconds); err == nil {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP date
	if t, err := time.Parse(time.RFC1123, retryAfter); err == nil {
		return time.Until(t)
	}

	return 60 * time.Second
}

// isRetryableError determines if an error should trigger a retry
func isRetryableError(err error) bool {
	switch err.(type) {
	case *networkError, *serverError, *rateLimitError:
		return true
	default:
		return false
	}
}

// paginatedResponse holds pagination information
type paginatedResponse struct {
	Pages struct {
		NextURL string `json:"next_url"`
	} `json:"pages"`
}

// Error types
type networkError struct {
	err error
}

func (e *networkError) Error() string {
	return fmt.Sprintf("network error: %v", e.err)
}

func (e *networkError) Unwrap() error {
	return e.err
}

type authError struct {
	message string
}

func (e *authError) Error() string {
	return e.message
}

type rateLimitError struct {
	retryAfter time.Duration
}

func (e *rateLimitError) Error() string {
	return fmt.Sprintf("rate limit exceeded, retry after %v", e.retryAfter)
}

type serverError struct {
	statusCode int
	body       string
}

func (e *serverError) Error() string {
	return fmt.Sprintf("server error %d: %s", e.statusCode, e.body)
}
