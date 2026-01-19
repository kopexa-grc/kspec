// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package atlassian

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider"
)

// rateLimitConfig defines the rate limiting for Atlassian API.
// Atlassian API: Rate limits vary by product but are generally generous.
// We use 10/s with burst of 20 for safe operation across all APIs.
var rateLimitConfig = &ratelimit.Config{
	RequestsPerSecond: 10,
	Burst:             20,
}

func init() {
	provider.Register(&provider.ProviderDefinition{
		Name:        "atlassian",
		Aliases:     []string{"jira", "confluence"},
		Description: "Scan Atlassian sites and organizations for security compliance (Jira, Confluence, Admin settings)",
		Flags: []provider.FlagDefinition{
			{
				Name:        "site",
				Description: "Atlassian site (e.g., yoursite.atlassian.net)",
				EnvVar:      "ATLASSIAN_SITE",
			},
			{
				Name:        "api-token",
				Description: "Atlassian API Token",
				EnvVar:      "ATLASSIAN_API_TOKEN",
			},
			{
				Name:        "email",
				Description: "Atlassian account email",
				EnvVar:      "ATLASSIAN_EMAIL",
			},
			{
				Name:        "org-id",
				Description: "Atlassian Organization ID (for admin APIs)",
				EnvVar:      "ATLASSIAN_ORG_ID",
			},
		},
		AssetTypes: []provider.AssetDefinition{
			{
				Name:        "site",
				Description: "Scan an Atlassian site",
				Args: []provider.ArgDefinition{
					{
						Name:        "site-name",
						Description: "Site name (e.g., yoursite.atlassian.net)",
						Required:    true,
						ConfigKey:   "site",
					},
				},
				ScannerKey: "atlassian-site",
			},
			{
				Name:        "org",
				Description: "Scan an Atlassian organization",
				Args: []provider.ArgDefinition{
					{
						Name:        "org-id",
						Description: "Organization ID",
						Required:    true,
						ConfigKey:   "org_id",
					},
				},
				ScannerKey: "atlassian-org",
			},
		},
		ConfigMapping: map[string]string{
			"api-token": "api_token",
			"org-id":    "org_id",
		},
		Factory: func() core.Provider {
			return NewProvider()
		},
		RateLimitConfig: rateLimitConfig,
	})
}
