// Package os provides local operating system resource scanning capabilities
// including services, packages, files, and macOS-specific resources.
package os

import (
	"context"

	"github.com/kopexa-grc/kspec/core"
)

type OSProvider struct{}

func New() *OSProvider {
	return &OSProvider{}
}

func (p *OSProvider) Name() string {
	return "os"
}

func (p *OSProvider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// OS provider usually operates on local system so no real connection setup
	return &OSConnection{}, nil
}

type OSConnection struct{}

func (c *OSConnection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&AppleCareResource{},
		&FileResource{},
		&PackageResource{},
		&ServiceResource{},
	}
}

func (p *OSProvider) Shutdown() error {
	return nil
}
