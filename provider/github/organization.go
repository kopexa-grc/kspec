package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"

	"github.com/kopexa-grc/kspec/core"
)

// OrganizationResource fetches GitHub organization data and settings.
type OrganizationResource struct {
	client *github.Client
}

// Name returns the resource type identifier for GitHub organizations.
func (r *OrganizationResource) Name() string {
	return "github_organization"
}

// Fetch retrieves organization details including security settings.
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

// DiscoveryResult holds discovery counts in order
type DiscoveryResult struct {
	ResourceType string
	Count        int
}

// Discover implements core.DiscoveryResource
// Scans the organization and discovers what resource types are present
func (r *OrganizationResource) Discover(ctx context.Context, asset core.Asset) (map[string]int, error) {
	// Only discover for github-org asset type
	if asset.Type != "github-org" {
		return nil, nil
	}

	config := asset.Config
	orgName, ok := config["owner"]
	if !ok {
		return nil, fmt.Errorf("missing 'owner' in config for organization discovery")
	}

	// Use ordered slice to maintain discovery order, then convert to map
	var results []DiscoveryResult

	// 1. Discover organization (always 1 for org scan)
	_, _, err := r.client.Organizations.Get(ctx, orgName)
	if err == nil {
		results = append(results, DiscoveryResult{"github_organization", 1})
	}

	// 2. Discover teams in the organization
	teams, _, err := r.client.Teams.ListTeams(ctx, orgName, nil)
	if err == nil && len(teams) > 0 {
		results = append(results, DiscoveryResult{"github_team", len(teams)})
	}

	// 3. Discover repositories in the organization
	// Note: Branches are discovered as sub-resources of repos, not at org level
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	repoCount := 0
	for {
		repos, resp, err := r.client.Repositories.ListByOrg(ctx, orgName, opts)
		if err != nil {
			break
		}
		repoCount += len(repos)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	if repoCount > 0 {
		results = append(results, DiscoveryResult{"github_repo", repoCount})
	}

	// Convert to map (order will be maintained by scan.go using tree.Root.Children order)
	discovered := make(map[string]int)
	for _, r := range results {
		discovered[r.ResourceType] = r.Count
	}

	return discovered, nil
}
