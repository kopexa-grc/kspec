// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/guardduty/types"

	"github.com/kopexa-grc/kspec/core"
)

// GuardDutyClient is the interface for GuardDuty operations.
type GuardDutyClient interface {
	ListDetectors(ctx context.Context, params *guardduty.ListDetectorsInput, optFns ...func(*guardduty.Options)) (*guardduty.ListDetectorsOutput, error)
	GetDetector(ctx context.Context, params *guardduty.GetDetectorInput, optFns ...func(*guardduty.Options)) (*guardduty.GetDetectorOutput, error)
	ListFindings(ctx context.Context, params *guardduty.ListFindingsInput, optFns ...func(*guardduty.Options)) (*guardduty.ListFindingsOutput, error)
	GetFindings(ctx context.Context, params *guardduty.GetFindingsInput, optFns ...func(*guardduty.Options)) (*guardduty.GetFindingsOutput, error)
}

// GuardDutyClientFactory creates GuardDuty clients for specific regions.
type GuardDutyClientFactory func(region string) GuardDutyClient

// GuardDutyDetector fetches GuardDuty detectors.
type GuardDutyDetector struct {
	clientFactory GuardDutyClientFactory
	accountID     string
	regions       []string
}

// NewGuardDutyDetector creates a new GuardDutyDetector resource.
func NewGuardDutyDetector(clientFactory GuardDutyClientFactory, accountID string, regions []string) *GuardDutyDetector {
	return &GuardDutyDetector{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *GuardDutyDetector) Name() string {
	return "aws_guardduty_detector"
}

// Fetch retrieves all GuardDuty detectors across configured regions.
func (r *GuardDutyDetector) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		// List detectors
		listResp, err := client.ListDetectors(ctx, &guardduty.ListDetectorsInput{})
		if err != nil {
			continue
		}

		for _, detectorID := range listResp.DetectorIds {
			// Get detector details
			detectorResp, err := client.GetDetector(ctx, &guardduty.GetDetectorInput{
				DetectorId: aws.String(detectorID),
			})
			if err != nil {
				continue
			}

			resource := make(core.Resource)
			resource["id"] = detectorID
			resource["region"] = region

			resource["status"] = string(detectorResp.Status)
			resource["finding_publishing_frequency"] = string(detectorResp.FindingPublishingFrequency)
			resource["service_role"] = aws.ToString(detectorResp.ServiceRole)

			// Created/Updated at
			if detectorResp.CreatedAt != nil {
				resource["created_at"] = aws.ToString(detectorResp.CreatedAt)
			}
			if detectorResp.UpdatedAt != nil {
				resource["updated_at"] = aws.ToString(detectorResp.UpdatedAt)
			}

			// Data sources
			if detectorResp.DataSources != nil { //nolint:staticcheck // DataSources deprecated but still functional
				ds := detectorResp.DataSources //nolint:staticcheck // DataSources deprecated but still functional

				// CloudTrail
				if ds.CloudTrail != nil {
					resource["cloudtrail_status"] = string(ds.CloudTrail.Status)
					resource["has_cloudtrail"] = ds.CloudTrail.Status == types.DataSourceStatusEnabled
				}

				// DNS logs
				if ds.DNSLogs != nil {
					resource["dnslogs_status"] = string(ds.DNSLogs.Status)
					resource["has_dnslogs"] = ds.DNSLogs.Status == types.DataSourceStatusEnabled
				}

				// Flow logs
				if ds.FlowLogs != nil {
					resource["flowlogs_status"] = string(ds.FlowLogs.Status)
					resource["has_flowlogs"] = ds.FlowLogs.Status == types.DataSourceStatusEnabled
				}

				// S3 logs
				if ds.S3Logs != nil {
					resource["s3logs_status"] = string(ds.S3Logs.Status)
					resource["has_s3logs"] = ds.S3Logs.Status == types.DataSourceStatusEnabled
				}

				// Kubernetes
				if ds.Kubernetes != nil && ds.Kubernetes.AuditLogs != nil {
					resource["kubernetes_audit_status"] = string(ds.Kubernetes.AuditLogs.Status)
					resource["has_kubernetes_audit"] = ds.Kubernetes.AuditLogs.Status == types.DataSourceStatusEnabled
				}

				// Malware protection
				if ds.MalwareProtection != nil && ds.MalwareProtection.ScanEc2InstanceWithFindings != nil {
					if ds.MalwareProtection.ScanEc2InstanceWithFindings.EbsVolumes != nil {
						resource["malware_ebs_status"] = string(ds.MalwareProtection.ScanEc2InstanceWithFindings.EbsVolumes.Status)
						resource["has_malware_ebs"] = ds.MalwareProtection.ScanEc2InstanceWithFindings.EbsVolumes.Status == types.DataSourceStatusEnabled
					}
				}
			}

			// Features
			var enabledFeatures []string
			for _, feature := range detectorResp.Features {
				if feature.Status == types.FeatureStatusEnabled {
					enabledFeatures = append(enabledFeatures, string(feature.Name))
				}
			}
			resource["enabled_features"] = enabledFeatures
			resource["enabled_feature_count"] = len(enabledFeatures)

			// Check for specific features
			resource["has_s3_protection"] = containsFeature(enabledFeatures, "S3_DATA_EVENTS")
			resource["has_eks_protection"] = containsFeature(enabledFeatures, "EKS_AUDIT_LOGS")
			resource["has_malware_protection"] = containsFeature(enabledFeatures, "EBS_MALWARE_PROTECTION")
			resource["has_rds_protection"] = containsFeature(enabledFeatures, "RDS_LOGIN_EVENTS")
			resource["has_lambda_protection"] = containsFeature(enabledFeatures, "LAMBDA_NETWORK_LOGS")
			resource["has_runtime_monitoring"] = containsFeature(enabledFeatures, "RUNTIME_MONITORING")

			// Tags
			resource["tags"] = detectorResp.Tags

			// Computed
			resource["is_enabled"] = detectorResp.Status == types.DetectorStatusEnabled
			resource["is_disabled"] = detectorResp.Status == types.DetectorStatusDisabled

			resources = append(resources, resource)
		}

		// If no detectors found, create a resource indicating GuardDuty is not enabled
		if len(listResp.DetectorIds) == 0 {
			resource := make(core.Resource)
			resource["id"] = "not_enabled"
			resource["region"] = region
			resource["is_enabled"] = false
			resource["status"] = "NOT_ENABLED"
			resources = append(resources, resource)
		}
	}

	return resources, nil
}

