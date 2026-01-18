// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"

	"github.com/kopexa-grc/kspec/core"
)

// SSMClient is the interface for SSM operations.
//
//nolint:dupl // AWS SDK client interfaces have similar structures by design
type SSMClient interface {
	DescribeInstanceInformation(ctx context.Context, params *ssm.DescribeInstanceInformationInput, optFns ...func(*ssm.Options)) (*ssm.DescribeInstanceInformationOutput, error)
	DescribeParameters(ctx context.Context, params *ssm.DescribeParametersInput, optFns ...func(*ssm.Options)) (*ssm.DescribeParametersOutput, error)
	ListDocuments(ctx context.Context, params *ssm.ListDocumentsInput, optFns ...func(*ssm.Options)) (*ssm.ListDocumentsOutput, error)
	DescribePatchBaselines(ctx context.Context, params *ssm.DescribePatchBaselinesInput, optFns ...func(*ssm.Options)) (*ssm.DescribePatchBaselinesOutput, error)
	GetPatchBaseline(ctx context.Context, params *ssm.GetPatchBaselineInput, optFns ...func(*ssm.Options)) (*ssm.GetPatchBaselineOutput, error)
}

// SSMClientFactory creates SSM clients for specific regions.
type SSMClientFactory func(region string) SSMClient

// SSMInstance fetches SSM managed instances.
type SSMInstance struct {
	clientFactory SSMClientFactory
	accountID     string
	regions       []string
}

