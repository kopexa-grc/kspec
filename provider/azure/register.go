// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package azure

import (
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider"
)

// rateLimitConfig defines the rate limiting for Azure ARM API.
// Azure ARM API rate limits vary by resource type (typically 200-1200 requests/5min).
// We use conservative limits: 10/s with burst of 20 for safety.
// Azure SDK has built-in retry, but proactive rate limiting helps avoid throttling.
var rateLimitConfig = &ratelimit.Config{
	RequestsPerSecond: 10,
	Burst:             20,
}

func init() {
	provider.Register(&provider.ProviderDefinition{
		Name:        "azure",
		Aliases:     []string{"az", "microsoft-azure"},
		Description: "Scan Azure subscriptions for security compliance (Resource Manager, Storage, Network, Compute, and more)",
		Flags: []provider.FlagDefinition{
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
		AssetTypes: []provider.AssetDefinition{
			{
				Name:        "subscription",
				Aliases:     []string{"sub"},
				Description: "Scan an Azure subscription",
				Args: []provider.ArgDefinition{
					{
						Name:        "subscription-id",
						Description: "Azure subscription ID",
						Required:    true,
						ConfigKey:   "subscription_id",
					},
				},
				ScannerKey: "azure-subscription",
			},
		},
		ConfigMapping: map[string]string{
			"client-id":      "client_id",
			"tenant-id":      "tenant_id",
			"client-secret":  "client_secret",
			"resource-group": "resource_group",
		},
		Factory: NewProvider,
		RateLimitConfig: rateLimitConfig,
	})
}
