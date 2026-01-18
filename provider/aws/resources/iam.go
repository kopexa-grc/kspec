// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

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

// IAMClient is the interface for IAM operations.
type IAMClient interface {
	ListUsers(ctx context.Context, params *iam.ListUsersInput, optFns ...func(*iam.Options)) (*iam.ListUsersOutput, error)
	GetLoginProfile(ctx context.Context, params *iam.GetLoginProfileInput, optFns ...func(*iam.Options)) (*iam.GetLoginProfileOutput, error)
	ListMFADevices(ctx context.Context, params *iam.ListMFADevicesInput, optFns ...func(*iam.Options)) (*iam.ListMFADevicesOutput, error)
	ListAccessKeys(ctx context.Context, params *iam.ListAccessKeysInput, optFns ...func(*iam.Options)) (*iam.ListAccessKeysOutput, error)
	ListGroupsForUser(ctx context.Context, params *iam.ListGroupsForUserInput, optFns ...func(*iam.Options)) (*iam.ListGroupsForUserOutput, error)
	ListAttachedUserPolicies(ctx context.Context, params *iam.ListAttachedUserPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedUserPoliciesOutput, error)
	ListUserPolicies(ctx context.Context, params *iam.ListUserPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListUserPoliciesOutput, error)
	ListRoles(ctx context.Context, params *iam.ListRolesInput, optFns ...func(*iam.Options)) (*iam.ListRolesOutput, error)
	ListAttachedRolePolicies(ctx context.Context, params *iam.ListAttachedRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedRolePoliciesOutput, error)
	ListRolePolicies(ctx context.Context, params *iam.ListRolePoliciesInput, optFns ...func(*iam.Options)) (*iam.ListRolePoliciesOutput, error)
	ListGroups(ctx context.Context, params *iam.ListGroupsInput, optFns ...func(*iam.Options)) (*iam.ListGroupsOutput, error)
	GetGroup(ctx context.Context, params *iam.GetGroupInput, optFns ...func(*iam.Options)) (*iam.GetGroupOutput, error)
	ListAttachedGroupPolicies(ctx context.Context, params *iam.ListAttachedGroupPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListAttachedGroupPoliciesOutput, error)
	ListGroupPolicies(ctx context.Context, params *iam.ListGroupPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListGroupPoliciesOutput, error)
	ListPolicies(ctx context.Context, params *iam.ListPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListPoliciesOutput, error)
	GetPolicyVersion(ctx context.Context, params *iam.GetPolicyVersionInput, optFns ...func(*iam.Options)) (*iam.GetPolicyVersionOutput, error)
}

// IAMUser fetches IAM users.
type IAMUser struct {
	client    IAMClient
	accountID string
}

// NewIAMUser creates a new IAMUser resource.
func NewIAMUser(client IAMClient, accountID string) *IAMUser {
	return &IAMUser{client: client, accountID: accountID}
}

// Name returns the resource type name.
func (r *IAMUser) Name() string {
	return "aws_iam_user"
}

