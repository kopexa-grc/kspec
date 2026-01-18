// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package network

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:            "network",
		Aliases:         []string{"host", "dns"},
		Description:     "Scan network hosts for security compliance (DNS, TLS, HTTP headers)",
		Flags:           []registry.FlagDefinition{},
		RateLimitConfig: &ratelimit.Config{RequestsPerSecond: 20, Burst: 40},
		AssetTypes: []registry.AssetDefinition{
			{
				Name:        "host",
				Description: "Scan a network host",
				Args: []registry.ArgDefinition{
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
