// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package resources provides Azure resource scanners for security compliance.
// Each resource type implements the core.ResourceSpec interface and uses
// the Azure SDK to fetch resource data for policy evaluation.
package resources

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"

	"github.com/kopexa-grc/kspec/core"
)

// StorageAccount represents an Azure Storage Account resource scanner.
type StorageAccount struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewStorageAccount creates a new StorageAccount resource scanner.
func NewStorageAccount(credential azcore.TokenCredential, subscriptionID string) *StorageAccount {
	return &StorageAccount{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *StorageAccount) Name() string {
	return "azure_storage_account"
}

// Fetch retrieves storage accounts from Azure.
func (r *StorageAccount) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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

// Pager represents an Azure SDK pager interface for listing resources.
type Pager[T any] interface {
	More() bool
	NextPage(ctx context.Context) (T, error)
}

// PageExtractor extracts individual items from a page result.
type PageExtractor[TPage any, TItem any] func(page TPage) []*TItem

// fetchWithPager is a generic helper for fetching Azure resources using a pager.
// It handles the common pattern of iterating through pages and converting
// each item to a core.Resource via JSON marshaling.
func fetchWithPager[TPage any, TItem any](
	ctx context.Context,
	pager Pager[TPage],
	extractor PageExtractor[TPage, TItem],
	resourceType string,
) ([]core.Resource, error) {
	var resources []core.Resource

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		items := extractor(page)
		for _, item := range items {
			resource, err := toResourceMap(item)
			if err != nil {
				continue
			}
			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// toResourceMap converts any struct to a map[string]interface{} via JSON marshaling.
func toResourceMap(item any) (map[string]interface{}, error) {
	return core.ToResource(item)
}
