// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

// DeviceConfiguration provides access to Intune device configurations.
type DeviceConfiguration struct {
	client *msgraphsdk.GraphServiceClient
}

// NewDeviceConfiguration creates a new DeviceConfiguration resource.
func NewDeviceConfiguration(client *msgraphsdk.GraphServiceClient) *DeviceConfiguration {
	return &DeviceConfiguration{client: client}
}

// Name returns the resource type identifier for device configurations.
func (r *DeviceConfiguration) Name() string {
	return "ms365_device_configuration"
}

// Fetch retrieves all device configuration policies from Intune.
func (r *DeviceConfiguration) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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

// DeviceCompliancePolicy provides access to Intune device compliance policies.
type DeviceCompliancePolicy struct {
	client *msgraphsdk.GraphServiceClient
}

// NewDeviceCompliancePolicy creates a new DeviceCompliancePolicy resource.
func NewDeviceCompliancePolicy(client *msgraphsdk.GraphServiceClient) *DeviceCompliancePolicy {
	return &DeviceCompliancePolicy{client: client}
}

// Name returns the resource type identifier for device compliance policies.
func (r *DeviceCompliancePolicy) Name() string {
	return "ms365_device_compliance_policy"
}

// Fetch retrieves all device compliance policies from Intune.
func (r *DeviceCompliancePolicy) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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
