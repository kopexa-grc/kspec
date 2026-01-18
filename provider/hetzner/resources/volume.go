// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// Volume fetches Hetzner Cloud volumes.
type Volume struct {
	client *hcloud.Client
}

// NewVolume creates a new Volume resource.
func NewVolume(client *hcloud.Client) *Volume {
	return &Volume{client: client}
}

// Name returns the resource type name.
func (r *Volume) Name() string {
	return "hcloud_volume"
}

// Fetch retrieves all Hetzner Cloud volumes.
func (r *Volume) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	volumes, err := r.client.Volume.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(volumes))

	for _, volume := range volumes {
		resource := make(core.Resource)
		resource["id"] = volume.ID
		resource["name"] = volume.Name
		resource["status"] = string(volume.Status)
		resource["created"] = volume.Created.String()
		resource["labels"] = volume.Labels
		resource["size"] = volume.Size
		resource["linux_device"] = volume.LinuxDevice
		resource["format"] = volume.Format

		// Location info
		if volume.Location != nil {
			resource["location"] = volume.Location.Name
			resource["location_id"] = volume.Location.ID
			resource["country"] = volume.Location.Country
			resource["city"] = volume.Location.City
		}

		// Server attachment
		if volume.Server != nil {
			resource["server_id"] = volume.Server.ID
			resource["is_attached"] = true
		} else {
			resource["is_attached"] = false
		}

		// Protection
		resource["delete_protection"] = volume.Protection.Delete

		// Security flags
		resource["is_available"] = volume.Status == hcloud.VolumeStatusAvailable
		resource["is_creating"] = volume.Status == hcloud.VolumeStatusCreating
		resource["has_delete_protection"] = volume.Protection.Delete
		resource["has_labels"] = len(volume.Labels) > 0

		resources = append(resources, resource)
	}

	return resources, nil
}
