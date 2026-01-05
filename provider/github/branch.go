package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/kopexa-grc/kspec/core"
)

type BranchResource struct {
	client *github.Client
}

func (r *BranchResource) Name() string {
	return "github_branch"
}

func (r *BranchResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	owner, ok := config["owner"]
	if !ok {
		return nil, fmt.Errorf("missing 'owner' in config")
	}

	// Check if this is a single repo scan or an org scan
	repoName, hasRepo := config["repo"]

	if hasRepo {
		// Single repo scan
		return r.fetchRepoBranches(ctx, owner, repoName)
	}

	// Org scan - fetch branches for all repos in the organization
	return r.fetchOrgBranches(ctx, owner)
}

func (r *BranchResource) fetchRepoBranches(ctx context.Context, owner, repoName string) ([]core.Resource, error) {
	// Get repo info to find default branch
	repo, _, err := r.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo %s/%s: %w", owner, repoName, err)
	}
	defaultBranchName := repo.GetDefaultBranch()

	// Only fetch the default branch (main/master/etc.) for faster scanning
	branch, _, err := r.client.Repositories.GetBranch(ctx, owner, repoName, defaultBranchName, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch default branch %s for %s/%s: %w", defaultBranchName, owner, repoName, err)
	}

	// Convert to the Branch type expected by branchToResource
	branchData := r.branchToResource(ctx, owner, repoName, branch, defaultBranchName)

	return []core.Resource{branchData}, nil
}

func (r *BranchResource) fetchOrgBranches(ctx context.Context, orgName string) ([]core.Resource, error) {
	var resources []core.Resource

	// List all repos in the org
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

			// Only fetch the default branch for faster scanning
			branch, _, err := r.client.Repositories.GetBranch(ctx, orgName, repoName, defaultBranchName, 0)
			if err != nil {
				// Skip repos we can't access
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

func (r *BranchResource) branchToResource(ctx context.Context, owner, repoName string, branch *github.Branch, defaultBranch string) core.Resource {
	branchName := branch.GetName()

	// Get branch protection
	protection, _, err := r.client.Repositories.GetBranchProtection(ctx, owner, repoName, branchName)

	branchData := map[string]interface{}{
		"name":       branchName,
		"is_default": branchName == defaultBranch,
		"protected":  branch.GetProtected(),
	}

	if err == nil && protection != nil {
		// Convert protection to map
		protectionData, err := json.Marshal(protection)
		if err == nil {
			var protectionMap map[string]interface{}
			if err := json.Unmarshal(protectionData, &protectionMap); err == nil {
				branchData["protection_rules"] = protectionMap

				// Add specific protection fields for easier querying
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
