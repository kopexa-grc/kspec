// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package sbom

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider"
)

func init() {
	provider.Register(&provider.ProviderDefinition{
		Name:        "sbom",
		Aliases:     []string{"bom", "cyclonedx", "spdx"},
		Description: "Scan Software Bill of Materials (SBOM) files for compliance (CycloneDX, SPDX)",
		Flags: []provider.FlagDefinition{
			{
				Name:        "sbom-path",
				Description: "Path to SBOM file or directory",
			},
		},
		// Both "file" and "dir" asset types intentionally share the same ConfigKey "sbom_path".
		// This key is a generic input path that may point to either a single SBOM file or a
		// directory containing multiple SBOM files. The provider determines at runtime whether
		// the path refers to a file or a directory, and the ScannerKey ("sbom-file" vs
		// "sbom-directory") is used to select the appropriate scan behavior. When adding new
		// asset types, either reuse "sbom_path" for the same generic path semantics or introduce
		// a distinct ConfigKey if different configuration is required.
		AssetTypes: []provider.AssetDefinition{
			{
				Name:        "file",
				Description: "Scan a single SBOM file",
				Args: []provider.ArgDefinition{
					{
						Name:        "path",
						Description: "Path to SBOM file",
						Required:    true,
						ConfigKey:   "sbom_path",
					},
				},
				ScannerKey: "sbom-file",
			},
			{
				Name:        "dir",
				Aliases:     []string{"directory"},
				Description: "Scan a directory of SBOM files",
				Args: []provider.ArgDefinition{
					{
						Name:        "path",
						Description: "Path to directory containing SBOM files",
						Required:    true,
						ConfigKey:   "sbom_path",
					},
				},
				ScannerKey: "sbom-directory",
			},
		},
		ConfigMapping: map[string]string{
			"sbom-path": "sbom_path",
		},
		Factory: func() core.Provider {
			return NewProvider()
		},
		// SBOM provider parses local files, no rate limiting needed
		RateLimitConfig: nil,
	})
}
