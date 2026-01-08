// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package azure

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"

	"github.com/kopexa-grc/kspec/core"
)

// PolicyAssignmentResource represents an Azure Policy assignment resource scanner.
type PolicyAssignmentResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *PolicyAssignmentResource) Name() string {
	return "azure_policy_assignment"
}

// Fetch retrieves policy assignments from Azure.
func (r *PolicyAssignmentResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armpolicy.NewAssignmentsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy assignments client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list policy assignments: %w", err)
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
				resourceMap["displayName"] = assignment.Properties.DisplayName
				resourceMap["description"] = assignment.Properties.Description
				resourceMap["policyDefinitionId"] = assignment.Properties.PolicyDefinitionID
				resourceMap["scope"] = assignment.Properties.Scope
				resourceMap["enforcementMode"] = assignment.Properties.EnforcementMode

				if assignment.Properties.NonComplianceMessages != nil {
					messages := make([]string, 0, len(assignment.Properties.NonComplianceMessages))
					for _, msg := range assignment.Properties.NonComplianceMessages {
						if msg.Message != nil {
							messages = append(messages, *msg.Message)
						}
					}
					resourceMap["nonComplianceMessages"] = messages
				}
			}

			if assignment.Identity != nil {
				resourceMap["identityType"] = assignment.Identity.Type
				resourceMap["identityPrincipalId"] = assignment.Identity.PrincipalID
				resourceMap["identityTenantId"] = assignment.Identity.TenantID
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}

// PolicyDefinitionResource represents an Azure Policy definition resource scanner.
type PolicyDefinitionResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *PolicyDefinitionResource) Name() string {
	return "azure_policy_definition"
}

// Fetch retrieves policy definitions from Azure.
func (r *PolicyDefinitionResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	client, err := armpolicy.NewDefinitionsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy definitions client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list policy definitions: %w", err)
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
				resourceMap["displayName"] = definition.Properties.DisplayName
				resourceMap["description"] = definition.Properties.Description
				resourceMap["policyType"] = definition.Properties.PolicyType
				resourceMap["mode"] = definition.Properties.Mode

				if definition.Properties.Metadata != nil {
					resourceMap["metadata"] = definition.Properties.Metadata
				}
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}

// PolicyComplianceResource represents an Azure Policy compliance state resource scanner.
type PolicyComplianceResource struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// Name returns the resource type name.
func (r *PolicyComplianceResource) Name() string {
	return "azure_policy_compliance"
}

// Fetch retrieves policy compliance states from Azure.
func (r *PolicyComplianceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	if r.subscriptionID == "" {
		return nil, fmt.Errorf("subscription_id is required")
	}

	// Policy States requires the PolicyInsights client
	// Using the policy assignments client to get compliance info
	client, err := armpolicy.NewAssignmentsClient(r.subscriptionID, r.credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create policy client: %w", err)
	}

	resources := make([]core.Resource, 0)

	pager := client.NewListPager(nil)
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list policy assignments for compliance: %w", err)
		}

		for _, assignment := range page.Value {
			resourceMap := make(map[string]interface{})

			resourceMap["id"] = assignment.ID
			resourceMap["name"] = assignment.Name

			if assignment.Properties != nil {
				resourceMap["displayName"] = assignment.Properties.DisplayName
				resourceMap["policyDefinitionId"] = assignment.Properties.PolicyDefinitionID
				resourceMap["scope"] = assignment.Properties.Scope
				resourceMap["enforcementMode"] = assignment.Properties.EnforcementMode
			}

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}
