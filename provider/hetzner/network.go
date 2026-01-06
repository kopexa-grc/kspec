package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// NetworkResource fetches Hetzner Cloud networks.
type NetworkResource struct {
	client *hcloud.Client
}

// Name returns the resource type name.
func (r *NetworkResource) Name() string {
	return "hcloud_network"
}

// Fetch retrieves all Hetzner Cloud networks.
func (r *NetworkResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	networks, err := r.client.Network.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(networks))

	for _, network := range networks {
		resource := make(core.Resource)
		resource["id"] = network.ID
		resource["name"] = network.Name
		resource["created"] = network.Created.String()
		resource["labels"] = network.Labels
		resource["ip_range"] = network.IPRange.String()
		resource["expose_routes_to_vswitch"] = network.ExposeRoutesToVSwitch

		// Subnets
		var subnets []map[string]interface{}
		for _, subnet := range network.Subnets {
			subnetMap := map[string]interface{}{
				"type":         string(subnet.Type),
				"ip_range":     subnet.IPRange.String(),
				"network_zone": string(subnet.NetworkZone),
				"gateway":      subnet.Gateway.String(),
			}
			if subnet.VSwitchID != 0 {
				subnetMap["vswitch_id"] = subnet.VSwitchID
			}
			subnets = append(subnets, subnetMap)
		}
		resource["subnets"] = subnets
		resource["subnet_count"] = len(subnets)

		// Routes
		var routes []map[string]interface{}
		for _, route := range network.Routes {
			routeMap := map[string]interface{}{
				"destination": route.Destination.String(),
				"gateway":     route.Gateway.String(),
			}
			routes = append(routes, routeMap)
		}
		resource["routes"] = routes
		resource["route_count"] = len(routes)

		// Servers in this network
		var serverIDs []int64
		for _, server := range network.Servers {
			serverIDs = append(serverIDs, server.ID)
		}
		resource["server_ids"] = serverIDs
		resource["server_count"] = len(serverIDs)

		// Load balancers in this network
		var lbIDs []int64
		for _, lb := range network.LoadBalancers {
			lbIDs = append(lbIDs, lb.ID)
		}
		resource["load_balancer_ids"] = lbIDs
		resource["load_balancer_count"] = len(lbIDs)

		// Protection
		resource["delete_protection"] = network.Protection.Delete

		// Security flags
		resource["has_subnets"] = len(network.Subnets) > 0
		resource["has_routes"] = len(network.Routes) > 0
		resource["has_servers"] = len(network.Servers) > 0
		resource["has_load_balancers"] = len(network.LoadBalancers) > 0
		resource["has_delete_protection"] = network.Protection.Delete
		resource["has_labels"] = len(network.Labels) > 0

		resources = append(resources, resource)
	}

	return resources, nil
}
