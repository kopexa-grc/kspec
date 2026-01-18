// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package cloudflare

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:        "cloudflare",
		Aliases:     []string{"cf"},
		Description: "Scan Cloudflare accounts and zones for security compliance (DNS, WAF, SSL/TLS, Access)",
		Flags: []registry.FlagDefinition{
			{
				Name:        "api-token",
				Description: "Cloudflare API Token",
				EnvVar:      "CLOUDFLARE_API_TOKEN",
			},
			{
				Name:        "api-key",
				Description: "Cloudflare API Key (legacy authentication)",
				EnvVar:      "CLOUDFLARE_API_KEY",
			},
			{
				Name:        "email",
				Description: "Cloudflare account email (required with api-key)",
				EnvVar:      "CLOUDFLARE_EMAIL",
			},
			{
				Name:        "account-id",
				Description: "Cloudflare Account ID",
				EnvVar:      "CLOUDFLARE_ACCOUNT_ID",
			},
		},
		AssetTypes: []registry.AssetDefinition{
			{
				Name:        "account",
				Description: "Scan a Cloudflare account",
				Args: []registry.ArgDefinition{
					{
						Name:        "account-id",
						Description: "Account ID (optional, scans all if omitted)",
						Required:    false,
						ConfigKey:   "account_id",
					},
				},
				DefaultName: "all",
				ScannerKey:  "cloudflare-account",
			},
			{
				Name:        "zone",
				Description: "Scan a specific Cloudflare zone",
				Args: []registry.ArgDefinition{
					{
						Name:        "zone-id",
						Description: "Zone ID",
						Required:    true,
						ConfigKey:   "zone_id",
					},
				},
				ScannerKey: "cloudflare-zone",
			},
		},
		ConfigMapping: map[string]string{
			"api-token":  "api_token",
			"api-key":    "api_key",
			"account-id": "account_id",
		},
		Factory: func() core.Provider {
			return NewProvider()
		},
		// Cloudflare SDK has built-in rate limiting with automatic backoff,
		// so no external rate limiting is needed.
		RateLimitConfig: nil,
	})
}
