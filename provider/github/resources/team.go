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

// Team fetches GitHub team data for an organization.
type Team struct {
	client *github.Client
}

// NewTeam creates a new Team resource.
func NewTeam(client *github.Client) *Team {
	return &Team{client: client}
}

// Name returns the resource type identifier for GitHub teams.
func (r *Team) Name() string {
	return "github_team"
}

// Fetch retrieves all teams for an organization.
func (r *Team) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	config := asset.Config
	org, ok := config["org"]
	if !ok {
		if val, exists := config["owner"]; exists {
			org = val
		} else {
			return nil, fmt.Errorf("missing 'org' (or 'owner') in config for github_team")
		}
	}

	teams, _, err := r.client.Teams.ListTeams(ctx, org, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams for org %s: %w", org, err)
	}

	resources := make([]core.Resource, 0, len(teams))

	for _, team := range teams {
		data, err := json.Marshal(team)
		if err != nil {
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
