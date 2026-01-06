package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

// DeviceResource provides access to Azure AD registered devices.
type DeviceResource struct {
	client *msgraphsdk.GraphServiceClient
}

// Name returns the resource type identifier for Azure AD devices.
func (r *DeviceResource) Name() string {
	return "ms365_device"
}

// Fetch retrieves all devices registered in Azure AD.
func (r *DeviceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	result, err := r.client.Devices().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get devices: %w", err)
	}

	devices := result.GetValue()
	if devices == nil {
		return []core.Resource{}, nil
	}

	resources := make([]core.Resource, 0, len(devices))
	for _, device := range devices {
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

// ManagedDeviceResource provides access to Intune managed devices.
type ManagedDeviceResource struct {
	client *msgraphsdk.GraphServiceClient
}

// Name returns the resource type identifier for Intune managed devices.
func (r *ManagedDeviceResource) Name() string {
	return "ms365_managed_device"
}

// Fetch retrieves all managed devices from Microsoft Intune.
func (r *ManagedDeviceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// Get managed devices from Intune
	result, err := r.client.DeviceManagement().ManagedDevices().Get(ctx, nil)
	if err != nil {
		// Intune may not be enabled
		return []core.Resource{}, err
	}

	devices := result.GetValue()
	if devices == nil {
		return []core.Resource{}, nil
	}

	resources := make([]core.Resource, 0, len(devices))
	for _, device := range devices {
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
