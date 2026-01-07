package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"

	"github.com/kopexa-grc/kspec/core"
)

// SecurityHubResource fetches Security Hub configuration.
type SecurityHubResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *SecurityHubResource) Name() string {
	return "aws_securityhub"
}

// Fetch retrieves Security Hub configuration across configured regions.
func (r *SecurityHubResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	resources := make([]core.Resource, 0, len(r.conn.regions))

	for _, region := range r.conn.regions {
		client := securityhub.NewFromConfig(r.conn.ConfigForRegion(region))

		// Get hub resource
		hubResp, err := client.DescribeHub(ctx, &securityhub.DescribeHubInput{})
		if err != nil {
			// Security Hub not enabled in this region
			resource := make(core.Resource)
			resource["id"] = "not_enabled"
			resource["region"] = region
			resource["is_enabled"] = false
			resources = append(resources, resource)
			continue
		}

		resource := make(core.Resource)
		resource["id"] = aws.ToString(hubResp.HubArn)
		resource["arn"] = aws.ToString(hubResp.HubArn)
		resource["region"] = region

		resource["is_enabled"] = true
		resource["auto_enable_controls"] = hubResp.AutoEnableControls
		resource["control_finding_generator"] = string(hubResp.ControlFindingGenerator)

		if hubResp.SubscribedAt != nil {
			resource["subscribed_at"] = aws.ToString(hubResp.SubscribedAt)
		}

		// Get enabled standards
		standardsResp, err := client.GetEnabledStandards(ctx, &securityhub.GetEnabledStandardsInput{})
		if err == nil {
			enabledStandards := make([]string, 0, len(standardsResp.StandardsSubscriptions))
			standardsArns := make([]string, 0, len(standardsResp.StandardsSubscriptions))
			for _, sub := range standardsResp.StandardsSubscriptions {
				enabledStandards = append(enabledStandards, aws.ToString(sub.StandardsArn))
				standardsArns = append(standardsArns, aws.ToString(sub.StandardsSubscriptionArn))
			}
			resource["enabled_standards"] = enabledStandards
			resource["standards_subscription_arns"] = standardsArns
			resource["enabled_standards_count"] = len(enabledStandards)

			// Check for specific standards
			resource["has_aws_foundational_security"] = containsStandard(enabledStandards, "aws-foundational-security-best-practices")
			resource["has_cis_benchmark"] = containsStandard(enabledStandards, "cis-aws-foundations-benchmark")
			resource["has_pci_dss"] = containsStandard(enabledStandards, "pci-dss")
			resource["has_nist"] = containsStandard(enabledStandards, "nist")
		} else {
			resource["enabled_standards_count"] = 0
		}

		// Get integrations
		integrationsResp, err := client.ListEnabledProductsForImport(ctx, &securityhub.ListEnabledProductsForImportInput{})
		if err == nil {
			resource["enabled_products"] = integrationsResp.ProductSubscriptions
			resource["enabled_products_count"] = len(integrationsResp.ProductSubscriptions)
		} else {
			resource["enabled_products_count"] = 0
		}

		// Get finding aggregator (cross-region)
		aggregatorResp, err := client.GetFindingAggregator(ctx, &securityhub.GetFindingAggregatorInput{
			FindingAggregatorArn: hubResp.HubArn, // Use hub ARN as starting point
		})
		if err == nil {
			resource["finding_aggregator_arn"] = aws.ToString(aggregatorResp.FindingAggregatorArn)
			resource["region_linking_mode"] = aws.ToString(aggregatorResp.RegionLinkingMode)
			resource["linked_regions"] = aggregatorResp.Regions
			resource["has_cross_region_aggregation"] = true
		} else {
			resource["has_cross_region_aggregation"] = false
		}

		// Get admin account status
		adminResp, err := client.GetAdministratorAccount(ctx, &securityhub.GetAdministratorAccountInput{})
		if err == nil && adminResp.Administrator != nil {
			resource["administrator_account_id"] = aws.ToString(adminResp.Administrator.AccountId)
			resource["administrator_status"] = aws.ToString(adminResp.Administrator.MemberStatus)
			resource["is_member_account"] = true
		} else {
			resource["is_member_account"] = false
		}

		// Get member accounts (if admin)
		membersResp, err := client.ListMembers(ctx, &securityhub.ListMembersInput{})
		if err == nil {
			resource["member_count"] = len(membersResp.Members)
			memberAccountIDs := make([]string, 0, len(membersResp.Members))
			for _, member := range membersResp.Members {
				memberAccountIDs = append(memberAccountIDs, aws.ToString(member.AccountId))
			}
			resource["member_account_ids"] = memberAccountIDs
			resource["is_admin_account"] = len(membersResp.Members) > 0
		} else {
			resource["member_count"] = 0
			resource["is_admin_account"] = false
		}

		// Get organization configuration
		orgConfigResp, err := client.DescribeOrganizationConfiguration(ctx, &securityhub.DescribeOrganizationConfigurationInput{})
		if err == nil {
			resource["auto_enable_organization"] = orgConfigResp.AutoEnable
			resource["auto_enable_standards"] = string(orgConfigResp.AutoEnableStandards)
			resource["is_organization_config"] = true
			if orgConfigResp.OrganizationConfiguration != nil {
				resource["organization_config_type"] = string(orgConfigResp.OrganizationConfiguration.ConfigurationType)
			}
		} else {
			resource["is_organization_config"] = false
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func containsStandard(standards []string, search string) bool {
	for _, s := range standards {
		if containsSubstr(s, search) {
			return true
		}
	}
	return false
}

func containsSubstr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && hasSubstring(s, substr))
}

func hasSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// SecurityHubFindingResource fetches Security Hub findings summary.
type SecurityHubFindingResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *SecurityHubFindingResource) Name() string {
	return "aws_securityhub_finding"
}

