// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package atlassian

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:        "atlassian",
		Aliases:     []string{"jira", "confluence"},
		Description: "Scan Atlassian sites and organizations for security compliance (Jira, Confluence, Admin settings)",
		Flags: []registry.FlagDefinition{
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
		AssetTypes: []registry.AssetDefinition{
			{
				Name:        "site",
				Description: "Scan an Atlassian site",
				Args: []registry.ArgDefinition{
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
				Args: []registry.ArgDefinition{
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
	})
}
