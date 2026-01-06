package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/keyvault/armkeyvault"

	"github.com/kopexa-grc/kspec/core"
)

type KeyVaultResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func (r *KeyVaultResource) Name() string {
	return "azure_keyvault_vault"
}

func (r *KeyVaultResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Create Key Vault client
	client, err := armkeyvault.NewVaultsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create key vaults client: %w", err)
	}

	var resources []core.Resource

	// List all key vaults in the subscription
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list key vaults: %w", err)
		}

		for _, vault := range page.Value {
			// Convert to map[string]interface{}
			data, err := json.Marshal(vault)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			// Fetch keys and secrets for the vault
			if vault.Name != nil {
				// Get resource group from asset config
				resourceGroup := asset.Config["resource_group"]
				if resourceGroup != "" {
					// Note: For keys and secrets, we'd need to use Azure Key Vault SDK (not ARM SDK)
					// This requires the vault URI and proper permissions
					resourceMap["keys"] = []interface{}{}
					resourceMap["secrets"] = []interface{}{}
					resourceMap["diagnosticSettings"] = []interface{}{}
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
