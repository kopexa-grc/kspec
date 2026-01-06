package atlassian

import (
	"context"

	"github.com/ctreminiom/go-atlassian/admin"

	"github.com/kopexa-grc/kspec/core"
)

// AdminOrganizationResource fetches Atlassian organization information.
type AdminOrganizationResource struct {
	client *admin.Client
	orgID  string
}

// Name returns the resource type name.
func (r *AdminOrganizationResource) Name() string {
	return "atlassian_organization"
}

// Fetch retrieves organization details.
func (r *AdminOrganizationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.orgID == "" {
		// No org ID configured, skip
		return nil, nil
	}

	var resources []core.Resource

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
	verifiedDomainCount, _ := resource["verified_domain_count"].(int)
	resource["has_verified_domains"] = verifiedDomainCount > 0

	resources = append(resources, resource)

	return resources, nil
}
