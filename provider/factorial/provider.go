// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package factorial provides a kspec provider for Factorial HR.
package factorial

import (
	"context"
	"fmt"
	"os"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/factorial/resources"
	"github.com/kopexa-grc/kspec/provider/registry"
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

	apiVersion := config["api_version"]
	if apiVersion == "" {
		apiVersion = os.Getenv("FACTORIAL_API_VERSION")
	}

	// Get rate limiter from provider definition
	def, ok := registry.Get("factorial")
	if !ok {
		return nil, fmt.Errorf("factorial provider not registered")
	}

	// Create client with rate limiting
	client := NewClient(ClientConfig{
		APIKey:      apiKey,
		AccessToken: accessToken,
		BaseURL:     baseURL,
		APIVersion:  apiVersion,
		Limiter:     def.NewRateLimiter(),
	})

	// Verify connection by fetching company info or employees
	_, err := client.DoRequest(ctx, "GET", "resources/employees/employees")
	if err != nil {
		return nil, fmt.Errorf("factorial: failed to connect: %w", err)
	}

	return &Connection{client: client}, nil
}

// Connection represents an active connection to Factorial HR.
type Connection struct {
	client *Client
}

// Resources returns all available Factorial HR resources.
func (c *Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		// Core HR
		resources.NewEmployee(c.client),
		resources.NewContract(c.client),
		// Organization
		resources.NewTeam(c.client),
		resources.NewTeamMembership(c.client),
		resources.NewLocation(c.client),
		// Time & Attendance
		resources.NewShift(c.client),
		resources.NewLeave(c.client),
		resources.NewAllowance(c.client),
		// Documents
		resources.NewDocument(c.client),
		resources.NewFolder(c.client),
	}
}

// Client returns the underlying Factorial client for direct access if needed.
func (c *Connection) Client() *Client {
	return c.client
}
