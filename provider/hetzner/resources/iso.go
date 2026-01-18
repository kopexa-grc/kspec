// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// ISO fetches Hetzner Cloud ISO images.
type ISO struct {
	client *hcloud.Client
}

// NewISO creates a new ISO resource.
func NewISO(client *hcloud.Client) *ISO {
	return &ISO{client: client}
}

// Name returns the resource type name.
func (r *ISO) Name() string {
	return "hcloud_iso"
}

// Fetch retrieves all Hetzner Cloud ISO images.
func (r *ISO) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	isos, err := r.client.ISO.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(isos))

	for _, iso := range isos {
		resource := make(core.Resource)
		resource["id"] = iso.ID
		resource["name"] = iso.Name
		resource["description"] = iso.Description
		resource["type"] = string(iso.Type)
		if iso.Architecture != nil {
			resource["architecture"] = string(*iso.Architecture)
		}

		// Deprecation info
		resource["is_deprecated"] = iso.IsDeprecated()
		if iso.Deprecation != nil {
			resource["deprecation_announced"] = iso.Deprecation.Announced.String()
			resource["unavailable_after"] = iso.Deprecation.UnavailableAfter.String()
		}

		// Security flags
		resource["is_public"] = iso.Type == hcloud.ISOTypePublic
		resource["is_private"] = iso.Type == hcloud.ISOTypePrivate
		if iso.Architecture != nil {
			resource["is_arm"] = *iso.Architecture == hcloud.ArchitectureARM
			resource["is_x86"] = *iso.Architecture == hcloud.ArchitectureX86
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
