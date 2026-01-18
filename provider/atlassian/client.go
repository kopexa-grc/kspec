// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package atlassian

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ctreminiom/go-atlassian/admin"
	"github.com/ctreminiom/go-atlassian/confluence"
	v3 "github.com/ctreminiom/go-atlassian/jira/v3"

	"github.com/kopexa-grc/kspec/pkg/ratelimit"
)

// Client wraps the Atlassian API clients with rate limiting support.
type Client struct {
	jira       *v3.Client
	confluence *confluence.Client
	admin      *admin.Client
	limiter    *ratelimit.Limiter
	site       string
	cloudID    string
	orgID      string
}

// ClientConfig holds configuration for creating an Atlassian client.
type ClientConfig struct {
	// Email is the Atlassian account email for authentication.
	Email string

	// APIToken is the Atlassian API token.
	APIToken string

	// Site is the Atlassian site (e.g., yoursite.atlassian.net).
	Site string

	// OrgID is the Atlassian organization ID (for admin APIs).
	OrgID string

	// Limiter is an optional rate limiter. If nil, no rate limiting is applied.
	Limiter *ratelimit.Limiter
}

// NewClient creates a new Atlassian client with the given configuration.
func NewClient(ctx context.Context, cfg ClientConfig) (*Client, error) {
	// Create Jira client
	jiraClient, err := v3.New(nil, cfg.Site)
	if err != nil {
		return nil, fmt.Errorf("atlassian: failed to create Jira client: %w", err)
	}
	jiraClient.Auth.SetBasicAuth(cfg.Email, cfg.APIToken)

	// Create Confluence client
	confluenceClient, err := confluence.New(nil, cfg.Site)
	if err != nil {
		return nil, fmt.Errorf("atlassian: failed to create Confluence client: %w", err)
	}
	confluenceClient.Auth.SetBasicAuth(cfg.Email, cfg.APIToken)

	// Create Admin client (uses different base URL)
	adminClient, err := admin.New(nil)
	if err != nil {
		return nil, fmt.Errorf("atlassian: failed to create Admin client: %w", err)
	}
	adminClient.Auth.SetBearerToken(cfg.APIToken)

	// Discover Cloud ID from site
	cloudID, err := discoverCloudID(cfg.Site, cfg.Email, cfg.APIToken)
	if err != nil {
		// Cloud ID is optional for some operations
		cloudID = ""
	}

	return &Client{
		jira:       jiraClient,
		confluence: confluenceClient,
		admin:      adminClient,
		limiter:    cfg.Limiter,
		site:       cfg.Site,
		cloudID:    cloudID,
		orgID:      cfg.OrgID,
	}, nil
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

// Jira returns the underlying Jira client.
func (c *Client) Jira() *v3.Client {
	return c.jira
}

// Confluence returns the underlying Confluence client.
func (c *Client) Confluence() *confluence.Client {
	return c.confluence
}

// Admin returns the underlying Admin client.
func (c *Client) Admin() *admin.Client {
	return c.admin
}

// OrgID returns the organization ID.
func (c *Client) OrgID() string {
	return c.orgID
}

// Site returns the site name.
func (c *Client) Site() string {
	return c.site
}

// CloudID returns the cloud ID.
func (c *Client) CloudID() string {
	return c.cloudID
}

// discoverCloudID fetches the Cloud ID from the Atlassian site.
func discoverCloudID(site, email, apiToken string) (string, error) {
	url := fmt.Sprintf("https://%s/_edge/tenant_info", site)

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth(email, apiToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to get tenant info: %s", resp.Status)
	}

	var info struct {
		CloudID string `json:"cloudId"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}

	return info.CloudID, nil
}
