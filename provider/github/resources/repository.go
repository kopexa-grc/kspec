// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-github/v62/github"

	"github.com/kopexa-grc/kspec/core"
)

// Repository fetches GitHub repository data and metadata.
type Repository struct {
	client *github.Client
}

// NewRepository creates a new Repository resource.
func NewRepository(client *github.Client) *Repository {
	return &Repository{client: client}
}

// Name returns the resource type identifier for GitHub repositories.
func (r *Repository) Name() string {
	return "github_repo"
}

// SubResources implements core.SubResourceProvider.
// Returns branch resources as sub-resources of repos.
func (r *Repository) SubResources() []core.ResourceSpec {
	return []core.ResourceSpec{
		NewBranch(r.client),
	}
}

// Fetch retrieves repository data for a single repo or all repos in an organization.
func (r *Repository) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config

	// Parse owner and repo from config
	owner, repo := r.parseOwnerRepo(config)
	if owner == "" {
		return nil, fmt.Errorf("missing 'owner' in config (use 'owner' key or 'repository' in owner/repo format)")
	}

	if repo != "" {
		// Single repo scan
		return r.fetchSingleRepo(ctx, owner, repo)
	}

	// Org scan - fetch all repos in the organization
	return r.fetchOrgRepos(ctx, owner)
}

// parseOwnerRepo extracts owner and repo from config.
func (r *Repository) parseOwnerRepo(config map[string]string) (owner, repo string) {
	owner = config["owner"]
	repo = config["repo"]

	if owner == "" {
		if repository, ok := config["repository"]; ok && repository != "" {
			parts := strings.SplitN(repository, "/", 2)
			if len(parts) == 2 {
				owner = parts[0]
				repo = parts[1]
			}
		}
	}

	return owner, repo
}

func (r *Repository) fetchSingleRepo(ctx context.Context, owner, repoName string) ([]core.Resource, error) {
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

func (r *Repository) fetchOrgRepos(ctx context.Context, orgName string) ([]core.Resource, error) {
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

func (r *Repository) repoToResource(ctx context.Context, owner string, repo *github.Repository) (core.Resource, error) {
	repoName := repo.GetName()

	data, err := json.Marshal(repo)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal repo: %w", err)
	}

	var resourceMap map[string]interface{}
	if err := json.Unmarshal(data, &resourceMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal repo to map: %w", err)
	}

	// Fetch collaborators
	collabs, _, err := r.client.Repositories.ListCollaborators(ctx, owner, repoName, nil)
	if err == nil {
		var collabList []map[string]interface{}
		for _, c := range collabs {
			collabList = append(collabList, map[string]interface{}{
				"user":        c.GetLogin(),
				"permissions": c.Permissions,
				"role_name":   c.GetRoleName(),
			})
		}
		resourceMap["collaborators"] = collabList
	}

	// Fetch repository file tree
	defaultBranch := repo.GetDefaultBranch()
	tree, _, err := r.client.Git.GetTree(ctx, owner, repoName, defaultBranch, true)
	if err == nil && tree != nil {
		var files []map[string]interface{}
		for _, entry := range tree.Entries {
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
		resourceMap["files"] = []map[string]interface{}{}
	}

	return resourceMap, nil
}

// Discover implements core.DiscoveryResource.
func (r *Repository) Discover(ctx context.Context, asset core.Asset) (map[string]int, error) {
	if asset.Type != "github-repo" {
		return nil, nil
	}

	config := asset.Config
	owner, repoName := r.parseOwnerRepo(config)
	if owner == "" {
		return nil, fmt.Errorf("missing 'owner' in config for repo discovery")
	}
	if repoName == "" {
		return nil, fmt.Errorf("missing 'repo' in config for repo discovery")
	}

	discovered := make(map[string]int)

	_, _, err := r.client.Repositories.Get(ctx, owner, repoName)
	if err == nil {
		discovered["github_repo"] = 1
	}

	branches, _, err := r.client.Repositories.ListBranches(ctx, owner, repoName, nil)
	if err == nil && len(branches) > 0 {
		discovered["github_branch"] = len(branches)
	}

	return discovered, nil
}
