// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package ms365 provides Microsoft 365 (Microsoft Graph API) scanning
// capabilities for security policy evaluation.
package ms365

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/ms365/resources"
	"github.com/kopexa-grc/kspec/provider/registry"
)

// Provider implements the core.Provider interface for Microsoft 365.
type Provider struct{}

// NewProvider creates a new MS365 provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "ms365"
}

// Connection represents an active connection to Microsoft 365 Graph API.
type Connection struct {
	client *Client
}

// Connect establishes a connection to Microsoft 365 Graph API.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	var credential azcore.TokenCredential
	var err error

	// Get tenant ID from config
	tenantID := config["tenant_id"]
	clientID := config["client_id"]
	clientSecret := config["client_secret"]

	if clientID != "" && clientSecret != "" && tenantID != "" {
		// Service Principal authentication with client secret
		credential, err = azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create client secret credential: %w", err)
		}
	} else {
		// Use default Azure credential (tries az cli, managed identity, env vars, etc.)
		credential, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create default credential: %w", err)
		}
	}

	// Create Graph client
	graphClient, err := msgraphsdk.NewGraphServiceClientWithCredentials(credential, []string{"https://graph.microsoft.com/.default"})
	if err != nil {
		return nil, fmt.Errorf("failed to create Graph client: %w", err)
	}

	// Get rate limiter from provider definition
	def, ok := registry.Get("ms365")
	if !ok {
		return nil, fmt.Errorf("ms365 provider not registered")
	}

	// Create client with rate limiting
	client := NewClient(ClientConfig{
		Graph:   graphClient,
		Limiter: def.NewRateLimiter(),
	})

	return &Connection{
		client: client,
	}, nil
}

// Resources returns all available Microsoft 365 resources.
func (c *Connection) Resources() []core.ResourceSpec {
	graph := c.client.Graph()
	return []core.ResourceSpec{
		// Tenant and organization
		resources.NewTenant(graph),

		// Users and groups
		resources.NewUser(graph),
		resources.NewGroup(graph),

		// Applications and service principals
		resources.NewApplication(graph),
		resources.NewServicePrincipal(graph),

		// Devices
		resources.NewDevice(graph),
		resources.NewManagedDevice(graph),
		resources.NewDeviceConfiguration(graph),
		resources.NewDeviceCompliancePolicy(graph),

		// Domains
		resources.NewDomain(graph),

		// Identity and access
		resources.NewConditionalAccessPolicy(graph),
		resources.NewNamedLocation(graph),

		// Directory roles
		resources.NewDirectoryRole(graph),

		// Policies
		resources.NewAuthorizationPolicy(graph),
		resources.NewAuthenticationMethodPolicy(graph),

		// Security
		resources.NewRiskyUser(graph),
		resources.NewSecureScore(graph),

		// Teams
		resources.NewTeam(graph),

		// Settings
		resources.NewDirectorySetting(graph),
	}
}

// Client returns the underlying MS365 client for direct access if needed.
func (c *Connection) Client() *Client {
	return c.client
}
