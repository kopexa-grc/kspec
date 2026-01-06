package cloudflare

import (
	"context"
	"encoding/json"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zones"
	"github.com/kopexa-grc/kspec/core"
)

// ZoneSettingsResource fetches zone security settings.
type ZoneSettingsResource struct {
	client *cloudflare.Client
}

// Name returns the resource type name.
func (r *ZoneSettingsResource) Name() string {
	return "cloudflare_zone_settings"
}

// Fetch retrieves security settings for a zone.
func (r *ZoneSettingsResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	zoneID := asset.Config["zone_id"]
	if zoneID == "" {
		return nil, nil
	}

	// Build a consolidated settings map
	settingsMap := make(map[string]interface{})
	settingsMap["zone_id"] = zoneID

	// Fetch zone to get its settings
	zone, err := r.client.Zones.Get(ctx, zones.ZoneGetParams{
		ZoneID: cloudflare.F(zoneID),
	})
	if err != nil {
		return nil, err
	}

	// Add zone-level settings
	data, err := json.Marshal(zone)
	if err == nil {
		var zoneData map[string]interface{}
		if json.Unmarshal(data, &zoneData) == nil {
			// Copy relevant settings
			for key, val := range zoneData {
				settingsMap[key] = val
			}
		}
	}

	// Add computed security flags
	r.addSecurityFlags(settingsMap)

	return []core.Resource{settingsMap}, nil
}

func (r *ZoneSettingsResource) addSecurityFlags(settings map[string]interface{}) {
	// Check zone status
	if status, ok := settings["status"].(string); ok {
		settings["is_active"] = status == "active"
		settings["is_pending"] = status == "pending"
	}

	// Check zone plan
	if plan, ok := settings["plan"].(map[string]interface{}); ok {
		if name, ok := plan["name"].(string); ok {
			settings["plan_name"] = name
			settings["is_free_plan"] = name == "Free Website"
			settings["is_pro_plan"] = name == "Pro Website"
			settings["is_business_plan"] = name == "Business Website"
			settings["is_enterprise_plan"] = name == "Enterprise Website"
		}
	}

	// Check if zone is paused
	if paused, ok := settings["paused"].(bool); ok {
		settings["is_paused"] = paused
	}

	// Check development mode
	if devMode, ok := settings["development_mode"].(float64); ok {
		settings["development_mode_enabled"] = devMode > 0
	}
}
