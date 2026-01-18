// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/organizations/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/kopexa-grc/kspec/core"
)

// STSClient is the interface for STS operations.
type STSClient interface {
	GetCallerIdentity(ctx context.Context, params *sts.GetCallerIdentityInput, optFns ...func(*sts.Options)) (*sts.GetCallerIdentityOutput, error)
}

// AccountIAMClient is the interface for IAM operations used by Account.
type AccountIAMClient interface {
	ListAccountAliases(ctx context.Context, params *iam.ListAccountAliasesInput, optFns ...func(*iam.Options)) (*iam.ListAccountAliasesOutput, error)
	GetAccountSummary(ctx context.Context, params *iam.GetAccountSummaryInput, optFns ...func(*iam.Options)) (*iam.GetAccountSummaryOutput, error)
	GetAccountPasswordPolicy(ctx context.Context, params *iam.GetAccountPasswordPolicyInput, optFns ...func(*iam.Options)) (*iam.GetAccountPasswordPolicyOutput, error)
}

// OrganizationsClient is the interface for Organizations operations.
type OrganizationsClient interface {
	DescribeOrganization(ctx context.Context, params *organizations.DescribeOrganizationInput, optFns ...func(*organizations.Options)) (*organizations.DescribeOrganizationOutput, error)
	ListAccounts(ctx context.Context, params *organizations.ListAccountsInput, optFns ...func(*organizations.Options)) (*organizations.ListAccountsOutput, error)
	ListRoots(ctx context.Context, params *organizations.ListRootsInput, optFns ...func(*organizations.Options)) (*organizations.ListRootsOutput, error)
	ListOrganizationalUnitsForParent(ctx context.Context, params *organizations.ListOrganizationalUnitsForParentInput, optFns ...func(*organizations.Options)) (*organizations.ListOrganizationalUnitsForParentOutput, error)
}

// Account fetches AWS account information.
type Account struct {
	stsClient STSClient
	iamClient AccountIAMClient
	orgClient OrganizationsClient
}

// NewAccount creates a new Account resource.
func NewAccount(stsClient STSClient, iamClient AccountIAMClient, orgClient OrganizationsClient) *Account {
	return &Account{
		stsClient: stsClient,
		iamClient: iamClient,
		orgClient: orgClient,
	}
}

// Name returns the resource type name.
func (r *Account) Name() string {
	return "aws_account"
}

// Fetch retrieves AWS account information.
func (r *Account) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	identity, err := r.stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, core.WrapError("aws", "account", "get_caller_identity", err)
	}

	resource := make(core.Resource)
	resource["id"] = aws.ToString(identity.Account)
	resource["account_id"] = aws.ToString(identity.Account)
	resource["arn"] = aws.ToString(identity.Arn)
	resource["user_id"] = aws.ToString(identity.UserId)

	// Get account aliases
	aliasResp, err := r.iamClient.ListAccountAliases(ctx, &iam.ListAccountAliasesInput{})
	if err == nil && len(aliasResp.AccountAliases) > 0 {
		resource["aliases"] = aliasResp.AccountAliases
		resource["alias"] = aliasResp.AccountAliases[0]
	}

	// Try to get organization info (may fail if not in an org)
	orgResp, err := r.orgClient.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err == nil && orgResp.Organization != nil {
		resource["organization_id"] = aws.ToString(orgResp.Organization.Id)
		resource["organization_arn"] = aws.ToString(orgResp.Organization.Arn)
		resource["is_organization_member"] = true
		resource["is_management_account"] = aws.ToString(orgResp.Organization.MasterAccountId) == aws.ToString(identity.Account)
	} else {
		resource["is_organization_member"] = false
		resource["is_management_account"] = false
	}

	// Get account summary for computed fields
	summaryResp, err := r.iamClient.GetAccountSummary(ctx, &iam.GetAccountSummaryInput{})
	if err == nil && summaryResp.SummaryMap != nil {
		resource["users_count"] = summaryResp.SummaryMap["Users"]
		resource["groups_count"] = summaryResp.SummaryMap["Groups"]
		resource["roles_count"] = summaryResp.SummaryMap["Roles"]
		resource["policies_count"] = summaryResp.SummaryMap["Policies"]
		resource["mfa_devices_count"] = summaryResp.SummaryMap["MFADevices"]
		resource["access_keys_count"] = summaryResp.SummaryMap["AccessKeysPerUserQuota"]

		// Check if account has MFA enabled (root account MFA)
		if mfaDevices, ok := summaryResp.SummaryMap["AccountMFAEnabled"]; ok {
			resource["account_mfa_enabled"] = mfaDevices > 0
		}
	}

	// Get password policy for compliance checks
	passwordPolicy, err := r.iamClient.GetAccountPasswordPolicy(ctx, &iam.GetAccountPasswordPolicyInput{})
	if err == nil && passwordPolicy.PasswordPolicy != nil {
		pp := passwordPolicy.PasswordPolicy
		resource["has_password_policy"] = true
		resource["password_policy"] = map[string]any{
			"minimum_length":            pp.MinimumPasswordLength,
			"require_symbols":           pp.RequireSymbols,
			"require_numbers":           pp.RequireNumbers,
			"require_uppercase":         pp.RequireUppercaseCharacters,
			"require_lowercase":         pp.RequireLowercaseCharacters,
			"allow_users_to_change":     pp.AllowUsersToChangePassword,
			"max_age_days":              pp.MaxPasswordAge,
			"password_reuse_prevention": pp.PasswordReusePrevention,
			"hard_expiry":               pp.HardExpiry,
		}

		// Computed: Check password policy strength
		isStrong := true
		if pp.MinimumPasswordLength == nil || *pp.MinimumPasswordLength < 14 {
			isStrong = false
		}
		if !pp.RequireSymbols || !pp.RequireNumbers || !pp.RequireUppercaseCharacters || !pp.RequireLowercaseCharacters {
			isStrong = false
		}
		resource["has_strong_password_policy"] = isStrong
	} else {
		resource["has_password_policy"] = false
		resource["has_strong_password_policy"] = false
	}

	return []core.Resource{resource}, nil
}

