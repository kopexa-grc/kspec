package hetzner

import (
	"context"
	"fmt"
	"os"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// Provider implements the core.Provider interface for Hetzner Cloud.
type Provider struct{}

// NewProvider creates a new Hetzner Cloud provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "hetzner"
}

// Connect establishes a connection to Hetzner Cloud API.
func (p *Provider) Connect(ctx context.Context, config map[string]string) (core.Connection, error) {
	// Get API token from config or environment
	apiToken := config["api_token"]
	if apiToken == "" {
		apiToken = os.Getenv("HCLOUD_TOKEN")
	}
	if apiToken == "" {
		apiToken = os.Getenv("HETZNER_API_TOKEN")
	}

	if apiToken == "" {
		return nil, fmt.Errorf("hetzner: no API token provided. Set api_token, HCLOUD_TOKEN, or HETZNER_API_TOKEN")
	}

	// Get optional project name for identification
	projectName := config["project"]
	if projectName == "" {
		projectName = os.Getenv("HCLOUD_PROJECT")
	}
	if projectName == "" {
		projectName = "default"
	}

	// Create Hetzner Cloud client
	client := hcloud.NewClient(hcloud.WithToken(apiToken))

	// Verify connection by fetching locations
	_, err := client.Location.All(ctx)
	if err != nil {
		return nil, fmt.Errorf("hetzner: failed to connect: %w", err)
	}

	return &Connection{
		client:      client,
		projectName: projectName,
	}, nil
}

// Connection represents an active connection to Hetzner Cloud.
type Connection struct {
	client      *hcloud.Client
	projectName string
}

// Resources returns all available Hetzner Cloud resources.
func (c *Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		// Infrastructure
		&LocationResource{client: c.client},
		&DatacenterResource{client: c.client},
		&ServerTypeResource{client: c.client},
		&ISOResource{client: c.client},
		// Compute
		&ServerResource{client: c.client},
		// Storage
		&VolumeResource{client: c.client},
		// Networking
		&NetworkResource{client: c.client},
		&FloatingIPResource{client: c.client},
		&PrimaryIPResource{client: c.client},
		// Security
		&FirewallResource{client: c.client},
		// Images
		&ImageResource{client: c.client},
		// SSH Keys
		&SSHKeyResource{client: c.client},
	}
}
