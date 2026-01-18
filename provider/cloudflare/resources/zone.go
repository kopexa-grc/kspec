// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zones"

	"github.com/kopexa-grc/kspec/core"
)

// Zone fetches Cloudflare zones (domains).
type Zone struct {
	client    *cloudflare.Client
	accountID string
}

// NewZone creates a new Zone resource.
func NewZone(client *cloudflare.Client, accountID string) *Zone {
	return &Zone{client: client, accountID: accountID}
}

// Name returns the resource type name.
func (r *Zone) Name() string {
	return "cloudflare_zone"
}

// Fetch retrieves all zones for the account.
func (r *Zone) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// Check if scanning a specific zone
	zoneID := asset.Config["zone_id"]
	if zoneID != "" {
		return r.fetchSingleZone(ctx, zoneID)
	}

	// List all zones
	params := zones.ZoneListParams{}
	if r.accountID != "" {
		params.Account = cloudflare.F(zones.ZoneListParamsAccount{
			ID: cloudflare.F(r.accountID),
		})
	}

	iter := r.client.Zones.ListAutoPaging(ctx, params)
	for iter.Next() {
		zone := iter.Current()

		data, err := json.Marshal(zone)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resources = append(resources, resourceMap)
	}

	if err := iter.Err(); err != nil {
		return nil, err
	}

	return resources, nil
}

func (r *Zone) fetchSingleZone(ctx context.Context, zoneID string) ([]core.Resource, error) {
	zone, err := r.client.Zones.Get(ctx, zones.ZoneGetParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(zone)
	if err != nil {
		return nil, err
	}

	var resourceMap map[string]interface{}
	if err := json.Unmarshal(data, &resourceMap); err != nil {
		return nil, err
	}

	return []core.Resource{resourceMap}, nil
}

// SubResources returns dependent resources for zones.
func (r *Zone) SubResources() []core.ResourceSpec {
	return []core.ResourceSpec{
		NewZoneSettings(r.client),
		NewDNSRecord(r.client),
		NewWAFRule(r.client),
		NewFirewallRule(r.client),
	}
}

// BuildChildAsset implements core.AssetContextProvider.
// Passes the zone_id to child resources.
func (r *Zone) BuildChildAsset(parentData core.Resource, baseAsset core.Asset) core.Asset {
	asset := core.Asset{
		Type:   baseAsset.Type,
		Name:   baseAsset.Name,
		Config: make(map[string]string),
	}

	// Copy base config
	for k, v := range baseAsset.Config {
		asset.Config[k] = v
	}

	// Add zone_id from parent data for sub-resource fetching
	if id, ok := parentData["id"]; ok {
		if idStr, ok := id.(string); ok {
			asset.Config["zone_id"] = idStr
		}
	}

	return asset
}
