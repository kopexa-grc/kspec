package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/groups"

	"github.com/kopexa-grc/kspec/core"
)

type GroupResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *GroupResource) Name() string {
	return "ms365_group"
}

func (r *GroupResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

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

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, group := range result.GetValue() {
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

// GroupLifecyclePolicyResource handles group lifecycle policies
type GroupLifecyclePolicyResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *GroupLifecyclePolicyResource) Name() string {
	return "ms365_group_lifecycle_policy"
}

func (r *GroupLifecyclePolicyResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	result, err := r.client.GroupLifecyclePolicies().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get group lifecycle policies: %w", err)
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
		resourceMap["groupLifetimeInDays"] = policy.GetGroupLifetimeInDays()
		resourceMap["managedGroupTypes"] = policy.GetManagedGroupTypes()
		resourceMap["alternateNotificationEmails"] = policy.GetAlternateNotificationEmails()

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
