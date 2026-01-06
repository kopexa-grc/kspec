package cloudflare

import (
	"context"
	"encoding/json"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/accounts"

	"github.com/kopexa-grc/kspec/core"
)

// AccountResource fetches Cloudflare account information.
type AccountResource struct {
	client *cloudflare.Client
}

// Name returns the resource type name.
func (r *AccountResource) Name() string {
	return "cloudflare_account"
}

// Fetch retrieves all accessible Cloudflare accounts.
func (r *AccountResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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
