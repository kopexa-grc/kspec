// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mariadb/armmariadb"

	"github.com/kopexa-grc/kspec/core"
)

// MariaDBServer represents an Azure MariaDB server resource scanner.
type MariaDBServer struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewMariaDBServer creates a new MariaDBServer resource scanner.
func NewMariaDBServer(credential azcore.TokenCredential, subscriptionID string) *MariaDBServer {
	return &MariaDBServer{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *MariaDBServer) Name() string {
	return "azure_mariadb_server"
}

// Fetch retrieves MariaDB servers from Azure.
func (r *MariaDBServer) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armmariadb.NewServersClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create MariaDB client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list MariaDB servers: %w", err)
		}

		for _, server := range page.Value {
			data, err := json.Marshal(server)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if server.Properties != nil {
				resourceMap["administratorLogin"] = server.Properties.AdministratorLogin
				resourceMap["version"] = server.Properties.Version
				resourceMap["sslEnforcement"] = server.Properties.SSLEnforcement
				resourceMap["minimalTlsVersion"] = server.Properties.MinimalTLSVersion
				resourceMap["publicNetworkAccess"] = server.Properties.PublicNetworkAccess
				resourceMap["userVisibleState"] = server.Properties.UserVisibleState
				resourceMap["fullyQualifiedDomainName"] = server.Properties.FullyQualifiedDomainName
				resourceMap["earliestRestoreDate"] = server.Properties.EarliestRestoreDate

				if server.Properties.StorageProfile != nil {
					resourceMap["storageMB"] = server.Properties.StorageProfile.StorageMB
					resourceMap["backupRetentionDays"] = server.Properties.StorageProfile.BackupRetentionDays
					resourceMap["geoRedundantBackup"] = server.Properties.StorageProfile.GeoRedundantBackup
					resourceMap["storageAutogrow"] = server.Properties.StorageProfile.StorageAutogrow
				}
			}

			if server.SKU != nil {
				resourceMap["skuName"] = server.SKU.Name
				resourceMap["skuTier"] = server.SKU.Tier
				resourceMap["skuCapacity"] = server.SKU.Capacity
				resourceMap["skuFamily"] = server.SKU.Family
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
