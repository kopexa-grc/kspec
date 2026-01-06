package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// PrimaryIPResource fetches Hetzner Cloud primary IPs.
type PrimaryIPResource struct {
	client *hcloud.Client
}

// Name returns the resource type name.
func (r *PrimaryIPResource) Name() string {
	return "hcloud_primary_ip"
}

// Fetch retrieves all Hetzner Cloud primary IPs.
func (r *PrimaryIPResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	primaryIPs, err := r.client.PrimaryIP.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(primaryIPs))

	for _, pip := range primaryIPs {
		resource := make(core.Resource)
		resource["id"] = pip.ID
		resource["name"] = pip.Name
		resource["created"] = pip.Created.String()
		resource["labels"] = pip.Labels
		resource["type"] = string(pip.Type)
		resource["blocked"] = pip.Blocked
		resource["auto_delete"] = pip.AutoDelete
		resource["assignee_type"] = pip.AssigneeType
		resource["assignee_id"] = pip.AssigneeID

		// IP address
		resource["ip"] = pip.IP.String()
		if pip.Network != nil {
			resource["network"] = pip.Network.String()
		}

		// Location (using Location instead of deprecated Datacenter)
		if pip.Location != nil {
			resource["location"] = pip.Location.Name
			resource["location_id"] = pip.Location.ID
			resource["country"] = pip.Location.Country
			resource["city"] = pip.Location.City
		}

		// DNS PTR records
		var dnsPtrs []map[string]interface{}
		for ip, ptr := range pip.DNSPtr {
			dnsPtrs = append(dnsPtrs, map[string]interface{}{
				"ip":  ip,
				"ptr": ptr,
			})
		}
		resource["dns_ptr"] = dnsPtrs

		// Protection
		resource["delete_protection"] = pip.Protection.Delete

		// Security flags
		resource["is_ipv4"] = pip.Type == hcloud.PrimaryIPTypeIPv4
		resource["is_ipv6"] = pip.Type == hcloud.PrimaryIPTypeIPv6
		resource["is_blocked"] = pip.Blocked
		resource["is_assigned"] = pip.AssigneeID != 0
		resource["has_auto_delete"] = pip.AutoDelete
		resource["has_delete_protection"] = pip.Protection.Delete
		resource["has_labels"] = len(pip.Labels) > 0
		resource["has_dns_ptr"] = len(pip.DNSPtr) > 0

		resources = append(resources, resource)
	}

	return resources, nil
}
