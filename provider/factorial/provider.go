// Package factorial provides a kspec provider for Factorial HR.
package factorial

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/kopexa-grc/kspec/core"
)

const (
	defaultBaseURL    = "https://api.factorialhr.com"
	defaultAPIVersion = "2025-07-01"
)

// Provider implements the core.Provider interface for Factorial HR.
type Provider struct{}

// NewProvider creates a new Factorial HR provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "factorial"
}

// Connect establishes a connection to Factorial HR API.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	apiKey := config["api_key"]
	if apiKey == "" {
		apiKey = os.Getenv("FACTORIAL_API_KEY")
	}

	accessToken := config["access_token"]
	if accessToken == "" {
		accessToken = os.Getenv("FACTORIAL_ACCESS_TOKEN")
	}

	if apiKey == "" && accessToken == "" {
		return nil, fmt.Errorf("factorial: no credentials provided. Set api_key/FACTORIAL_API_KEY or access_token/FACTORIAL_ACCESS_TOKEN")
	}

	baseURL := config["base_url"]
	if baseURL == "" {
		baseURL = os.Getenv("FACTORIAL_BASE_URL")
	}
	if baseURL == "" {
		baseURL = defaultBaseURL
	}

	apiVersion := config["api_version"]
	if apiVersion == "" {
		apiVersion = os.Getenv("FACTORIAL_API_VERSION")
	}
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}

	conn := &Connection{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		apiKey:      apiKey,
		accessToken: accessToken,
		baseURL:     baseURL,
		apiVersion:  apiVersion,
	}

	// Verify connection by fetching company info or employees
	_, err := conn.doRequest(ctx, "GET", "resources/employees/employees")
	if err != nil {
		return nil, fmt.Errorf("factorial: failed to connect: %w", err)
	}

	return conn, nil
}

// Connection represents an active connection to Factorial HR.
type Connection struct {
	client      *http.Client
	apiKey      string
	accessToken string
	baseURL     string
	apiVersion  string
}

// Resources returns all available Factorial HR resources.
func (c *Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		// Core HR
		&EmployeeResource{conn: c},
		&ContractResource{conn: c},
		// Organization
		&TeamResource{conn: c},
		&TeamMembershipResource{conn: c},
		&LocationResource{conn: c},
		// Time & Attendance
		&ShiftResource{conn: c},
		&LeaveResource{conn: c},
		&AllowanceResource{conn: c},
		// Documents
		&DocumentResource{conn: c},
		&FolderResource{conn: c},
	}
}

// doRequest performs an HTTP request to the Factorial API.
func (c *Connection) doRequest(ctx context.Context, method, path string) ([]byte, error) {
	url := fmt.Sprintf("%s/api/%s/%s", c.baseURL, c.apiVersion, path)

	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("factorial: failed to create request: %w", err)
	}

	if c.apiKey != "" {
		req.Header.Set("x-api-key", c.apiKey)
	} else if c.accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("factorial: request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("factorial: failed to read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("factorial: API error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// fetchPaginated handles pagination for list endpoints.
func (c *Connection) fetchPaginated(ctx context.Context, path string) ([]map[string]any, error) {
	var allResults []map[string]any
	page := 1
	perPage := 100

	for {
		paginatedPath := fmt.Sprintf("%s?page=%d&per_page=%d", path, page, perPage)
		body, err := c.doRequest(ctx, "GET", paginatedPath)
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
