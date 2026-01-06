package cloudflare

import (
	"context"

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
	accountID, err := resolveAccountID(r.accountID, asset, "Workers")
	if err != nil {
		return nil, err
	}

	result, err := r.client.Workers.Scripts.List(ctx, workers.ScriptListParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		return nil, err
	}

	return itemsToResources(result.Result, accountID, r.addComputedFields), nil
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
