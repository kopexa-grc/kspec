// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

// Package provider manages the registry of available security scanning providers.
package provider

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/kspec/core"
	_ "github.com/kopexa-grc/kspec/provider/all" // Import all providers to register them
	"github.com/kopexa-grc/kspec/provider/registry"
)

// GetProviders returns all available provider factories.
func GetProviders() []core.Provider {
	defs := registry.All()
	providers := make([]core.Provider, 0, len(defs))
	for _, def := range defs {
		providers = append(providers, def.Factory())
	}
	return providers
}

// GetProviderByName returns a specific provider by name or alias.
func GetProviderByName(name string) (core.Provider, error) {
	def, ok := registry.Get(name)
	if !ok {
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
	return def.Factory(), nil
}

// InitProvider initializes a single provider by name with the given config.
// Returns the connection and resource registry.
func InitProvider(ctx context.Context, providerName string, config map[string]string) (core.Connection, map[string]core.ResourceSpec, error) {
	provider, err := GetProviderByName(providerName)
	if err != nil {
		return nil, nil, err
	}

	conn, err := provider.Connect(ctx, config)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to connect to provider %s: %w", provider.Name(), err)
	}

	resourceRegistry := make(map[string]core.ResourceSpec)

	// Register all primary resources
	for _, r := range conn.Resources() {
		resourceRegistry[r.Name()] = r

		// Check if this resource provides sub-resources
		if subProvider, ok := r.(core.SubResourceProvider); ok {
			subResources := subProvider.SubResources()
			for _, subRes := range subResources {
				// Register sub-resources
				resourceRegistry[subRes.Name()] = subRes
			}
		}
	}

	return conn, resourceRegistry, nil
}

// InitProviders initializes all providers with the given config and returns a unified resource map.
// This is a helper for CLIs that want "all default providers".
//
// Deprecated: Use InitProvider for lazy-loading individual providers.
func InitProviders(ctx context.Context, config map[string]string) (map[string]core.ResourceSpec, error) {
	resourceRegistry := make(map[string]core.ResourceSpec)

	providers := GetProviders()
	for _, p := range providers {
		conn, err := p.Connect(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to provider %s: %w", p.Name(), err)
		}
		for _, r := range conn.Resources() {
			// Last writer wins for duplicate resource names
			resourceRegistry[r.Name()] = r
		}
	}

	return resourceRegistry, nil
}
