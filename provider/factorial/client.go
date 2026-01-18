// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package factorial

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/kopexa-grc/kspec/pkg/ratelimit"
)

const (
	defaultBaseURL    = "https://api.factorialhr.com"
	defaultAPIVersion = "2025-07-01"
)

// Client wraps the Factorial HR REST API client with rate limiting support.
type Client struct {
	httpClient  *http.Client
	limiter     *ratelimit.Limiter
	apiKey      string
	accessToken string
	baseURL     string
	apiVersion  string
}

// ClientConfig holds configuration for creating a Factorial client.
type ClientConfig struct {
	// APIKey is the Factorial HR API key.
	APIKey string
	// AccessToken is the OAuth access token.
	AccessToken string
	// BaseURL is the API base URL (defaults to https://api.factorialhr.com).
	BaseURL string
	// APIVersion is the API version to use (defaults to 2025-07-01).
	APIVersion string
	// Limiter is an optional rate limiter. If nil, no rate limiting is applied.
	Limiter *ratelimit.Limiter
}

// NewClient creates a new Factorial HR client with the given configuration.
func NewClient(cfg ClientConfig) *Client {
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	apiVersion := cfg.APIVersion
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}

	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		limiter:     cfg.Limiter,
		apiKey:      cfg.APIKey,
		accessToken: cfg.AccessToken,
		baseURL:     baseURL,
		apiVersion:  apiVersion,
	}
}

// Do executes a function with rate limiting.
// If no rate limiter is configured, the function is executed immediately.
func (c *Client) Do(ctx context.Context, fn func() error) error {
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}
	}
	return fn()
}

// DoRequest performs an HTTP request to the Factorial API.
func (c *Client) DoRequest(ctx context.Context, method, path string) ([]byte, error) {
	var body []byte
	err := c.Do(ctx, func() error {
		url := fmt.Sprintf("%s/api/%s/%s", c.baseURL, c.apiVersion, path)

		req, err := http.NewRequestWithContext(ctx, method, url, http.NoBody)
		if err != nil {
			return fmt.Errorf("factorial: failed to create request: %w", err)
		}

		if c.apiKey != "" {
			req.Header.Set("x-api-key", c.apiKey)
		} else if c.accessToken != "" {
			req.Header.Set("Authorization", "Bearer "+c.accessToken)
		}
		req.Header.Set("Accept", "application/json")
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("factorial: request failed: %w", err)
		}
		defer resp.Body.Close()

		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("factorial: failed to read response: %w", err)
		}

		if resp.StatusCode >= 400 {
			return fmt.Errorf("factorial: API error %d: %s", resp.StatusCode, string(body))
		}

		return nil
	})

	return body, err
}

// FetchPaginated handles pagination for list endpoints.
func (c *Client) FetchPaginated(ctx context.Context, path string) ([]map[string]any, error) {
	var allResults []map[string]any
	page := 1
	perPage := 100

	for {
		paginatedPath := fmt.Sprintf("%s?page=%d&per_page=%d", path, page, perPage)
		body, err := c.DoRequest(ctx, "GET", paginatedPath)
		if err != nil {
			return nil, err
		}

		var results []map[string]any
		if err := json.Unmarshal(body, &results); err != nil {
			// Try single object response (some endpoints return object instead of array)
			var singleResult map[string]any
			if err2 := json.Unmarshal(body, &singleResult); err2 == nil {
				// Check if it has a "data" key with array
				if data, ok := singleResult["data"].([]any); ok {
					for _, item := range data {
						if m, ok := item.(map[string]any); ok {
							allResults = append(allResults, m)
						}
					}
					// Check for more pages
					if meta, ok := singleResult["meta"].(map[string]any); ok {
						if total, ok := meta["total_pages"].(float64); ok && float64(page) < total {
							page++
							continue
						}
					}
					return allResults, nil
				}
				return []map[string]any{singleResult}, nil
			}
			return nil, fmt.Errorf("factorial: failed to parse response: %w", err)
		}

		if len(results) == 0 {
			break
		}

		allResults = append(allResults, results...)

		if len(results) < perPage {
			break
		}
		page++
	}

	return allResults, nil
}
