// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package network

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:        "network",
		Aliases:     []string{"host", "dns"},
		Description: "Scan network hosts for security compliance (DNS, TLS, HTTP headers)",
		Flags:       []registry.FlagDefinition{},
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
