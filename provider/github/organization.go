package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/juliankoehn/kspec/core"
)

type OrganizationResource struct {
	client *github.Client
}

func (r *OrganizationResource) Name() string {
	return "github_organization"
}

func (r *OrganizationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	orgName, ok := config["owner"]
	if !ok {
		return nil, fmt.Errorf("missing 'owner' in config for organization resource")
	}

	// Fetch Organization details
	org, _, err := r.client.Organizations.Get(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organization %s: %w", orgName, err)
	}

	// Convert to map
	data, err := json.Marshal(org)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal organization: %w", err)
	}

	var resourceMap map[string]interface{}
	if err := json.Unmarshal(data, &resourceMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal organization to map: %w", err)
	}

	// Add additional fields that might not be in the Organization struct
	// Two-Factor Requirement
	resourceMap["two_factor_requirement_enabled"] = org.GetTwoFactorRequirementEnabled()
	
	// Default Repository Permission
	resourceMap["default_repository_permission"] = org.GetDefaultRepoPermission()
	
	// Verified status
	resourceMap["is_verified"] = org.GetIsVerified()

	return []core.Resource{resourceMap}, nil
}
