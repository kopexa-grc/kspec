package aws

import (
	"context"
	"encoding/json"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"

	"github.com/kopexa-grc/kspec/core"
)

// IAMUserResource fetches IAM users.
type IAMUserResource struct {
	conn   *Connection
	client IAMAPI // optional, for testing
}

// Name returns the resource type name.
func (r *IAMUserResource) Name() string {
	return "aws_iam_user"
}

// getClient returns the IAM client, using the injected mock if available.
func (r *IAMUserResource) getClient() IAMAPI {
	if r.client != nil {
		return r.client
	}
	return iam.NewFromConfig(r.conn.cfg)
}

// Fetch retrieves all IAM users with their security details.
func (r *IAMUserResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	client := r.getClient()

	var resources []core.Resource
	paginator := iam.NewListUsersPaginator(client, &iam.ListUsersInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, core.WrapError("aws", "iam_user", "list", err)
		}

		for _, user := range page.Users {
			resource := make(core.Resource)
			resource["id"] = aws.ToString(user.UserId)
			resource["arn"] = aws.ToString(user.Arn)
			resource["name"] = aws.ToString(user.UserName)
			resource["path"] = aws.ToString(user.Path)

			if user.CreateDate != nil {
				resource["created_at"] = user.CreateDate.Format(time.RFC3339)
				resource["age_days"] = int(time.Since(*user.CreateDate).Hours() / 24)
			}
			if user.PasswordLastUsed != nil {
				resource["password_last_used"] = user.PasswordLastUsed.Format(time.RFC3339)
				resource["password_last_used_days"] = int(time.Since(*user.PasswordLastUsed).Hours() / 24)
			}

			// Get login profile (console access)
			_, err := client.GetLoginProfile(ctx, &iam.GetLoginProfileInput{
				UserName: user.UserName,
			})
			resource["has_console_access"] = err == nil

			// Get MFA devices
			mfaResp, err := client.ListMFADevices(ctx, &iam.ListMFADevicesInput{
				UserName: user.UserName,
			})
			if err == nil {
				resource["has_mfa"] = len(mfaResp.MFADevices) > 0
				resource["mfa_device_count"] = len(mfaResp.MFADevices)

				mfaTypes := make([]string, 0, len(mfaResp.MFADevices))
				for _, device := range mfaResp.MFADevices {
					serial := aws.ToString(device.SerialNumber)
					if strings.Contains(serial, "mfa/") {
						mfaTypes = append(mfaTypes, "virtual")
					} else {
						mfaTypes = append(mfaTypes, "hardware")
					}
				}
				resource["mfa_types"] = mfaTypes
			} else {
				resource["has_mfa"] = false
				resource["mfa_device_count"] = 0
			}

			// Get access keys
			keysResp, err := client.ListAccessKeys(ctx, &iam.ListAccessKeysInput{
				UserName: user.UserName,
			})
			if err == nil {
				resource["has_access_keys"] = len(keysResp.AccessKeyMetadata) > 0
				resource["access_key_count"] = len(keysResp.AccessKeyMetadata)

				var oldestKeyAge int
				var hasActiveKeys bool
				for _, key := range keysResp.AccessKeyMetadata {
					if key.Status == types.StatusTypeActive {
						hasActiveKeys = true
					}
					if key.CreateDate != nil {
						keyAge := int(time.Since(*key.CreateDate).Hours() / 24)
						if keyAge > oldestKeyAge {
							oldestKeyAge = keyAge
						}
					}
				}
				resource["has_active_access_keys"] = hasActiveKeys
				resource["access_key_age_days"] = oldestKeyAge
				resource["access_keys_need_rotation"] = oldestKeyAge > 90
			} else {
				resource["has_access_keys"] = false
				resource["access_key_count"] = 0
			}

			// Get user groups
			groupsResp, err := client.ListGroupsForUser(ctx, &iam.ListGroupsForUserInput{
				UserName: user.UserName,
			})
			if err == nil {
				groupNames := make([]string, 0, len(groupsResp.Groups))
				for _, group := range groupsResp.Groups {
					groupNames = append(groupNames, aws.ToString(group.GroupName))
				}
				resource["groups"] = groupNames
				resource["group_count"] = len(groupNames)
			}

			// Get attached policies
			attachedPolicies, err := client.ListAttachedUserPolicies(ctx, &iam.ListAttachedUserPoliciesInput{
				UserName: user.UserName,
			})
			if err == nil {
				policyArns := make([]string, 0, len(attachedPolicies.AttachedPolicies))
				var isAdmin bool
				for _, policy := range attachedPolicies.AttachedPolicies {
					arn := aws.ToString(policy.PolicyArn)
					policyArns = append(policyArns, arn)
					if strings.Contains(arn, "AdministratorAccess") {
						isAdmin = true
					}
				}
				resource["attached_policies"] = policyArns
				resource["attached_policy_count"] = len(policyArns)
				resource["is_admin"] = isAdmin
			}

			// Get inline policies
			inlinePolicies, err := client.ListUserPolicies(ctx, &iam.ListUserPoliciesInput{
				UserName: user.UserName,
			})
			if err == nil {
				resource["inline_policies"] = inlinePolicies.PolicyNames
				resource["inline_policy_count"] = len(inlinePolicies.PolicyNames)
				resource["has_inline_policies"] = len(inlinePolicies.PolicyNames) > 0
			}

			// Computed: User without MFA but with console access
			hasConsole, _ := resource["has_console_access"].(bool) //nolint:errcheck // zero value is intentional default
			hasMFA, _ := resource["has_mfa"].(bool)                //nolint:errcheck // zero value is intentional default
			resource["console_without_mfa"] = hasConsole && !hasMFA

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// IAMRoleResource fetches IAM roles.
type IAMRoleResource struct {
	conn   *Connection
	client IAMAPI // optional, for testing
}

// Name returns the resource type name.
func (r *IAMRoleResource) Name() string {
	return "aws_iam_role"
}

// getClient returns the IAM client, using the injected mock if available.
func (r *IAMRoleResource) getClient() IAMAPI {
	if r.client != nil {
		return r.client
	}
	return iam.NewFromConfig(r.conn.cfg)
}

// Fetch retrieves all IAM roles.
func (r *IAMRoleResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	client := r.getClient()

	var resources []core.Resource
	paginator := iam.NewListRolesPaginator(client, &iam.ListRolesInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, core.WrapError("aws", "iam_role", "list", err)
		}

		for _, role := range page.Roles {
			resource := make(core.Resource)
			resource["id"] = aws.ToString(role.RoleId)
			resource["arn"] = aws.ToString(role.Arn)
			resource["name"] = aws.ToString(role.RoleName)
			resource["path"] = aws.ToString(role.Path)
			resource["description"] = aws.ToString(role.Description)

			if role.CreateDate != nil {
				resource["created_at"] = role.CreateDate.Format(time.RFC3339)
				resource["age_days"] = int(time.Since(*role.CreateDate).Hours() / 24)
			}

			if role.MaxSessionDuration != nil {
				resource["max_session_duration"] = *role.MaxSessionDuration
			}

			// Parse assume role policy
			if role.AssumeRolePolicyDocument != nil {
				policyDoc := aws.ToString(role.AssumeRolePolicyDocument)
				decoded, err := url.QueryUnescape(policyDoc)
				if err == nil {
					policyDoc = decoded
				}
				resource["assume_role_policy"] = policyDoc

				// Analyze trust policy
				var policy map[string]any
				if err := json.Unmarshal([]byte(policyDoc), &policy); err == nil {
					// Check for overly permissive trust
					resource["allows_external_assume"] = checkExternalTrust(policy, r.conn.accountID)
					resource["allows_cross_account"] = checkCrossAccountTrust(policy, r.conn.accountID)
					resource["trusted_principals"] = extractTrustedPrincipals(policy)
				}
			}

			// Get attached policies
			attachedPolicies, err := client.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
				RoleName: role.RoleName,
			})
			if err == nil {
				var policyArns []string
				var isAdmin bool
				for _, policy := range attachedPolicies.AttachedPolicies {
					arn := aws.ToString(policy.PolicyArn)
					policyArns = append(policyArns, arn)
					if strings.Contains(arn, "AdministratorAccess") {
						isAdmin = true
					}
				}
				resource["attached_policies"] = policyArns
				resource["attached_policy_count"] = len(policyArns)
				resource["is_admin"] = isAdmin
			}

			// Get inline policies
			inlinePolicies, err := client.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
				RoleName: role.RoleName,
			})
			if err == nil {
				resource["inline_policies"] = inlinePolicies.PolicyNames
				resource["inline_policy_count"] = len(inlinePolicies.PolicyNames)
				resource["has_inline_policies"] = len(inlinePolicies.PolicyNames) > 0
			}

			// Computed fields
			isServiceLinkedRole := strings.HasPrefix(aws.ToString(role.Path), "/aws-service-role/")
			resource["is_service_role"] = isServiceLinkedRole
			resource["is_service_linked_role"] = isServiceLinkedRole

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// IAMGroupResource fetches IAM groups.
type IAMGroupResource struct {
	conn   *Connection
	client IAMAPI // optional, for testing
}

