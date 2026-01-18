// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	v3 "github.com/ctreminiom/go-atlassian/jira/v3"
	"github.com/ctreminiom/go-atlassian/pkg/infra/models"

	"github.com/kopexa-grc/kspec/core"
)

// JiraProject fetches Jira projects.
type JiraProject struct {
	client *v3.Client
}

// NewJiraProject creates a new JiraProject resource.
func NewJiraProject(client *v3.Client) *JiraProject {
	return &JiraProject{client: client}
}

// Name returns the resource type name.
func (r *JiraProject) Name() string {
	return "atlassian_jira_project"
}

// Fetch retrieves all Jira projects.
func (r *JiraProject) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// Search for all projects
	options := &models.ProjectSearchOptionsScheme{
		Expand: []string{"description", "lead", "insight"},
	}

	startAt := 0
	maxResults := 50

	for {
		projects, response, err := r.client.Project.Search(ctx, options, startAt, maxResults)
		if err != nil {
			if response != nil && response.Code == 404 {
				// No projects found
				break
			}
			return nil, err
		}

		for _, project := range projects.Values {
			resource := make(core.Resource)
			resource["id"] = project.ID
			resource["key"] = project.Key
			resource["name"] = project.Name
			resource["description"] = project.Description
			resource["project_type_key"] = project.ProjectTypeKey
			resource["style"] = project.Style
			resource["is_private"] = project.IsPrivate

			// Add lead information
			if project.Lead != nil {
				resource["lead_account_id"] = project.Lead.AccountID
				resource["lead_display_name"] = project.Lead.DisplayName
				resource["lead_email"] = project.Lead.EmailAddress
			}

			// Add insight information
			if project.Insight != nil {
				resource["insight_total_issues"] = project.Insight.TotalIssueCount
				resource["insight_last_issue_update"] = project.Insight.LastIssueUpdateTime
			}

			// Computed security flags
			resource["has_lead"] = project.Lead != nil && project.Lead.AccountID != ""

			resources = append(resources, resource)
		}

		if projects.IsLast {
			break
		}
		startAt += maxResults
	}

	return resources, nil
}
