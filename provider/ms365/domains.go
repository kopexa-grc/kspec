package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

// DomainResource provides access to Azure AD verified domains.
type DomainResource struct {
	client *msgraphsdk.GraphServiceClient
}

// Name returns the resource type identifier for domains.
func (r *DomainResource) Name() string {
	return "ms365_domain"
}

// Fetch retrieves all verified domains with their configuration.
func (r *DomainResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	result, err := r.client.Domains().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get domains: %w", err)
	}

	domains := result.GetValue()
	if domains == nil {
		return []core.Resource{}, nil
	}

	resources := make([]core.Resource, 0, len(domains))
	for _, domain := range domains {
		data, err := json.Marshal(domain)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = domain.GetId()
		resourceMap["name"] = domain.GetId() // Domain ID is the domain name
		resourceMap["authenticationType"] = domain.GetAuthenticationType()
		resourceMap["isAdminManaged"] = domain.GetIsAdminManaged()
		resourceMap["isDefault"] = domain.GetIsDefault()
		resourceMap["isInitial"] = domain.GetIsInitial()
		resourceMap["isRoot"] = domain.GetIsRoot()
		resourceMap["isVerified"] = domain.GetIsVerified()
		resourceMap["availabilityStatus"] = domain.GetAvailabilityStatus()
		resourceMap["supportedServices"] = domain.GetSupportedServices()

		// Password validity and notification
		resourceMap["passwordValidityPeriodInDays"] = domain.GetPasswordValidityPeriodInDays()
		resourceMap["passwordNotificationWindowInDays"] = domain.GetPasswordNotificationWindowInDays()

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
