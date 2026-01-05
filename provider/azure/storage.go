package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"github.com/kopexa-grc/kspec/core"
)

type StorageAccountResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func (r *StorageAccountResource) Name() string {
	return "azure_storage_account"
}

func (r *StorageAccountResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Create storage accounts client
	client, err := armstorage.NewAccountsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	var resources []core.Resource

	// List all storage accounts in the subscription
	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list storage accounts: %w", err)
		}

		for _, account := range page.Value {
			// Convert to map[string]interface{}
			data, err := json.Marshal(account)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
