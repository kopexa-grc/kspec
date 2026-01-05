package provider

import (
	"context"
	"fmt"

	"github.com/juliankoehn/kspec/core"
	"github.com/juliankoehn/kspec/provider/github"
	"github.com/juliankoehn/kspec/provider/network"
	"github.com/juliankoehn/kspec/provider/os"
)

// GetProviders returns all available provider factories
func GetProviders() []core.Provider {
	return []core.Provider{
		os.New(),
		network.NewNetworkProvider(),
		github.NewGithubProvider(),
	}
}

// InitProviders initializes all providers with the given config and returns a unified resource map.
// This is a helper for CLIs that want "all default providers".
func InitProviders(ctx context.Context, config map[string]string) (map[string]core.ResourceSpec, error) {
	registry := make(map[string]core.ResourceSpec)

	providers := GetProviders()
	for _, p := range providers {
		conn, err := p.Connect(ctx, config)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to provider %s: %w", p.Name(), err)
		}
		for _, r := range conn.Resources() {
			if _, exists := registry[r.Name()]; exists {
				// Warn or Error? For now, last writer wins or error?
				// Simple overwrite.
			}
			registry[r.Name()] = r
		}
	}

	return registry, nil
}
