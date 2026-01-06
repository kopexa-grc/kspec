package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// ServerResource fetches Hetzner Cloud servers.
type ServerResource struct {
	client *hcloud.Client
}

// Name returns the resource type name.
func (r *ServerResource) Name() string {
	return "hcloud_server"
}

// Fetch retrieves all Hetzner Cloud servers.
func (r *ServerResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	servers, err := r.client.Server.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(servers))

	for _, server := range servers {
		resource := make(core.Resource)
		resource["id"] = server.ID
		resource["name"] = server.Name
		resource["status"] = string(server.Status)
		resource["created"] = server.Created.String()
		resource["labels"] = server.Labels
		resource["locked"] = server.Locked
		resource["rescue_enabled"] = server.RescueEnabled
		resource["backup_window"] = server.BackupWindow
		resource["outgoing_traffic"] = server.OutgoingTraffic
		resource["ingoing_traffic"] = server.IngoingTraffic
		resource["included_traffic"] = server.IncludedTraffic
		resource["primary_disk_size"] = server.PrimaryDiskSize
		resource["load_balancers"] = len(server.LoadBalancers)

		// Protection
		if server.Protection.Delete {
			resource["delete_protection"] = true
		}
		if server.Protection.Rebuild {
			resource["rebuild_protection"] = true
		}

		// Server type info
		if server.ServerType != nil {
			resource["server_type"] = server.ServerType.Name
			resource["server_type_id"] = server.ServerType.ID
			resource["instance_cores"] = server.ServerType.Cores
			resource["instance_memory"] = server.ServerType.Memory
			resource["instance_disk"] = server.ServerType.Disk
			resource["cpu_type"] = string(server.ServerType.CPUType)
			resource["architecture"] = string(server.ServerType.Architecture)
		}

		// Location info (using Location instead of deprecated Datacenter)
		if server.Location != nil {
			resource["location"] = server.Location.Name
			resource["location_id"] = server.Location.ID
			resource["country"] = server.Location.Country
			resource["city"] = server.Location.City
			resource["network_zone"] = string(server.Location.NetworkZone)
		}

		// Image info
		if server.Image != nil {
			resource["image_id"] = server.Image.ID
			resource["image_name"] = server.Image.Name
			resource["image_type"] = string(server.Image.Type)
			resource["os_flavor"] = server.Image.OSFlavor
			resource["os_version"] = server.Image.OSVersion
		}

		// Public networking
		if server.PublicNet.IPv4.IP != nil {
			resource["public_ipv4"] = server.PublicNet.IPv4.IP.String()
			resource["ipv4_blocked"] = server.PublicNet.IPv4.Blocked
		}
		if server.PublicNet.IPv6.IP != nil {
			resource["public_ipv6"] = server.PublicNet.IPv6.IP.String()
			resource["ipv6_blocked"] = server.PublicNet.IPv6.Blocked
		}

		// Private networks
		var privateNetworks []map[string]interface{}
		for _, pn := range server.PrivateNet {
			pnMap := map[string]interface{}{
				"network_id":  pn.Network.ID,
				"ip":          pn.IP.String(),
				"mac_address": pn.MACAddress,
			}
			if len(pn.Aliases) > 0 {
				var aliases []string
				for _, a := range pn.Aliases {
					aliases = append(aliases, a.String())
				}
				pnMap["aliases"] = aliases
			}
			privateNetworks = append(privateNetworks, pnMap)
		}
		resource["private_networks"] = privateNetworks
		resource["private_network_count"] = len(privateNetworks)

		// Volumes
		var volumeIDs []int64
		for _, vol := range server.Volumes {
			volumeIDs = append(volumeIDs, vol.ID)
		}
		resource["volume_ids"] = volumeIDs
		resource["volume_count"] = len(volumeIDs)

		// Security flags
		resource["is_running"] = server.Status == hcloud.ServerStatusRunning
		resource["is_locked"] = server.Locked
		resource["has_backup_window"] = server.BackupWindow != ""
		resource["has_public_ipv4"] = server.PublicNet.IPv4.IP != nil
		resource["has_public_ipv6"] = server.PublicNet.IPv6.IP != nil
		resource["has_private_network"] = len(server.PrivateNet) > 0
		resource["has_volumes"] = len(server.Volumes) > 0
		resource["has_delete_protection"] = server.Protection.Delete
		resource["has_rebuild_protection"] = server.Protection.Rebuild

		resources = append(resources, resource)
	}

	return resources, nil
}
