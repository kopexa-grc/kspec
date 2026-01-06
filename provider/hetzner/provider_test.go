package hetzner

import (
	"context"
	"os"
	"testing"
)

func TestProvider_Name(t *testing.T) {
	p := NewProvider()
	if got := p.Name(); got != "hetzner" {
		t.Errorf("Name() = %v, want %v", got, "hetzner")
	}
}

func TestProvider_Connect_MissingCredentials(t *testing.T) {
	// Clear any environment variables
	os.Unsetenv("HCLOUD_TOKEN")
	os.Unsetenv("HETZNER_API_TOKEN")

	p := NewProvider()
	_, err := p.Connect(context.Background(), map[string]string{})

	if err == nil {
		t.Error("Connect() expected error for missing credentials")
	}
}

func TestProvider_Connect_WithEnvVar(t *testing.T) {
	// This test will fail to actually connect but should accept the token
	os.Setenv("HCLOUD_TOKEN", "test-token")
	defer os.Unsetenv("HCLOUD_TOKEN")

	p := NewProvider()
	_, err := p.Connect(context.Background(), map[string]string{})

	// Should fail on actual connection (invalid token) but not on missing credentials
	if err != nil && containsString(err.Error(), "no API token") {
		t.Error("Connect() should have accepted credentials from environment")
	}
}

func TestProvider_Connect_WithConfig(t *testing.T) {
	os.Unsetenv("HCLOUD_TOKEN")
	os.Unsetenv("HETZNER_API_TOKEN")

	p := NewProvider()
	_, err := p.Connect(context.Background(), map[string]string{
		"api_token": "test-token",
	})

	// Should fail on actual connection (invalid token) but not on missing credentials
	if err != nil && containsString(err.Error(), "no API token") {
		t.Error("Connect() should have accepted credentials from config")
	}
}

// Resource name tests
func TestLocationResource_Name(t *testing.T) {
	r := &LocationResource{}
	if got := r.Name(); got != "hcloud_location" {
		t.Errorf("Name() = %v, want hcloud_location", got)
	}
}

func TestDatacenterResource_Name(t *testing.T) {
	r := &DatacenterResource{}
	if got := r.Name(); got != "hcloud_datacenter" {
		t.Errorf("Name() = %v, want hcloud_datacenter", got)
	}
}

func TestServerTypeResource_Name(t *testing.T) {
	r := &ServerTypeResource{}
	if got := r.Name(); got != "hcloud_server_type" {
		t.Errorf("Name() = %v, want hcloud_server_type", got)
	}
}

func TestISOResource_Name(t *testing.T) {
	r := &ISOResource{}
	if got := r.Name(); got != "hcloud_iso" {
		t.Errorf("Name() = %v, want hcloud_iso", got)
	}
}

func TestServerResource_Name(t *testing.T) {
	r := &ServerResource{}
	if got := r.Name(); got != "hcloud_server" {
		t.Errorf("Name() = %v, want hcloud_server", got)
	}
}

func TestVolumeResource_Name(t *testing.T) {
	r := &VolumeResource{}
	if got := r.Name(); got != "hcloud_volume" {
		t.Errorf("Name() = %v, want hcloud_volume", got)
	}
}

func TestNetworkResource_Name(t *testing.T) {
	r := &NetworkResource{}
	if got := r.Name(); got != "hcloud_network" {
		t.Errorf("Name() = %v, want hcloud_network", got)
	}
}

func TestFloatingIPResource_Name(t *testing.T) {
	r := &FloatingIPResource{}
	if got := r.Name(); got != "hcloud_floating_ip" {
		t.Errorf("Name() = %v, want hcloud_floating_ip", got)
	}
}

func TestPrimaryIPResource_Name(t *testing.T) {
	r := &PrimaryIPResource{}
	if got := r.Name(); got != "hcloud_primary_ip" {
		t.Errorf("Name() = %v, want hcloud_primary_ip", got)
	}
}

func TestFirewallResource_Name(t *testing.T) {
	r := &FirewallResource{}
	if got := r.Name(); got != "hcloud_firewall" {
		t.Errorf("Name() = %v, want hcloud_firewall", got)
	}
}

func TestImageResource_Name(t *testing.T) {
	r := &ImageResource{}
	if got := r.Name(); got != "hcloud_image" {
		t.Errorf("Name() = %v, want hcloud_image", got)
	}
}

func TestSSHKeyResource_Name(t *testing.T) {
	r := &SSHKeyResource{}
	if got := r.Name(); got != "hcloud_ssh_key" {
		t.Errorf("Name() = %v, want hcloud_ssh_key", got)
	}
}

func TestConnection_Resources(t *testing.T) {
	conn := &Connection{
		client:      nil,
		projectName: "test-project",
	}

	resources := conn.Resources()

	expectedNames := []string{
		"hcloud_location",
		"hcloud_datacenter",
		"hcloud_server_type",
		"hcloud_iso",
		"hcloud_server",
		"hcloud_volume",
		"hcloud_network",
		"hcloud_floating_ip",
		"hcloud_primary_ip",
		"hcloud_firewall",
		"hcloud_image",
		"hcloud_ssh_key",
	}

	if len(resources) != len(expectedNames) {
		t.Errorf("Resources() returned %d resources, want %d", len(resources), len(expectedNames))
	}

	nameMap := make(map[string]bool)
	for _, r := range resources {
		nameMap[r.Name()] = true
	}

	for _, name := range expectedNames {
		if !nameMap[name] {
			t.Errorf("Resources() missing resource: %s", name)
		}
	}
}

func TestTruncateKey(t *testing.T) {
	tests := []struct {
		name   string
		key    string
		maxLen int
		want   string
	}{
		{
			name:   "short key",
			key:    "ssh-ed25519 AAAA",
			maxLen: 50,
			want:   "ssh-ed25519 AAAA",
		},
		{
			name:   "long key",
			key:    "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC1234567890abcdefghij",
			maxLen: 20,
			want:   "ssh-rsa AAAAB3NzaC1y...",
		},
		{
			name:   "exact length",
			key:    "12345678901234567890",
			maxLen: 20,
			want:   "12345678901234567890",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := truncateKey(tt.key, tt.maxLen); got != tt.want {
				t.Errorf("truncateKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper function
func containsString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
