// Package azure provides Azure cloud resource scanning capabilities
// for security policy evaluation.
package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/kopexa-grc/kspec/core"
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
	credential     azcore.TokenCredential
	subscriptionID string
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

	return &Connection{
		credential:     azureCred,
		subscriptionID: subscriptionID,
	}, nil
}

// Resources returns the list of available Azure resource types that can be scanned.
func (c *Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		// Storage
		&StorageAccountResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// Databases
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
		&MariaDBServerResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&CosmosDBAccountResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&RedisCacheResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// Security
		&KeyVaultResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&SecurityAssessmentResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&SecurityAlertResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&DefenderSettingResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&SecureScoreResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// Networking
		&NetworkSecurityGroupResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// Compute
		&VirtualMachineResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&AKSClusterResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// Web
		&AppServiceResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// IoT
		&IoTHubResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// IAM
		&RoleAssignmentResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&RoleDefinitionResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// Policy
		&PolicyAssignmentResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&PolicyDefinitionResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		&PolicyComplianceResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// Advisor
		&AdvisorRecommendationResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
		// Subscription
		&SubscriptionResource{
			credential:     c.credential,
			subscriptionID: c.subscriptionID,
		},
	}
}
