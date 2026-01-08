// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/iothub/armiothub"

	"github.com/kopexa-grc/kspec/core"
)

// IoTHubResource represents an Azure IoT Hub resource scanner.
type IoTHubResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *IoTHubResource) Name() string {
	return "azure_iothub"
}

// Fetch retrieves IoT Hubs from Azure.
func (r *IoTHubResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armiothub.NewResourceClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create IoT Hub client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListBySubscriptionPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list IoT Hubs: %w", err)
		}

		for _, hub := range page.Value {
			data, err := json.Marshal(hub)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if hub.Properties != nil {
				resourceMap["state"] = hub.Properties.State
				resourceMap["hostName"] = hub.Properties.HostName
				resourceMap["provisioningState"] = hub.Properties.ProvisioningState
				resourceMap["publicNetworkAccess"] = hub.Properties.PublicNetworkAccess
				resourceMap["minTlsVersion"] = hub.Properties.MinTLSVersion
				resourceMap["disableLocalAuth"] = hub.Properties.DisableLocalAuth
				resourceMap["disableDeviceSAS"] = hub.Properties.DisableDeviceSAS
				resourceMap["disableModuleSAS"] = hub.Properties.DisableModuleSAS
				resourceMap["enableDataResidency"] = hub.Properties.EnableDataResidency
				resourceMap["restrictOutboundNetworkAccess"] = hub.Properties.RestrictOutboundNetworkAccess

				if hub.Properties.Features != nil {
					resourceMap["features"] = hub.Properties.Features
				}

				// Network rule sets
				if hub.Properties.NetworkRuleSets != nil {
					resourceMap["defaultAction"] = hub.Properties.NetworkRuleSets.DefaultAction
					resourceMap["applyToBuiltInEventHubEndpoint"] = hub.Properties.NetworkRuleSets.ApplyToBuiltInEventHubEndpoint
				}

				// IP filter rules
				if hub.Properties.IPFilterRules != nil {
					rules := make([]map[string]interface{}, 0, len(hub.Properties.IPFilterRules))
					for _, rule := range hub.Properties.IPFilterRules {
						ruleMap := map[string]interface{}{
							"filterName": rule.FilterName,
							"action":     rule.Action,
							"ipMask":     rule.IPMask,
						}
						rules = append(rules, ruleMap)
					}
					resourceMap["ipFilterRules"] = rules
				}
			}

			if hub.SKU != nil {
				resourceMap["skuName"] = hub.SKU.Name
				resourceMap["skuTier"] = hub.SKU.Tier
				resourceMap["skuCapacity"] = hub.SKU.Capacity
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
