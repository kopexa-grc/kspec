package atlassian

import (
	"context"

	"github.com/ctreminiom/go-atlassian/admin"

	"github.com/kopexa-grc/kspec/core"
)

// SCIMGroupResource fetches SCIM-provisioned groups.
// Note: SCIM requires a directory ID to list groups. Without directory access,
// this resource returns empty.
type SCIMGroupResource struct {
	client *admin.Client
	orgID  string
}

// Name returns the resource type name.
func (r *SCIMGroupResource) Name() string {
	return "atlassian_scim_group"
}

// Fetch retrieves SCIM groups.
// Note: SCIM API requires directory ID which must be obtained separately.
// This is typically configured through the Atlassian Admin portal.
func (r *SCIMGroupResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// SCIM requires a directory ID to list groups.
	// The directory ID is typically obtained from the Admin console
	// and passed via asset configuration.
	directoryID := asset.Config["directory_id"]
	if directoryID == "" {
		// Without a directory ID, we cannot list SCIM groups
		return nil, nil
	}

	var resources []core.Resource

	groups, response, err := r.client.SCIM.Group.Gets(ctx, directoryID, "", 1, 100)
	if err != nil {
		if response != nil && (response.Code == 404 || response.Code == 403) {
			// SCIM not enabled or no access
			return resources, nil
		}
		return nil, err
	}

	for _, group := range groups.Resources {
		resource := make(core.Resource)
		resource["id"] = group.ID
		resource["external_id"] = group.ExternalID
		resource["display_name"] = group.DisplayName
		resource["directory_id"] = directoryID

		// Member count
		memberCount := 0
		if group.Members != nil {
			memberCount = len(group.Members)
		}
		resource["member_count"] = memberCount

		// Security flags
		resource["is_provisioned"] = true
		resource["has_external_id"] = group.ExternalID != ""
		resource["has_members"] = memberCount > 0
		resource["is_empty"] = memberCount == 0

		resources = append(resources, resource)
	}

	return resources, nil
}
