// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package resources provides Hetzner Cloud resource types for security scanning.
package resources

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// Location fetches Hetzner Cloud locations.
type Location struct {
	client *hcloud.Client
}

// NewLocation creates a new Location resource.
func NewLocation(client *hcloud.Client) *Location {
	return &Location{client: client}
}

// Name returns the resource type name.
func (r *Location) Name() string {
	return "hcloud_location"
}

// Fetch retrieves all Hetzner Cloud locations.
func (r *Location) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	locations, err := r.client.Location.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(locations))

	for _, location := range locations {
		resource := make(core.Resource)
		resource["id"] = location.ID
		resource["name"] = location.Name
		resource["description"] = location.Description
		resource["country"] = location.Country
		resource["city"] = location.City
		resource["latitude"] = location.Latitude
		resource["longitude"] = location.Longitude
		resource["network_zone"] = string(location.NetworkZone)

		resources = append(resources, resource)
	}

	return resources, nil
}
