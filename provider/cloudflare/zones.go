package cloudflare

import (
	"context"
	"encoding/json"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zones"

	"github.com/kopexa-grc/kspec/core"
)

// ZoneResource fetches Cloudflare zones (domains).
type ZoneResource struct {
	client    *cloudflare.Client
	accountID string
}

// Name returns the resource type name.
func (r *ZoneResource) Name() string {
	return "cloudflare_zone"
}

// Fetch retrieves all zones for the account.
func (r *ZoneResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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

func (r *ZoneResource) fetchSingleZone(ctx context.Context, zoneID string) ([]core.Resource, error) {
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
func (r *ZoneResource) SubResources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&ZoneSettingsResource{client: r.client},
		&DNSRecordResource{client: r.client},
		&WAFRuleResource{client: r.client},
		&FirewallRuleResource{client: r.client},
	}
}
