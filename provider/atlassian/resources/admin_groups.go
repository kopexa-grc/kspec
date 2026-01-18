// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/ctreminiom/go-atlassian/admin"

	"github.com/kopexa-grc/kspec/core"
)

// AdminGroup fetches organization groups via Admin API.
// Note: The Atlassian Admin API doesn't provide a direct group listing endpoint.
// Groups are typically managed through SCIM or Jira/Confluence directly.
type AdminGroup struct {
	client *admin.Client
	orgID  string
}

// NewAdminGroup creates a new AdminGroup resource.
func NewAdminGroup(client *admin.Client, orgID string) *AdminGroup {
	return &AdminGroup{client: client, orgID: orgID}
}

// Name returns the resource type name.
func (r *AdminGroup) Name() string {
	return "atlassian_admin_group"
}

// Fetch retrieves organization groups.
// Note: This is a placeholder as the Admin API doesn't support group listing directly.
// Use SCIM groups or Jira/Confluence groups instead.
func (r *AdminGroup) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// The Atlassian Admin API v2 doesn't provide a direct group listing endpoint.
	// Groups are managed through:
	// 1. SCIM API (if SCIM provisioning is enabled)
	// 2. Jira Cloud API for Jira groups
	// 3. Confluence Cloud API for Confluence groups
	//
	// This resource returns empty for now. Use atlassian_scim_group for SCIM groups.
	return nil, nil
}
