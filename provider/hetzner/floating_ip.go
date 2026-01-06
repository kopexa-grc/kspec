package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// FloatingIPResource fetches Hetzner Cloud floating IPs.
type FloatingIPResource struct {
	client *hcloud.Client
}

// Name returns the resource type name.
func (r *FloatingIPResource) Name() string {
	return "hcloud_floating_ip"
}

// Fetch retrieves all Hetzner Cloud floating IPs.
func (r *FloatingIPResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	floatingIPs, err := r.client.FloatingIP.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(floatingIPs))

	for _, fip := range floatingIPs {
		resource := make(core.Resource)
		resource["id"] = fip.ID
		resource["name"] = fip.Name
		resource["description"] = fip.Description
		resource["created"] = fip.Created.String()
		resource["labels"] = fip.Labels
		resource["type"] = string(fip.Type)
		resource["blocked"] = fip.Blocked

		// IP address
		resource["ip"] = fip.IP.String()
		if fip.Network != nil {
			resource["network"] = fip.Network.String()
		}

		// Location
		if fip.HomeLocation != nil {
			resource["home_location"] = fip.HomeLocation.Name
			resource["home_location_id"] = fip.HomeLocation.ID
			resource["country"] = fip.HomeLocation.Country
		}

		// Server assignment
		if fip.Server != nil {
			resource["server_id"] = fip.Server.ID
			resource["is_assigned"] = true
		} else {
			resource["is_assigned"] = false
		}

		// DNS PTR records
		var dnsPtrs []map[string]interface{}
		for ip, ptr := range fip.DNSPtr {
			dnsPtrs = append(dnsPtrs, map[string]interface{}{
				"ip":  ip,
				"ptr": ptr,
			})
		}
		resource["dns_ptr"] = dnsPtrs

		// Protection
		resource["delete_protection"] = fip.Protection.Delete

		// Security flags
		resource["is_ipv4"] = fip.Type == hcloud.FloatingIPTypeIPv4
		resource["is_ipv6"] = fip.Type == hcloud.FloatingIPTypeIPv6
		resource["is_blocked"] = fip.Blocked
		resource["has_delete_protection"] = fip.Protection.Delete
		resource["has_labels"] = len(fip.Labels) > 0
		resource["has_dns_ptr"] = len(fip.DNSPtr) > 0

		resources = append(resources, resource)
	}

	return resources, nil
}
