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
