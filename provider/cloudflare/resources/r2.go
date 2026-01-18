// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/r2"

	"github.com/kopexa-grc/kspec/core"
)

// R2Bucket fetches R2 storage buckets.
type R2Bucket struct {
	client    *cloudflare.Client
	accountID string
}

// NewR2Bucket creates a new R2Bucket resource.
func NewR2Bucket(client *cloudflare.Client, accountID string) *R2Bucket {
	return &R2Bucket{client: client, accountID: accountID}
}

// Name returns the resource type name.
func (r *R2Bucket) Name() string {
	return "cloudflare_r2_bucket"
}

// Fetch retrieves all R2 buckets for the account.
func (r *R2Bucket) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	accountID, err := resolveAccountID(r.accountID, asset, "R2 bucket")
	if err != nil {
		return nil, err
	}

	result, err := r.client.R2.Buckets.List(ctx, r2.BucketListParams{
		AccountID: cloudflare.F(accountID),
	})
	if err != nil {
		// R2 might not be enabled for this account
		return nil, err
	}

	return itemsToResources(result.Buckets, accountID, r.addSecurityFields), nil
}

func (r *R2Bucket) addSecurityFields(bucket map[string]interface{}) {
	// Check storage class
	if storageClass, ok := bucket["storage_class"].(string); ok {
		bucket["is_standard_storage"] = storageClass == "Standard"
		bucket["is_infrequent_access"] = storageClass == "InfrequentAccess"
	}

	// R2 buckets are private by default
	// Public access requires explicit configuration via custom domains or r2.dev
	bucket["default_private"] = true
}
