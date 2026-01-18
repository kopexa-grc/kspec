// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"

	"github.com/kopexa-grc/kspec/core"
)

// VirtualMachine represents an Azure Virtual Machine resource scanner.
type VirtualMachine struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewVirtualMachine creates a new VirtualMachine resource scanner.
func NewVirtualMachine(credential azcore.TokenCredential, subscriptionID string) *VirtualMachine {
	return &VirtualMachine{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *VirtualMachine) Name() string {
	return "azure_virtual_machine"
}

// Fetch retrieves virtual machines from Azure.
func (r *VirtualMachine) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Create VM client
	client, err := armcompute.NewVirtualMachinesClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create VM client: %w", err)
	}

	var resources []core.Resource

	// List all VMs in the subscription
	pager := client.NewListAllPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list VMs: %w", err)
		}

		for _, vm := range page.Value {
			// Convert to map[string]interface{}
			data, err := json.Marshal(vm)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			// Extract OS disk information for easier access
			if vm.Properties != nil && vm.Properties.StorageProfile != nil && vm.Properties.StorageProfile.OSDisk != nil {
				osDiskData, err := json.Marshal(vm.Properties.StorageProfile.OSDisk)
				if err == nil {
					var osDiskMap map[string]interface{}
					if err := json.Unmarshal(osDiskData, &osDiskMap); err == nil {
						resourceMap["osDisk"] = osDiskMap
					}
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
