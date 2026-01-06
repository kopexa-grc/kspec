package azure

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"github.com/kopexa-grc/kspec/core"
)

// StorageAccountResource represents an Azure Storage Account resource scanner.
type StorageAccountResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *StorageAccountResource) Name() string {
	return "azure_storage_account"
}

// Fetch retrieves storage accounts from Azure.
func (r *StorageAccountResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armstorage.NewAccountsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage accounts client: %w", err)
	}

	pager := client.NewListPager(nil)
	return fetchWithPager(ctx, pager, func(page armstorage.AccountsClientListResponse) []*armstorage.Account {
		return page.Value
	}, r.Name())
}
