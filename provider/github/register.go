// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package github

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider"
)

// rateLimitConfig defines the rate limiting for GitHub API.
// GitHub API: 5000 requests/hour for authenticated users = ~1.4/s
// We use 5/s with burst of 10 for safety margin.
var rateLimitConfig = &ratelimit.Config{
	RequestsPerSecond: 5,
	Burst:             10,
}

func init() {
	provider.Register(&provider.ProviderDefinition{
		Name:        "github",
		Aliases:     []string{"gh"},
		Description: "Scan GitHub organizations and repositories for security compliance",
		Flags: []provider.FlagDefinition{
			{
				Name:        "token",
				Description: "GitHub personal access token",
				EnvVar:      "GITHUB_TOKEN",
			},
			{
				Name:        "credential-type",
				Description: "Credential type (bearer, env)",
				Default:     "env",
			},
			{
				Name:        "env-var",
				Description: "Environment variable for token",
				Default:     "GITHUB_TOKEN",
			},
		},
		AssetTypes: []provider.AssetDefinition{
			{
				Name:        "org",
				Description: "Scan a GitHub organization",
				Args: []provider.ArgDefinition{
					{
						Name:        "owner",
						Description: "Organization name",
						Required:    true,
						ConfigKey:   "owner",
					},
				},
				ScannerKey: "github-org",
			},
			{
				Name:        "repo",
				Description: "Scan a GitHub repository",
				Args: []provider.ArgDefinition{
					{
						Name:        "repository",
						Description: "Repository in format owner/repo",
						Required:    true,
						ConfigKey:   "repository",
					},
				},
				ScannerKey: "github-repo",
			},
		},
		ConfigMapping: map[string]string{
			"credential-type": "credential_type",
			"env-var":         "env_var",
		},
		Factory: func() core.Provider {
			return NewProvider()
		},
		RateLimitConfig: rateLimitConfig,
	})
}
