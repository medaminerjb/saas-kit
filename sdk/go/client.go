package saaskit

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// Config holds the configuration for the SaaSKit client.
type Config struct {
	// BaseURL is the base URL of the SaaSKit API.
	BaseURL string
	// APIKey is the optional API key for authentication.
	APIKey string
	// HTTPClient is the custom HTTP client to use. If nil, a default client is used.
	HTTPClient *http.Client
	// Timeout is the timeout for HTTP requests. If zero, a default timeout is used.
	Timeout time.Duration
	// MaxRetries is the maximum number of retries for failed requests.
	MaxRetries int
	// RetryDelay is the initial delay between retries.
	RetryDelay time.Duration
}

// Client is the SaaSKit API client.
type Client struct {
	config     *Config
	httpClient *http.Client

	// API clients
	Auth    *AuthClient
	Users   *UsersClient
	Tenants *TenantsClient
	Metadata *MetadataClient
}

// NewClient creates a new SaaSKit client with the given configuration.
func NewClient(cfg *Config) (*Client, error) {
	if cfg == nil {
		cfg = &Config{}
	}

	if cfg.BaseURL == "" {
		cfg.BaseURL = "http://localhost:8080"
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	if cfg.RetryDelay == 0 {
		cfg.RetryDelay = 1 * time.Second
	}

	httpClient := cfg.HTTPClient
	if httpClient == nil {
		httpClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	client := &Client{
		config:     cfg,
		httpClient: httpClient,
	}

	// Initialize API clients
	client.Auth = NewAuthClient(client)
	client.Users = NewUsersClient(client)
	client.Tenants = NewTenantsClient(client)
	client.Metadata = NewMetadataClient(client)

	return client, nil
}

// BaseURL returns the configured base URL.
func (c *Client) BaseURL() string {
	return c.config.BaseURL
}

// HTTPClient returns the HTTP client.
func (c *Client) HTTPClient() *http.Client {
	return c.httpClient
}

// Do executes an HTTP request with retry logic.
func (c *Client) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	var lastErr error
	
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(c.backoff(attempt)):
			}
		}

		resp, err := c.httpClient.Do(req.WithContext(ctx))
		if err != nil {
			lastErr = err
			continue
		}

		// Don't retry on success or client errors (4xx)
		if resp.StatusCode < 500 {
			return resp, nil
		}

		resp.Body.Close()
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil, lastErr
}

// backoff calculates the delay for retry attempts using exponential backoff.
func (c *Client) backoff(attempt int) time.Duration {
	// Exponential backoff: delay * 2^(attempt-1)
	delay := c.config.RetryDelay * time.Duration(1<<(attempt-1))
	
	// Cap the delay at 30 seconds
	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	
	return delay
}
