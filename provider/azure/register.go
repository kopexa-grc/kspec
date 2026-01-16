// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package azure

import "github.com/kopexa-grc/kspec/provider/registry"

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:        "azure",
		Aliases:     []string{"az", "microsoft-azure"},
		Description: "Scan Azure subscriptions for security compliance (Resource Manager, Storage, Network, Compute, and more)",
		Flags: []registry.FlagDefinition{
			{
				Name:        "client-id",
				Description: "Azure Client ID (Service Principal)",
				EnvVar:      "AZURE_CLIENT_ID",
			},
			{
				Name:        "tenant-id",
				Description: "Azure Tenant ID",
				EnvVar:      "AZURE_TENANT_ID",
			},
			{
				Name:        "client-secret",
				Description: "Azure Client Secret (Service Principal)",
				EnvVar:      "AZURE_CLIENT_SECRET",
			},
			{
				Name:        "resource-group",
				Description: "Azure Resource Group (optional filter)",
			},
		},
		AssetTypes: []registry.AssetDefinition{
			{
				Name:        "subscription",
				Aliases:     []string{"sub"},
				Description: "Scan an Azure subscription",
				Args: []registry.ArgDefinition{
					{
						Name:        "subscription-id",
						Description: "Azure subscription ID",
						Required:    true,
						ConfigKey:   "subscription_id",
					},
				},
				ScannerKey: "azure",
			},
		},
		ConfigMapping: map[string]string{
			"client-id":      "client_id",
			"tenant-id":      "tenant_id",
			"client-secret":  "client_secret",
			"resource-group": "resource_group",
		},
		Factory: NewProvider,
	})
}
