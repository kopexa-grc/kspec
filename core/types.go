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
}

// CheckResult represents the outcome of a policy check.
type CheckResult struct {
	TenantID   string                 `json:"tenant_id"`
	ResourceID string                 `json:"resource_id"`
	Passed     bool                   `json:"passed"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}
