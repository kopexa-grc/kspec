package ms365

import (
	"context"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

// DeviceConfigurationResource provides access to Intune device configurations.
type DeviceConfigurationResource struct {
	client *msgraphsdk.GraphServiceClient
}

// Name returns the resource type identifier for device configurations.
func (r *DeviceConfigurationResource) Name() string {
	return "ms365_device_configuration"
}

// Fetch retrieves all device configuration policies from Intune.
func (r *DeviceConfigurationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	result, err := r.client.DeviceManagement().DeviceConfigurations().Get(ctx, nil)
	if err != nil {
		// Device management might not be licensed
		return nil, err
	}

	if result.GetValue() == nil {
		return nil, nil
	}

	return fetchDeviceManagementResources(ctx, result.GetValue(), "configurationType"), nil
}

// DeviceCompliancePolicyResource provides access to Intune device compliance policies.
type DeviceCompliancePolicyResource struct {
	client *msgraphsdk.GraphServiceClient
}

// Name returns the resource type identifier for device compliance policies.
func (r *DeviceCompliancePolicyResource) Name() string {
	return "ms365_device_compliance_policy"
}

// Fetch retrieves all device compliance policies from Intune.
func (r *DeviceCompliancePolicyResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	result, err := r.client.DeviceManagement().DeviceCompliancePolicies().Get(ctx, nil)
	if err != nil {
		// Device management might not be licensed
		return nil, err
	}

	if result.GetValue() == nil {
		return nil, nil
	}

	return fetchDeviceManagementResources(ctx, result.GetValue(), "policyType"), nil
}
