package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

type TenantResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *TenantResource) Name() string {
	return "ms365_tenant"
}

func (r *TenantResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// Get organization details
	org, err := r.client.Organization().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}

	if org.GetValue() == nil || len(org.GetValue()) == 0 {
		return nil, fmt.Errorf("no organization found")
	}

	organization := org.GetValue()[0]

	// Convert to map
	data, err := json.Marshal(organization)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal organization: %w", err)
	}

	var resourceMap map[string]interface{}
	if err := json.Unmarshal(data, &resourceMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal organization: %w", err)
	}

	// Add additional fields
	resourceMap["id"] = organization.GetId()
	resourceMap["displayName"] = organization.GetDisplayName()
	resourceMap["tenantType"] = organization.GetTenantType()

	if organization.GetVerifiedDomains() != nil {
		var domains []map[string]interface{}
		for _, domain := range organization.GetVerifiedDomains() {
			domains = append(domains, map[string]interface{}{
				"name":      domain.GetName(),
				"isDefault": domain.GetIsDefault(),
				"isInitial": domain.GetIsInitial(),
				"type":      domain.GetTypeEscaped(),
			})
		}
		resourceMap["verifiedDomains"] = domains
	}

	if organization.GetAssignedPlans() != nil {
		var plans []map[string]interface{}
		for _, plan := range organization.GetAssignedPlans() {
			plans = append(plans, map[string]interface{}{
				"service":          plan.GetService(),
				"capabilityStatus": plan.GetCapabilityStatus(),
				"servicePlanId":    plan.GetServicePlanId(),
			})
		}
		resourceMap["assignedPlans"] = plans
	}

	return []core.Resource{resourceMap}, nil
}

// Discover implements core.DiscoveryResource
func (r *TenantResource) Discover(ctx context.Context, asset core.Asset) (map[string]int, error) {
	if asset.Type != "ms365" && asset.Type != "ms365-tenant" {
		return nil, nil
	}

	discovered := make(map[string]int)

	// Tenant is always 1
	discovered["ms365_tenant"] = 1

	// Discover users
	users, err := r.client.Users().Get(ctx, nil)
	if err == nil && users.GetValue() != nil {
		discovered["ms365_user"] = len(users.GetValue())
	}

	// Discover groups
	groups, err := r.client.Groups().Get(ctx, nil)
	if err == nil && groups.GetValue() != nil {
		discovered["ms365_group"] = len(groups.GetValue())
	}

	// Discover applications
	apps, err := r.client.Applications().Get(ctx, nil)
	if err == nil && apps.GetValue() != nil {
		discovered["ms365_application"] = len(apps.GetValue())
	}

	// Discover service principals
	sps, err := r.client.ServicePrincipals().Get(ctx, nil)
	if err == nil && sps.GetValue() != nil {
		discovered["ms365_service_principal"] = len(sps.GetValue())
	}

	// Discover devices
	devices, err := r.client.Devices().Get(ctx, nil)
	if err == nil && devices.GetValue() != nil {
		discovered["ms365_device"] = len(devices.GetValue())
	}

	// Discover managed devices (Intune)
	managedDevices, err := r.client.DeviceManagement().ManagedDevices().Get(ctx, nil)
	if err == nil && managedDevices.GetValue() != nil {
		discovered["ms365_managed_device"] = len(managedDevices.GetValue())
	}

	// Discover device configurations (Intune)
	deviceConfigs, err := r.client.DeviceManagement().DeviceConfigurations().Get(ctx, nil)
	if err == nil && deviceConfigs.GetValue() != nil {
		discovered["ms365_device_configuration"] = len(deviceConfigs.GetValue())
	}

	// Discover device compliance policies (Intune)
	compliancePolicies, err := r.client.DeviceManagement().DeviceCompliancePolicies().Get(ctx, nil)
	if err == nil && compliancePolicies.GetValue() != nil {
		discovered["ms365_device_compliance_policy"] = len(compliancePolicies.GetValue())
	}

	// Discover domains
	domains, err := r.client.Domains().Get(ctx, nil)
	if err == nil && domains.GetValue() != nil {
		discovered["ms365_domain"] = len(domains.GetValue())
	}

	// Discover conditional access policies
	caps, err := r.client.Identity().ConditionalAccess().Policies().Get(ctx, nil)
	if err == nil && caps.GetValue() != nil {
		discovered["ms365_conditional_access_policy"] = len(caps.GetValue())
	}

	// Discover directory roles
	roles, err := r.client.DirectoryRoles().Get(ctx, nil)
	if err == nil && roles.GetValue() != nil {
		discovered["ms365_directory_role"] = len(roles.GetValue())
	}

	// Discover teams
	teams, err := r.client.Teams().Get(ctx, nil)
	if err == nil && teams.GetValue() != nil {
		discovered["ms365_team"] = len(teams.GetValue())
	}

	return discovered, nil
}
