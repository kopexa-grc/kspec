// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package ms365

import (
	"context"
	"testing"
	"time"
)

// Mock implementation of DeviceManagementItem for testing
type mockDeviceManagementItem struct {
	id                   *string
	displayName          *string
	description          *string
	createdDateTime      *time.Time
	lastModifiedDateTime *time.Time
	version              *int32
	odataType            *string
	additionalData       map[string]any
}

func (m *mockDeviceManagementItem) GetId() *string                 { return m.id }
func (m *mockDeviceManagementItem) GetDisplayName() *string        { return m.displayName }
func (m *mockDeviceManagementItem) GetDescription() *string        { return m.description }
func (m *mockDeviceManagementItem) GetCreatedDateTime() *time.Time { return m.createdDateTime }
func (m *mockDeviceManagementItem) GetLastModifiedDateTime() *time.Time {
	return m.lastModifiedDateTime
}
func (m *mockDeviceManagementItem) GetVersion() *int32                { return m.version }
func (m *mockDeviceManagementItem) GetOdataType() *string             { return m.odataType }
func (m *mockDeviceManagementItem) GetAdditionalData() map[string]any { return m.additionalData }

func strPtr(s string) *string        { return &s }
func int32Ptr(i int32) *int32        { return &i }
func timePtr(t time.Time) *time.Time { return &t }

func TestFetchDeviceManagementResources(t *testing.T) {
	now := time.Now()

	t.Run("converts items to resources", func(t *testing.T) {
		items := []*mockDeviceManagementItem{
			{
				id:          strPtr("item-1"),
				displayName: strPtr("Test Policy 1"),
				description: strPtr("Description 1"),
				version:     int32Ptr(1),
				odataType:   strPtr("#microsoft.graph.deviceCompliancePolicy"),
			},
			{
				id:          strPtr("item-2"),
				displayName: strPtr("Test Policy 2"),
				description: strPtr("Description 2"),
				version:     int32Ptr(2),
				odataType:   strPtr("#microsoft.graph.deviceConfiguration"),
			},
		}

		resources := fetchDeviceManagementResources(context.Background(), items, "policyType")

		if len(resources) != 2 {
			t.Fatalf("expected 2 resources, got %d", len(resources))
		}

		// Check first resource
		if *resources[0]["id"].(*string) != "item-1" {
			t.Errorf("expected id=item-1, got %v", resources[0]["id"])
		}
		if *resources[0]["displayName"].(*string) != "Test Policy 1" {
			t.Errorf("expected displayName=Test Policy 1, got %v", resources[0]["displayName"])
		}
		if resources[0]["policyType"] != "#microsoft.graph.deviceCompliancePolicy" {
			t.Errorf("expected policyType=#microsoft.graph.deviceCompliancePolicy, got %v", resources[0]["policyType"])
		}
	})

	t.Run("includes additional data", func(t *testing.T) {
		items := []*mockDeviceManagementItem{
			{
				id:          strPtr("item-1"),
				displayName: strPtr("Policy with extras"),
				additionalData: map[string]any{
					"customField1": "value1",
					"customField2": 123,
				},
			},
		}

		resources := fetchDeviceManagementResources(context.Background(), items, "policyType")

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}

		if resources[0]["customField1"] != "value1" {
			t.Errorf("expected customField1=value1, got %v", resources[0]["customField1"])
		}
		if resources[0]["customField2"] != 123 {
			t.Errorf("expected customField2=123, got %v", resources[0]["customField2"])
		}
	})

	t.Run("handles empty slice", func(t *testing.T) {
		items := []*mockDeviceManagementItem{}

		resources := fetchDeviceManagementResources(context.Background(), items, "policyType")

		if len(resources) != 0 {
			t.Errorf("expected 0 resources, got %d", len(resources))
		}
	})

	t.Run("handles nil pointers gracefully", func(t *testing.T) {
		items := []*mockDeviceManagementItem{
			{
				id:          nil,
				displayName: nil,
				description: nil,
				version:     nil,
				odataType:   nil,
			},
		}

		resources := fetchDeviceManagementResources(context.Background(), items, "policyType")

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}

		// Should have nil values for the nil pointers (stored as (*string)(nil))
		if resources[0]["id"] != (*string)(nil) {
			t.Errorf("expected id=(*string)(nil), got %v (type: %T)", resources[0]["id"], resources[0]["id"])
		}
	})

	t.Run("includes timestamps", func(t *testing.T) {
		items := []*mockDeviceManagementItem{
			{
				id:                   strPtr("item-1"),
				displayName:          strPtr("Timed Policy"),
				createdDateTime:      timePtr(now),
				lastModifiedDateTime: timePtr(now),
			},
		}

		resources := fetchDeviceManagementResources(context.Background(), items, "policyType")

		if len(resources) != 1 {
			t.Fatalf("expected 1 resource, got %d", len(resources))
		}

		if resources[0]["createdDateTime"] == nil {
			t.Error("expected createdDateTime to be set")
		}
		if resources[0]["lastModifiedDateTime"] == nil {
			t.Error("expected lastModifiedDateTime to be set")
		}
	})
}

func BenchmarkFetchDeviceManagementResources(b *testing.B) {
	items := make([]*mockDeviceManagementItem, 100)
	for i := range items {
		items[i] = &mockDeviceManagementItem{
			id:          strPtr("item-" + string(rune(i))),
			displayName: strPtr("Policy " + string(rune(i))),
			description: strPtr("Description for policy"),
			version:     int32Ptr(int32(i)),
			odataType:   strPtr("#microsoft.graph.deviceCompliancePolicy"),
			additionalData: map[string]any{
				"extra1": "value",
				"extra2": 123,
			},
		}
	}

	b.ResetTimer()
	for b.Loop() {
		_ = fetchDeviceManagementResources(context.Background(), items, "policyType")
	}
}
