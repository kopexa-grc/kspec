// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package atlassian provides Atlassian Cloud (Jira, Confluence, Admin) scanning
// capabilities for security policy evaluation.
package atlassian

import (
	"context"
	"fmt"
	"os"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/atlassian/resources"
	"github.com/kopexa-grc/kspec/provider/registry"
)

// Provider implements the core.Provider interface for Atlassian Cloud.
type Provider struct{}

// NewProvider creates a new Atlassian provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "atlassian"
}

// Connect establishes a connection to Atlassian Cloud APIs.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Get authentication credentials
	email := config["email"]
	if email == "" {
		email = os.Getenv("ATLASSIAN_EMAIL")
	}

	apiToken := config["api_token"]
	if apiToken == "" {
		apiToken = os.Getenv("ATLASSIAN_API_TOKEN")
	}

	site := config["site"]
	if site == "" {
		site = os.Getenv("ATLASSIAN_SITE")
	}

	if email == "" || apiToken == "" {
		return nil, fmt.Errorf("atlassian: no valid credentials provided. Set email and api_token or ATLASSIAN_EMAIL and ATLASSIAN_API_TOKEN")
	}

	if site == "" {
		return nil, fmt.Errorf("atlassian: site not specified. Set site or ATLASSIAN_SITE (e.g., yoursite.atlassian.net)")
	}

	// Get org ID from config or environment
	orgID := config["org_id"]
	if orgID == "" {
		orgID = os.Getenv("ATLASSIAN_ORG_ID")
	}

	// Get rate limiter from provider definition
	def, ok := registry.Get("atlassian")
	if !ok {
		return nil, fmt.Errorf("atlassian provider not registered")
	}

	// Create client with rate limiting
	client, err := NewClient(ctx, ClientConfig{
		Email:    email,
		APIToken: apiToken,
		Site:     site,
		OrgID:    orgID,
		Limiter:  def.NewRateLimiter(),
	})
	if err != nil {
		return nil, err
	}

	return &Connection{client: client}, nil
}

// Connection represents an active connection to Atlassian Cloud.
type Connection struct {
	client *Client
}

// Resources returns all available Atlassian resources.
func (c *Connection) Resources() []core.ResourceSpec {
	jira := c.client.Jira()
	confluence := c.client.Confluence()
	admin := c.client.Admin()
	orgID := c.client.OrgID()

	return []core.ResourceSpec{
		// Jira resources
		resources.NewJiraProject(jira),
		resources.NewJiraPermissionScheme(jira),
		resources.NewJiraSecurityScheme(jira),
		resources.NewJiraUser(jira),
		// Confluence resources
		resources.NewConfluenceSpace(confluence),
		resources.NewConfluencePage(confluence),
		// Admin resources (require org ID)
		resources.NewAdminOrganization(admin, orgID),
		resources.NewAdminUser(admin, orgID),
		resources.NewAdminGroup(admin, orgID),
		resources.NewAdminPolicy(admin, orgID),
		// SCIM resources
		resources.NewSCIMUser(admin, orgID),
		resources.NewSCIMGroup(admin, orgID),
	}
}

// EntryResourceType returns the entry point resource type for a given asset type.
func (c *Connection) EntryResourceType(assetType string) string {
	switch assetType {
	case "atlassian-site":
		return "atlassian_jira_project"
	case "atlassian-org":
		return "atlassian_organization"
	default:
		return ""
	}
}

// Client returns the underlying Atlassian client for direct access if needed.
func (c *Connection) Client() *Client {
	return c.client
}
