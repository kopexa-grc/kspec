package atlassian

import (
	"context"

	"github.com/ctreminiom/go-atlassian/confluence"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"

	"github.com/kopexa-grc/kspec/core"
)

// ConfluencePageResource fetches Confluence pages.
type ConfluencePageResource struct {
	client *confluence.Client
}

// Name returns the resource type name.
func (r *ConfluencePageResource) Name() string {
	return "atlassian_confluence_page"
}

// Fetch retrieves Confluence pages.
func (r *ConfluencePageResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// Get pages using CQL search
	options := &models.SearchContentOptions{
		Expand: []string{"space", "version"},
		Limit:  25,
	}

	results, response, err := r.client.Search.Content(ctx, "type=page", options)
	if err != nil {
		if response != nil && response.Code == 404 {
			return nil, nil
		}
		return nil, err
	}

	resources := make([]core.Resource, 0, len(results.Results))
	for _, result := range results.Results {
		if result.Content == nil {
			continue
		}

		content := result.Content
		resource := make(core.Resource)
		resource["id"] = content.ID
		resource["title"] = content.Title
		resource["type"] = content.Type
		resource["status"] = content.Status

		// Add space information
		if content.Space != nil {
			resource["space_key"] = content.Space.Key
			resource["space_name"] = content.Space.Name
		}

		// Add version information
		if content.Version != nil {
			resource["version_number"] = content.Version.Number
			if content.Version.By != nil {
				resource["version_by"] = content.Version.By.DisplayName
			}
		}

		// Add search result metadata
		resource["last_modified"] = result.LastModified
		resource["url"] = result.URL

		// Security flags
		resource["is_published"] = content.Status == "current"
		resource["is_draft"] = content.Status == "draft"
		resource["is_archived"] = content.Status == "archived"

		resources = append(resources, resource)
	}

	return resources, nil
}
