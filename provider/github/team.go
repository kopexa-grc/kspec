package github

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v62/github"
	"github.com/juliankoehn/kspec/core"
)

type TeamResource struct {
	client *github.Client
}

func (r *TeamResource) Name() string {
	return "github_team"
}

func (r *TeamResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	org, ok := config["org"]
	if !ok {
		// Fallback to owner if org is not set, assuming owner is an org
		if val, exists := config["owner"]; exists {
			org = val
		} else {
			return nil, fmt.Errorf("missing 'org' (or 'owner') in config for github_team")
		}
	}

	// List Teams for the Organization
	teams, _, err := r.client.Teams.ListTeams(ctx, org, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams for org %s: %w", org, err)
	}

	var resources []core.Resource

	for _, team := range teams {
		// Convert to map
		data, err := json.Marshal(team)
		if err != nil {
			// Skip malformed? or error?
			continue
		}
		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