// Name returns the resource type name.
func (r *IAMGroupResource) Name() string {
	return "aws_iam_group"
}

// getClient returns the IAM client, using the injected mock if available.
func (r *IAMGroupResource) getClient() IAMAPI {
	if r.client != nil {
		return r.client
	}
	return iam.NewFromConfig(r.conn.cfg)
}

// Fetch retrieves all IAM groups.
func (r *IAMGroupResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	client := r.getClient()

	var resources []core.Resource
	paginator := iam.NewListGroupsPaginator(client, &iam.ListGroupsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, core.WrapError("aws", "iam_group", "list", err)
		}

		for _, group := range page.Groups {
			resource := make(core.Resource)
			resource["id"] = aws.ToString(group.GroupId)
			resource["arn"] = aws.ToString(group.Arn)
			resource["name"] = aws.ToString(group.GroupName)
			resource["path"] = aws.ToString(group.Path)

			if group.CreateDate != nil {
				resource["created_at"] = group.CreateDate.Format(time.RFC3339)
			}

			// Get group members
			membersResp, err := client.GetGroup(ctx, &iam.GetGroupInput{
				GroupName: group.GroupName,
			})
			if err == nil {
				var memberNames []string
				for _, user := range membersResp.Users {
					memberNames = append(memberNames, aws.ToString(user.UserName))
				}
				resource["members"] = memberNames
				resource["member_count"] = len(memberNames)
				resource["has_members"] = len(memberNames) > 0
			}

			// Get attached policies
			attachedPolicies, err := client.ListAttachedGroupPolicies(ctx, &iam.ListAttachedGroupPoliciesInput{
				GroupName: group.GroupName,
			})
			if err == nil {
				var policyArns []string
				var isAdmin bool
				for _, policy := range attachedPolicies.AttachedPolicies {
					arn := aws.ToString(policy.PolicyArn)
					policyArns = append(policyArns, arn)
					if strings.Contains(arn, "AdministratorAccess") {
						isAdmin = true
					}
				}
				resource["attached_policies"] = policyArns
				resource["attached_policy_count"] = len(policyArns)
				resource["is_admin_group"] = isAdmin
			}

			// Get inline policies
			inlinePolicies, err := client.ListGroupPolicies(ctx, &iam.ListGroupPoliciesInput{
				GroupName: group.GroupName,
			})
			if err == nil {
				resource["inline_policies"] = inlinePolicies.PolicyNames
				resource["inline_policy_count"] = len(inlinePolicies.PolicyNames)
				resource["has_inline_policies"] = len(inlinePolicies.PolicyNames) > 0
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// IAMPolicyResource fetches IAM policies.
type IAMPolicyResource struct {
	conn   *Connection
	client IAMAPI // optional, for testing
}

// Name returns the resource type name.
func (r *IAMPolicyResource) Name() string {
	return "aws_iam_policy"
}

// getClient returns the IAM client, using the injected mock if available.
func (r *IAMPolicyResource) getClient() IAMAPI {
	if r.client != nil {
		return r.client
	}
	return iam.NewFromConfig(r.conn.cfg)
}

// Fetch retrieves all customer-managed IAM policies.
func (r *IAMPolicyResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	client := r.getClient()

	var resources []core.Resource
	paginator := iam.NewListPoliciesPaginator(client, &iam.ListPoliciesInput{
		Scope: types.PolicyScopeTypeLocal, // Only customer-managed policies
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, core.WrapError("aws", "iam_policy", "list", err)
		}

		for _, policy := range page.Policies {
			resource := make(core.Resource)
			resource["id"] = aws.ToString(policy.PolicyId)
			resource["arn"] = aws.ToString(policy.Arn)
			resource["name"] = aws.ToString(policy.PolicyName)
			resource["path"] = aws.ToString(policy.Path)
			resource["description"] = aws.ToString(policy.Description)

			if policy.CreateDate != nil {
				resource["created_at"] = policy.CreateDate.Format(time.RFC3339)
			}
			if policy.UpdateDate != nil {
				resource["updated_at"] = policy.UpdateDate.Format(time.RFC3339)
			}

			resource["attachment_count"] = policy.AttachmentCount
			resource["is_attachable"] = policy.IsAttachable
			resource["default_version_id"] = aws.ToString(policy.DefaultVersionId)

			// Computed fields
			resource["is_attached"] = policy.AttachmentCount != nil && *policy.AttachmentCount > 0

			// Get policy document
			if policy.DefaultVersionId != nil {
				versionResp, err := client.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{
					PolicyArn: policy.Arn,
					VersionId: policy.DefaultVersionId,
				})
				if err == nil && versionResp.PolicyVersion != nil && versionResp.PolicyVersion.Document != nil {
					policyDoc := aws.ToString(versionResp.PolicyVersion.Document)
					decoded, err := url.QueryUnescape(policyDoc)
					if err == nil {
						policyDoc = decoded
					}
					resource["policy_document"] = policyDoc

					// Analyze policy for security issues
					var policyObj map[string]any
					if err := json.Unmarshal([]byte(policyDoc), &policyObj); err == nil {
						hasStarAction, hasStarResource := analyzePolicyPermissions(policyObj)
						resource["has_wildcard_actions"] = hasStarAction
						resource["has_wildcard_resources"] = hasStarResource
						resource["is_overly_permissive"] = hasStarAction && hasStarResource
					}
				}
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// checkExternalTrust checks if a trust policy allows external entities.
func checkExternalTrust(policy map[string]any, accountID string) bool {
	statements, ok := policy["Statement"].([]any)
	if !ok {
		return false
	}

	for _, stmt := range statements {
		statement, ok := stmt.(map[string]any)
		if !ok {
			continue
		}

		principal := statement["Principal"]
		if principal == "*" {
			return true
		}

		if principalMap, ok := principal.(map[string]any); ok {
			if awsPrincipal, ok := principalMap["AWS"]; ok {
				if awsPrincipal == "*" {
					return true
				}
				// Check for external account IDs
				if arnStr, ok := awsPrincipal.(string); ok {
					if !strings.Contains(arnStr, accountID) && !strings.HasPrefix(arnStr, "arn:aws:iam::"+accountID) {
						return true
					}
				}
			}
		}
	}

	return false
}

// checkCrossAccountTrust checks if a trust policy allows cross-account access.
func checkCrossAccountTrust(policy map[string]any, accountID string) bool {
	statements, ok := policy["Statement"].([]any)
	if !ok {
		return false
	}

	for _, stmt := range statements {
		statement, ok := stmt.(map[string]any)
		if !ok {
			continue
		}

		principal := statement["Principal"]
		if principalMap, ok := principal.(map[string]any); ok {
			if awsPrincipal, ok := principalMap["AWS"]; ok {
				// Check for different account IDs
				checkArn := func(arn string) bool {
					if strings.HasPrefix(arn, "arn:aws:iam::") {
						parts := strings.Split(arn, ":")
						if len(parts) >= 5 && parts[4] != accountID {
							return true
						}
					}
					return false
				}

				switch v := awsPrincipal.(type) {
				case string:
					if checkArn(v) {
						return true
					}
				case []any:
					for _, a := range v {
						if arnStr, ok := a.(string); ok && checkArn(arnStr) {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

// extractTrustedPrincipals extracts all trusted principals from a trust policy.
func extractTrustedPrincipals(policy map[string]any) []string {
	var principals []string
	statements, ok := policy["Statement"].([]any)
	if !ok {
		return principals
	}

	for _, stmt := range statements {
		statement, ok := stmt.(map[string]any)
		if !ok {
			continue
		}

		principal := statement["Principal"]
		if principal == "*" {
			principals = append(principals, "*")
			continue
		}

		if principalMap, ok := principal.(map[string]any); ok {
			for _, v := range principalMap {
				switch val := v.(type) {
				case string:
					principals = append(principals, val)
				case []any:
					for _, item := range val {
						if s, ok := item.(string); ok {
							principals = append(principals, s)
						}
					}
				}
			}
		}
	}

	return principals
}

// analyzePolicyPermissions checks for wildcard actions and resources.
func analyzePolicyPermissions(policy map[string]any) (hasStarAction, hasStarResource bool) {
	statements, ok := policy["Statement"].([]any)
	if !ok {
		return false, false
	}

	for _, stmt := range statements {
		statement, ok := stmt.(map[string]any)
		if !ok {
			continue
		}

		// Only check Allow statements
		if effect, ok := statement["Effect"].(string); ok && effect != policyEffectAllow {
			continue
		}

		// Check actions
		if action := statement["Action"]; action != nil {
			if action == "*" {
				hasStarAction = true
			} else if actionStr, ok := action.(string); ok && (actionStr == "*" || strings.HasSuffix(actionStr, ":*")) {
				hasStarAction = true
			} else if actionList, ok := action.([]any); ok {
				for _, a := range actionList {
					if aStr, ok := a.(string); ok && (aStr == "*" || strings.HasSuffix(aStr, ":*")) {
						hasStarAction = true
						break
					}
				}
			}
		}

		// Check resources
		if resource := statement["Resource"]; resource != nil {
			if resource == "*" {
				hasStarResource = true
			} else if resourceStr, ok := resource.(string); ok && resourceStr == "*" {
				hasStarResource = true
			} else if resourceList, ok := resource.([]any); ok {
				for _, r := range resourceList {
					if rStr, ok := r.(string); ok && rStr == "*" {
						hasStarResource = true
						break
					}
				}
			}
		}
	}

	return hasStarAction, hasStarResource
}
