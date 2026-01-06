package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/kopexa-grc/kspec/core"
)

// TunnelResource fetches Cloudflare Tunnels.
type TunnelResource struct {
	client    *cloudflare.Client
	accountID string
}

// Name returns the resource type name.
func (r *TunnelResource) Name() string {
	return "cloudflare_tunnel"
}

// Fetch retrieves all Cloudflare Tunnels for the account.
func (r *TunnelResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	accountID := r.accountID
	if accountID == "" {
		accountID = asset.Config["account_id"]
	}
	if accountID == "" {
		return nil, fmt.Errorf("account_id is required for Tunnel scanning")
	}

	var resources []core.Resource

	// List all tunnels
	iter := r.client.ZeroTrust.Tunnels.ListAutoPaging(ctx, zero_trust.TunnelListParams{
		AccountID: cloudflare.F(accountID),
	})

	for iter.Next() {
		tunnel := iter.Current()

		data, err := json.Marshal(tunnel)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["account_id"] = accountID

		// Add computed fields
		r.addComputedFields(resourceMap)

		resources = append(resources, resourceMap)
	}

	if err := iter.Err(); err != nil {
		return resources, nil
	}

	return resources, nil
}

func (r *TunnelResource) addComputedFields(tunnel map[string]interface{}) {
	// Check tunnel status
	if status, ok := tunnel["status"].(string); ok {
		tunnel["is_healthy"] = status == "healthy"
		tunnel["is_degraded"] = status == "degraded"
		tunnel["is_down"] = status == "down" || status == "inactive"
	}

	// Check tunnel type
	if tunnelType, ok := tunnel["tun_type"].(string); ok {
		tunnel["is_cfd_tunnel"] = tunnelType == "cfd_tunnel"
		tunnel["is_warp_connector"] = tunnelType == "warp_connector"
	}

	// Check connections
	if connections, ok := tunnel["connections"].([]interface{}); ok {
		tunnel["connection_count"] = len(connections)
		tunnel["has_connections"] = len(connections) > 0

		// Count healthy connections
		healthyCount := 0
		for _, conn := range connections {
			if c, ok := conn.(map[string]interface{}); ok {
				if c["is_pending_reconnect"] == false {
					healthyCount++
				}
			}
		}
		tunnel["healthy_connection_count"] = healthyCount
	}

	// Check if remotely managed
	if remoteConfig, ok := tunnel["remote_config"].(bool); ok {
		tunnel["is_remotely_managed"] = remoteConfig
	}
}
