package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/kopexa-grc/kspec/core"
)

type RepoResource struct {
	client *github.Client
}

// SubResources implements core.SubResourceProvider
// Returns branch resources as sub-resources of repos
func (r *RepoResource) SubResources() []core.ResourceSpec {
	return []core.ResourceSpec{
		&BranchResource{client: r.client},
	}
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

	// Check if this is a single repo scan or an org scan
	repoName, hasRepo := config["repo"]

	if hasRepo {
		// Single repo scan
		return r.fetchSingleRepo(ctx, owner, repoName)
	}

	// Org scan - fetch all repos in the organization
	return r.fetchOrgRepos(ctx, owner)
}

func (r *RepoResource) fetchSingleRepo(ctx context.Context, owner, repoName string) ([]core.Resource, error) {
	// Fetch Repository details
	repo, _, err := r.client.Repositories.Get(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo %s/%s: %w", owner, repoName, err)
	}

	resourceMap, err := r.repoToResource(ctx, owner, repo)
	if err != nil {
		return nil, err
	}

	return []core.Resource{resourceMap}, nil
}

func (r *RepoResource) fetchOrgRepos(ctx context.Context, orgName string) ([]core.Resource, error) {
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
			resourceMap, err := r.repoToResource(ctx, orgName, repo)
			if err != nil {
				// Log but continue with other repos
				continue
			}
			resources = append(resources, resourceMap)
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return resources, nil
}

func (r *RepoResource) repoToResource(ctx context.Context, owner string, repo *github.Repository) (core.Resource, error) {
	repoName := repo.GetName()

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

	return resourceMap, nil
}

// Discover implements core.DiscoveryResource
// Scans the repository and discovers what resource types are present
func (r *RepoResource) Discover(ctx context.Context, asset core.Asset) (map[string]int, error) {
	// Only discover for github-repo asset type
	if asset.Type != "github-repo" {
		return nil, nil
	}

	config := asset.Config
	owner, ok := config["owner"]
	if !ok {
		return nil, fmt.Errorf("missing 'owner' in config for repo discovery")
	}
	repoName, ok := config["repo"]
	if !ok {
		return nil, fmt.Errorf("missing 'repo' in config for repo discovery")
	}

	discovered := make(map[string]int)

	// Discover repository (always 1 for repo scan)
	_, _, err := r.client.Repositories.Get(ctx, owner, repoName)
	if err == nil {
		discovered["github_repo"] = 1
	}

	// Discover branches
	branches, _, err := r.client.Repositories.ListBranches(ctx, owner, repoName, nil)
	if err == nil && len(branches) > 0 {
		discovered["github_branch"] = len(branches)
	}

	return discovered, nil
}
