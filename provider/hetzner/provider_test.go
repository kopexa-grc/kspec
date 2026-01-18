package hetzner

import (
	"context"
	"os"
	"testing"

	"github.com/kopexa-grc/kspec/provider/hetzner/resources"
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
func TestLocation_Name(t *testing.T) {
	r := resources.NewLocation(nil)
	if got := r.Name(); got != "hcloud_location" {
		t.Errorf("Name() = %v, want hcloud_location", got)
	}
}

func TestDatacenter_Name(t *testing.T) {
	r := resources.NewDatacenter(nil)
	if got := r.Name(); got != "hcloud_datacenter" {
		t.Errorf("Name() = %v, want hcloud_datacenter", got)
	}
}

func TestServerType_Name(t *testing.T) {
	r := resources.NewServerType(nil)
	if got := r.Name(); got != "hcloud_server_type" {
		t.Errorf("Name() = %v, want hcloud_server_type", got)
	}
}

func TestISO_Name(t *testing.T) {
	r := resources.NewISO(nil)
	if got := r.Name(); got != "hcloud_iso" {
		t.Errorf("Name() = %v, want hcloud_iso", got)
	}
}

func TestServer_Name(t *testing.T) {
	r := resources.NewServer(nil)
	if got := r.Name(); got != "hcloud_server" {
		t.Errorf("Name() = %v, want hcloud_server", got)
	}
}

func TestVolume_Name(t *testing.T) {
	r := resources.NewVolume(nil)
	if got := r.Name(); got != "hcloud_volume" {
		t.Errorf("Name() = %v, want hcloud_volume", got)
	}
}

func TestNetwork_Name(t *testing.T) {
	r := resources.NewNetwork(nil)
	if got := r.Name(); got != "hcloud_network" {
		t.Errorf("Name() = %v, want hcloud_network", got)
	}
}

func TestFloatingIP_Name(t *testing.T) {
	r := resources.NewFloatingIP(nil)
	if got := r.Name(); got != "hcloud_floating_ip" {
		t.Errorf("Name() = %v, want hcloud_floating_ip", got)
	}
}

func TestPrimaryIP_Name(t *testing.T) {
	r := resources.NewPrimaryIP(nil)
	if got := r.Name(); got != "hcloud_primary_ip" {
		t.Errorf("Name() = %v, want hcloud_primary_ip", got)
	}
}

func TestFirewall_Name(t *testing.T) {
	r := resources.NewFirewall(nil)
	if got := r.Name(); got != "hcloud_firewall" {
		t.Errorf("Name() = %v, want hcloud_firewall", got)
	}
}

func TestImage_Name(t *testing.T) {
	r := resources.NewImage(nil)
	if got := r.Name(); got != "hcloud_image" {
		t.Errorf("Name() = %v, want hcloud_image", got)
	}
}

func TestSSHKey_Name(t *testing.T) {
	r := resources.NewSSHKey(nil)
	if got := r.Name(); got != "hcloud_ssh_key" {
		t.Errorf("Name() = %v, want hcloud_ssh_key", got)
	}
}

func TestConnection_Resources(t *testing.T) {
	// Create a client wrapper with nil hcloud client for testing
	// The Resources() method requires a valid client to access HCloud()
	client := &Client{hc: nil, limiter: nil}
	conn := &Connection{
		client:      client,
		projectName: "test-project",
	}

	// Skip if we can't get resources without a real client
	// This tests only works with a mock or real client
	t.Skip("TestConnection_Resources requires a valid hcloud client")

	res := conn.Resources()

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

	if len(res) != len(expectedNames) {
		t.Errorf("Resources() returned %d resources, want %d", len(res), len(expectedNames))
	}

	nameMap := make(map[string]bool)
	for _, r := range res {
		nameMap[r.Name()] = true
	}

	for _, name := range expectedNames {
		if !nameMap[name] {
			t.Errorf("Resources() missing resource: %s", name)
		}
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
