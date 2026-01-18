// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package ms365

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:        "ms365",
		Aliases:     []string{"m365", "microsoft365", "office365"},
		Description: "Scan Microsoft 365 tenants for security compliance (Exchange, SharePoint, Teams, Azure AD)",
		Flags: []registry.FlagDefinition{
			{
				Name:        "client-id",
				Description: "Microsoft 365 Client ID (Application)",
				EnvVar:      "MS365_CLIENT_ID",
			},
			{
				Name:        "tenant-id",
				Description: "Microsoft 365 Tenant ID",
				EnvVar:      "MS365_TENANT_ID",
			},
			{
				Name:        "client-secret",
				Description: "Microsoft 365 Client Secret",
				EnvVar:      "MS365_CLIENT_SECRET",
			},
		},
		AssetTypes: []registry.AssetDefinition{
			{
				Name:        "tenant",
				Description: "Scan a Microsoft 365 tenant",
				Args: []registry.ArgDefinition{
					{
						Name:        "tenant-id",
						Description: "Tenant ID",
						Required:    false,
						ConfigKey:   "tenant_id",
					},
				},
				DefaultName: "default",
				ScannerKey:  "ms365",
			},
		},
		ConfigMapping: map[string]string{
			"client-id":     "client_id",
			"tenant-id":     "tenant_id",
			"client-secret": "client_secret",
		},
		Factory: func() core.Provider {
			return NewProvider()
		},
		// Microsoft Graph API: throttling limits vary by endpoint
		// Most endpoints: 10,000 requests per 10 minutes
		// We use 10/s with burst of 20 for safety margin
		// Graph SDK has built-in throttling handling, but rate limiting adds safety
		RateLimitConfig: &ratelimit.Config{
			RequestsPerSecond: 10,
			Burst:             20,
		},
	})
}
