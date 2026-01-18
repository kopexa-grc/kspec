// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	v3 "github.com/ctreminiom/go-atlassian/jira/v3"

	"github.com/kopexa-grc/kspec/core"
)

// JiraSecurityScheme fetches Jira issue security schemes.
// Note: The go-atlassian v3 API doesn't currently expose issue security schemes directly.
// This resource serves as a placeholder for future implementation.
type JiraSecurityScheme struct {
	client *v3.Client
}

// NewJiraSecurityScheme creates a new JiraSecurityScheme resource.
func NewJiraSecurityScheme(client *v3.Client) *JiraSecurityScheme {
	return &JiraSecurityScheme{client: client}
}

// Name returns the resource type name.
func (r *JiraSecurityScheme) Name() string {
	return "atlassian_jira_security_scheme"
}

// Fetch retrieves Jira issue security schemes.
// Note: This is a placeholder as the current SDK version doesn't support
// fetching issue security schemes directly.
func (r *JiraSecurityScheme) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// The go-atlassian SDK v3 doesn't currently expose issue security schemes.
	// This would typically be accessed via:
	// GET /rest/api/3/issuesecurityschemes
	//
	// For now, return empty to avoid breaking policy evaluations.
	return nil, nil
}
