package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// LocationResource fetches Hetzner Cloud locations.
type LocationResource struct {
	client *hcloud.Client
}

// Name returns the resource type name.
func (r *LocationResource) Name() string {
	return "hcloud_location"
}

// Fetch retrieves all Hetzner Cloud locations.
func (r *LocationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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
