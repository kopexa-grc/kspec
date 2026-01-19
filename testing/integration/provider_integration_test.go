// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package integration_test

import (
	"testing"

	"github.com/kopexa-grc/kspec/provider"

	// Import all providers to trigger registration for tests.
	_ "github.com/kopexa-grc/kspec/provider/all"
)

func TestGetProviders(t *testing.T) {
	providers := provider.GetProviders()

	if len(providers) == 0 {
		t.Error("GetProviders() returned empty list")
	}

	// Check that all providers have names
	for _, p := range providers {
		if p.Name() == "" {
			t.Error("Provider has empty name")
		}
	}

	// Check for expected providers
	expectedProviders := []string{
		"os",
		"network",
		"github",
		"azure",
		"ms365",
		"cloudflare",
		"atlassian",
		"sbom",
		"hetzner",
		"factorial",
	}

	providerNames := make(map[string]bool)
	for _, p := range providers {
		providerNames[p.Name()] = true
	}

	for _, expected := range expectedProviders {
		if !providerNames[expected] {
			t.Errorf("GetProviders() missing expected provider %q", expected)
		}
	}
}

func TestGetProviderByName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantName string
		wantErr  bool
	}{
		// OS provider
		{name: "os provider", input: "os", wantName: "os", wantErr: false},
		{name: "local alias", input: "local", wantName: "os", wantErr: false},

		// Network provider
		{name: "network provider", input: "network", wantName: "network", wantErr: false},
		{name: "host alias", input: "host", wantName: "network", wantErr: false},

		// GitHub provider
		{name: "github provider", input: "github", wantName: "github", wantErr: false},

		// Azure provider
		{name: "azure provider", input: "azure", wantName: "azure", wantErr: false},

		// MS365 provider
		{name: "ms365 provider", input: "ms365", wantName: "ms365", wantErr: false},
		{name: "microsoft365 alias", input: "microsoft365", wantName: "ms365", wantErr: false},
		{name: "m365 alias", input: "m365", wantName: "ms365", wantErr: false},

		// Cloudflare provider
		{name: "cloudflare provider", input: "cloudflare", wantName: "cloudflare", wantErr: false},
		{name: "cf alias", input: "cf", wantName: "cloudflare", wantErr: false},

		// Atlassian provider
		{name: "atlassian provider", input: "atlassian", wantName: "atlassian", wantErr: false},
		{name: "jira alias", input: "jira", wantName: "atlassian", wantErr: false},
		{name: "confluence alias", input: "confluence", wantName: "atlassian", wantErr: false},

		// SBOM provider
		{name: "sbom provider", input: "sbom", wantName: "sbom", wantErr: false},
		{name: "bom alias", input: "bom", wantName: "sbom", wantErr: false},
		{name: "cyclonedx alias", input: "cyclonedx", wantName: "sbom", wantErr: false},
		{name: "spdx alias", input: "spdx", wantName: "sbom", wantErr: false},

		// Hetzner provider
		{name: "hetzner provider", input: "hetzner", wantName: "hetzner", wantErr: false},
		{name: "hcloud alias", input: "hcloud", wantName: "hetzner", wantErr: false},

		// Factorial provider
		{name: "factorial provider", input: "factorial", wantName: "factorial", wantErr: false},
		{name: "factorial-hr alias", input: "factorial-hr", wantName: "factorial", wantErr: false},
		{name: "hris alias", input: "hris", wantName: "factorial", wantErr: false},

		// Unknown provider
		{name: "unknown provider", input: "unknown", wantName: "", wantErr: true},
		{name: "empty provider", input: "", wantName: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := provider.GetProviderByName(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("GetProviderByName(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("GetProviderByName(%q) unexpected error: %v", tt.input, err)
				return
			}

			if p == nil {
				t.Errorf("GetProviderByName(%q) returned nil provider", tt.input)
				return
			}

			if p.Name() != tt.wantName {
				t.Errorf("GetProviderByName(%q).Name() = %q, want %q", tt.input, p.Name(), tt.wantName)
			}
		})
	}
}

func TestGetProviderByName_Uniqueness(t *testing.T) {
	// Test that the same provider is returned for aliases
	tests := []struct {
		name    string
		aliases []string
	}{
		{name: "os aliases", aliases: []string{"os", "local"}},
		{name: "network aliases", aliases: []string{"network", "host"}},
		{name: "ms365 aliases", aliases: []string{"ms365", "microsoft365", "m365"}},
		{name: "cloudflare aliases", aliases: []string{"cloudflare", "cf"}},
		{name: "atlassian aliases", aliases: []string{"atlassian", "jira", "confluence"}},
		{name: "sbom aliases", aliases: []string{"sbom", "bom", "cyclonedx", "spdx"}},
		{name: "hetzner aliases", aliases: []string{"hetzner", "hcloud"}},
		{name: "factorial aliases", aliases: []string{"factorial", "factorial-hr", "hris"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if len(tt.aliases) < 2 {
				return
			}

			// Get first provider
			p1, err := provider.GetProviderByName(tt.aliases[0])
			if err != nil {
				t.Fatalf("GetProviderByName(%q) error: %v", tt.aliases[0], err)
			}

			// All aliases should return provider with same name
			for _, alias := range tt.aliases[1:] {
				p2, err := provider.GetProviderByName(alias)
				if err != nil {
					t.Fatalf("GetProviderByName(%q) error: %v", alias, err)
				}

				if p1.Name() != p2.Name() {
					t.Errorf("GetProviderByName(%q).Name() = %q, want %q (same as %q)",
						alias, p2.Name(), p1.Name(), tt.aliases[0])
				}
			}
		})
	}
}

func TestGetProviders_Count(t *testing.T) {
	providers := provider.GetProviders()

	// We expect at least 10 providers (os, network, github, azure, ms365, cloudflare, atlassian, sbom, hetzner, factorial)
	minExpected := 10
	if len(providers) < minExpected {
		t.Errorf("GetProviders() returned %d providers, want at least %d", len(providers), minExpected)
	}
}

func TestGetProviders_NoDuplicates(t *testing.T) {
	providers := provider.GetProviders()

	seen := make(map[string]bool)
	for _, p := range providers {
		name := p.Name()
		if seen[name] {
			t.Errorf("GetProviders() contains duplicate provider %q", name)
		}
		seen[name] = true
	}
}