// Organization fetches AWS Organizations information.
type Organization struct {
	orgClient OrganizationsClient
}

// NewOrganization creates a new Organization resource.
func NewOrganization(orgClient OrganizationsClient) *Organization {
	return &Organization{orgClient: orgClient}
}

// Name returns the resource type name.
func (r *Organization) Name() string {
	return "aws_organization"
}

// Fetch retrieves AWS Organizations information.
func (r *Organization) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	// Get organization details
	orgResp, err := r.orgClient.DescribeOrganization(ctx, &organizations.DescribeOrganizationInput{})
	if err != nil {
		// Not in an organization - return empty (this is expected behavior)
		return []core.Resource{}, nil //nolint:nilerr // intentionally returning empty slice when not in org
	}

	if orgResp.Organization == nil {
		return []core.Resource{}, nil
	}

	org := orgResp.Organization
	resource := make(core.Resource)
	resource["id"] = aws.ToString(org.Id)
	resource["arn"] = aws.ToString(org.Arn)
	resource["master_account_id"] = aws.ToString(org.MasterAccountId)
	resource["master_account_arn"] = aws.ToString(org.MasterAccountArn)
	resource["master_account_email"] = aws.ToString(org.MasterAccountEmail)
	resource["feature_set"] = string(org.FeatureSet)

	// Computed: Check if all features are enabled
	resource["all_features_enabled"] = org.FeatureSet == "ALL"

	// List accounts in the organization
	accountsPaginator := organizations.NewListAccountsPaginator(r.orgClient, &organizations.ListAccountsInput{})
	var accounts []map[string]any
	var activeCount, suspendedCount int

	for accountsPaginator.HasMorePages() {
		page, err := accountsPaginator.NextPage(ctx)
		if err != nil {
			break
		}
		for _, account := range page.Accounts {
			acc := map[string]any{
				"id":     aws.ToString(account.Id),
				"arn":    aws.ToString(account.Arn),
				"name":   aws.ToString(account.Name),
				"email":  aws.ToString(account.Email),
				"status": string(account.Status),
			}
			if account.JoinedTimestamp != nil {
				acc["joined_at"] = account.JoinedTimestamp.String()
			}
			accounts = append(accounts, acc)

			switch account.Status {
			case types.AccountStatusActive:
				activeCount++
			case types.AccountStatusSuspended:
				suspendedCount++
			case types.AccountStatusPendingClosure:
				// Pending closure accounts are not counted
			}
		}
	}

	resource["accounts"] = accounts
	resource["account_count"] = len(accounts)
	resource["active_account_count"] = activeCount
	resource["suspended_account_count"] = suspendedCount

	// List organizational units
	rootsResp, err := r.orgClient.ListRoots(ctx, &organizations.ListRootsInput{})
	if err == nil && len(rootsResp.Roots) > 0 {
		root := rootsResp.Roots[0]
		resource["root_id"] = aws.ToString(root.Id)
		resource["root_arn"] = aws.ToString(root.Arn)
		resource["root_name"] = aws.ToString(root.Name)

		// Count OUs
		ouPaginator := organizations.NewListOrganizationalUnitsForParentPaginator(r.orgClient, &organizations.ListOrganizationalUnitsForParentInput{
			ParentId: root.Id,
		})
		var ouCount int
		for ouPaginator.HasMorePages() {
			page, err := ouPaginator.NextPage(ctx)
			if err != nil {
				break
			}
			ouCount += len(page.OrganizationalUnits)
		}
		resource["organizational_unit_count"] = ouCount
	}

	// List enabled policy types
	if err == nil && len(rootsResp.Roots) > 0 {
		root := rootsResp.Roots[0]
		var policyTypes []string
		for _, pt := range root.PolicyTypes {
			if pt.Status == "ENABLED" {
				policyTypes = append(policyTypes, string(pt.Type))
			}
		}
		resource["enabled_policy_types"] = policyTypes
		resource["has_scp_enabled"] = contains(policyTypes, "SERVICE_CONTROL_POLICY")
		resource["has_tag_policy_enabled"] = contains(policyTypes, "TAG_POLICY")
		resource["has_backup_policy_enabled"] = contains(policyTypes, "BACKUP_POLICY")
		resource["has_aiservices_opt_out_policy_enabled"] = contains(policyTypes, "AISERVICES_OPT_OUT_POLICY")
	}

	return []core.Resource{resource}, nil
}

// contains checks if a string slice contains a value.
func contains(slice []string, val string) bool {
	for _, s := range slice {
		if s == val {
			return true
		}
	}
	return false
}
