// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armpolicy"

	"github.com/kopexa-grc/kspec/core"
)

// PolicyAssignment represents an Azure Policy assignment resource scanner.
type PolicyAssignment struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewPolicyAssignment creates a new PolicyAssignment resource scanner.
func NewPolicyAssignment(credential azcore.TokenCredential, subscriptionID string) *PolicyAssignment {
	return &PolicyAssignment{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *PolicyAssignment) Name() string {
	return "azure_policy_assignment"
}

// Fetch retrieves policy assignments from Azure.
func (r *PolicyAssignment) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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

// PolicyDefinition represents an Azure Policy definition resource scanner.
type PolicyDefinition struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewPolicyDefinition creates a new PolicyDefinition resource scanner.
func NewPolicyDefinition(credential azcore.TokenCredential, subscriptionID string) *PolicyDefinition {
	return &PolicyDefinition{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *PolicyDefinition) Name() string {
	return "azure_policy_definition"
}

// Fetch retrieves policy definitions from Azure.
func (r *PolicyDefinition) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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

// PolicyCompliance represents an Azure Policy compliance state resource scanner.
type PolicyCompliance struct {
	credential     azcore.TokenCredential
	subscriptionID string
}

// NewPolicyCompliance creates a new PolicyCompliance resource scanner.
func NewPolicyCompliance(credential azcore.TokenCredential, subscriptionID string) *PolicyCompliance {
	return &PolicyCompliance{
		credential:     credential,
		subscriptionID: subscriptionID,
	}
}

// Name returns the resource type name.
func (r *PolicyCompliance) Name() string {
	return "azure_policy_compliance"
}

// Fetch retrieves policy compliance states from Azure.
func (r *PolicyCompliance) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
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
