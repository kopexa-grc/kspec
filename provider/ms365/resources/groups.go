// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"

	"github.com/kopexa-grc/kspec/core"
)

// Group provides access to Microsoft 365 group resources via the Graph API.
type Group struct {
	client *msgraphsdk.GraphServiceClient
}

// NewGroup creates a new Group resource.
func NewGroup(client *msgraphsdk.GraphServiceClient) *Group {
	return &Group{client: client}
}

// Name returns the resource type identifier for MS365 groups.
func (r *Group) Name() string {
	return "ms365_group"
}

// Fetch retrieves all groups from Microsoft 365 with their properties and membership details.
func (r *Group) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	requestConfig := &groups.GroupsRequestBuilderGetRequestConfiguration{
		QueryParameters: &groups.GroupsRequestBuilderGetQueryParameters{
			Select: []string{
				"id", "displayName", "description", "mail", "mailEnabled", "mailNickname",
				"securityEnabled", "groupTypes", "visibility", "createdDateTime",
				"membershipRule", "membershipRuleProcessingState", "isAssignableToRole",
				"onPremisesSyncEnabled", "onPremisesLastSyncDateTime",
			},
		},
	}

	result, err := r.client.Groups().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups: %w", err)
	}

	groupList := result.GetValue()
	if groupList == nil {
		return []core.Resource{}, nil
	}

	resources := make([]core.Resource, 0, len(groupList))
	for _, group := range groupList {
		data, err := json.Marshal(group)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		// Add computed fields
		resourceMap["id"] = group.GetId()
		resourceMap["displayName"] = group.GetDisplayName()
		resourceMap["description"] = group.GetDescription()
		resourceMap["mail"] = group.GetMail()
		resourceMap["mailEnabled"] = group.GetMailEnabled()
		resourceMap["securityEnabled"] = group.GetSecurityEnabled()
		resourceMap["visibility"] = group.GetVisibility()
		resourceMap["groupTypes"] = group.GetGroupTypes()
		resourceMap["isAssignableToRole"] = group.GetIsAssignableToRole()
		resourceMap["onPremisesSyncEnabled"] = group.GetOnPremisesSyncEnabled()

		// Determine group type
		groupTypes := group.GetGroupTypes()
		isUnified := false
		isDynamic := false
		for _, gt := range groupTypes {
			if gt == "Unified" {
				isUnified = true
			}
			if gt == "DynamicMembership" {
				isDynamic = true
			}
		}
		resourceMap["isUnifiedGroup"] = isUnified
		resourceMap["isDynamicGroup"] = isDynamic
		resourceMap["isMicrosoft365Group"] = isUnified
		resourceMap["isSecurityGroup"] = group.GetSecurityEnabled() != nil && *group.GetSecurityEnabled() && !isUnified

		// Get member count
		members, err := r.client.Groups().ByGroupId(*group.GetId()).Members().Get(ctx, nil)
		if err == nil && members.GetValue() != nil {
			resourceMap["memberCount"] = len(members.GetValue())
		}

		// Get owner count
		owners, err := r.client.Groups().ByGroupId(*group.GetId()).Owners().Get(ctx, nil)
		if err == nil && owners.GetValue() != nil {
			resourceMap["ownerCount"] = len(owners.GetValue())
			resourceMap["hasOwners"] = len(owners.GetValue()) > 0
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

// GroupLifecyclePolicy handles group lifecycle policies.
type GroupLifecyclePolicy struct {
	client *msgraphsdk.GraphServiceClient
}

// NewGroupLifecyclePolicy creates a new GroupLifecyclePolicy resource.
func NewGroupLifecyclePolicy(client *msgraphsdk.GraphServiceClient) *GroupLifecyclePolicy {
	return &GroupLifecyclePolicy{client: client}
}

// Name returns the resource type identifier for group lifecycle policies.
func (r *GroupLifecyclePolicy) Name() string {
	return "ms365_group_lifecycle_policy"
}

// Fetch retrieves all group lifecycle policies from Microsoft 365.
func (r *GroupLifecyclePolicy) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	result, err := r.client.GroupLifecyclePolicies().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get group lifecycle policies: %w", err)
	}

	policies := result.GetValue()
	if policies == nil {
		return []core.Resource{}, nil
	}

	resources := make([]core.Resource, 0, len(policies))
	for _, policy := range policies {
		data, err := json.Marshal(policy)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = policy.GetId()
		resourceMap["groupLifetimeInDays"] = policy.GetGroupLifetimeInDays()
		resourceMap["managedGroupTypes"] = policy.GetManagedGroupTypes()
		resourceMap["alternateNotificationEmails"] = policy.GetAlternateNotificationEmails()

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
