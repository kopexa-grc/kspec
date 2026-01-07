package azure

import (
	"context"

	"github.com/kopexa-grc/kspec/core"
)

// Pager represents an Azure SDK pager interface for listing resources.
type Pager[T any] interface {
	More() bool
	NextPage(ctx context.Context) (T, error)
}

// PageExtractor extracts individual items from a page result.
type PageExtractor[TPage any, TItem any] func(page TPage) []*TItem

// fetchWithPager is a generic helper for fetching Azure resources using a pager.
// It handles the common pattern of iterating through pages and converting
// each item to a core.Resource via JSON marshaling.
func fetchWithPager[TPage any, TItem any](
	ctx context.Context,
	pager Pager[TPage],
	extractor PageExtractor[TPage, TItem],
	resourceType string,
) ([]core.Resource, error) {
	var resources []core.Resource

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		items := extractor(page)
		for _, item := range items {
			resource, err := toResourceMap(item)
			if err != nil {
				continue
			}
			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// toResourceMap converts any struct to a map[string]interface{} via JSON marshaling.
func toResourceMap(item any) (map[string]interface{}, error) {
	return core.ToResource(item)
}
