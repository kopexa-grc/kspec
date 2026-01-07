package cloudflare

import (
	"fmt"

	"github.com/kopexa-grc/kspec/core"
)

// resolveAccountID returns the account ID from the resource config or asset config.
// Returns an error if no account ID is found.
func resolveAccountID(resourceAccountID string, asset core.Asset, resourceName string) (string, error) {
	accountID := resourceAccountID
	if accountID == "" {
		accountID = asset.Config["account_id"]
	}
	if accountID == "" {
		return "", fmt.Errorf("account_id is required for %s scanning", resourceName)
	}
	return accountID, nil
}

// itemsToResources converts a slice of items to core.Resource objects via JSON marshaling.
// It adds the accountID to each resource and optionally applies a callback for additional fields.
func itemsToResources[T any](
	items []T,
	accountID string,
	addFields func(map[string]interface{}),
) []core.Resource {
	resources := make([]core.Resource, 0, len(items))

	for _, item := range items {
		resourceMap, err := core.ToResource(item)
		if err != nil {
			continue
		}

		resourceMap["account_id"] = accountID

		if addFields != nil {
			addFields(resourceMap)
		}

		resources = append(resources, resourceMap)
	}

	return resources
}
