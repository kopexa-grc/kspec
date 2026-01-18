// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"

	"github.com/kopexa-grc/kspec/core"
)

// RedisCache represents an Azure Redis Cache resource scanner.
type RedisCache struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewRedisCache creates a new RedisCache resource scanner.
func NewRedisCache(credential azcore.TokenCredential, subscriptionID string) *RedisCache {
	return &RedisCache{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *RedisCache) Name() string {
	return "azure_redis_cache"
}

// Fetch retrieves Redis Cache instances from Azure.
func (r *RedisCache) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armredis.NewClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Redis client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list Redis caches: %w", err)
		}

		for _, cache := range page.Value {
			data, err := json.Marshal(cache)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if cache.Properties != nil {
				resourceMap["hostName"] = cache.Properties.HostName
				resourceMap["port"] = cache.Properties.Port
				resourceMap["sslPort"] = cache.Properties.SSLPort
				resourceMap["provisioningState"] = cache.Properties.ProvisioningState
				resourceMap["redisVersion"] = cache.Properties.RedisVersion
				resourceMap["publicNetworkAccess"] = cache.Properties.PublicNetworkAccess
				resourceMap["enableNonSslPort"] = cache.Properties.EnableNonSSLPort
				resourceMap["minimumTlsVersion"] = cache.Properties.MinimumTLSVersion
				resourceMap["replicasPerMaster"] = cache.Properties.ReplicasPerMaster
				resourceMap["replicasPerPrimary"] = cache.Properties.ReplicasPerPrimary
				resourceMap["shardCount"] = cache.Properties.ShardCount

				if cache.Properties.RedisConfiguration != nil {
					resourceMap["maxMemoryPolicy"] = cache.Properties.RedisConfiguration.MaxmemoryPolicy
					resourceMap["rdbBackupEnabled"] = cache.Properties.RedisConfiguration.RdbBackupEnabled
					resourceMap["rdbBackupFrequency"] = cache.Properties.RedisConfiguration.RdbBackupFrequency
					resourceMap["rdbBackupMaxSnapshotCount"] = cache.Properties.RedisConfiguration.RdbBackupMaxSnapshotCount
				}

				// Private endpoint connections
				if cache.Properties.PrivateEndpointConnections != nil {
					peCount := len(cache.Properties.PrivateEndpointConnections)
					resourceMap["privateEndpointConnectionCount"] = peCount
					resourceMap["hasPrivateEndpoint"] = peCount > 0
				}

				// Linked servers
				if cache.Properties.LinkedServers != nil {
					resourceMap["linkedServerCount"] = len(cache.Properties.LinkedServers)
				}

				// SKU info from properties
				if cache.Properties.SKU != nil {
					resourceMap["skuName"] = cache.Properties.SKU.Name
					resourceMap["skuFamily"] = cache.Properties.SKU.Family
					resourceMap["skuCapacity"] = cache.Properties.SKU.Capacity
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
