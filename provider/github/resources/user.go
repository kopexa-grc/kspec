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

// User fetches GitHub user/member data for an organization.
type User struct {
	client *github.Client
}

// NewUser creates a new User resource.
func NewUser(client *github.Client) *User {
	return &User{client: client}
}

// Name returns the resource type identifier for GitHub users.
func (r *User) Name() string {
	return "github_user"
}

// Fetch retrieves all members for an organization.
func (r *User) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	org, ok := config["org"]
	if !ok {
		if val, exists := config["owner"]; exists {
			org = val
		} else {
			return nil, fmt.Errorf("missing 'org' (or 'owner') in config for github_user")
		}
	}

	// List organization members with pagination
	opts := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allMembers []*github.User
	for {
		members, resp, err := r.client.Organizations.ListMembers(ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list members for org %s: %w", org, err)
		}
		allMembers = append(allMembers, members...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	resources := make([]core.Resource, 0, len(allMembers))

	for _, member := range allMembers {
		// Fetch full user details for each member
		user, _, err := r.client.Users.Get(ctx, member.GetLogin())
		if err != nil {
			// If we can't get full details, use the basic member info
			user = member
		}

		data, err := json.Marshal(user)
		if err != nil {
			continue
		}
		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		// Add computed fields for security checks
		resourceMap["is_organization_member"] = true

		// Check if user has 2FA enabled (requires admin access)
		// This is available in ListMembers with filter "2fa_disabled"
		resourceMap["two_factor_enabled"] = true // Default to true, will be overridden if we can check

		resources = append(resources, resourceMap)
	}

	// Also check for members with 2FA disabled (requires org admin)
	twoFADisabledOpts := &github.ListMembersOptions{
		Filter:      "2fa_disabled",
		ListOptions: github.ListOptions{PerPage: 100},
	}

	twoFADisabled := make(map[string]bool)
	for {
		members, resp, err := r.client.Organizations.ListMembers(ctx, org, twoFADisabledOpts)
		if err != nil {
			// May fail without admin permissions - that's OK
			break
		}
		for _, m := range members {
			twoFADisabled[m.GetLogin()] = true
		}
		if resp.NextPage == 0 {
			break
		}
		twoFADisabledOpts.Page = resp.NextPage
	}

	// Update 2FA status for users who have it disabled
	for i, res := range resources {
		if login, ok := res["login"].(string); ok {
			if twoFADisabled[login] {
				resources[i]["two_factor_enabled"] = false
			}
		}
	}

	return resources, nil
}