// Fetch retrieves all IAM users with their security details.
func (r *IAMUser) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource
	paginator := iam.NewListUsersPaginator(r.client, &iam.ListUsersInput{})

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
			_, err := r.client.GetLoginProfile(ctx, &iam.GetLoginProfileInput{
				UserName: user.UserName,
			})
			resource["has_console_access"] = err == nil

			// Get MFA devices
			mfaResp, err := r.client.ListMFADevices(ctx, &iam.ListMFADevicesInput{
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
			keysResp, err := r.client.ListAccessKeys(ctx, &iam.ListAccessKeysInput{
				UserName: user.UserName,
			})
			if err == nil {
				resource["has_access_keys"] = len(keysResp.AccessKeyMetadata) > 0
				resource["access_key_count"] = len(keysResp.AccessKeyMetadata)

				var oldestKeyAge int
				var hasActiveKeys bool
				var activeKeyCount int
				for _, key := range keysResp.AccessKeyMetadata {
					if key.Status == types.StatusTypeActive {
						hasActiveKeys = true
						activeKeyCount++
					}
					if key.CreateDate != nil {
						keyAge := int(time.Since(*key.CreateDate).Hours() / 24)
						if keyAge > oldestKeyAge {
							oldestKeyAge = keyAge
						}
					}
				}
				resource["has_active_access_keys"] = hasActiveKeys
				resource["active_access_key_count"] = activeKeyCount
				resource["access_key_age_days"] = oldestKeyAge
				resource["access_keys_need_rotation"] = oldestKeyAge > 90
			} else {
				resource["has_access_keys"] = false
				resource["access_key_count"] = 0
				resource["active_access_key_count"] = 0
			}

			// Get user groups
			groupsResp, err := r.client.ListGroupsForUser(ctx, &iam.ListGroupsForUserInput{
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
			attachedPolicies, err := r.client.ListAttachedUserPolicies(ctx, &iam.ListAttachedUserPoliciesInput{
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
			inlinePolicies, err := r.client.ListUserPolicies(ctx, &iam.ListUserPoliciesInput{
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

			// Computed: Days since last activity (use password_last_used or create date)
			switch {
			case user.PasswordLastUsed != nil:
				resource["last_activity_days"] = int(time.Since(*user.PasswordLastUsed).Hours() / 24)
			case user.CreateDate != nil:
				// If never logged in, use account creation date
				resource["last_activity_days"] = int(time.Since(*user.CreateDate).Hours() / 24)
			default:
				resource["last_activity_days"] = -1 // Unknown
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// IAMRole fetches IAM roles.
type IAMRole struct {
	client    IAMClient
	accountID string
}

// NewIAMRole creates a new IAMRole resource.
func NewIAMRole(client IAMClient, accountID string) *IAMRole {
	return &IAMRole{client: client, accountID: accountID}
}

// Name returns the resource type name.
func (r *IAMRole) Name() string {
	return "aws_iam_role"
}

// Fetch retrieves all IAM roles.
func (r *IAMRole) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource
	paginator := iam.NewListRolesPaginator(r.client, &iam.ListRolesInput{})

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
					resource["allows_external_assume"] = checkExternalTrust(policy, r.accountID)
					resource["allows_cross_account"] = checkCrossAccountTrust(policy, r.accountID)
					resource["trusted_principals"] = extractTrustedPrincipals(policy)
				}
			}

			// Get attached policies
			attachedPolicies, err := r.client.ListAttachedRolePolicies(ctx, &iam.ListAttachedRolePoliciesInput{
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
			inlinePolicies, err := r.client.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
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

// IAMGroup fetches IAM groups.
type IAMGroup struct {
	client IAMClient
}

// NewIAMGroup creates a new IAMGroup resource.
func NewIAMGroup(client IAMClient) *IAMGroup {
	return &IAMGroup{client: client}
}

// Name returns the resource type name.
func (r *IAMGroup) Name() string {
	return "aws_iam_group"
}

// Fetch retrieves all IAM groups.
func (r *IAMGroup) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource
	paginator := iam.NewListGroupsPaginator(r.client, &iam.ListGroupsInput{})

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
			membersResp, err := r.client.GetGroup(ctx, &iam.GetGroupInput{
				GroupName: group.GroupName,
			})
			if err == nil {
				var memberNames []string
				for _, user := range membersResp.Users {
					memberNames = append(memberNames, aws.ToString(user.UserName))
				}
				resource["members"] = memberNames
				resource["member_count"] = len(memberNames)
				resource["user_count"] = len(memberNames) // Alias for policy compatibility
				resource["has_members"] = len(memberNames) > 0
			} else {
				resource["member_count"] = 0
				resource["user_count"] = 0
				resource["has_members"] = false
			}

			// Get attached policies
			attachedPolicies, err := r.client.ListAttachedGroupPolicies(ctx, &iam.ListAttachedGroupPoliciesInput{
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
			inlinePolicies, err := r.client.ListGroupPolicies(ctx, &iam.ListGroupPoliciesInput{
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

// IAMPolicy fetches IAM policies.
type IAMPolicy struct {
	client IAMClient
}

// NewIAMPolicy creates a new IAMPolicy resource.
func NewIAMPolicy(client IAMClient) *IAMPolicy {
	return &IAMPolicy{client: client}
}

// Name returns the resource type name.
func (r *IAMPolicy) Name() string {
	return "aws_iam_policy"
}

// Fetch retrieves all customer-managed IAM policies.
func (r *IAMPolicy) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource
	paginator := iam.NewListPoliciesPaginator(r.client, &iam.ListPoliciesInput{
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
				versionResp, err := r.client.GetPolicyVersion(ctx, &iam.GetPolicyVersionInput{
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
						// is_admin_policy: true if policy allows all actions on all resources
						resource["is_admin_policy"] = hasStarAction && hasStarResource
					} else {
						resource["is_admin_policy"] = false
					}
				} else {
					resource["is_admin_policy"] = false
				}
			} else {
				resource["is_admin_policy"] = false
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
