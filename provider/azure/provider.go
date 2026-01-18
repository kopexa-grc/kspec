// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package azure provides Azure cloud resource scanning capabilities
// for security policy evaluation.
package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/azure/resources"
	"github.com/kopexa-grc/kspec/provider/registry"
)

// Provider implements the core.Provider interface for Azure cloud resources.
type Provider struct{}

// NewProvider creates a new Azure provider instance for scanning Azure resources.
func NewProvider() core.Provider {
	return &Provider{}
}

// Name returns the provider name identifier.
func (p *Provider) Name() string {
	return "azure"
}

// Connection represents an authenticated connection to Azure services.
type Connection struct {
	client *Client
}

// Connect establishes a connection to Azure using the provided configuration.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Get subscription ID from config (support both formats)
	subscriptionID, ok := config["subscription_id"]
	if !ok {
		subscriptionID, ok = config["subscription-id"]
		if !ok {
			subscriptionID = "" // Will use default subscription
		}
	}

	var azureCred azcore.TokenCredential
	var err error

	// Check if we have client_id and tenant_id for Service Principal auth
	clientID, hasClientID := config["client_id"]
	tenantID, hasTenantID := config["tenant_id"]

	if hasClientID && hasTenantID {
		// Service Principal authentication
		// Parse credentials from config
		credential, err := core.ParseCredentialFromConfig(config)
		if err != nil {
			return nil, fmt.Errorf("failed to parse credential: %w", err)
		}

		// Resolve the actual secret value
		secret, err := credential.ResolveSecret()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve credential: %w", err)
		}

		azureCred, err = azidentity.NewClientSecretCredential(tenantID, clientID, secret, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create client secret credential: %w", err)
		}
	} else {
		// Use default Azure credential (tries az cli, managed identity, env vars, etc.)
		// This does NOT require any env var to be set
		azureCred, err = azidentity.NewDefaultAzureCredential(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create default Azure credential: %w", err)
		}
	}

	// Get rate limiter from provider definition
	def, ok := registry.Get("azure")
	if !ok {
		return nil, fmt.Errorf("azure provider not registered")
	}

	// Create client with rate limiting
	client := NewClient(ClientConfig{
		Credential:     azureCred,
		SubscriptionID: subscriptionID,
		Limiter:        def.NewRateLimiter(),
	})

	return &Connection{
		client: client,
	}, nil
}

// Resources returns the list of available Azure resource types that can be scanned.
func (c *Connection) Resources() []core.ResourceSpec {
	cred := c.client.Credential()
	subID := c.client.SubscriptionID()
	return []core.ResourceSpec{
		// Storage
		resources.NewStorageAccount(cred, subID),
		// Databases
		resources.NewSqlServer(cred, subID),
		resources.NewMySQLServer(cred, subID),
		resources.NewPostgreSQLServer(cred, subID),
		resources.NewMariaDBServer(cred, subID),
		resources.NewCosmosDBAccount(cred, subID),
		resources.NewRedisCache(cred, subID),
		// Security
		resources.NewKeyVault(cred, subID),
		resources.NewSecurityAssessment(cred, subID),
		resources.NewSecurityAlert(cred, subID),
		resources.NewDefenderSetting(cred, subID),
		resources.NewSecureScore(cred, subID),
		// Networking
		resources.NewNetworkSecurityGroup(cred, subID),
		// Compute
		resources.NewVirtualMachine(cred, subID),
		resources.NewAKSCluster(cred, subID),
		// Web
		resources.NewAppService(cred, subID),
		// IoT
		resources.NewIoTHub(cred, subID),
		// IAM
		resources.NewRoleAssignment(cred, subID),
		resources.NewRoleDefinition(cred, subID),
		// Policy
		resources.NewPolicyAssignment(cred, subID),
		resources.NewPolicyDefinition(cred, subID),
		resources.NewPolicyCompliance(cred, subID),
		// Advisor
		resources.NewAdvisorRecommendation(cred, subID),
		// Subscription
		resources.NewSubscription(cred, subID),
	}
}

// Client returns the underlying Azure client for direct access if needed.
func (c *Connection) Client() *Client {
	return c.client
}
