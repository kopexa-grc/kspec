// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/cosmos/armcosmos"

	"github.com/kopexa-grc/kspec/core"
)

// CosmosDBAccount represents an Azure Cosmos DB account resource scanner.
type CosmosDBAccount struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewCosmosDBAccount creates a new CosmosDBAccount resource scanner.
func NewCosmosDBAccount(credential azcore.TokenCredential, subscriptionID string) *CosmosDBAccount {
	return &CosmosDBAccount{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *CosmosDBAccount) Name() string {
	return "azure_cosmosdb_account"
}

// Fetch retrieves Cosmos DB accounts from Azure.
func (r *CosmosDBAccount) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armcosmos.NewDatabaseAccountsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Cosmos DB client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Cosmos DB accounts: %w", err)
		}

		for _, account := range page.Value {
			data, err := json.Marshal(account)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if account.Properties != nil {
				resourceMap["documentEndpoint"] = account.Properties.DocumentEndpoint
				resourceMap["databaseAccountOfferType"] = account.Properties.DatabaseAccountOfferType
				resourceMap["enableAutomaticFailover"] = account.Properties.EnableAutomaticFailover
				resourceMap["enableMultipleWriteLocations"] = account.Properties.EnableMultipleWriteLocations
				resourceMap["isVirtualNetworkFilterEnabled"] = account.Properties.IsVirtualNetworkFilterEnabled
				resourceMap["publicNetworkAccess"] = account.Properties.PublicNetworkAccess
				resourceMap["disableKeyBasedMetadataWriteAccess"] = account.Properties.DisableKeyBasedMetadataWriteAccess
				resourceMap["enableAnalyticalStorage"] = account.Properties.EnableAnalyticalStorage
				resourceMap["enableFreeTier"] = account.Properties.EnableFreeTier
				resourceMap["disableLocalAuth"] = account.Properties.DisableLocalAuth

				if account.Properties.ConsistencyPolicy != nil {
					resourceMap["defaultConsistencyLevel"] = account.Properties.ConsistencyPolicy.DefaultConsistencyLevel
				}

				if account.Properties.BackupPolicy != nil {
					resourceMap["backupPolicyType"] = account.Properties.BackupPolicy.GetBackupPolicy().Type
				}

				// IP rules
				if account.Properties.IPRules != nil {
					ipRules := make([]string, 0, len(account.Properties.IPRules))
					for _, rule := range account.Properties.IPRules {
						if rule.IPAddressOrRange != nil {
							ipRules = append(ipRules, *rule.IPAddressOrRange)
						}
					}
					resourceMap["ipRules"] = ipRules
				}

				// Capabilities
				if account.Properties.Capabilities != nil {
					caps := make([]string, 0, len(account.Properties.Capabilities))
					for _, cap := range account.Properties.Capabilities {
						if cap.Name != nil {
							caps = append(caps, *cap.Name)
						}
					}
					resourceMap["capabilities"] = caps
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
