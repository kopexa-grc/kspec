// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package hetzner

import (
	"context"
	"fmt"
	"os"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider/hetzner/resources"
)

// Provider implements the core.Provider interface for Hetzner Cloud.
type Provider struct{}

// NewProvider creates a new Hetzner Cloud provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "hetzner"
}

// Connect establishes a connection to Hetzner Cloud API.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Get API token from config or environment
	apiToken := config["api_token"]
	if apiToken == "" {
		apiToken = os.Getenv("HCLOUD_TOKEN")
	}
	if apiToken == "" {
		apiToken = os.Getenv("HETZNER_API_TOKEN")
	}

	if apiToken == "" {
		return nil, fmt.Errorf("hetzner: no API token provided. Set api_token, HCLOUD_TOKEN, or HETZNER_API_TOKEN")
	}

	// Get optional project name for identification
	projectName := config["project"]
	if projectName == "" {
		projectName = os.Getenv("HCLOUD_PROJECT")
	}
	if projectName == "" {
		projectName = "default"
	}

	// Create client with rate limiting
	client := NewClient(ctx, ClientConfig{
		Token:   apiToken,
		Limiter: ratelimit.New("hetzner", *rateLimitConfig),
	})

	// Verify connection by fetching locations
	_, err := client.HCloud().Location.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("hetzner: failed to connect: %w", err)
	}

	return &Connection{
		client:      client,
		projectName: projectName,
	}, nil
}

// Connection represents an active connection to Hetzner Cloud.
type Connection struct {
	client      *Client
	projectName string
}

// Resources returns all available Hetzner Cloud resources.
// Note: Only project-specific resources are included. Global catalog resources
// (Location, Datacenter, ServerType, ISO, public Images) are excluded as they
// are not owned by the customer and not relevant for security scanning.
func (c *Connection) Resources() []core.ResourceSpec {
	hc := c.client.HCloud()
	return []core.ResourceSpec{
		// Project (root)
		resources.NewProject(hc),
		// Compute
		resources.NewServer(hc),
		// Storage
		resources.NewVolume(hc),
		// Networking
		resources.NewNetwork(hc),
		resources.NewFloatingIP(hc),
		resources.NewPrimaryIP(hc),
		// Security
		resources.NewFirewall(hc),
		// SSH Keys
		resources.NewSSHKey(hc),
	}
}

// EntryResourceType maps CLI asset types to internal resource type identifiers.
func (c *Connection) EntryResourceType(assetType string) string {
	switch assetType {
	case "hetzner-project":
		return "hetzner_project"
	default:
		return ""
	}
}

// Client returns the underlying Hetzner client for direct access if needed.
func (c *Connection) Client() *Client {
	return c.client
}
