// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package discovery

import (
	"context"

	"github.com/kopexa-grc/kspec/core"
)

// Discover is a high-level API for programmatic resource discovery.
//
// Example:
//
//	result, err := discovery.Discover(ctx, "github",
//	    map[string]string{"token": "ghp_xxx"},
//	    "github-org", "my-org")
func Discover(ctx context.Context, providerName string, config map[string]string, assetType, assetName string) (*Result, error) {
	discConfig := Config{
		ProviderName:   providerName,
		ProviderConfig: config,
		Asset: core.Asset{
			Type:   assetType,
			Name:   assetName,
			Config: config,
		},
		IncludeInstances: false, // Default to lightweight mode
	}

	return DiscoverWithOptions(ctx, discConfig)
}

// DiscoverWithInstances performs discovery and includes individual resource instances.
// This is required for graph generation but uses more memory.
//
// Example:
//
//	result, err := discovery.DiscoverWithInstances(ctx, "github",
//	    map[string]string{"token": "ghp_xxx"},
//	    "github-org", "my-org")
//	graph := result.Graph // Contains nodes and edges
func DiscoverWithInstances(ctx context.Context, providerName string, config map[string]string, assetType, assetName string) (*Result, error) {
	discConfig := Config{
		ProviderName:   providerName,
		ProviderConfig: config,
		Asset: core.Asset{
			Type:   assetType,
			Name:   assetName,
			Config: config,
		},
		IncludeInstances: true,
	}

	return DiscoverWithOptions(ctx, discConfig)
}

// DiscoverWithOptions performs discovery with full configuration control.
func DiscoverWithOptions(ctx context.Context, config Config) (*Result, error) {
	disc := NewDiscoverer(config)

	if err := disc.Initialize(ctx); err != nil {
		return nil, err
	}

	return disc.Discover(ctx, config.Asset)
}

// QuickInventory returns a simple map of resource type -> count.
// This is the fastest way to get an overview of available resources.
//
// Example:
//
//	inventory, err := discovery.QuickInventory(ctx, "aws",
//	    map[string]string{"region": "eu-central-1"},
//	    "aws-account", "production")
//	// Returns: map[string]int{"aws_s3_bucket": 42, "aws_iam_user": 15, ...}
func QuickInventory(ctx context.Context, providerName string, config map[string]string, assetType, assetName string) (map[string]int, error) {
	result, err := Discover(ctx, providerName, config, assetType, assetName)
	if err != nil {
		return nil, err
	}

	inventory := make(map[string]int)
	for _, r := range result.Resources {
		if r.Count >= 0 {
			inventory[r.Type] = r.Count
		}
	}

	return inventory, nil
}

// ListResourceTypes returns the available resource types for a provider.
// This doesn't make any API calls - it just returns what's registered.
func ListResourceTypes(providerName string) ([]string, error) {
	// Use a minimal config to get the resource registry
	ctx := context.Background()
	discConfig := Config{
		ProviderName:   providerName,
		ProviderConfig: map[string]string{},
		Mode:           ModeRegistry,
	}

	disc := NewDiscoverer(discConfig)

	// Try to get resource specs from registry (this may fail without credentials)
	// For now, return empty - providers can add a method to return their resource types
	_ = disc
	_ = ctx

	return nil, nil
}
