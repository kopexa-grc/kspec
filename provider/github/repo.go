package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/juliankoehn/kspec/core"
)

type RepoResource struct {
	client *github.Client
}

func (r *RepoResource) Name() string {
	return "github_repo"
}

func (r *RepoResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	owner, ok := config["owner"]
	if !ok {
		return nil, fmt.Errorf("missing 'owner' in config")
	}
	repoName, ok := config["repo"]
	if !ok {
		return nil, fmt.Errorf("missing 'repo' in config")
	}

	// Fetch Repository details
	repo, _, err := r.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo %s/%s: %w", owner, repoName, err)
	}

	// Convert structs to map[string]interface{}.
	data, err := json.Marshal(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal repo: %w", err)
	}

	var resourceMap map[string]interface{}
	if err := json.Unmarshal(data, &resourceMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal repo to map: %w", err)
	}

	// Attempt to fetch collaborators (simplified)
	collabs, _, err := r.client.Repositories.ListCollaborators(ctx, owner, repoName, nil)
	if err == nil {
		var collabList []map[string]interface{}
		for _, c := range collabs {
			collabList = append(collabList, map[string]interface{}{
				"user":        c.GetLogin(),
				"permissions": c.Permissions, // map[string]bool
				"role_name":   c.GetRoleName(),
			})
		}
		resourceMap["collaborators"] = collabList
	}

	// Fetch repository file tree (for file-based checks like Dependabot)
	// Get the default branch tree
	defaultBranch := repo.GetDefaultBranch()
	tree, _, err := r.client.Git.GetTree(ctx, owner, repoName, defaultBranch, true)
	if err == nil && tree != nil {
		var files []map[string]interface{}
		for _, entry := range tree.Entries {
			// Only include files, not trees (directories)
			if entry.GetType() == "blob" {
				files = append(files, map[string]interface{}{
					"path": entry.GetPath(),
					"type": entry.GetType(),
					"size": entry.GetSize(),
					"sha":  entry.GetSHA(),
				})
			}
		}
		resourceMap["files"] = files
	} else {
		// If we can't fetch the tree, set empty array
		resourceMap["files"] = []map[string]interface{}{}
	}

	return []core.Resource{resourceMap}, nil
}
