// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// Project represents a Hetzner Cloud project resource scanner.
type Project struct {
	client *hcloud.Client
}

// NewProject creates a new Project resource scanner.
func NewProject(client *hcloud.Client) *Project {
	return &Project{client: client}
}

// Name returns the resource type name.
func (r *Project) Name() string {
	return "hetzner_project"
}

// ChildResources implements core.ChildResourceProvider.
// Returns the child resource types that belong to a Hetzner project.
func (r *Project) ChildResources() []string {
	return []string{
		// Compute
		"hcloud_server",
		// Storage
		"hcloud_volume",
		// Networking
		"hcloud_network",
		"hcloud_floating_ip",
		"hcloud_primary_ip",
		// Security
		"hcloud_firewall",
		// SSH Keys
		"hcloud_ssh_key",
	}
}

// Fetch retrieves project-level information from Hetzner Cloud.
func (r *Project) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	resource := make(core.Resource)
	resource["id"] = asset.Name
	resource["name"] = asset.Name

	// Get resource counts for overview
	servers, err := r.client.Server.All(ctx)
	if err == nil {
		resource["server_count"] = len(servers)
	}

	volumes, err := r.client.Volume.All(ctx)
	if err == nil {
		resource["volume_count"] = len(volumes)
	}

	networks, err := r.client.Network.All(ctx)
	if err == nil {
		resource["network_count"] = len(networks)
	}

	firewalls, err := r.client.Firewall.All(ctx)
	if err == nil {
		resource["firewall_count"] = len(firewalls)
	}

	sshKeys, err := r.client.SSHKey.All(ctx)
	if err == nil {
		resource["ssh_key_count"] = len(sshKeys)
	}

	floatingIPs, err := r.client.FloatingIP.All(ctx)
	if err == nil {
		resource["floating_ip_count"] = len(floatingIPs)
	}

	primaryIPs, err := r.client.PrimaryIP.All(ctx)
	if err == nil {
		resource["primary_ip_count"] = len(primaryIPs)
	}

	return []core.Resource{resource}, nil
}
