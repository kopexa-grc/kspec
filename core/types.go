// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package core

import (
	"context"
)

// Resource represents the data we are checking (e.g., a GitHub Repository).
// It's a simple map so CEL can dynamically access its fields.
type Resource map[string]interface{}

// ResourceSpec defines a specific resource type that can be fetched.
type ResourceSpec interface {
	Name() string
	Fetch(ctx context.Context, asset Asset) ([]Resource, error)
}

// SubResourceProvider is an optional interface that ResourceSpecs can implement
// to dynamically register additional resource types at runtime.
// This allows resources to provide dependent sub-resources (e.g., SQL Server -> Databases).
type SubResourceProvider interface {
	ResourceSpec
	// SubResources returns additional resource specs that should be registered
	// These are typically dependent resources discovered during connection
	SubResources() []ResourceSpec
}

// ChildResourceProvider is an optional interface for resources that have
// top-level child resources in the hierarchy (e.g., Organization → Teams, Repos).
// Unlike SubResourceProvider which returns ResourceSpecs, this returns type names
// that are already registered in the provider's registry.
type ChildResourceProvider interface {
	ResourceSpec
	// ChildResources returns the names of child resource types.
	// These types must be registered in the provider's registry.
	ChildResources() []string
}

// DiscoveryResource is an optional interface for resources that can discover
// what sub-resources actually exist before scanning them.
// This enables smart scanning that only checks resources that are present.
type DiscoveryResource interface {
	ResourceSpec
	// Discover scans the environment and returns which resource types are present
	// Returns a map of resource type name -> count of instances found
	Discover(ctx context.Context, asset Asset) (map[string]int, error)
}

// Provider defines the interface for modular data sources (GitHub, AWS, etc.).
type Provider interface {
	Name() string
	// Connect establishes a session with the provider using the given config
	Connect(ctx context.Context, config map[string]string) (Connection, error)
}

// Connection represents an active session with a provider
type Connection interface {
	// Resources returns the list of resources available on this connection
	Resources() []ResourceSpec
	// EntryResourceType returns the entry point resource type for a given asset type.
	// For example, "github-org" -> "github_organization", "aws-account" -> "aws_account"
	EntryResourceType(assetType string) string
}

// AssetContextProvider is an optional interface that ResourceSpecs can implement
// to provide context when building assets for sub-resource fetching.
// This allows parent resources to pass necessary context (e.g., repo name) to children.
type AssetContextProvider interface {
	ResourceSpec
	// BuildChildAsset creates an asset configuration for fetching child/sub resources.
	// The parent resource data and base asset are provided to build the child asset.
	BuildChildAsset(parentData Resource, baseAsset Asset) Asset
}

