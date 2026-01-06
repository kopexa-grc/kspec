package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// ImageResource fetches Hetzner Cloud images.
type ImageResource struct {
	client *hcloud.Client
}

// Name returns the resource type name.
func (r *ImageResource) Name() string {
	return "hcloud_image"
}

// Fetch retrieves all Hetzner Cloud images.
func (r *ImageResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	images, err := r.client.Image.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(images))

	for _, img := range images {
		resource := make(core.Resource)
		resource["id"] = img.ID
		resource["name"] = img.Name
		resource["description"] = img.Description
		resource["created"] = img.Created.String()
		resource["labels"] = img.Labels
		resource["type"] = string(img.Type)
		resource["status"] = string(img.Status)
		resource["os_flavor"] = img.OSFlavor
		resource["os_version"] = img.OSVersion
		resource["architecture"] = string(img.Architecture)
		resource["disk_size"] = img.DiskSize
		resource["image_size"] = img.ImageSize
		resource["rapid_deploy"] = img.RapidDeploy

		// Deprecation info (img.Deprecated is a time.Time)
		resource["is_deprecated"] = img.IsDeprecated()
		if !img.Deprecated.IsZero() {
			resource["deprecated_at"] = img.Deprecated.String()
		}

		// Created from info
		if img.CreatedFrom != nil {
			resource["created_from_id"] = img.CreatedFrom.ID
			resource["created_from_name"] = img.CreatedFrom.Name
		}

		// Bound to server
		if img.BoundTo != nil {
			resource["bound_to"] = img.BoundTo.ID
			resource["is_bound"] = true
		} else {
			resource["is_bound"] = false
		}

		// Protection
		resource["delete_protection"] = img.Protection.Delete

		// Security flags
		resource["is_system"] = img.Type == hcloud.ImageTypeSystem
		resource["is_app"] = img.Type == hcloud.ImageTypeApp
		resource["is_snapshot"] = img.Type == hcloud.ImageTypeSnapshot
		resource["is_backup"] = img.Type == hcloud.ImageTypeBackup
		resource["is_available"] = img.Status == hcloud.ImageStatusAvailable
		resource["is_creating"] = img.Status == hcloud.ImageStatusCreating
		resource["is_arm"] = img.Architecture == hcloud.ArchitectureARM
		resource["is_x86"] = img.Architecture == hcloud.ArchitectureX86
		resource["has_rapid_deploy"] = img.RapidDeploy
		resource["has_delete_protection"] = img.Protection.Delete
		resource["has_labels"] = len(img.Labels) > 0

		resources = append(resources, resource)
	}

	return resources, nil
}
