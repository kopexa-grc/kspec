// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// Datacenter fetches Hetzner Cloud datacenters.
type Datacenter struct {
	client *hcloud.Client
}

// NewDatacenter creates a new Datacenter resource.
func NewDatacenter(client *hcloud.Client) *Datacenter {
	return &Datacenter{client: client}
}

// Name returns the resource type name.
func (r *Datacenter) Name() string {
	return "hcloud_datacenter"
}

// Fetch retrieves all Hetzner Cloud datacenters.
func (r *Datacenter) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	datacenters, err := r.client.Datacenter.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(datacenters))

	for _, dc := range datacenters {
		resource := make(core.Resource)
		resource["id"] = dc.ID
		resource["name"] = dc.Name
		resource["description"] = dc.Description

		// Location info
		if dc.Location != nil {
			resource["location_id"] = dc.Location.ID
			resource["location_name"] = dc.Location.Name
			resource["location_country"] = dc.Location.Country
			resource["location_city"] = dc.Location.City
			resource["network_zone"] = string(dc.Location.NetworkZone)
		}

		// Server types available in this datacenter
		var availableServerTypes []string
		var supportedServerTypes []string
		for _, st := range dc.ServerTypes.Available {
			availableServerTypes = append(availableServerTypes, st.Name)
		}
		for _, st := range dc.ServerTypes.Supported {
			supportedServerTypes = append(supportedServerTypes, st.Name)
		}
		resource["available_server_types"] = availableServerTypes
		resource["supported_server_types"] = supportedServerTypes
		resource["available_server_type_count"] = len(availableServerTypes)
		resource["supported_server_type_count"] = len(supportedServerTypes)

		resources = append(resources, resource)
	}

	return resources, nil
}
