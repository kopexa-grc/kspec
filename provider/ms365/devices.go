package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/juliankoehn/kspec/core"
)

type DeviceResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *DeviceResource) Name() string {
	return "ms365_device"
}

func (r *DeviceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	result, err := r.client.Devices().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, device := range result.GetValue() {
		data, err := json.Marshal(device)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = device.GetId()
		resourceMap["displayName"] = device.GetDisplayName()
		resourceMap["deviceId"] = device.GetDeviceId()
		resourceMap["accountEnabled"] = device.GetAccountEnabled()
		resourceMap["operatingSystem"] = device.GetOperatingSystem()
		resourceMap["operatingSystemVersion"] = device.GetOperatingSystemVersion()
		resourceMap["trustType"] = device.GetTrustType()
		resourceMap["isCompliant"] = device.GetIsCompliant()
		resourceMap["isManaged"] = device.GetIsManaged()
		resourceMap["approximateLastSignInDateTime"] = device.GetApproximateLastSignInDateTime()
		resourceMap["registeredOwners"] = device.GetRegisteredOwners()
		resourceMap["registeredUsers"] = device.GetRegisteredUsers()

		// Device management type
		resourceMap["mdmAppId"] = device.GetMdmAppId()
		resourceMap["profileType"] = device.GetProfileType()

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

type ManagedDeviceResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *ManagedDeviceResource) Name() string {
	return "ms365_managed_device"
}

func (r *ManagedDeviceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// Get managed devices from Intune
	result, err := r.client.DeviceManagement().ManagedDevices().Get(ctx, nil)
	if err != nil {
		// Intune may not be enabled
		return resources, nil
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, device := range result.GetValue() {
		data, err := json.Marshal(device)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = device.GetId()
		resourceMap["deviceName"] = device.GetDeviceName()
		resourceMap["managedDeviceOwnerType"] = device.GetManagedDeviceOwnerType()
		resourceMap["operatingSystem"] = device.GetOperatingSystem()
		resourceMap["complianceState"] = device.GetComplianceState()
		resourceMap["jailBroken"] = device.GetJailBroken()
		resourceMap["managementAgent"] = device.GetManagementAgent()
		resourceMap["isEncrypted"] = device.GetIsEncrypted()
		resourceMap["isSupervised"] = device.GetIsSupervised()
		resourceMap["azureADRegistered"] = device.GetAzureADRegistered()
		resourceMap["deviceEnrollmentType"] = device.GetDeviceEnrollmentType()
		resourceMap["lastSyncDateTime"] = device.GetLastSyncDateTime()
		resourceMap["userPrincipalName"] = device.GetUserPrincipalName()
		resourceMap["model"] = device.GetModel()
		resourceMap["manufacturer"] = device.GetManufacturer()

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
