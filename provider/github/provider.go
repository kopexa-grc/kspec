// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package github provides GitHub repository and organization scanning
// capabilities for security policy evaluation.
package github

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/github/resources"
	"github.com/kopexa-grc/kspec/provider/registry"
)

// Provider implements the core.Provider interface for GitHub.
type Provider struct{}

// NewProvider creates a new GitHub provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "github"
}

// Connect establishes a connection to GitHub APIs.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Resolve token from config
	token, err := resolveToken(config)
	if err != nil {
		return nil, err
	}

	// Get rate limiter from provider definition
	def, ok := registry.Get("github")
	if !ok {
		return nil, fmt.Errorf("github provider not registered")
	}

	// Create client with rate limiting
	client := NewClient(ctx, ClientConfig{
		Token:   token,
		Limiter: def.NewRateLimiter(),
	})

	return &Connection{client: client}, nil
}

// resolveToken extracts the token from config using various methods.
func resolveToken(config map[string]string) (string, error) {
	// Try to parse credentials from config
	cred, err := core.ParseCredentialFromConfig(config)
	if err != nil {
		// Fall back to legacy token support if credential parsing fails
		return config["token"], nil //nolint:nilerr // Intentional fallback to legacy config
	}

	// Resolve token based on credential type
	switch cred.Type { //nolint:exhaustive // Only specific credential types are supported for GitHub
	case core.CredentialTypeBearer:
		return cred.Secret, nil
	case core.CredentialTypeEnv:
		return cred.ResolveSecret()
	case core.CredentialTypePassword:
		return cred.Secret, nil
	default:
		return "", fmt.Errorf("unsupported credential type for GitHub provider: %s", cred.Type)
	}
}

// Connection represents an active connection to GitHub APIs.
type Connection struct {
	client *Client
}

// Resources returns all available GitHub resources.
func (c *Connection) Resources() []core.ResourceSpec {
	gh := c.client.GitHub()
	return []core.ResourceSpec{
		resources.NewOrganization(gh),
		resources.NewRepository(gh),
		resources.NewTeam(gh),
		resources.NewBranch(gh),
	}
}

// Client returns the underlying GitHub client for direct access if needed.
func (c *Connection) Client() *Client {
	return c.client
}
