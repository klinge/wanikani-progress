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
	mu         sync.RWMutex // protects apiToken and rateLimitInfo
	rateLimit  domain.RateLimitInfo
}

// NewClient creates a new WaniKani API client
func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetAPIToken sets the API token for authentication
func (c *Client) SetAPIToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.apiToken = token
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
	}

	var allSubjects []domain.Subject
	nextURL := fmt.Sprintf("%s/subjects?%s", baseURL, params.Encode())

	for nextURL != "" {
		var response paginatedResponse
		var subjects []domain.Subject

		err := c.fetchWithRetry(ctx, nextURL, &response, &subjects)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch subjects: %w", err)
		}

		allSubjects = append(allSubjects, subjects...)
		nextURL = response.Pages.NextURL
	}

	return allSubjects, nil
}

// FetchAssignments retrieves assignments from the WaniKani API
func (c *Client) FetchAssignments(ctx context.Context, updatedAfter *time.Time) ([]domain.Assignment, error) {
	params := url.Values{}
	if updatedAfter != nil {
		params.Set("updated_after", updatedAfter.Format(time.RFC3339))
	}

	var allAssignments []domain.Assignment
	nextURL := fmt.Sprintf("%s/assignments?%s", baseURL, params.Encode())

	for nextURL != "" {
		var response paginatedResponse
		var assignments []domain.Assignment

		err := c.fetchWithRetry(ctx, nextURL, &response, &assignments)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch assignments: %w", err)
		}

		allAssignments = append(allAssignments, assignments...)
		nextURL = response.Pages.NextURL
	}

	return allAssignments, nil
}

// FetchReviews retrieves reviews from the WaniKani API
func (c *Client) FetchReviews(ctx context.Context, updatedAfter *time.Time) ([]domain.Review, error) {
	params := url.Values{}
	if updatedAfter != nil {
		params.Set("updated_after", updatedAfter.Format(time.RFC3339))
	}

	var allReviews []domain.Review
	nextURL := fmt.Sprintf("%s/reviews?%s", baseURL, params.Encode())

	for nextURL != "" {
		var response paginatedResponse
		var reviews []domain.Review

		err := c.fetchWithRetry(ctx, nextURL, &response, &reviews)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch reviews: %w", err)
		}

		allReviews = append(allReviews, reviews...)
		nextURL = response.Pages.NextURL
	}

	return allReviews, nil
}

// FetchStatistics retrieves the current statistics snapshot from the WaniKani API
func (c *Client) FetchStatistics(ctx context.Context) (*domain.Statistics, error) {
	endpoint := fmt.Sprintf("%s/summary", baseURL)

	var stats domain.Statistics
	err := c.fetchWithRetry(ctx, endpoint, nil, &stats)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch statistics: %w", err)
	}

	return &stats, nil
}

// fetchWithRetry performs an HTTP request with retry logic and exponential backoff
func (c *Client) fetchWithRetry(ctx context.Context, url string, paginationInfo *paginatedResponse, data interface{}) error {
	var lastErr error
	backoff := initialBackoff

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
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
			return err
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// doRequest performs a single HTTP request
func (c *Client) doRequest(ctx context.Context, url string, paginationInfo *paginatedResponse, data interface{}) error {
	c.mu.RLock()
	token := c.apiToken
	c.mu.RUnlock()

	if token == "" {
		return fmt.Errorf("API token not set")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Wanikani-Revision", "20170710")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return &networkError{err: err}
	}
	defer resp.Body.Close()

	// Update rate limit information
	c.updateRateLimitInfo(resp)

	// Handle HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		if resp.StatusCode == http.StatusUnauthorized {
			return &authError{message: "Invalid API token"}
		}
		if resp.StatusCode == http.StatusTooManyRequests {
			return &rateLimitError{retryAfter: parseRetryAfter(resp)}
		}
		if resp.StatusCode >= 500 {
			return &serverError{statusCode: resp.StatusCode, body: string(body)}
		}
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
		// For non-paginated responses (like statistics), parse directly
		var wrapper struct {
			Data json.RawMessage `json:"data"`
		}
		if err := json.Unmarshal(body, &wrapper); err != nil {
			return fmt.Errorf("failed to parse response: %w", err)
		}

		if err := json.Unmarshal(wrapper.Data, data); err != nil {
			return fmt.Errorf("failed to parse data: %w", err)
		}
	}

	return nil
}

// updateRateLimitInfo updates the rate limit information from response headers
func (c *Client) updateRateLimitInfo(resp *http.Response) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if remaining := resp.Header.Get("RateLimit-Remaining"); remaining != "" {
		fmt.Sscanf(remaining, "%d", &c.rateLimit.Remaining)
	}

	if reset := resp.Header.Get("RateLimit-Reset"); reset != "" {
		var timestamp int64
		fmt.Sscanf(reset, "%d", &timestamp)
		c.rateLimit.ResetAt = time.Unix(timestamp, 0)
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
