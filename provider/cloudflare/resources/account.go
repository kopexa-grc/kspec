// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/accounts"

	"github.com/kopexa-grc/kspec/core"
)

// Account fetches Cloudflare account information.
type Account struct {
	client *cloudflare.Client
}

// NewAccount creates a new Account resource.
func NewAccount(client *cloudflare.Client) *Account {
	return &Account{client: client}
}

// Name returns the resource type name.
func (r *Account) Name() string {
	return "cloudflare_account"
}

// Fetch retrieves all accessible Cloudflare accounts.
func (r *Account) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// List all accounts accessible to this API token
	iter := r.client.Accounts.ListAutoPaging(ctx, accounts.AccountListParams{})
	for iter.Next() {
		account := iter.Current()

		data, err := json.Marshal(account)
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
