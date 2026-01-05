package ms365

import (
	"context"
	"encoding/json"
	"fmt"

	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"

	"github.com/juliankoehn/kspec/core"
)

type ApplicationResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *ApplicationResource) Name() string {
	return "ms365_application"
}

func (r *ApplicationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	result, err := r.client.Applications().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get applications: %w", err)
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, app := range result.GetValue() {
		data, err := json.Marshal(app)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = app.GetId()
		resourceMap["appId"] = app.GetAppId()
		resourceMap["displayName"] = app.GetDisplayName()
		resourceMap["createdDateTime"] = app.GetCreatedDateTime()
		resourceMap["signInAudience"] = app.GetSignInAudience()

		// Check for password credentials (secrets)
		if app.GetPasswordCredentials() != nil {
			var secrets []map[string]interface{}
			for _, cred := range app.GetPasswordCredentials() {
				secrets = append(secrets, map[string]interface{}{
					"keyId":       cred.GetKeyId(),
					"displayName": cred.GetDisplayName(),
					"startDateTime": cred.GetStartDateTime(),
					"endDateTime":   cred.GetEndDateTime(),
				})
			}
			resourceMap["passwordCredentials"] = secrets
			resourceMap["hasSecrets"] = len(secrets) > 0
		}

		// Check for certificate credentials
		if app.GetKeyCredentials() != nil {
			var certs []map[string]interface{}
			for _, cred := range app.GetKeyCredentials() {
				certs = append(certs, map[string]interface{}{
					"keyId":         cred.GetKeyId(),
					"displayName":   cred.GetDisplayName(),
					"startDateTime": cred.GetStartDateTime(),
					"endDateTime":   cred.GetEndDateTime(),
					"type":          cred.GetTypeEscaped(),
					"usage":         cred.GetUsage(),
				})
			}
			resourceMap["keyCredentials"] = certs
			resourceMap["hasCertificates"] = len(certs) > 0
		}

		// API permissions
		if app.GetRequiredResourceAccess() != nil {
			var permissions []map[string]interface{}
			for _, access := range app.GetRequiredResourceAccess() {
				resourceAccess := map[string]interface{}{
					"resourceAppId": access.GetResourceAppId(),
				}
				var perms []map[string]interface{}
				for _, perm := range access.GetResourceAccess() {
					perms = append(perms, map[string]interface{}{
						"id":   perm.GetId(),
						"type": perm.GetTypeEscaped(),
					})
				}
				resourceAccess["resourceAccess"] = perms
				permissions = append(permissions, resourceAccess)
			}
			resourceMap["requiredResourceAccess"] = permissions
		}

		// Web configuration
		if app.GetWeb() != nil {
			web := app.GetWeb()
			webConfig := map[string]interface{}{
				"redirectUris": web.GetRedirectUris(),
			}
			if web.GetImplicitGrantSettings() != nil {
				webConfig["enableIdTokenIssuance"] = web.GetImplicitGrantSettings().GetEnableIdTokenIssuance()
				webConfig["enableAccessTokenIssuance"] = web.GetImplicitGrantSettings().GetEnableAccessTokenIssuance()
			}
			resourceMap["web"] = webConfig
		}

		// Public client configuration
		if app.GetPublicClient() != nil {
			resourceMap["publicClient"] = map[string]interface{}{
				"redirectUris": app.GetPublicClient().GetRedirectUris(),
			}
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}

type ServicePrincipalResource struct {
	client *msgraphsdk.GraphServiceClient
}

func (r *ServicePrincipalResource) Name() string {
	return "ms365_service_principal"
}

func (r *ServicePrincipalResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	result, err := r.client.ServicePrincipals().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get service principals: %w", err)
	}

	if result.GetValue() == nil {
		return resources, nil
	}

	for _, sp := range result.GetValue() {
		data, err := json.Marshal(sp)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["id"] = sp.GetId()
		resourceMap["appId"] = sp.GetAppId()
		resourceMap["displayName"] = sp.GetDisplayName()
		resourceMap["servicePrincipalType"] = sp.GetServicePrincipalType()
		resourceMap["accountEnabled"] = sp.GetAccountEnabled()
		resourceMap["appOwnerOrganizationId"] = sp.GetAppOwnerOrganizationId()

		// App role assignment required
		resourceMap["appRoleAssignmentRequired"] = sp.GetAppRoleAssignmentRequired()

		// Tags
		resourceMap["tags"] = sp.GetTags()

		// Check for credentials
		if sp.GetPasswordCredentials() != nil {
			resourceMap["hasSecrets"] = len(sp.GetPasswordCredentials()) > 0
		}
		if sp.GetKeyCredentials() != nil {
			resourceMap["hasCertificates"] = len(sp.GetKeyCredentials()) > 0
		}

		// OAuth2 permission grants (delegated permissions)
		oauth2Grants, err := r.client.ServicePrincipals().ByServicePrincipalId(*sp.GetId()).Oauth2PermissionGrants().Get(ctx, nil)
		if err == nil && oauth2Grants.GetValue() != nil {
			var grants []map[string]interface{}
			for _, grant := range oauth2Grants.GetValue() {
				grants = append(grants, map[string]interface{}{
					"id":          grant.GetId(),
					"consentType": grant.GetConsentType(),
					"scope":       grant.GetScope(),
					"resourceId":  grant.GetResourceId(),
				})
			}
			resourceMap["oauth2PermissionGrants"] = grants
		}

		// App role assignments
		appRoles, err := r.client.ServicePrincipals().ByServicePrincipalId(*sp.GetId()).AppRoleAssignments().Get(ctx, nil)
		if err == nil && appRoles.GetValue() != nil {
			var roles []map[string]interface{}
			for _, role := range appRoles.GetValue() {
				roles = append(roles, map[string]interface{}{
					"id":                   role.GetId(),
					"appRoleId":            role.GetAppRoleId(),
					"resourceDisplayName":  role.GetResourceDisplayName(),
					"resourceId":           role.GetResourceId(),
				})
			}
			resourceMap["appRoleAssignments"] = roles
		}

		resources = append(resources, resourceMap)
	}

	return resources, nil
}
