// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package factorial

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:        "factorial",
		Aliases:     []string{"factorial-hr", "hris"},
		Description: "Scan Factorial HR for compliance (Employee data, Policies, Documents)",
		Flags: []registry.FlagDefinition{
			{
				Name:        "factorial-api-key",
				Description: "Factorial HR API Key",
				EnvVar:      "FACTORIAL_API_KEY",
			},
			{
				Name:        "access-token",
				Description: "Factorial HR OAuth Access Token",
				EnvVar:      "FACTORIAL_ACCESS_TOKEN",
			},
		},
		AssetTypes: []registry.AssetDefinition{
			{
				Name:        "company",
				Description: "Scan Factorial HR company data",
				DefaultName: "default",
				ScannerKey:  "factorial-company",
			},
		},
		ConfigMapping: map[string]string{
			"factorial-api-key": "api_key",
			"api-key":           "api_key",
			"access-token":      "access_token",
		},
		Factory: func() core.Provider {
			return NewProvider()
		},
		// Factorial HR API: Conservative rate limiting for HR API
		// We use 5/s with burst of 10 to be respectful of the API
		RateLimitConfig: &ratelimit.Config{
			RequestsPerSecond: 5,
			Burst:             10,
		},
	})
}
