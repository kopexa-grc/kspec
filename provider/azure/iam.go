// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"

	"github.com/kopexa-grc/kspec/core"
)

// RoleAssignmentResource represents an Azure role assignment resource scanner.
type RoleAssignmentResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *RoleAssignmentResource) Name() string {
	return "azure_role_assignment"
}

// Fetch retrieves role assignments from Azure.
func (r *RoleAssignmentResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	scope := fmt.Sprintf("/subscriptions/%s", r.subscriptionID)
	client, err := armauthorization.NewRoleAssignmentsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create role assignments client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListForScopePager(scope, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list role assignments: %w", err)
		}

		for _, assignment := range page.Value {
			data, err := json.Marshal(assignment)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if assignment.Properties != nil {
				resourceMap["principalId"] = assignment.Properties.PrincipalID
				resourceMap["roleDefinitionId"] = assignment.Properties.RoleDefinitionID
				resourceMap["scope"] = assignment.Properties.Scope
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}

// RoleDefinitionResource represents an Azure role definition resource scanner.
type RoleDefinitionResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *RoleDefinitionResource) Name() string {
	return "azure_role_definition"
}

// Fetch retrieves role definitions from Azure.
func (r *RoleDefinitionResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	scope := fmt.Sprintf("/subscriptions/%s", r.subscriptionID)
	client, err := armauthorization.NewRoleDefinitionsClient(r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create role definitions client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(scope, nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list role definitions: %w", err)
		}

		for _, definition := range page.Value {
			data, err := json.Marshal(definition)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			if definition.Properties != nil {
				resourceMap["roleName"] = definition.Properties.RoleName
				resourceMap["roleType"] = definition.Properties.RoleType
				resourceMap["description"] = definition.Properties.Description

				// Permissions
				if definition.Properties.Permissions != nil {
					perms := make([]map[string]interface{}, 0, len(definition.Properties.Permissions))
					for _, perm := range definition.Properties.Permissions {
						permMap := map[string]interface{}{
							"actions":    perm.Actions,
							"notActions": perm.NotActions,
						}
						perms = append(perms, permMap)
					}
					resourceMap["permissions"] = perms
				}

				resourceMap["assignableScopes"] = definition.Properties.AssignableScopes
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
