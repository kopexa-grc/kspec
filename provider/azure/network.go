package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"

	"github.com/kopexa-grc/kspec/core"
)

type NetworkSecurityGroupResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

func (r *NetworkSecurityGroupResource) Name() string {
	return "azure_network_security_group"
}

func (r *NetworkSecurityGroupResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Create NSG client
	client, err := armnetwork.NewSecurityGroupsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create NSG client: %w", err)
	}

	var resources []core.Resource

	// List all NSGs in the subscription
	pager := client.NewListAllPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list NSGs: %w", err)
		}

		for _, nsg := range page.Value {
			// Convert to map[string]interface{}
			data, err := json.Marshal(nsg)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			// Security rules are already included in the NSG properties
			// Extract them for easier access in policies
			if nsg.Properties != nil && nsg.Properties.SecurityRules != nil {
				var securityRules []map[string]interface{}
				for _, rule := range nsg.Properties.SecurityRules {
					ruleData, err := json.Marshal(rule)
					if err != nil {
						continue
					}
					var ruleMap map[string]interface{}
					if err := json.Unmarshal(ruleData, &ruleMap); err != nil {
						continue
					}
					securityRules = append(securityRules, ruleMap)
				}
				resourceMap["securityRules"] = securityRules
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
