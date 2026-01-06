package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/r2"
	"github.com/kopexa-grc/kspec/core"
)

// R2BucketResource fetches R2 storage buckets.
type R2BucketResource struct {
	client    *cloudflare.Client
	accountID string
}

// Name returns the resource type name.
func (r *R2BucketResource) Name() string {
	return "cloudflare_r2_bucket"
}

// Fetch retrieves all R2 buckets for the account.
func (r *R2BucketResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	accountID := r.accountID
	if accountID == "" {
		accountID = asset.Config["account_id"]
	}
	if accountID == "" {
		return nil, fmt.Errorf("account_id is required for R2 bucket scanning")
	}

	var resources []core.Resource

	// List all R2 buckets
	result, err := r.client.R2.Buckets.List(ctx, r2.BucketListParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		// R2 might not be enabled for this account
		return resources, nil
	}

	for _, bucket := range result.Buckets {
		data, err := json.Marshal(bucket)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["account_id"] = accountID

		// Add security-relevant fields
		r.addSecurityFields(resourceMap)

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

func (r *R2BucketResource) addSecurityFields(bucket map[string]interface{}) {
	// Check storage class
	if storageClass, ok := bucket["storage_class"].(string); ok {
		bucket["is_standard_storage"] = storageClass == "Standard"
		bucket["is_infrequent_access"] = storageClass == "InfrequentAccess"
	}

	// R2 buckets are private by default
	// Public access requires explicit configuration via custom domains or r2.dev
	bucket["default_private"] = true
}
