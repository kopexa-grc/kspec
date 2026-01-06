package atlassian

import (
	"context"

	"github.com/ctreminiom/go-atlassian/admin"
	"github.com/kopexa-grc/kspec/core"
)

// AdminPolicyResource fetches organization policies via Admin API.
type AdminPolicyResource struct {
	client *admin.Client
	orgID  string
}

// Name returns the resource type name.
func (r *AdminPolicyResource) Name() string {
	return "atlassian_auth_policy"
}

// Fetch retrieves organization policies.
func (r *AdminPolicyResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.orgID == "" {
		return nil, nil
	}

	var resources []core.Resource

	// Get authentication policies
	policies, response, err := r.client.Organization.Policy.Gets(ctx, r.orgID, "", "")
	if err != nil {
		if response != nil && response.Code == 404 {
			return resources, nil
		}
		return nil, err
	}

	for _, policy := range policies.Data {
		resource := make(core.Resource)
		resource["id"] = policy.ID
		resource["type"] = policy.Type

		if policy.Attributes != nil {
			resource["name"] = policy.Attributes.Name
			resource["policy_type"] = policy.Attributes.Type
			resource["status"] = policy.Attributes.Status
			resource["created_at"] = policy.Attributes.CreatedAt
			resource["updated_at"] = policy.Attributes.UpdatedAt

			// Count resources
			resource["resource_count"] = len(policy.Attributes.Resources)
		}

		// Security flags
		resource["is_active"] = resource["status"] == "enabled"
		resource["is_configured"] = policy.Attributes != nil && policy.Attributes.Name != ""

		resources = append(resources, resource)
	}

	return resources, nil
}
