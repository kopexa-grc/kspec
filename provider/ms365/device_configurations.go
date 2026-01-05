package ms365

import (
	"context"
	"encoding/json"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/juliankoehn/kspec/core"
)

type DeviceConfigurationResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *DeviceConfigurationResource) Name() string {
	return "ms365_device_configuration"
}

func (r *DeviceConfigurationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// Get device configurations from Intune
	result, err := r.client.DeviceManagement().DeviceConfigurations().Get(ctx, nil)
	if err != nil {
		// Device management might not be licensed
		return resources, nil
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, config := range result.GetValue() {
		data, err := json.Marshal(config)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = config.GetId()
		resourceMap["displayName"] = config.GetDisplayName()
		resourceMap["description"] = config.GetDescription()
		resourceMap["createdDateTime"] = config.GetCreatedDateTime()
		resourceMap["lastModifiedDateTime"] = config.GetLastModifiedDateTime()
		resourceMap["version"] = config.GetVersion()

		// The @odata.type tells us what kind of configuration this is
		if odataType := config.GetOdataType(); odataType != nil {
			resourceMap["configurationType"] = *odataType
		}

		// Get additional data which contains platform-specific settings
		additionalData := config.GetAdditionalData()
		if additionalData != nil {
			for key, value := range additionalData {
				resourceMap[key] = value
			}
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

type DeviceCompliancePolicyResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *DeviceCompliancePolicyResource) Name() string {
	return "ms365_device_compliance_policy"
}

func (r *DeviceCompliancePolicyResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// Get device compliance policies from Intune
	result, err := r.client.DeviceManagement().DeviceCompliancePolicies().Get(ctx, nil)
	if err != nil {
		// Device management might not be licensed
		return resources, nil
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, policy := range result.GetValue() {
		data, err := json.Marshal(policy)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = policy.GetId()
		resourceMap["displayName"] = policy.GetDisplayName()
		resourceMap["description"] = policy.GetDescription()
		resourceMap["createdDateTime"] = policy.GetCreatedDateTime()
		resourceMap["lastModifiedDateTime"] = policy.GetLastModifiedDateTime()
		resourceMap["version"] = policy.GetVersion()

		// The @odata.type tells us what kind of compliance policy this is
		if odataType := policy.GetOdataType(); odataType != nil {
			resourceMap["policyType"] = *odataType
		}

		// Get additional data which contains platform-specific settings
		additionalData := policy.GetAdditionalData()
		if additionalData != nil {
			for key, value := range additionalData {
				resourceMap[key] = value
			}
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
