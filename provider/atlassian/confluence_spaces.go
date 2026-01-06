package atlassian

import (
	"context"

	"github.com/ctreminiom/go-atlassian/confluence"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"
	"github.com/kopexa-grc/kspec/core"
)

// ConfluenceSpaceResource fetches Confluence spaces.
type ConfluenceSpaceResource struct {
	client *confluence.Client
}

// Name returns the resource type name.
func (r *ConfluenceSpaceResource) Name() string {
	return "atlassian_confluence_space"
}

// Fetch retrieves all Confluence spaces.
func (r *ConfluenceSpaceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	options := &models.GetSpacesOptionScheme{
		Expand: []string{"permissions", "homepage"},
	}

	startAt := 0
	maxResults := 25

	for {
		spaces, response, err := r.client.Space.Gets(ctx, options, startAt, maxResults)
		if err != nil {
			if response != nil && response.Code == 404 {
				break
			}
			return nil, err
		}

		for _, space := range spaces.Results {
			resource := make(core.Resource)
			resource["id"] = space.ID
			resource["key"] = space.Key
			resource["name"] = space.Name
			resource["type"] = space.Type
			resource["status"] = space.Status

			// Analyze permissions
			hasAnonymousAccess := false
			hasUnlicensedAccess := false
			permissionCount := 0

			if space.Permissions != nil {
				permissionCount = len(space.Permissions)
				for _, perm := range space.Permissions {
					if perm.AnonymousAccess {
						hasAnonymousAccess = true
					}
					if perm.UnlicensedAccess {
						hasUnlicensedAccess = true
					}
				}
			}

			resource["permission_count"] = permissionCount

			// Security flags
			resource["has_anonymous_access"] = hasAnonymousAccess
			resource["has_unlicensed_access"] = hasUnlicensedAccess
			resource["is_global"] = space.Type == "global"
			resource["is_personal"] = space.Type == "personal"
			resource["is_active"] = space.Status == "current"

			resources = append(resources, resource)
		}

		// Check if there are more results
		if len(spaces.Results) < maxResults {
			break
		}
		startAt += maxResults
	}

	return resources, nil
}
