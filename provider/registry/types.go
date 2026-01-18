// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package registry provides a dynamic provider registration system.
package registry

import (
	"strings"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
)

// ProviderDefinition describes a provider and its configuration requirements.
type ProviderDefinition struct {
	// Name is the primary identifier for the provider (e.g., "aws", "azure").
	Name string

	// Aliases are alternative names that map to this provider (e.g., "amazon", "ec2" for aws).
	Aliases []string

	// Description provides help text for the provider.
	Description string

	// Flags defines provider-specific CLI flags.
	Flags []FlagDefinition

	// AssetTypes defines the asset types this provider supports.
	AssetTypes []AssetDefinition

	// Factory creates a new instance of the provider.
	Factory ProviderFactory

	// ConfigMapping maps flag names to config keys if they differ.
	// If not set, flag names are used directly as config keys.
	ConfigMapping map[string]string

	// RateLimitConfig defines provider-specific rate limiting.
	// If nil, the provider's SDK handles rate limiting internally.
	RateLimitConfig *ratelimit.Config

	// ResourceHierarchy defines parent-child relationships between resource types.
	// Maps child resource type -> parent resource type.
	// Example: {"github_repo": "github_organization", "github_team": "github_organization"}
	ResourceHierarchy map[string]string

	// ResourceGroups defines visual groupings for resource types.
	// Maps resource type -> group name (e.g., "compute", "storage", "identity").
	ResourceGroups map[string]string

	// RootResourceType is the top-level resource type for this provider.
	// Example: "github_organization" for GitHub, "aws_account" for AWS.
	RootResourceType string
}

// FlagDefinition describes a CLI flag for a provider.
type FlagDefinition struct {
	// Name is the flag name (without dashes).
	Name string

	// Short is the single-character shorthand (optional).
	Short string

	// Description provides help text for the flag.
	Description string

	// Default is the default value for the flag.
	Default string

	// Required indicates whether the flag must be provided.
	Required bool

	// EnvVar is the environment variable to use as fallback.
	EnvVar string

	// Type specifies the flag type (string, bool, stringSlice).
	// Defaults to "string" if not specified.
	Type FlagType
}

// FlagType represents the type of a flag value.
type FlagType string

// Flag type constants.
const (
	FlagTypeString      FlagType = "string"
	FlagTypeBool        FlagType = "bool"
	FlagTypeStringSlice FlagType = "stringSlice"
)

// AssetDefinition describes an asset type that a provider can scan.
type AssetDefinition struct {
	// Name is the asset type identifier (e.g., "account", "subscription", "org").
	Name string

	// Aliases are alternative names for this asset type.
	Aliases []string

	// Description provides help text for the asset type.
	Description string

	// Args defines positional arguments for this asset type.
	Args []ArgDefinition

	// DefaultName is the default asset name if none is provided.
	DefaultName string

	// ScannerKey is the asset type key used by the scanner (e.g., "aws-account", "github-org").
	// If not set, it will be derived from provider name + asset name.
	ScannerKey string
}

// ArgDefinition describes a positional argument for an asset type.
type ArgDefinition struct {
	// Name is the argument name for help text.
	Name string

	// Description provides help text for the argument.
	Description string

	// Required indicates whether this argument must be provided.
	Required bool

	// ConfigKey is the key to store this argument in the config map.
	ConfigKey string
}

// ProviderFactory is a function that creates a new provider instance.
type ProviderFactory func() core.Provider

// GetConfigKey returns the config key for a flag.
// If a ConfigMapping is defined and contains this flag, use that.
// Otherwise, convert dashes to underscores.
func (p *ProviderDefinition) GetConfigKey(flagName string) string {
	if p.ConfigMapping != nil {
		if key, ok := p.ConfigMapping[flagName]; ok {
			return key
		}
	}
	// Convert dashes to underscores for config keys
	return dashToUnderscore(flagName)
}

// GetAssetType returns the asset definition for the given name or alias.
func (p *ProviderDefinition) GetAssetType(name string) (*AssetDefinition, bool) {
	for i := range p.AssetTypes {
		if p.AssetTypes[i].Name == name {
			return &p.AssetTypes[i], true
		}
		for _, alias := range p.AssetTypes[i].Aliases {
			if alias == name {
				return &p.AssetTypes[i], true
			}
		}
	}
	return nil, false
}

// dashToUnderscore converts dashes to underscores in a string.
func dashToUnderscore(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

// GetScannerKey returns the scanner key for an asset type.
// If ScannerKey is set on the asset definition, it returns that.
// Otherwise, it builds a default key from provider and asset name.
func (p *ProviderDefinition) GetScannerKey(assetTypeName string) string {
	// Find the asset definition
	assetDef, ok := p.GetAssetType(assetTypeName)
	if ok && assetDef.ScannerKey != "" {
		return assetDef.ScannerKey
	}

	// Default: provider-assettype or just provider if assetType is empty
	if assetTypeName == "" {
		return p.Name
	}
	return p.Name + "-" + assetTypeName
}

// NewRateLimiter creates a rate limiter for this provider.
// Returns nil if the provider handles rate limiting internally (RateLimitConfig is nil).
func (p *ProviderDefinition) NewRateLimiter() *ratelimit.Limiter {
	if p.RateLimitConfig == nil {
		return nil
	}
	return ratelimit.New(p.Name, *p.RateLimitConfig)
}

// NeedsRateLimiting returns true if the provider requires external rate limiting.
func (p *ProviderDefinition) NeedsRateLimiting() bool {
	return p.RateLimitConfig != nil
}

// GetParentType returns the parent resource type for a child type.
// Returns empty string if no parent is defined.
func (p *ProviderDefinition) GetParentType(childType string) string {
	if p.ResourceHierarchy == nil {
		return ""
	}
	return p.ResourceHierarchy[childType]
}

// GetResourceGroup returns the visual group for a resource type.
// Returns "other" if no group is defined.
func (p *ProviderDefinition) GetResourceGroup(resourceType string) string {
	if p.ResourceGroups == nil {
		return "other"
	}
	if group, ok := p.ResourceGroups[resourceType]; ok {
		return group
	}
	return "other"
}

// GetRootType returns the root resource type for this provider.
func (p *ProviderDefinition) GetRootType() string {
	return p.RootResourceType
}
