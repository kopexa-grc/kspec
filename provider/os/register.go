// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package os

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:        "os",
		Aliases:     []string{"local", "localhost"},
		Description: "Scan local operating system for security compliance",
		Flags:       []registry.FlagDefinition{},
		AssetTypes: []registry.AssetDefinition{
			{
				Name:        "local",
				Description: "Scan local system",
				DefaultName: "localhost",
				ScannerKey:  "local",
			},
		},
		Factory: func() core.Provider {
			return New()
		},
		// No rate limiting needed for local operations
		RateLimitConfig: nil,
	})
}
