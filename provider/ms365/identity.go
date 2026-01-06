package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

type ConditionalAccessPolicyResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *ConditionalAccessPolicyResource) Name() string {
	return "ms365_conditional_access_policy"
}

func (r *ConditionalAccessPolicyResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	result, err := r.client.Identity().ConditionalAccess().Policies().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get conditional access policies: %w", err)
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
		resourceMap["displayName"] = policy.GetDisplayName()
		resourceMap["state"] = policy.GetState()
		resourceMap["createdDateTime"] = policy.GetCreatedDateTime()
		resourceMap["modifiedDateTime"] = policy.GetModifiedDateTime()

		// Conditions
		if policy.GetConditions() != nil {
			conditions := policy.GetConditions()
			conditionsMap := make(map[string]interface{})

			// Users
			if conditions.GetUsers() != nil {
				users := conditions.GetUsers()
				conditionsMap["includeUsers"] = users.GetIncludeUsers()
				conditionsMap["excludeUsers"] = users.GetExcludeUsers()
				conditionsMap["includeGroups"] = users.GetIncludeGroups()
				conditionsMap["excludeGroups"] = users.GetExcludeGroups()
				conditionsMap["includeRoles"] = users.GetIncludeRoles()
				conditionsMap["excludeRoles"] = users.GetExcludeRoles()
			}

			// Applications
			if conditions.GetApplications() != nil {
				apps := conditions.GetApplications()
				conditionsMap["includeApplications"] = apps.GetIncludeApplications()
				conditionsMap["excludeApplications"] = apps.GetExcludeApplications()
			}

			// Platforms
			if conditions.GetPlatforms() != nil {
				platforms := conditions.GetPlatforms()
				conditionsMap["includePlatforms"] = platforms.GetIncludePlatforms()
				conditionsMap["excludePlatforms"] = platforms.GetExcludePlatforms()
			}

			// Locations
			if conditions.GetLocations() != nil {
				locations := conditions.GetLocations()
				conditionsMap["includeLocations"] = locations.GetIncludeLocations()
				conditionsMap["excludeLocations"] = locations.GetExcludeLocations()
			}

			// Sign-in risk
			conditionsMap["signInRiskLevels"] = conditions.GetSignInRiskLevels()
			conditionsMap["userRiskLevels"] = conditions.GetUserRiskLevels()

			resourceMap["conditions"] = conditionsMap
		}

		// Grant controls
		if policy.GetGrantControls() != nil {
			grant := policy.GetGrantControls()
			grantMap := map[string]interface{}{
				"operator":                    grant.GetOperator(),
				"builtInControls":             grant.GetBuiltInControls(),
				"customAuthenticationFactors": grant.GetCustomAuthenticationFactors(),
				"termsOfUse":                  grant.GetTermsOfUse(),
			}
			resourceMap["grantControls"] = grantMap

			// Check for MFA requirement
			builtInControls := grant.GetBuiltInControls()
			requiresMfa := false
			for _, control := range builtInControls {
				if control.String() == "mfa" {
					requiresMfa = true
					break
				}
			}
			resourceMap["requiresMfa"] = requiresMfa
		}

		// Session controls
		if policy.GetSessionControls() != nil {
			session := policy.GetSessionControls()
			sessionMap := make(map[string]interface{})

			if session.GetSignInFrequency() != nil {
				sessionMap["signInFrequency"] = map[string]interface{}{
					"value":     session.GetSignInFrequency().GetValue(),
					"type":      session.GetSignInFrequency().GetTypeEscaped(),
					"isEnabled": session.GetSignInFrequency().GetIsEnabled(),
				}
			}

			if session.GetPersistentBrowser() != nil {
				sessionMap["persistentBrowser"] = map[string]interface{}{
					"mode":      session.GetPersistentBrowser().GetMode(),
					"isEnabled": session.GetPersistentBrowser().GetIsEnabled(),
				}
			}

			resourceMap["sessionControls"] = sessionMap
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

type NamedLocationResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *NamedLocationResource) Name() string {
	return "ms365_named_location"
}

func (r *NamedLocationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	result, err := r.client.Identity().ConditionalAccess().NamedLocations().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get named locations: %w", err)
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, location := range result.GetValue() {
		data, err := json.Marshal(location)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = location.GetId()
		resourceMap["displayName"] = location.GetDisplayName()
		resourceMap["createdDateTime"] = location.GetCreatedDateTime()
		resourceMap["modifiedDateTime"] = location.GetModifiedDateTime()

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
