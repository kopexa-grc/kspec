// Package os provides local operating system resource scanning capabilities
// including services, packages, files, and macOS-specific resources.
package os

import (
	"context"

	"github.com/kopexa-grc/kspec/core"
)

// Provider implements the Provider interface for local operating system resources.
type Provider struct{}

// New creates a new Provider instance.
func New() *Provider {
	return &Provider{}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "os"
}

// Connect establishes a connection to the local operating system.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// OS provider usually operates on local system so no real connection setup
	return &Connection{}, nil
}

// Connection represents a connection to the local operating system.
type Connection struct{}

// Resources returns the list of available OS resource types.
func (c *Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&AppleCareResource{},
		&FileResource{},
		&PackageResource{},
		&ServiceResource{},
	}
}

// Shutdown releases any resources held by the provider.
func (p *Provider) Shutdown() error {
	return nil
}
