package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/juliankoehn/kspec/core"
)

type AzureProvider struct{}

func NewAzureProvider() core.Provider {
	return &AzureProvider{}
}

func (p *AzureProvider) Name() string {
	return "azure"
}

type AzureConnection struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func (p *AzureProvider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Get subscription ID from config
	subscriptionID, ok := config["subscription_id"]
	if !ok {
		subscriptionID = "" // Will use default subscription
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

	return &AzureConnection{
		credential:     azureCred,
		subscriptionID: subscriptionID,
	}, nil
}

func (c *AzureConnection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&StorageAccountResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&SqlServerResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&MySQLServerResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&PostgreSQLServerResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&KeyVaultResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&NetworkSecurityGroupResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&VirtualMachineResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&AppServiceResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&SubscriptionResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
	}
}
