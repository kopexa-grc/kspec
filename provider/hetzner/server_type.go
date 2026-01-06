package hetzner

import (
	"context"

	"github.com/hetznercloud/hcloud-go/v2/hcloud"

	"github.com/kopexa-grc/kspec/core"
)

// ServerTypeResource fetches Hetzner Cloud server types.
type ServerTypeResource struct {
	client *hcloud.Client
}

// Name returns the resource type name.
func (r *ServerTypeResource) Name() string {
	return "hcloud_server_type"
}

// Fetch retrieves all Hetzner Cloud server types.
func (r *ServerTypeResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	serverTypes, err := r.client.ServerType.All(ctx)
	if err != nil {
		return nil, err
	}

	resources := make([]core.Resource, 0, len(serverTypes))

	for _, st := range serverTypes {
		resource := make(core.Resource)
		resource["id"] = st.ID
		resource["name"] = st.Name
		resource["description"] = st.Description
		resource["cores"] = st.Cores
		resource["memory"] = st.Memory
		resource["disk"] = st.Disk
		resource["storage_type"] = string(st.StorageType)
		resource["cpu_type"] = string(st.CPUType)
		resource["architecture"] = string(st.Architecture)
		// Note: IncludedTraffic is deprecated, use Pricings for traffic info

		// Deprecation info
		resource["is_deprecated"] = st.IsDeprecated()
		if st.Deprecation != nil {
			resource["deprecation_announced"] = st.Deprecation.Announced.String()
			resource["unavailable_after"] = st.Deprecation.UnavailableAfter.String()
		}

		// Pricing
		var prices []map[string]interface{}
		for _, price := range st.Pricings {
			priceMap := map[string]interface{}{
				"location":       price.Location.Name,
				"price_hourly":   price.Hourly.Gross,
				"price_monthly":  price.Monthly.Gross,
				"hourly_net":     price.Hourly.Net,
				"monthly_net":    price.Monthly.Net,
				"per_tb_traffic": price.PerTBTraffic.Gross,
			}
			prices = append(prices, priceMap)
		}
		resource["prices"] = prices

		// Security flags
		resource["is_arm"] = st.Architecture == hcloud.ArchitectureARM
		resource["is_x86"] = st.Architecture == hcloud.ArchitectureX86
		resource["is_shared"] = st.CPUType == hcloud.CPUTypeShared
		resource["is_dedicated"] = st.CPUType == hcloud.CPUTypeDedicated
		resource["has_local_storage"] = st.StorageType == hcloud.StorageTypeLocal
		resource["has_ceph_storage"] = st.StorageType == hcloud.StorageTypeCeph

		resources = append(resources, resource)
	}

	return resources, nil
}
