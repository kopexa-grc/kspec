// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"

	"github.com/kopexa-grc/kspec/core"
)

// Organization fetches GitHub organization data and settings.
type Organization struct {
	client *github.Client
}

// NewOrganization creates a new Organization resource.
func NewOrganization(client *github.Client) *Organization {
	return &Organization{client: client}
}

// Name returns the resource type identifier for GitHub organizations.
func (r *Organization) Name() string {
	return "github_organization"
}

// ChildResources implements core.ChildResourceProvider.
// Returns the child resource types that belong to an organization.
func (r *Organization) ChildResources() []string {
	return []string{"github_repo", "github_team", "github_user"}
}

// Fetch retrieves organization details including security settings.
func (r *Organization) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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
	resourceMap["two_factor_requirement_enabled"] = org.GetTwoFactorRequirementEnabled()
	resourceMap["default_repository_permission"] = org.GetDefaultRepoPermission()
	resourceMap["is_verified"] = org.GetIsVerified()

	return []core.Resource{resourceMap}, nil
}

// Discover implements core.DiscoveryResource.
// Scans the organization and discovers what resource types are present.
func (r *Organization) Discover(ctx context.Context, asset core.Asset) (map[string]int, error) {
	// Only discover for github-org asset type
	if asset.Type != "github-org" {
		return nil, nil
	}

	config := asset.Config
	orgName, ok := config["owner"]
	if !ok {
		return nil, fmt.Errorf("missing 'owner' in config for organization discovery")
	}

	discovered := make(map[string]int)

	// 1. Discover organization (always 1 for org scan)
	_, _, err := r.client.Organizations.Get(ctx, orgName)
	if err == nil {
		discovered["github_organization"] = 1
	}

	// 2. Discover teams in the organization
	teams, _, err := r.client.Teams.ListTeams(ctx, orgName, nil)
	if err == nil && len(teams) > 0 {
		discovered["github_team"] = len(teams)
	}

	// 3. Discover repositories in the organization
	repoOpts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	repoCount := 0
	for {
		repos, resp, err := r.client.Repositories.ListByOrg(ctx, orgName, repoOpts)
		if err != nil {
			break
		}
		repoCount += len(repos)

		if resp.NextPage == 0 {
			break
		}
		repoOpts.Page = resp.NextPage
	}

	if repoCount > 0 {
		discovered["github_repo"] = repoCount
	}

	// 4. Discover members in the organization
	memberOpts := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	memberCount := 0
	for {
		members, resp, err := r.client.Organizations.ListMembers(ctx, orgName, memberOpts)
		if err != nil {
			break
		}
		memberCount += len(members)

		if resp.NextPage == 0 {
			break
		}
		memberOpts.Page = resp.NextPage
	}

	if memberCount > 0 {
		discovered["github_user"] = memberCount
	}

	return discovered, nil
}
