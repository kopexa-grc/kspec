// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package github

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:        "github",
		Aliases:     []string{"gh"},
		Description: "Scan GitHub organizations and repositories for security compliance",
		Flags: []registry.FlagDefinition{
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
		AssetTypes: []registry.AssetDefinition{
			{
				Name:        "org",
				Description: "Scan a GitHub organization",
				Args: []registry.ArgDefinition{
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
				Args: []registry.ArgDefinition{
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
	})
}