func containsFeature(features []string, feature string) bool {
	for _, f := range features {
		if f == feature {
			return true
		}
	}
	return false
}

// GuardDutyFinding fetches GuardDuty findings.
type GuardDutyFinding struct {
	clientFactory GuardDutyClientFactory
	accountID     string
	regions       []string
}

// NewGuardDutyFinding creates a new GuardDutyFinding resource.
func NewGuardDutyFinding(clientFactory GuardDutyClientFactory, accountID string, regions []string) *GuardDutyFinding {
	return &GuardDutyFinding{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *GuardDutyFinding) Name() string {
	return "aws_guardduty_finding"
}

// Fetch retrieves GuardDuty findings across configured regions.
func (r *GuardDutyFinding) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		// List detectors
		listResp, err := client.ListDetectors(ctx, &guardduty.ListDetectorsInput{})
		if err != nil {
			continue
		}

		for _, detectorID := range listResp.DetectorIds {
			// List findings with a filter for active findings
			findingsResp, err := client.ListFindings(ctx, &guardduty.ListFindingsInput{
				DetectorId: aws.String(detectorID),
				FindingCriteria: &types.FindingCriteria{
					Criterion: map[string]types.Condition{
						"service.archived": {
							Equals: []string{"false"},
						},
					},
				},
				MaxResults: aws.Int32(50), // Limit to avoid excessive API calls
			})
			if err != nil {
				continue
			}

			if len(findingsResp.FindingIds) == 0 {
				continue
			}

			// Get finding details
			detailsResp, err := client.GetFindings(ctx, &guardduty.GetFindingsInput{
				DetectorId: aws.String(detectorID),
				FindingIds: findingsResp.FindingIds,
			})
			if err != nil {
				continue
			}

			for _, finding := range detailsResp.Findings {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(finding.Id)
				resource["arn"] = aws.ToString(finding.Arn)
				resource["detector_id"] = detectorID
				resource["region"] = aws.ToString(finding.Region)

				resource["type"] = aws.ToString(finding.Type)
				resource["title"] = aws.ToString(finding.Title)
				resource["description"] = aws.ToString(finding.Description)
				resource["severity"] = finding.Severity

				// Categorize severity
				var severity float64
				if finding.Severity != nil {
					severity = *finding.Severity
				}
				resource["is_low_severity"] = severity >= 0 && severity < 4
				resource["is_medium_severity"] = severity >= 4 && severity < 7
				resource["is_high_severity"] = severity >= 7 && severity < 9
				resource["is_critical_severity"] = severity >= 9

				// Timestamps
				if finding.CreatedAt != nil {
					resource["created_at"] = aws.ToString(finding.CreatedAt)
				}
				if finding.UpdatedAt != nil {
					resource["updated_at"] = aws.ToString(finding.UpdatedAt)
				}

				// Service info
				if finding.Service != nil {
					resource["count"] = finding.Service.Count
					resource["is_archived"] = finding.Service.Archived
					resource["detector_id"] = aws.ToString(finding.Service.DetectorId)

					if finding.Service.Action != nil {
						action := finding.Service.Action
						resource["action_type"] = aws.ToString(action.ActionType)
					}
				}

				// Resource info
				if finding.Resource != nil {
					resource["resource_type"] = aws.ToString(finding.Resource.ResourceType)

					if finding.Resource.InstanceDetails != nil {
						resource["instance_id"] = aws.ToString(finding.Resource.InstanceDetails.InstanceId)
						resource["instance_type"] = aws.ToString(finding.Resource.InstanceDetails.InstanceType)
					}

					if finding.Resource.AccessKeyDetails != nil {
						resource["access_key_id"] = aws.ToString(finding.Resource.AccessKeyDetails.AccessKeyId)
						resource["principal_id"] = aws.ToString(finding.Resource.AccessKeyDetails.PrincipalId)
						resource["user_name"] = aws.ToString(finding.Resource.AccessKeyDetails.UserName)
						resource["user_type"] = aws.ToString(finding.Resource.AccessKeyDetails.UserType)
					}

					if len(finding.Resource.S3BucketDetails) > 0 {
						bucket := finding.Resource.S3BucketDetails[0]
						resource["bucket_name"] = aws.ToString(bucket.Name)
						resource["bucket_arn"] = aws.ToString(bucket.Arn)
					}
				}

				// Account ID
				resource["account_id"] = aws.ToString(finding.AccountId)

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
