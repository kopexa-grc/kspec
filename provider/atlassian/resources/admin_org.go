// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

// Package resources provides Atlassian Cloud resource fetchers for Jira,
// Confluence, and Admin APIs used in security policy evaluation.
package resources

import (
	"context"

	"github.com/ctreminiom/go-atlassian/admin"

	"github.com/kopexa-grc/kspec/core"
)

// AdminOrganization fetches Atlassian organization information.
type AdminOrganization struct {
	client *admin.Client
	orgID  string
}

// NewAdminOrganization creates a new AdminOrganization resource.
func NewAdminOrganization(client *admin.Client, orgID string) *AdminOrganization {
	return &AdminOrganization{client: client, orgID: orgID}
}

// Name returns the resource type name.
func (r *AdminOrganization) Name() string {
	return "atlassian_organization"
}

// Fetch retrieves organization details.
func (r *AdminOrganization) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.orgID == "" {
		// No org ID configured, skip
		return nil, nil
	}

	resources := make([]core.Resource, 0, 1)

	// Get organization details
	org, response, err := r.client.Organization.Get(ctx, r.orgID)
	if err != nil {
		if response != nil && response.Code == 404 {
			return resources, nil
		}
		return nil, err
	}

	if org == nil || org.Data == nil {
		return resources, nil
	}

	resource := make(core.Resource)
	resource["id"] = org.Data.ID
	resource["type"] = org.Data.Type

	if org.Data.Attributes != nil {
		resource["name"] = org.Data.Attributes.Name
	}

	// Get organization domains
	domains, _, err := r.client.Organization.Domains(ctx, r.orgID, "")
	if err == nil && domains != nil {
		var domainList []map[string]interface{}
		verifiedDomainsCount := 0

		for _, domain := range domains.Data {
			domainMap := map[string]interface{}{
				"id": domain.ID,
			}

			if domain.Attributes != nil {
				domainMap["name"] = domain.Attributes.Name
				if domain.Attributes.Claim != nil {
					domainMap["status"] = domain.Attributes.Claim.Status
					if domain.Attributes.Claim.Status == "VERIFIED" {
						verifiedDomainsCount++
					}
				}
			}

			domainList = append(domainList, domainMap)
		}

		resource["domains"] = domainList
		resource["domain_count"] = len(domains.Data)
		resource["verified_domain_count"] = verifiedDomainsCount
	}

	// Security flags
	verifiedDomainCount := 0
	if vdc, ok := resource["verified_domain_count"].(int); ok {
		verifiedDomainCount = vdc
	}
	resource["has_verified_domains"] = verifiedDomainCount > 0

	resources = append(resources, resource)

	return resources, nil
}
