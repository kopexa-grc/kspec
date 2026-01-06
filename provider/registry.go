// Package provider manages the registry of available security scanning providers.
package provider

import (
	"context"
	"fmt"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/atlassian"
	"github.com/kopexa-grc/kspec/provider/azure"
	"github.com/kopexa-grc/kspec/provider/cloudflare"
	"github.com/kopexa-grc/kspec/provider/github"
	"github.com/kopexa-grc/kspec/provider/hetzner"
	"github.com/kopexa-grc/kspec/provider/ms365"
	"github.com/kopexa-grc/kspec/provider/network"
	"github.com/kopexa-grc/kspec/provider/os"
	"github.com/kopexa-grc/kspec/provider/sbom"
)

// GetProviders returns all available provider factories
func GetProviders() []core.Provider {
	return []core.Provider{
		os.New(),
		network.NewProvider(),
		github.NewProvider(),
		azure.NewProvider(),
		ms365.NewProvider(),
		cloudflare.NewProvider(),
		atlassian.NewProvider(),
		sbom.NewProvider(),
		hetzner.NewProvider(),
	}
}

// GetProviderByName returns a specific provider by name
func GetProviderByName(name string) (core.Provider, error) {
	switch name {
	case "os", "local":
		return os.New(), nil
	case "network", "host":
		return network.NewProvider(), nil
	case "github":
		return github.NewProvider(), nil
	case "azure":
		return azure.NewProvider(), nil
	case "ms365", "microsoft365", "m365":
		return ms365.NewProvider(), nil
	case "cloudflare", "cf":
		return cloudflare.NewProvider(), nil
	case "atlassian", "jira", "confluence":
		return atlassian.NewProvider(), nil
	case "sbom", "bom", "cyclonedx", "spdx":
		return sbom.NewProvider(), nil
	case "hetzner", "hcloud":
		return hetzner.NewProvider(), nil
	default:
		return nil, fmt.Errorf("unknown provider: %s", name)
	}
}

// InitProvider initializes a single provider by name with the given config
func InitProvider(ctx context.Context, providerName string, config map[string]string) (map[string]core.ResourceSpec, error) {
	provider, err := GetProviderByName(providerName)
	if err != nil {
		return nil, err
	}

	conn, err := provider.Connect(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to provider %s: %w", provider.Name(), err)
	}

	registry := make(map[string]core.ResourceSpec)

	// Register all primary resources
	for _, r := range conn.Resources() {
		registry[r.Name()] = r

		// Check if this resource provides sub-resources
		if subProvider, ok := r.(core.SubResourceProvider); ok {
			subResources := subProvider.SubResources()
			for _, subRes := range subResources {
				// Register sub-resources
				registry[subRes.Name()] = subRes
			}
		}
	}

	return registry, nil
}

// InitProviders initializes all providers with the given config and returns a unified resource map.
// This is a helper for CLIs that want "all default providers".
//
// Deprecated: Use InitProvider for lazy-loading individual providers.
func InitProviders(ctx context.Context, config map[string]string) (map[string]core.ResourceSpec, error) {
	registry := make(map[string]core.ResourceSpec)

	providers := GetProviders()
	for _, p := range providers {
		conn, err := p.Connect(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to provider %s: %w", p.Name(), err)
		}
		for _, r := range conn.Resources() {
			// Last writer wins for duplicate resource names
			registry[r.Name()] = r
		}
	}

	return registry, nil
}
