// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/kopexa-grc/kspec/core"
)

// AuthorizationPolicy provides access to Azure AD authorization policies.
type AuthorizationPolicy struct {
	client *msgraphsdk.GraphServiceClient
}

// NewAuthorizationPolicy creates a new AuthorizationPolicy resource.
func NewAuthorizationPolicy(client *msgraphsdk.GraphServiceClient) *AuthorizationPolicy {
	return &AuthorizationPolicy{client: client}
}

// Name returns the resource type identifier for authorization policies.
func (r *AuthorizationPolicy) Name() string {
	return "ms365_authorization_policy"
}

// Fetch retrieves the tenant authorization policy with user permissions.
func (r *AuthorizationPolicy) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	resources := make([]core.Resource, 0, 1)

	result, err := r.client.Policies().AuthorizationPolicy().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get authorization policy: %w", err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authorization policy: %w", err)
	}

	var resourceMap map[string]interface{}
	if err := json.Unmarshal(data, &resourceMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authorization policy: %w", err)
	}

	resourceMap["id"] = result.GetId()
	resourceMap["displayName"] = result.GetDisplayName()
	resourceMap["description"] = result.GetDescription()

	// Key security settings
	resourceMap["allowedToSignUpEmailBasedSubscriptions"] = result.GetAllowedToSignUpEmailBasedSubscriptions()
	resourceMap["allowedToUseSSPR"] = result.GetAllowedToUseSSPR()
	resourceMap["allowEmailVerifiedUsersToJoinOrganization"] = result.GetAllowEmailVerifiedUsersToJoinOrganization()
	resourceMap["allowInvitesFrom"] = result.GetAllowInvitesFrom()
	resourceMap["blockMsolPowerShell"] = result.GetBlockMsolPowerShell()
	resourceMap["guestUserRoleId"] = result.GetGuestUserRoleId()

	// Default user role permissions
	if result.GetDefaultUserRolePermissions() != nil {
		perms := result.GetDefaultUserRolePermissions()
		resourceMap["defaultUserRolePermissions"] = map[string]interface{}{
			"allowedToCreateApps":                      perms.GetAllowedToCreateApps(),
			"allowedToCreateSecurityGroups":            perms.GetAllowedToCreateSecurityGroups(),
			"allowedToCreateTenants":                   perms.GetAllowedToCreateTenants(),
			"allowedToReadBitlockerKeysForOwnedDevice": perms.GetAllowedToReadBitlockerKeysForOwnedDevice(),
			"allowedToReadOtherUsers":                  perms.GetAllowedToReadOtherUsers(),
		}
	}

	resources = append(resources, resourceMap)
	return resources, nil
}

// AuthenticationMethodPolicy provides access to authentication method policies.
type AuthenticationMethodPolicy struct {
	client *msgraphsdk.GraphServiceClient
}

// NewAuthenticationMethodPolicy creates a new AuthenticationMethodPolicy resource.
func NewAuthenticationMethodPolicy(client *msgraphsdk.GraphServiceClient) *AuthenticationMethodPolicy {
	return &AuthenticationMethodPolicy{client: client}
}

// Name returns the resource type identifier for authentication method policies.
func (r *AuthenticationMethodPolicy) Name() string {
	return "ms365_authentication_method_policy"
}

// Fetch retrieves the authentication methods policy with method configurations.
func (r *AuthenticationMethodPolicy) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	resources := make([]core.Resource, 0, 1)

	result, err := r.client.Policies().AuthenticationMethodsPolicy().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication methods policy: %w", err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal authentication methods policy: %w", err)
	}

	var resourceMap map[string]interface{}
	if err := json.Unmarshal(data, &resourceMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal authentication methods policy: %w", err)
	}

	resourceMap["id"] = result.GetId()
	resourceMap["displayName"] = result.GetDisplayName()
	resourceMap["description"] = result.GetDescription()
	resourceMap["policyVersion"] = result.GetPolicyVersion()
	resourceMap["policyMigrationState"] = result.GetPolicyMigrationState()

	// Registration enforcement
	if result.GetRegistrationEnforcement() != nil {
		enforcement := result.GetRegistrationEnforcement()
		if enforcement.GetAuthenticationMethodsRegistrationCampaign() != nil {
			campaign := enforcement.GetAuthenticationMethodsRegistrationCampaign()
			resourceMap["registrationCampaign"] = map[string]interface{}{
				"state":                campaign.GetState(),
				"snoozeDurationInDays": campaign.GetSnoozeDurationInDays(),
			}
		}
	}

	// Authentication method configurations
	if result.GetAuthenticationMethodConfigurations() != nil {
		var methods []map[string]interface{}
		for _, method := range result.GetAuthenticationMethodConfigurations() {
			methodData, err := json.Marshal(method)
			if err != nil {
				continue
			}
			var methodMap map[string]interface{}
			if err := json.Unmarshal(methodData, &methodMap); err != nil {
				continue
			}
			methodMap["id"] = method.GetId()
			methodMap["state"] = method.GetState()
			methods = append(methods, methodMap)
		}
		resourceMap["authenticationMethodConfigurations"] = methods
	}

	resources = append(resources, resourceMap)
	return resources, nil
}

// IdentitySecurityDefaultsEnforcementPolicy provides access to security defaults.
type IdentitySecurityDefaultsEnforcementPolicy struct {
	client *msgraphsdk.GraphServiceClient
}

// NewIdentitySecurityDefaultsEnforcementPolicy creates a new IdentitySecurityDefaultsEnforcementPolicy resource.
func NewIdentitySecurityDefaultsEnforcementPolicy(client *msgraphsdk.GraphServiceClient) *IdentitySecurityDefaultsEnforcementPolicy {
	return &IdentitySecurityDefaultsEnforcementPolicy{client: client}
}

// Name returns the resource type identifier for security defaults policy.
func (r *IdentitySecurityDefaultsEnforcementPolicy) Name() string {
	return "ms365_security_defaults_policy"
}

// Fetch retrieves the security defaults enforcement policy.
func (r *IdentitySecurityDefaultsEnforcementPolicy) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	resources := make([]core.Resource, 0, 1)

	result, err := r.client.Policies().IdentitySecurityDefaultsEnforcementPolicy().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get security defaults policy: %w", err)
	}

	resourceMap := make(map[string]interface{})
	resourceMap["id"] = result.GetId()
	resourceMap["displayName"] = result.GetDisplayName()
	resourceMap["description"] = result.GetDescription()
	resourceMap["isEnabled"] = result.GetIsEnabled()

	resources = append(resources, resourceMap)
	return resources, nil
}
