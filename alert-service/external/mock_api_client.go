package external

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

// ExternalAlert represents an alert from the third-party API
type ExternalAlert struct {
	Source      string    `json:"source"`
	Severity    string    `json:"severity"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// ExternalAlertsResponse is the response from the mock API
type ExternalAlertsResponse struct {
	Alerts []ExternalAlert `json:"alerts"`
}

// MockAPIClient handles communication with the mock alerts API
type MockAPIClient struct {
	baseURL string
	client  *retryablehttp.Client
}

// NewMockAPIClient creates a new mock API client with retry support
func NewMockAPIClient(baseURL string) *MockAPIClient {
	retryClient := retryablehttp.NewClient()
	retryClient.RetryMax = 3
	retryClient.RetryWaitMin = 1 * time.Second
	retryClient.RetryWaitMax = 30 * time.Second
	retryClient.Logger = &RetryLogger{}

	retryClient.CheckRetry = func(ctx context.Context, resp *http.Response, err error) (bool, error) {
		if ctx.Err() != nil {
			return false, ctx.Err()
		}

		if err != nil {
			fmt.Printf("[RETRY] Connection error, will retry: %v\n", err)
			return true, nil
		}

		if resp.StatusCode >= 500 {
			fmt.Printf("[RETRY] Server error %d, will retry\n", resp.StatusCode)
			return true, nil
		}

		if resp.StatusCode == 429 {
			fmt.Printf("[RETRY] Rate limited, will retry\n")
			return true, nil
		}

		return false, nil
	}

	return &MockAPIClient{
		baseURL: baseURL,
		client:  retryClient,
	}
}

// RetryLogger logs retry attempts
type RetryLogger struct{}

func (l *RetryLogger) Printf(format string, args ...interface{}) {
	fmt.Printf("[RETRY] "+format+"\n", args...)
}

// CheckHealth verifies the mock API is available
func (c *MockAPIClient) CheckHealth(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	return nil
}

// FetchAlertsSince fetches alerts from the mock API since the given time
func (c *MockAPIClient) FetchAlertsSince(ctx context.Context, since time.Time) ([]ExternalAlert, error) {
	url := fmt.Sprintf("%s/alerts?since=%s", c.baseURL, since.Format(time.RFC3339))
	return c.fetchAlerts(ctx, url)
}

// FetchAllAlerts fetches all alerts from the mock API
func (c *MockAPIClient) FetchAllAlerts(ctx context.Context) ([]ExternalAlert, error) {
	url := fmt.Sprintf("%s/alerts", c.baseURL)
	return c.fetchAlerts(ctx, url)
}

// fetchAlerts is a helper that performs the actual HTTP request
func (c *MockAPIClient) fetchAlerts(ctx context.Context, url string) ([]ExternalAlert, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error calling mock API after retries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("mock API returned status %d: %s", resp.StatusCode, string(body))
	}

	var response ExternalAlertsResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return response.Alerts, nil
}
