// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package sbom

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/registry"
)

func init() {
	registry.Register(&registry.ProviderDefinition{
		Name:        "sbom",
		Aliases:     []string{"bom", "cyclonedx", "spdx"},
		Description: "Scan Software Bill of Materials (SBOM) files for compliance (CycloneDX, SPDX)",
		Flags: []registry.FlagDefinition{
			{
				Name:        "sbom-path",
				Description: "Path to SBOM file or directory",
			},
		},
		AssetTypes: []registry.AssetDefinition{
			{
				Name:        "file",
				Description: "Scan a single SBOM file",
				Args: []registry.ArgDefinition{
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
				Args: []registry.ArgDefinition{
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
	})
}
