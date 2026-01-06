package atlassian

import (
	"context"

	"github.com/ctreminiom/go-atlassian/admin"
	"github.com/kopexa-grc/kspec/core"
)

// SCIMUserResource fetches SCIM-provisioned users.
// Note: SCIM requires a directory ID to list users. Without directory access,
// this resource returns empty.
type SCIMUserResource struct {
	client *admin.Client
	orgID  string
}

// Name returns the resource type name.
func (r *SCIMUserResource) Name() string {
	return "atlassian_scim_user"
}

// Fetch retrieves SCIM users.
// Note: SCIM API requires directory ID which must be obtained separately.
// This is typically configured through the Atlassian Admin portal.
func (r *SCIMUserResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// SCIM requires a directory ID to list users.
	// The directory ID is typically obtained from the Admin console
	// and passed via asset configuration.
	directoryID := asset.Config["directory_id"]
	if directoryID == "" {
		// Without a directory ID, we cannot list SCIM users
		return nil, nil
	}

	var resources []core.Resource

	users, response, err := r.client.SCIM.User.Gets(ctx, directoryID, nil, 1, 100)
	if err != nil {
		if response != nil && (response.Code == 404 || response.Code == 403) {
			// SCIM not enabled or no access
			return resources, nil
		}
		return nil, err
	}

	for _, user := range users.Resources {
		resource := make(core.Resource)
		resource["id"] = user.ID
		resource["external_id"] = user.ExternalID
		resource["user_name"] = user.UserName
		resource["display_name"] = user.DisplayName
		resource["active"] = user.Active
		resource["directory_id"] = directoryID

		// Name components
		if user.Name != nil {
			resource["family_name"] = user.Name.FamilyName
			resource["given_name"] = user.Name.GivenName
			resource["formatted_name"] = user.Name.Formatted
		}

		// Email
		if len(user.Emails) > 0 {
			resource["email"] = user.Emails[0].Value
			resource["email_type"] = user.Emails[0].Type
			resource["email_primary"] = user.Emails[0].Primary
		}

		// Security flags
		resource["is_active"] = user.Active
		resource["is_provisioned"] = true
		resource["has_external_id"] = user.ExternalID != ""
		resource["has_email"] = len(user.Emails) > 0

		resources = append(resources, resource)
	}

	return resources, nil
}
