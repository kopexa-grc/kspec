package ms365

import (
	"context"
	"encoding/json"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

type DirectorySettingResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *DirectorySettingResource) Name() string {
	return "ms365_directory_setting"
}

func (r *DirectorySettingResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// Get group settings
	result, err := r.client.GroupSettings().Get(ctx, nil)
	if err != nil {
		// Group settings might not be configured
		return resources, nil
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, setting := range result.GetValue() {
		data, err := json.Marshal(setting)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = setting.GetId()
		resourceMap["displayName"] = setting.GetDisplayName()
		resourceMap["templateId"] = setting.GetTemplateId()

		// Parse values into a more accessible format
		if setting.GetValues() != nil {
			values := make(map[string]interface{})
			for _, value := range setting.GetValues() {
				if value.GetName() != nil && value.GetValue() != nil {
					values[*value.GetName()] = *value.GetValue()
				}
			}
			resourceMap["values"] = values

			// Extract common security settings
			if v, ok := values["EnableGroupCreation"]; ok {
				resourceMap["enableGroupCreation"] = v
			}
			if v, ok := values["AllowGuestsToBeGroupOwner"]; ok {
				resourceMap["allowGuestsToBeGroupOwner"] = v
			}
			if v, ok := values["AllowGuestsToAccessGroups"]; ok {
				resourceMap["allowGuestsToAccessGroups"] = v
			}
			if v, ok := values["GuestUsageGuidelinesUrl"]; ok {
				resourceMap["guestUsageGuidelinesUrl"] = v
			}
			if v, ok := values["GroupCreationAllowedGroupId"]; ok {
				resourceMap["groupCreationAllowedGroupId"] = v
			}
			if v, ok := values["AllowToAddGuests"]; ok {
				resourceMap["allowToAddGuests"] = v
			}
			if v, ok := values["UsageGuidelinesUrl"]; ok {
				resourceMap["usageGuidelinesUrl"] = v
			}
			if v, ok := values["ClassificationDescriptions"]; ok {
				resourceMap["classificationDescriptions"] = v
			}
			if v, ok := values["DefaultClassification"]; ok {
				resourceMap["defaultClassification"] = v
			}
			if v, ok := values["PrefixSuffixNamingRequirement"]; ok {
				resourceMap["prefixSuffixNamingRequirement"] = v
			}
			if v, ok := values["CustomBlockedWordsList"]; ok {
				resourceMap["customBlockedWordsList"] = v
			}
			if v, ok := values["EnableMSStandardBlockedWords"]; ok {
				resourceMap["enableMSStandardBlockedWords"] = v
			}
			if v, ok := values["ClassificationList"]; ok {
				resourceMap["classificationList"] = v
			}
			if v, ok := values["EnableMIPLabels"]; ok {
				resourceMap["enableMIPLabels"] = v
			}
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

type ExternalIdentityPolicyResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *ExternalIdentityPolicyResource) Name() string {
	return "ms365_external_identity_policy"
}

func (r *ExternalIdentityPolicyResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// Get B2B collaboration settings via cross-tenant access settings
	result, err := r.client.Policies().CrossTenantAccessPolicy().Get(ctx, nil)
	if err != nil {
		return resources, nil
	}

	resourceMap := make(map[string]interface{})
	resourceMap["id"] = result.GetId()
	resourceMap["displayName"] = result.GetDisplayName()
	resourceMap["allowedCloudEndpoints"] = result.GetAllowedCloudEndpoints()

	// Get default settings using the DefaultEscaped method (Default is a reserved word)
	defaultSettings, err := r.client.Policies().CrossTenantAccessPolicy().DefaultEscaped().Get(ctx, nil)
	if err == nil && defaultSettings != nil {
		defaultMap := make(map[string]interface{})

		if defaultSettings.GetB2bCollaborationInbound() != nil {
			inbound := defaultSettings.GetB2bCollaborationInbound()
			defaultMap["b2bCollaborationInbound"] = map[string]interface{}{
				"usersAndGroups": inbound.GetUsersAndGroups(),
				"applications":   inbound.GetApplications(),
			}
		}

		if defaultSettings.GetB2bCollaborationOutbound() != nil {
			outbound := defaultSettings.GetB2bCollaborationOutbound()
			defaultMap["b2bCollaborationOutbound"] = map[string]interface{}{
				"usersAndGroups": outbound.GetUsersAndGroups(),
				"applications":   outbound.GetApplications(),
			}
		}

		if defaultSettings.GetB2bDirectConnectInbound() != nil {
			inbound := defaultSettings.GetB2bDirectConnectInbound()
			defaultMap["b2bDirectConnectInbound"] = map[string]interface{}{
				"usersAndGroups": inbound.GetUsersAndGroups(),
				"applications":   inbound.GetApplications(),
			}
		}

		if defaultSettings.GetB2bDirectConnectOutbound() != nil {
			outbound := defaultSettings.GetB2bDirectConnectOutbound()
			defaultMap["b2bDirectConnectOutbound"] = map[string]interface{}{
				"usersAndGroups": outbound.GetUsersAndGroups(),
				"applications":   outbound.GetApplications(),
			}
		}

		resourceMap["default"] = defaultMap
	}

	resources = append(resources, resourceMap)
	return resources, nil
}
