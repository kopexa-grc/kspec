package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/users"

	"github.com/kopexa-grc/kspec/core"
)

type UserResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *UserResource) Name() string {
	return "ms365_user"
}

func (r *UserResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// Request specific properties for users
	requestConfig := &users.UsersRequestBuilderGetRequestConfiguration{
		QueryParameters: &users.UsersRequestBuilderGetQueryParameters{
			Select: []string{
				"id", "displayName", "userPrincipalName", "mail", "accountEnabled",
				"createdDateTime", "lastSignInDateTime", "userType", "assignedLicenses",
				"assignedPlans", "mfaDetail", "passwordPolicies", "passwordProfile",
				"onPremisesSyncEnabled", "onPremisesLastSyncDateTime",
			},
		},
	}

	result, err := r.client.Users().Get(ctx, requestConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, user := range result.GetValue() {
		// Convert to map
		data, err := json.Marshal(user)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		// Add computed fields
		resourceMap["id"] = user.GetId()
		resourceMap["displayName"] = user.GetDisplayName()
		resourceMap["userPrincipalName"] = user.GetUserPrincipalName()
		resourceMap["mail"] = user.GetMail()
		resourceMap["accountEnabled"] = user.GetAccountEnabled()
		resourceMap["userType"] = user.GetUserType()

		// Password policies
		if user.GetPasswordPolicies() != nil {
			resourceMap["passwordPolicies"] = user.GetPasswordPolicies()
		}

		// Check if user is synced from on-premises
		resourceMap["onPremisesSyncEnabled"] = user.GetOnPremisesSyncEnabled()

		// Get assigned licenses
		if user.GetAssignedLicenses() != nil {
			var licenses []map[string]interface{}
			for _, license := range user.GetAssignedLicenses() {
				licenses = append(licenses, map[string]interface{}{
					"skuId": license.GetSkuId(),
				})
			}
			resourceMap["assignedLicenses"] = licenses
		}

		// Try to get authentication methods for the user
		authMethods, err := r.client.Users().ByUserId(*user.GetId()).Authentication().Methods().Get(ctx, nil)
		if err == nil && authMethods.GetValue() != nil {
			var methods []map[string]interface{}
			for _, method := range authMethods.GetValue() {
				methods = append(methods, map[string]interface{}{
					"id":   method.GetId(),
					"type": method.GetOdataType(),
				})
			}
			resourceMap["authenticationMethods"] = methods
			resourceMap["mfaEnabled"] = len(methods) > 1 // More than just password
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