// Fetch retrieves Security Hub findings summary across configured regions.
func (r *SecurityHubFindingResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := securityhub.NewFromConfig(r.conn.ConfigForRegion(region))

		// Get findings with specific filter for active findings
		findingsResp, err := client.GetFindings(ctx, &securityhub.GetFindingsInput{
			Filters: &types.AwsSecurityFindingFilters{
				RecordState: []types.StringFilter{
					{
						Comparison: types.StringFilterComparisonEquals,
						Value:      aws.String("ACTIVE"),
					},
				},
				WorkflowStatus: []types.StringFilter{
					{
						Comparison: types.StringFilterComparisonNotEquals,
						Value:      aws.String("SUPPRESSED"),
					},
				},
			},
			MaxResults: aws.Int32(100),
		})
		if err != nil {
			continue
		}

		for _, finding := range findingsResp.Findings {
			resource := make(core.Resource)
			resource["id"] = aws.ToString(finding.Id)
			resource["region"] = region

			resource["title"] = aws.ToString(finding.Title)
			resource["description"] = aws.ToString(finding.Description)
			resource["product_arn"] = aws.ToString(finding.ProductArn)
			resource["product_name"] = aws.ToString(finding.ProductName)
			resource["company_name"] = aws.ToString(finding.CompanyName)
			resource["generator_id"] = aws.ToString(finding.GeneratorId)
			resource["aws_account_id"] = aws.ToString(finding.AwsAccountId)

			// Types
			resource["types"] = finding.Types
			resource["type_count"] = len(finding.Types)

			// Severity
			if finding.Severity != nil {
				resource["severity_label"] = string(finding.Severity.Label)
				resource["severity_normalized"] = finding.Severity.Normalized

				resource["is_critical"] = finding.Severity.Label == types.SeverityLabelCritical
				resource["is_high"] = finding.Severity.Label == types.SeverityLabelHigh
				resource["is_medium"] = finding.Severity.Label == types.SeverityLabelMedium
				resource["is_low"] = finding.Severity.Label == types.SeverityLabelLow
				resource["is_informational"] = finding.Severity.Label == types.SeverityLabelInformational
			}

			// Compliance
			if finding.Compliance != nil {
				resource["compliance_status"] = string(finding.Compliance.Status)
				resource["is_passed"] = finding.Compliance.Status == types.ComplianceStatusPassed
				resource["is_failed"] = finding.Compliance.Status == types.ComplianceStatusFailed
				resource["is_warning"] = finding.Compliance.Status == types.ComplianceStatusWarning
				resource["is_not_available"] = finding.Compliance.Status == types.ComplianceStatusNotAvailable

				relatedRequirements := make([]string, 0, len(finding.Compliance.RelatedRequirements))
				relatedRequirements = append(relatedRequirements, finding.Compliance.RelatedRequirements...)
				resource["related_requirements"] = relatedRequirements
			}

			// Workflow
			if finding.Workflow != nil {
				resource["workflow_status"] = string(finding.Workflow.Status)
				resource["is_new"] = finding.Workflow.Status == types.WorkflowStatusNew
				resource["is_notified"] = finding.Workflow.Status == types.WorkflowStatusNotified
				resource["is_resolved"] = finding.Workflow.Status == types.WorkflowStatusResolved
				resource["is_suppressed"] = finding.Workflow.Status == types.WorkflowStatusSuppressed
			}

			// Record state
			resource["record_state"] = string(finding.RecordState)

			// Resources
			if len(finding.Resources) > 0 {
				resourceTypes := make([]string, 0, len(finding.Resources))
				resourceIds := make([]string, 0, len(finding.Resources))
				for _, res := range finding.Resources {
					resourceTypes = append(resourceTypes, aws.ToString(res.Type))
					resourceIds = append(resourceIds, aws.ToString(res.Id))
				}
				resource["resource_types"] = resourceTypes
				resource["resource_ids"] = resourceIds
				resource["affected_resource_count"] = len(finding.Resources)
			}

			// Timestamps
			if finding.CreatedAt != nil {
				resource["created_at"] = aws.ToString(finding.CreatedAt)
			}
			if finding.UpdatedAt != nil {
				resource["updated_at"] = aws.ToString(finding.UpdatedAt)
			}
			if finding.FirstObservedAt != nil {
				resource["first_observed_at"] = aws.ToString(finding.FirstObservedAt)
			}
			if finding.LastObservedAt != nil {
				resource["last_observed_at"] = aws.ToString(finding.LastObservedAt)
			}

			// Remediation
			if finding.Remediation != nil && finding.Remediation.Recommendation != nil {
				resource["remediation_text"] = aws.ToString(finding.Remediation.Recommendation.Text)
				resource["remediation_url"] = aws.ToString(finding.Remediation.Recommendation.Url)
				resource["has_remediation"] = true
			} else {
				resource["has_remediation"] = false
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// SecurityHubStandardResource fetches Security Hub standards and their controls.
type SecurityHubStandardResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *SecurityHubStandardResource) Name() string {
	return "aws_securityhub_standard"
}

// Fetch retrieves Security Hub standards across configured regions.
func (r *SecurityHubStandardResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := securityhub.NewFromConfig(r.conn.ConfigForRegion(region))

		// Get enabled standards
		standardsResp, err := client.GetEnabledStandards(ctx, &securityhub.GetEnabledStandardsInput{})
		if err != nil {
			continue
		}

		for _, sub := range standardsResp.StandardsSubscriptions {
			resource := make(core.Resource)
			resource["id"] = aws.ToString(sub.StandardsSubscriptionArn)
			resource["arn"] = aws.ToString(sub.StandardsSubscriptionArn)
			resource["standards_arn"] = aws.ToString(sub.StandardsArn)
			resource["region"] = region

			resource["status"] = string(sub.StandardsStatus)
			resource["status_reason"] = string(sub.StandardsStatusReason.StatusReasonCode)

			// Computed
			resource["is_ready"] = sub.StandardsStatus == types.StandardsStatusReady
			resource["is_incomplete"] = sub.StandardsStatus == types.StandardsStatusIncomplete
			resource["is_deleting"] = sub.StandardsStatus == types.StandardsStatusDeleting
			resource["is_pending"] = sub.StandardsStatus == types.StandardsStatusPending
			resource["is_failed"] = sub.StandardsStatus == types.StandardsStatusFailed

			// Get controls for this standard
			controlsResp, err := client.DescribeStandardsControls(ctx, &securityhub.DescribeStandardsControlsInput{
				StandardsSubscriptionArn: sub.StandardsSubscriptionArn,
			})
			if err == nil {
				enabledCount := 0
				disabledCount := 0
				var disabledControls []string

				for _, control := range controlsResp.Controls {
					if control.ControlStatus == types.ControlStatusEnabled {
						enabledCount++
					} else {
						disabledCount++
						disabledControls = append(disabledControls, aws.ToString(control.ControlId))
					}
				}

				resource["total_control_count"] = len(controlsResp.Controls)
				resource["enabled_control_count"] = enabledCount
				resource["disabled_control_count"] = disabledCount
				resource["disabled_controls"] = disabledControls
				resource["all_controls_enabled"] = disabledCount == 0
			}

			resources = append(resources, resource)
		}
	}

	return resources, nil
}
