// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package resources contains GitHub resource implementations.
package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"

	"github.com/kopexa-grc/kspec/core"
)

// Branch fetches GitHub branch data and protection rules.
type Branch struct {
	client *github.Client
}

// NewBranch creates a new Branch resource.
func NewBranch(client *github.Client) *Branch {
	return &Branch{client: client}
}

// Name returns the resource type identifier for GitHub branches.
func (r *Branch) Name() string {
	return "github_branch"
}

// Fetch retrieves branch data for a repository or all repositories in an organization.
func (r *Branch) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	owner, ok := config["owner"]
	if !ok {
		return nil, fmt.Errorf("missing 'owner' in config")
	}

	repoName, hasRepo := config["repo"]

	if hasRepo {
		return r.fetchRepoBranches(ctx, owner, repoName)
	}

	return r.fetchOrgBranches(ctx, owner)
}

func (r *Branch) fetchRepoBranches(ctx context.Context, owner, repoName string) ([]core.Resource, error) {
	repo, _, err := r.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo %s/%s: %w", owner, repoName, err)
	}
	defaultBranchName := repo.GetDefaultBranch()

	branch, _, err := r.client.Repositories.GetBranch(ctx, owner, repoName, defaultBranchName, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch default branch %s for %s/%s: %w", defaultBranchName, owner, repoName, err)
	}

	branchData := r.branchToResource(ctx, owner, repoName, branch, defaultBranchName)

	return []core.Resource{branchData}, nil
}

func (r *Branch) fetchOrgBranches(ctx context.Context, orgName string) ([]core.Resource, error) {
	var resources []core.Resource

	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := r.client.Repositories.ListByOrg(ctx, orgName, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repos for org %s: %w", orgName, err)
		}

		for _, repo := range repos {
			repoName := repo.GetName()
			defaultBranchName := repo.GetDefaultBranch()

			branch, _, err := r.client.Repositories.GetBranch(ctx, orgName, repoName, defaultBranchName, 0)
			if err != nil {
				continue
			}

			branchData := r.branchToResource(ctx, orgName, repoName, branch, defaultBranchName)
			branchData["repo"] = repoName
			branchData["repo_full_name"] = orgName + "/" + repoName

			resources = append(resources, branchData)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return resources, nil
}

func (r *Branch) branchToResource(ctx context.Context, owner, repoName string, branch *github.Branch, defaultBranch string) core.Resource {
	branchName := branch.GetName()

	protection, _, err := r.client.Repositories.GetBranchProtection(ctx, owner, repoName, branchName)

	branchData := map[string]interface{}{
		"name":       branchName,
		"is_default": branchName == defaultBranch,
		"protected":  branch.GetProtected(),
	}

	if err == nil && protection != nil {
		protectionData, err := json.Marshal(protection)
		if err == nil {
			var protectionMap map[string]interface{}
			if err := json.Unmarshal(protectionData, &protectionMap); err == nil {
				branchData["protection_rules"] = protectionMap

				if protection.RequiredStatusChecks != nil {
					branchData["required_status_checks"] = protection.RequiredStatusChecks.Contexts
				}
				if protection.EnforceAdmins != nil {
					branchData["enforce_admins"] = map[string]interface{}{
						"enabled": protection.EnforceAdmins.Enabled,
					}
				}
				if protection.RequiredPullRequestReviews != nil {
					branchData["required_pull_request_reviews"] = true
				}
				if protection.AllowForcePushes != nil {
					branchData["allow_force_pushes"] = map[string]interface{}{
						"enabled": protection.AllowForcePushes.Enabled,
					}
				}
				if protection.RequiredConversationResolution != nil {
					branchData["required_conversation_resolution"] = map[string]interface{}{
						"enabled": protection.RequiredConversationResolution.Enabled,
					}
				}
				branchData["required_signatures"] = protection.GetRequiredSignatures().GetEnabled()
			}
		}
	} else {
		branchData["protection_rules"] = nil
	}

	return branchData
}
