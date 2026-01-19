// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package provider

import (
	"testing"
)

func TestDashToUnderscore(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"simple", "simple"},
		{"with-dash", "with_dash"},
		{"multiple-dashes-here", "multiple_dashes_here"},
		{"already_underscore", "already_underscore"},
		{"mixed-dash_and_underscore", "mixed_dash_and_underscore"},
		{"-leading-dash", "_leading_dash"},
		{"trailing-dash-", "trailing_dash_"},
		{"---multiple---", "___multiple___"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := dashToUnderscore(tt.input)
			if got != tt.want {
				t.Errorf("dashToUnderscore(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestProviderDefinition_GetConfigKey(t *testing.T) {
	tests := []struct {
		name     string
		def      *ProviderDefinition
		flagName string
		want     string
	}{
		{
			name: "no mapping uses dash to underscore",
			def: &ProviderDefinition{
				Name: "test",
			},
			flagName: "my-flag",
			want:     "my_flag",
		},
		{
			name: "empty mapping uses dash to underscore",
			def: &ProviderDefinition{
				Name:          "test",
				ConfigMapping: map[string]string{},
			},
			flagName: "another-flag",
			want:     "another_flag",
		},
		{
			name: "mapping overrides conversion",
			def: &ProviderDefinition{
				Name: "test",
				ConfigMapping: map[string]string{
					"my-flag": "custom_key",
				},
			},
			flagName: "my-flag",
			want:     "custom_key",
		},
		{
			name: "unmapped flag uses dash to underscore",
			def: &ProviderDefinition{
				Name: "test",
				ConfigMapping: map[string]string{
					"other-flag": "other_key",
				},
			},
			flagName: "my-flag",
			want:     "my_flag",
		},
		{
			name: "flag without dashes unchanged",
			def: &ProviderDefinition{
				Name: "test",
			},
			flagName: "simple",
			want:     "simple",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.def.GetConfigKey(tt.flagName)
			if got != tt.want {
				t.Errorf("GetConfigKey(%q) = %q, want %q", tt.flagName, got, tt.want)
			}
		})
	}
}

func TestProviderDefinition_GetAssetType(t *testing.T) {
	def := &ProviderDefinition{
		Name: "test-provider",
		AssetTypes: []AssetDefinition{
			{
				Name:    "account",
				Aliases: []string{"acc", "acct"},
			},
			{
				Name:    "subscription",
				Aliases: []string{"sub"},
			},
			{
				Name: "standalone",
				// No aliases
			},
		},
	}

	tests := []struct {
		name      string
		query     string
		wantFound bool
		wantName  string
	}{
		{
			name:      "find by primary name",
			query:     "account",
			wantFound: true,
			wantName:  "account",
		},
		{
			name:      "find by first alias",
			query:     "acc",
			wantFound: true,
			wantName:  "account",
		},
		{
			name:      "find by second alias",
			query:     "acct",
			wantFound: true,
			wantName:  "account",
		},
		{
			name:      "find second asset type by alias",
			query:     "sub",
			wantFound: true,
			wantName:  "subscription",
		},
		{
			name:      "find asset with no aliases",
			query:     "standalone",
			wantFound: true,
			wantName:  "standalone",
		},
		{
			name:      "not found returns false",
			query:     "nonexistent",
			wantFound: false,
		},
		{
			name:      "empty string not found",
			query:     "",
			wantFound: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			asset, ok := def.GetAssetType(tt.query)
			if ok != tt.wantFound {
				t.Errorf("GetAssetType(%q) found = %v, want %v", tt.query, ok, tt.wantFound)
			}
			if ok && asset.Name != tt.wantName {
				t.Errorf("GetAssetType(%q).Name = %q, want %q", tt.query, asset.Name, tt.wantName)
			}
		})
	}
}

func TestProviderDefinition_GetScannerKey(t *testing.T) {
	tests := []struct {
		name          string
		def           *ProviderDefinition
		assetTypeName string
		want          string
	}{
		{
			name: "explicit scanner key is used",
			def: &ProviderDefinition{
				Name: "github",
				AssetTypes: []AssetDefinition{
					{Name: "org", ScannerKey: "github-org"},
				},
			},
			assetTypeName: "org",
			want:          "github-org",
		},
		{
			name: "default key from provider and asset",
			def: &ProviderDefinition{
				Name: "aws",
				AssetTypes: []AssetDefinition{
					{Name: "account"},
				},
			},
			assetTypeName: "account",
			want:          "aws-account",
		},
		{
			name: "empty asset type returns provider name",
			def: &ProviderDefinition{
				Name: "simple",
			},
			assetTypeName: "",
			want:          "simple",
		},
		{
			name: "unknown asset type uses default format",
			def: &ProviderDefinition{
				Name: "test",
				AssetTypes: []AssetDefinition{
					{Name: "known"},
				},
			},
			assetTypeName: "unknown",
			want:          "test-unknown",
		},
		{
			name: "alias resolves to scanner key",
			def: &ProviderDefinition{
				Name: "github",
				AssetTypes: []AssetDefinition{
					{Name: "org", Aliases: []string{"organization"}, ScannerKey: "github-org"},
				},
			},
			assetTypeName: "organization",
			want:          "github-org",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.def.GetScannerKey(tt.assetTypeName)
			if got != tt.want {
				t.Errorf("GetScannerKey(%q) = %q, want %q", tt.assetTypeName, got, tt.want)
			}
		})
	}
}

func TestFlagTypeConstants(t *testing.T) {
	// Verify flag type constants have expected values
	if FlagTypeString != "string" {
		t.Errorf("FlagTypeString = %q, want %q", FlagTypeString, "string")
	}
	if FlagTypeBool != "bool" {
		t.Errorf("FlagTypeBool = %q, want %q", FlagTypeBool, "bool")
	}
	if FlagTypeStringSlice != "stringSlice" {
		t.Errorf("FlagTypeStringSlice = %q, want %q", FlagTypeStringSlice, "stringSlice")
	}
}
