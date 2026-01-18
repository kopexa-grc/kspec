// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package discovery

import (
	"github.com/kopexa-grc/kspec/provider/registry"
)

// GetParentType returns the parent resource type for a given child type and provider.
// Returns empty string if no parent is defined.
func GetParentType(providerName, childType string) string {
	def, ok := registry.Get(providerName)
	if !ok {
		return ""
	}
	return def.GetParentType(childType)
}

// GetResourceGroup returns the visual group for a resource type.
// Returns "other" if no group is defined.
func GetResourceGroup(providerName, resourceType string) string {
	def, ok := registry.Get(providerName)
	if !ok {
		return "other"
	}
	return def.GetResourceGroup(resourceType)
}

// GetRootType returns the root resource type for a provider.
// Returns empty string if the provider is not found or has no root type defined.
func GetRootType(providerName string) string {
	def, ok := registry.Get(providerName)
	if !ok {
		return ""
	}
	return def.GetRootType()
}

// GetRelationType returns the relationship type between parent and child.
// Currently returns "contains" as default.
func GetRelationType(parentType, childType string) string {
	// Simple default for now - can be extended via registry if needed
	return "contains"
}
