// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package network

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider"
)

func init() {
	provider.Register(&provider.ProviderDefinition{
		Name:            "network",
		Aliases:         []string{"host", "dns"},
		Description:     "Scan network hosts for security compliance (DNS, TLS, HTTP headers)",
		Flags:           []provider.FlagDefinition{},
		RateLimitConfig: &ratelimit.Config{RequestsPerSecond: 20, Burst: 40},
		AssetTypes: []provider.AssetDefinition{
			{
				Name:        "host",
				Description: "Scan a network host",
				Args: []provider.ArgDefinition{
					{
						Name:        "target",
						Description: "Host to scan (domain or IP)",
						Required:    true,
						ConfigKey:   "target",
					},
				},
				ScannerKey: "host",
			},
		},
		Factory: func() core.Provider {
			return NewProvider()
		},
	})
}
