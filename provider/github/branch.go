package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/juliankoehn/kspec/core"
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
	repoName, ok := config["repo"]
	if !ok {
		return nil, fmt.Errorf("missing 'repo' in config")
	}

	// Fetch all branches
	branches, _, err := r.client.Repositories.ListBranches(ctx, owner, repoName, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch branches for %s/%s: %w", owner, repoName, err)
	}

	// Get default branch
	repo, _, err := r.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo %s/%s: %w", owner, repoName, err)
	}
	defaultBranch := repo.GetDefaultBranch()

	var resources []core.Resource
	for _, branch := range branches {
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

		resources = append(resources, branchData)
	}

	return resources, nil
}
