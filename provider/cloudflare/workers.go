package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/workers"
	"github.com/kopexa-grc/kspec/core"
)

// WorkerResource fetches Workers scripts.
type WorkerResource struct {
	client    *cloudflare.Client
	accountID string
}

// Name returns the resource type name.
func (r *WorkerResource) Name() string {
	return "cloudflare_worker"
}

// Fetch retrieves all Workers scripts for the account.
func (r *WorkerResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	accountID := r.accountID
	if accountID == "" {
		accountID = asset.Config["account_id"]
	}
	if accountID == "" {
		return nil, fmt.Errorf("account_id is required for Workers scanning")
	}

	var resources []core.Resource

	// List all Workers scripts
	result, err := r.client.Workers.Scripts.List(ctx, workers.ScriptListParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		return resources, nil
	}

	for _, script := range result.Result {
		data, err := json.Marshal(script)
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

	return resources, nil
}

func (r *WorkerResource) addComputedFields(worker map[string]interface{}) {
	// Check if worker has routes configured
	if routes, ok := worker["routes"].([]interface{}); ok {
		worker["has_routes"] = len(routes) > 0
		worker["route_count"] = len(routes)
	}

	// Check for environment bindings
	if bindings, ok := worker["bindings"].([]interface{}); ok {
		worker["has_bindings"] = len(bindings) > 0
		worker["binding_count"] = len(bindings)

		// Check for sensitive binding types
		hasSecrets := false
		hasKV := false
		hasR2 := false
		hasD1 := false

		for _, b := range bindings {
			if binding, ok := b.(map[string]interface{}); ok {
				if bindingType, ok := binding["type"].(string); ok {
					switch bindingType {
					case "secret_text":
						hasSecrets = true
					case "kv_namespace":
						hasKV = true
					case "r2_bucket":
						hasR2 = true
					case "d1":
						hasD1 = true
					}
				}
			}
		}

		worker["has_secrets"] = hasSecrets
		worker["has_kv_binding"] = hasKV
		worker["has_r2_binding"] = hasR2
		worker["has_d1_binding"] = hasD1
	}
}
