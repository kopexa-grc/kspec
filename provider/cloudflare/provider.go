// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package cloudflare provides Cloudflare resource scanning capabilities
// for security policy evaluation.
package cloudflare

import (
	"context"
	"fmt"
	"os"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/option"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/cloudflare/resources"
)

// Provider implements the core.Provider interface for Cloudflare.
type Provider struct{}

// NewProvider creates a new Cloudflare provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "cloudflare"
}

// Connect establishes a connection to Cloudflare API.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	var opts []option.RequestOption

	// Try API Token first (preferred method)
	apiToken := config["api_token"]
	if apiToken == "" {
		apiToken = os.Getenv("CLOUDFLARE_API_TOKEN")
	}

	if apiToken != "" {
		opts = append(opts, option.WithAPIToken(apiToken))
	} else {
		// Fall back to API Key + Email (legacy)
		apiKey := config["api_key"]
		if apiKey == "" {
			apiKey = os.Getenv("CLOUDFLARE_API_KEY")
		}

		email := config["email"]
		if email == "" {
			email = os.Getenv("CLOUDFLARE_EMAIL")
		}

		if apiKey != "" && email != "" {
			opts = append(opts, option.WithAPIKey(apiKey), option.WithAPIEmail(email))
		} else {
			return nil, fmt.Errorf("cloudflare: no valid credentials provided. Set api_token or (api_key + email)")
		}
	}

	cfClient := cloudflare.NewClient(opts...)

	// Get account ID from config or discover it
	accountID := config["account_id"]
	if accountID == "" {
		accountID = os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	}

	// Create client wrapper (no rate limiter needed - SDK handles it internally)
	client := NewClient(ClientConfig{
		CF:        cfClient,
		AccountID: accountID,
	})

	return &Connection{
		client:    client,
		accountID: accountID,
	}, nil
}

// Connection represents an active connection to Cloudflare.
type Connection struct {
	client    *Client
	accountID string
}

// Resources returns all available Cloudflare resources.
func (c *Connection) Resources() []core.ResourceSpec {
	cf := c.client.Cloudflare()
	return []core.ResourceSpec{
		// Account resources
		resources.NewAccount(cf),
		// Zone resources
		resources.NewZone(cf, c.accountID),
		resources.NewZoneSettings(cf),
		resources.NewDNSRecord(cf),
		// Security resources
		resources.NewWAFRule(cf),
		resources.NewFirewallRule(cf),
		// Platform resources
		resources.NewR2Bucket(cf, c.accountID),
		resources.NewWorker(cf, c.accountID),
		resources.NewPagesProject(cf, c.accountID),
		resources.NewTunnel(cf, c.accountID),
		// Zero Trust resources
		resources.NewAccessApplication(cf, c.accountID),
		resources.NewAccessPolicy(cf, c.accountID),
	}
}

// Client returns the underlying Cloudflare client for direct access if needed.
func (c *Connection) Client() *Client {
	return c.client
}
