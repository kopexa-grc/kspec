// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package hetzner

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider"
)

// rateLimitConfig defines the rate limiting for Hetzner API.
// Hetzner API has generous rate limits (around 3600 requests per hour).
// We use 10/s with burst of 20 for safety margin.
var rateLimitConfig = &ratelimit.Config{
	RequestsPerSecond: 10,
	Burst:             20,
}

func init() {
	provider.Register(&provider.ProviderDefinition{
		Name:        "hetzner",
		Aliases:     []string{"hcloud"},
		Description: "Scan Hetzner Cloud projects for security compliance (Servers, Networks, Firewalls, Load Balancers)",
		Flags: []provider.FlagDefinition{
			{
				Name:        "hcloud-token",
				Description: "Hetzner Cloud API Token",
				EnvVar:      "HCLOUD_TOKEN",
			},
			{
				Name:        "project",
				Description: "Hetzner Cloud project name (for identification)",
			},
		},
		AssetTypes: []provider.AssetDefinition{
			{
				Name:        "project",
				Description: "Scan a Hetzner Cloud project",
				Args: []provider.ArgDefinition{
					{
						Name:        "project-name",
						Description: "Project name (for identification)",
						Required:    false,
						ConfigKey:   "project",
					},
				},
				DefaultName: "default",
				ScannerKey:  "hetzner-project",
			},
		},
		ConfigMapping: map[string]string{
			"hcloud-token": "api_token",
			"api-token":    "api_token",
		},
		Factory: func() core.Provider {
			return NewProvider()
		},
		RateLimitConfig: rateLimitConfig,
	})
}
