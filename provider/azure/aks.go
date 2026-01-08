// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"

	"github.com/kopexa-grc/kspec/core"
)

// AKSClusterResource represents an Azure Kubernetes Service cluster resource scanner.
type AKSClusterResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *AKSClusterResource) Name() string {
	return "azure_aks_cluster"
}

// Fetch retrieves AKS clusters from Azure.
func (r *AKSClusterResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armcontainerservice.NewManagedClustersClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create AKS client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list AKS clusters: %w", err)
		}

		for _, cluster := range page.Value {
			data, err := json.Marshal(cluster)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			// Extract key security properties
			if cluster.Properties != nil {
				resourceMap["kubernetesVersion"] = cluster.Properties.KubernetesVersion
				resourceMap["enableRBAC"] = cluster.Properties.EnableRBAC
				resourceMap["provisioningState"] = cluster.Properties.ProvisioningState

				if cluster.Properties.NetworkProfile != nil {
					resourceMap["networkPlugin"] = cluster.Properties.NetworkProfile.NetworkPlugin
					resourceMap["networkPolicy"] = cluster.Properties.NetworkProfile.NetworkPolicy
				}

				if cluster.Properties.AADProfile != nil {
					resourceMap["aadEnabled"] = true
					resourceMap["enableAzureRBAC"] = cluster.Properties.AADProfile.EnableAzureRBAC
				} else {
					resourceMap["aadEnabled"] = false
				}

				if cluster.Properties.APIServerAccessProfile != nil {
					resourceMap["enablePrivateCluster"] = cluster.Properties.APIServerAccessProfile.EnablePrivateCluster
					resourceMap["authorizedIPRanges"] = cluster.Properties.APIServerAccessProfile.AuthorizedIPRanges
				}

				if cluster.Properties.AutoUpgradeProfile != nil {
					resourceMap["upgradeChannel"] = cluster.Properties.AutoUpgradeProfile.UpgradeChannel
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
