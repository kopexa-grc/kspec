package atlassian

import (
	"context"
	"strings"

	v3 "github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/kopexa-grc/kspec/core"
)

// JiraPermissionSchemeResource fetches Jira permission schemes.
type JiraPermissionSchemeResource struct {
	client *v3.Client
}

// Name returns the resource type name.
func (r *JiraPermissionSchemeResource) Name() string {
	return "atlassian_jira_permission_scheme"
}

// Fetch retrieves all Jira permission schemes.
func (r *JiraPermissionSchemeResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// Get all permission schemes
	schemes, response, err := r.client.Permission.Scheme.Gets(ctx)
	if err != nil {
		if response != nil && response.Code == 404 {
			return resources, nil
		}
		return nil, err
	}

	for _, scheme := range schemes.PermissionSchemes {
		resource := make(core.Resource)
		resource["id"] = scheme.ID
		resource["name"] = scheme.Name
		resource["description"] = scheme.Description

		// Analyze permissions
		var permissions []map[string]interface{}
		hasAnonymousAccess := false
		hasGroupAnyoneAccess := false
		hasApplicationRoleAccess := false

		for _, perm := range scheme.Permissions {
			permMap := map[string]interface{}{
				"id":         perm.ID,
				"permission": perm.Permission,
			}

			if perm.Holder != nil {
				permMap["holder_type"] = perm.Holder.Type
				permMap["holder_parameter"] = perm.Holder.Parameter

				// Check for anonymous access
				if perm.Holder.Type == "anyone" {
					hasAnonymousAccess = true
				}

				// Check for "anyone" group access
				if perm.Holder.Type == "group" && strings.EqualFold(perm.Holder.Parameter, "anyone") {
					hasGroupAnyoneAccess = true
				}

				// Check for application role access
				if perm.Holder.Type == "applicationRole" {
					hasApplicationRoleAccess = true
				}
			}

			permissions = append(permissions, permMap)
		}

		resource["permissions"] = permissions
		resource["permission_count"] = len(scheme.Permissions)

		// Security flags
		resource["has_anonymous_access"] = hasAnonymousAccess
		resource["has_group_anyone_access"] = hasGroupAnyoneAccess
		resource["has_application_role_access"] = hasApplicationRoleAccess
		resource["is_default"] = scheme.Name == "Default Permission Scheme"

		resources = append(resources, resource)
	}

	return resources, nil
}