// NewSSMInstance creates a new SSMInstance resource.
func NewSSMInstance(clientFactory SSMClientFactory, accountID string, regions []string) *SSMInstance {
	return &SSMInstance{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *SSMInstance) Name() string {
	return "aws_ssm_instance"
}

// Fetch retrieves all SSM managed instances across configured regions.
func (r *SSMInstance) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := ssm.NewDescribeInstanceInformationPaginator(client, &ssm.DescribeInstanceInformationInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, instance := range page.InstanceInformationList {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(instance.InstanceId)
				resource["instance_id"] = aws.ToString(instance.InstanceId)
				resource["region"] = region

				resource["name"] = aws.ToString(instance.Name)
				resource["computer_name"] = aws.ToString(instance.ComputerName)
				resource["platform_type"] = string(instance.PlatformType)
				resource["platform_name"] = aws.ToString(instance.PlatformName)
				resource["platform_version"] = aws.ToString(instance.PlatformVersion)
				resource["agent_version"] = aws.ToString(instance.AgentVersion)
				resource["ip_address"] = aws.ToString(instance.IPAddress)

				// Ping status
				resource["ping_status"] = string(instance.PingStatus)
				resource["is_online"] = instance.PingStatus == types.PingStatusOnline
				resource["is_connection_lost"] = instance.PingStatus == types.PingStatusConnectionLost

				// Last ping
				if instance.LastPingDateTime != nil {
					resource["last_ping_time"] = instance.LastPingDateTime.String()
					hoursSinceLastPing := int(time.Since(*instance.LastPingDateTime).Hours())
					resource["hours_since_last_ping"] = hoursSinceLastPing
					resource["is_stale"] = hoursSinceLastPing > 24
				}

				// Association status
				resource["association_status"] = aws.ToString(instance.AssociationStatus)
				resource["has_associations"] = aws.ToString(instance.AssociationStatus) != ""

				// Resource type (EC2, on-premises, etc.)
				resource["resource_type"] = string(instance.ResourceType)
				resource["is_ec2"] = instance.ResourceType == types.ResourceTypeEc2Instance
				resource["is_managed_instance"] = instance.ResourceType == types.ResourceTypeManagedInstance

				// IAM role
				resource["iam_role"] = aws.ToString(instance.IamRole)
				resource["has_iam_role"] = instance.IamRole != nil && aws.ToString(instance.IamRole) != ""

				// Registration date
				if instance.RegistrationDate != nil {
					resource["registration_date"] = instance.RegistrationDate.String()
				}

				// Source type and ID
				resource["source_type"] = string(instance.SourceType)
				resource["source_id"] = aws.ToString(instance.SourceId)

				// Computed
				resource["is_windows"] = instance.PlatformType == types.PlatformTypeWindows
				resource["is_linux"] = instance.PlatformType == types.PlatformTypeLinux
				resource["is_macos"] = instance.PlatformType == types.PlatformTypeMacos

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// SSMParameter fetches SSM parameters.
type SSMParameter struct {
	clientFactory SSMClientFactory
	accountID     string
	regions       []string
}

// NewSSMParameter creates a new SSMParameter resource.
func NewSSMParameter(clientFactory SSMClientFactory, accountID string, regions []string) *SSMParameter {
	return &SSMParameter{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *SSMParameter) Name() string {
	return "aws_ssm_parameter"
}

// Fetch retrieves all SSM parameters across configured regions.
func (r *SSMParameter) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := ssm.NewDescribeParametersPaginator(client, &ssm.DescribeParametersInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, param := range page.Parameters {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(param.Name)
				resource["name"] = aws.ToString(param.Name)
				resource["region"] = region

				resource["type"] = string(param.Type)
				resource["description"] = aws.ToString(param.Description)
				resource["tier"] = string(param.Tier)
				resource["version"] = param.Version
				resource["data_type"] = aws.ToString(param.DataType)

				// Key ID (for SecureString)
				resource["key_id"] = aws.ToString(param.KeyId)
				resource["has_custom_kms"] = param.KeyId != nil && aws.ToString(param.KeyId) != "" && aws.ToString(param.KeyId) != "alias/aws/ssm"

				// Allowed pattern
				resource["allowed_pattern"] = aws.ToString(param.AllowedPattern)
				resource["has_validation"] = param.AllowedPattern != nil && aws.ToString(param.AllowedPattern) != ""

				// Last modified
				if param.LastModifiedDate != nil {
					resource["last_modified_date"] = param.LastModifiedDate.String()
					daysSinceModified := int(time.Since(*param.LastModifiedDate).Hours() / 24)
					resource["days_since_modified"] = daysSinceModified
				}

				resource["last_modified_user"] = aws.ToString(param.LastModifiedUser)

				// Policies
				policyNames := make([]string, 0, len(param.Policies))
				for _, policy := range param.Policies {
					policyNames = append(policyNames, aws.ToString(policy.PolicyType))
				}
				resource["policies"] = policyNames
				resource["policy_count"] = len(param.Policies)
				resource["has_policies"] = len(param.Policies) > 0

				// Computed type checks
				resource["is_string"] = param.Type == types.ParameterTypeString
				resource["is_string_list"] = param.Type == types.ParameterTypeStringList
				resource["is_secure_string"] = param.Type == types.ParameterTypeSecureString

				// Tier checks
				resource["is_standard"] = param.Tier == types.ParameterTierStandard
				resource["is_advanced"] = param.Tier == types.ParameterTierAdvanced
				resource["is_intelligent_tiering"] = param.Tier == types.ParameterTierIntelligentTiering

				// Security check: SecureString should use CMK
				isSecure, _ := resource["is_secure_string"].(bool)   //nolint:errcheck // zero value ok
				hasCustomKMS, _ := resource["has_custom_kms"].(bool) //nolint:errcheck // zero value ok
				resource["secure_with_cmk"] = isSecure && hasCustomKMS

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// SSMDocument fetches SSM documents.
type SSMDocument struct {
	clientFactory SSMClientFactory
	accountID     string
	regions       []string
}

// NewSSMDocument creates a new SSMDocument resource.
func NewSSMDocument(clientFactory SSMClientFactory, accountID string, regions []string) *SSMDocument {
	return &SSMDocument{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *SSMDocument) Name() string {
	return "aws_ssm_document"
}

// Fetch retrieves SSM documents owned by the account across configured regions.
func (r *SSMDocument) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		// Only get documents owned by this account
		paginator := ssm.NewListDocumentsPaginator(client, &ssm.ListDocumentsInput{
			Filters: []types.DocumentKeyValuesFilter{
				{
					Key:    aws.String("Owner"),
					Values: []string{"Self"},
				},
			},
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, doc := range page.DocumentIdentifiers {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(doc.Name)
				resource["name"] = aws.ToString(doc.Name)
				resource["region"] = region

				resource["owner"] = aws.ToString(doc.Owner)
				resource["document_version"] = aws.ToString(doc.DocumentVersion)
				resource["document_type"] = string(doc.DocumentType)
				resource["document_format"] = string(doc.DocumentFormat)
				resource["schema_version"] = aws.ToString(doc.SchemaVersion)
				resource["target_type"] = aws.ToString(doc.TargetType)

				// Platform types
				platforms := make([]string, 0, len(doc.PlatformTypes))
				for _, p := range doc.PlatformTypes {
					platforms = append(platforms, string(p))
				}
				resource["platform_types"] = platforms

				// Requires
				requires := make([]string, 0, len(doc.Requires))
				for _, req := range doc.Requires {
					requires = append(requires, aws.ToString(req.Name))
				}
				resource["requires"] = requires
				resource["has_dependencies"] = len(requires) > 0

				// Tags
				tags := make(map[string]string)
				for _, tag := range doc.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				// Created date
				if doc.CreatedDate != nil {
					resource["created_date"] = doc.CreatedDate.String()
				}

				// Review status
				resource["review_status"] = string(doc.ReviewStatus)
				resource["is_approved"] = doc.ReviewStatus == types.ReviewStatusApproved
				resource["is_pending"] = doc.ReviewStatus == types.ReviewStatusPending
				resource["is_rejected"] = doc.ReviewStatus == types.ReviewStatusRejected

				// Document type checks
				resource["is_command"] = doc.DocumentType == types.DocumentTypeCommand
				resource["is_policy"] = doc.DocumentType == types.DocumentTypePolicy
				resource["is_automation"] = doc.DocumentType == types.DocumentTypeAutomation
				resource["is_session"] = doc.DocumentType == types.DocumentTypeSession
				resource["is_package"] = doc.DocumentType == types.DocumentTypePackage

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// SSMPatchBaseline fetches SSM patch baselines.
type SSMPatchBaseline struct {
	clientFactory SSMClientFactory
	accountID     string
	regions       []string
}

// NewSSMPatchBaseline creates a new SSMPatchBaseline resource.
func NewSSMPatchBaseline(clientFactory SSMClientFactory, accountID string, regions []string) *SSMPatchBaseline {
	return &SSMPatchBaseline{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *SSMPatchBaseline) Name() string {
	return "aws_ssm_patchbaseline"
}

// Fetch retrieves SSM patch baselines across configured regions.
func (r *SSMPatchBaseline) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := ssm.NewDescribePatchBaselinesPaginator(client, &ssm.DescribePatchBaselinesInput{
			Filters: []types.PatchOrchestratorFilter{
				{
					Key:    aws.String("OWNER"),
					Values: []string{"Self"},
				},
			},
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, baseline := range page.BaselineIdentities {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(baseline.BaselineId)
				resource["baseline_id"] = aws.ToString(baseline.BaselineId)
				resource["name"] = aws.ToString(baseline.BaselineName)
				resource["region"] = region

				resource["description"] = aws.ToString(baseline.BaselineDescription)
				resource["operating_system"] = string(baseline.OperatingSystem)
				resource["is_default_baseline"] = baseline.DefaultBaseline

				// Get full baseline details
				baselineResp, err := client.GetPatchBaseline(ctx, &ssm.GetPatchBaselineInput{
					BaselineId: baseline.BaselineId,
				})
				if err == nil {
					resource["approved_patches"] = baselineResp.ApprovedPatches
					resource["approved_patch_count"] = len(baselineResp.ApprovedPatches)
					resource["rejected_patches"] = baselineResp.RejectedPatches
					resource["rejected_patch_count"] = len(baselineResp.RejectedPatches)
					resource["approved_patches_compliance_level"] = string(baselineResp.ApprovedPatchesComplianceLevel)
					resource["approved_patches_enable_non_security"] = baselineResp.ApprovedPatchesEnableNonSecurity

					// Patch filters (rules)
					if baselineResp.GlobalFilters != nil {
						resource["global_filter_count"] = len(baselineResp.GlobalFilters.PatchFilters)
					}

					// Approval rules
					if baselineResp.ApprovalRules != nil {
						resource["approval_rule_count"] = len(baselineResp.ApprovalRules.PatchRules)
						resource["has_approval_rules"] = len(baselineResp.ApprovalRules.PatchRules) > 0
					}

					// Sources (for Linux)
					resource["source_count"] = len(baselineResp.Sources)
					resource["has_custom_sources"] = len(baselineResp.Sources) > 0

					// Created/modified
					if baselineResp.CreatedDate != nil {
						resource["created_date"] = baselineResp.CreatedDate.String()
					}
					if baselineResp.ModifiedDate != nil {
						resource["modified_date"] = baselineResp.ModifiedDate.String()
					}
				}

				// OS type checks
				resource["is_windows"] = baseline.OperatingSystem == types.OperatingSystemWindows
				resource["is_amazon_linux"] = baseline.OperatingSystem == types.OperatingSystemAmazonLinux ||
					baseline.OperatingSystem == types.OperatingSystemAmazonLinux2 ||
					baseline.OperatingSystem == types.OperatingSystemAmazonLinux2022 ||
					baseline.OperatingSystem == types.OperatingSystemAmazonLinux2023
				resource["is_ubuntu"] = baseline.OperatingSystem == types.OperatingSystemUbuntu
				resource["is_redhat"] = baseline.OperatingSystem == types.OperatingSystemRedhatEnterpriseLinux
				resource["is_centos"] = baseline.OperatingSystem == types.OperatingSystemCentOS
				resource["is_debian"] = baseline.OperatingSystem == types.OperatingSystemDebian
				resource["is_macos"] = baseline.OperatingSystem == types.OperatingSystemMacOS

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
