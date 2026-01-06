package cloudflare

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/zero_trust"
	"github.com/kopexa-grc/kspec/core"
)

// AccessApplicationResource fetches Zero Trust Access applications.
type AccessApplicationResource struct {
	client    *cloudflare.Client
	accountID string
}

// Name returns the resource type name.
func (r *AccessApplicationResource) Name() string {
	return "cloudflare_access_application"
}

// Fetch retrieves all Access applications for the account.
func (r *AccessApplicationResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	accountID := r.accountID
	if accountID == "" {
		accountID = asset.Config["account_id"]
	}
	if accountID == "" {
		return nil, fmt.Errorf("account_id is required for Access application scanning")
	}

	var resources []core.Resource

	// List all Access applications
	iter := r.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(accountID),
	})

	for iter.Next() {
		app := iter.Current()

		data, err := json.Marshal(app)
		if err != nil {
			continue
		}

		var resourceMap map[string]interface{}
		if err := json.Unmarshal(data, &resourceMap); err != nil {
			continue
		}

		resourceMap["account_id"] = accountID

		// Add security-relevant computed fields
		r.addSecurityFields(resourceMap)

		resources = append(resources, resourceMap)
	}

	if err := iter.Err(); err != nil {
		return resources, nil
	}

	return resources, nil
}

func (r *AccessApplicationResource) addSecurityFields(app map[string]interface{}) {
	// Check application type
	if appType, ok := app["type"].(string); ok {
		app["is_self_hosted"] = appType == "self_hosted"
		app["is_saas"] = appType == "saas"
		app["is_ssh"] = appType == "ssh"
		app["is_vnc"] = appType == "vnc"
		app["is_bookmark"] = appType == "bookmark"
	}

	// Session duration check
	if sessionDuration, ok := app["session_duration"].(string); ok {
		app["has_session_limit"] = sessionDuration != "" && sessionDuration != "0"
	}

	// Check for CORS settings
	if corsHeaders, ok := app["cors_headers"].(map[string]interface{}); ok {
		app["has_cors_config"] = len(corsHeaders) > 0
		if allowAllOrigins, ok := corsHeaders["allow_all_origins"].(bool); ok {
			app["cors_allow_all_origins"] = allowAllOrigins
		}
	}

	// Check service auth settings
	if serviceAuth, ok := app["service_auth_401_redirect"].(bool); ok {
		app["has_service_auth_redirect"] = serviceAuth
	}

	// Check skip interstitial
	if skipInterstitial, ok := app["skip_interstitial"].(bool); ok {
		app["skips_interstitial"] = skipInterstitial
	}

	// Check auto redirect to IdP
	if autoRedirect, ok := app["auto_redirect_to_identity"].(bool); ok {
		app["auto_redirects_to_idp"] = autoRedirect
	}
}

// AccessPolicyResource fetches Zero Trust Access policies.
type AccessPolicyResource struct {
	client    *cloudflare.Client
	accountID string
}

// Name returns the resource type name.
func (r *AccessPolicyResource) Name() string {
	return "cloudflare_access_policy"
}

// Fetch retrieves all Access policies for applications.
func (r *AccessPolicyResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	accountID := r.accountID
	if accountID == "" {
		accountID = asset.Config["account_id"]
	}
	if accountID == "" {
		return nil, fmt.Errorf("account_id is required for Access policy scanning")
	}

	var resources []core.Resource

	// First, list all applications to get their policies
	appIter := r.client.ZeroTrust.Access.Applications.ListAutoPaging(ctx, zero_trust.AccessApplicationListParams{
		AccountID: cloudflare.F(accountID),
	})

	for appIter.Next() {
		app := appIter.Current()
		appID := app.ID

		// List policies for this application
		policyIter := r.client.ZeroTrust.Access.Applications.Policies.ListAutoPaging(ctx, appID, zero_trust.AccessApplicationPolicyListParams{
			AccountID: cloudflare.F(accountID),
		})

		for policyIter.Next() {
			policy := policyIter.Current()

			data, err := json.Marshal(policy)
			if err != nil {
				continue
			}

			var resourceMap map[string]interface{}
			if err := json.Unmarshal(data, &resourceMap); err != nil {
				continue
			}

			resourceMap["account_id"] = accountID
			resourceMap["application_id"] = appID

			// Add security-relevant computed fields
			r.addSecurityFields(resourceMap)

			resources = append(resources, resourceMap)
		}
	}

	return resources, nil
}

func (r *AccessPolicyResource) addSecurityFields(policy map[string]interface{}) {
	// Check policy decision
	if decision, ok := policy["decision"].(string); ok {
		policy["is_allow"] = decision == "allow"
		policy["is_deny"] = decision == "deny"
		policy["is_bypass"] = decision == "bypass"
		policy["is_non_identity"] = decision == "non_identity"
	}

	// Check for MFA requirement in include rules
	hasMFA := false
	if includes, ok := policy["include"].([]interface{}); ok {
		for _, rule := range includes {
			if r, ok := rule.(map[string]interface{}); ok {
				if authMethod, ok := r["auth_method"].(map[string]interface{}); ok {
					if method, ok := authMethod["auth_method"].(string); ok {
						if method == "mfa" {
							hasMFA = true
							break
						}
					}
				}
			}
		}
	}
	policy["requires_mfa"] = hasMFA

	// Check for geo restrictions
	hasGeoRestriction := false
	if includes, ok := policy["include"].([]interface{}); ok {
		for _, rule := range includes {
			if r, ok := rule.(map[string]interface{}); ok {
				if _, ok := r["geo"].(map[string]interface{}); ok {
					hasGeoRestriction = true
					break
				}
			}
		}
	}
	policy["has_geo_restriction"] = hasGeoRestriction

	// Check for device posture requirements
	hasDevicePosture := false
	if includes, ok := policy["include"].([]interface{}); ok {
		for _, rule := range includes {
			if r, ok := rule.(map[string]interface{}); ok {
				if _, ok := r["device_posture"].(map[string]interface{}); ok {
					hasDevicePosture = true
					break
				}
			}
		}
	}
	policy["has_device_posture"] = hasDevicePosture

	// Count rule groups
	if includes, ok := policy["include"].([]interface{}); ok {
		policy["include_rule_count"] = len(includes)
	}
	if excludes, ok := policy["exclude"].([]interface{}); ok {
		policy["exclude_rule_count"] = len(excludes)
	}
	if requires, ok := policy["require"].([]interface{}); ok {
		policy["require_rule_count"] = len(requires)
	}
}
