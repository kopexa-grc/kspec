// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package network provides network-related scanning capabilities including
// TLS certificates, HTTP security headers, and DNS configuration.
package network

import (
	"context"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/network/resources"
)

// Provider implements a network-based provider for scanning remote hosts
// for TLS certificates, HTTP security headers, and DNS configuration.
type Provider struct{}

// NewProvider creates a new Provider instance.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name identifier.
func (p *Provider) Name() string {
	return "network"
}

// Connect establishes a network connection with the given configuration.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Network provider currently stateless, potentially could take proxy config etc
	return &Connection{}, nil
}

// Connection represents an active network connection that provides
// access to network-related resources such as certificates, DNS, HTTP, and TLS.
type Connection struct{}

// Resources returns the list of available resource types for network scanning.
func (c *Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		resources.NewCertificate(),
		resources.NewDNS(),
		resources.NewHTTP(),
		resources.NewTLS(),
	}
}
