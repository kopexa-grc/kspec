// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/configservice/types"

	"github.com/kopexa-grc/kspec/core"
)

// ConfigClient is the interface for AWS Config operations.
type ConfigClient interface {
	DescribeConfigurationRecorders(ctx context.Context, params *configservice.DescribeConfigurationRecordersInput, optFns ...func(*configservice.Options)) (*configservice.DescribeConfigurationRecordersOutput, error)
	DescribeConfigurationRecorderStatus(ctx context.Context, params *configservice.DescribeConfigurationRecorderStatusInput, optFns ...func(*configservice.Options)) (*configservice.DescribeConfigurationRecorderStatusOutput, error)
	DescribeConfigRules(ctx context.Context, params *configservice.DescribeConfigRulesInput, optFns ...func(*configservice.Options)) (*configservice.DescribeConfigRulesOutput, error)
	DescribeComplianceByConfigRule(ctx context.Context, params *configservice.DescribeComplianceByConfigRuleInput, optFns ...func(*configservice.Options)) (*configservice.DescribeComplianceByConfigRuleOutput, error)
	DescribeDeliveryChannels(ctx context.Context, params *configservice.DescribeDeliveryChannelsInput, optFns ...func(*configservice.Options)) (*configservice.DescribeDeliveryChannelsOutput, error)
	DescribeDeliveryChannelStatus(ctx context.Context, params *configservice.DescribeDeliveryChannelStatusInput, optFns ...func(*configservice.Options)) (*configservice.DescribeDeliveryChannelStatusOutput, error)
	DescribeConformancePacks(ctx context.Context, params *configservice.DescribeConformancePacksInput, optFns ...func(*configservice.Options)) (*configservice.DescribeConformancePacksOutput, error)
	DescribeConformancePackCompliance(ctx context.Context, params *configservice.DescribeConformancePackComplianceInput, optFns ...func(*configservice.Options)) (*configservice.DescribeConformancePackComplianceOutput, error)
	DescribeConformancePackStatus(ctx context.Context, params *configservice.DescribeConformancePackStatusInput, optFns ...func(*configservice.Options)) (*configservice.DescribeConformancePackStatusOutput, error)
}

// ConfigClientFactory creates Config clients for specific regions.
type ConfigClientFactory func(region string) ConfigClient

// ConfigRecorder fetches AWS Config recorders.
type ConfigRecorder struct {
	clientFactory ConfigClientFactory
	accountID     string
	regions       []string
}

// NewConfigRecorder creates a new ConfigRecorder resource.
func NewConfigRecorder(clientFactory ConfigClientFactory, accountID string, regions []string) *ConfigRecorder {
	return &ConfigRecorder{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *ConfigRecorder) Name() string {
	return "aws_config_recorder"
}

// Fetch retrieves all Config recorders across configured regions.
func (r *ConfigRecorder) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		resp, err := client.DescribeConfigurationRecorders(ctx, &configservice.DescribeConfigurationRecordersInput{})
		if err != nil {
			continue
		}

		// Get recorder status
		statusResp, err := client.DescribeConfigurationRecorderStatus(ctx, &configservice.DescribeConfigurationRecorderStatusInput{})
		statusMap := make(map[string]types.ConfigurationRecorderStatus)
		if err == nil {
			for _, status := range statusResp.ConfigurationRecordersStatus {
				statusMap[aws.ToString(status.Name)] = status
			}
		}

		for _, recorder := range resp.ConfigurationRecorders {
			resource := make(core.Resource)
			recorderName := aws.ToString(recorder.Name)
			resource["id"] = recorderName
			resource["name"] = recorderName
			resource["region"] = region

			resource["role_arn"] = aws.ToString(recorder.RoleARN)

			// Recording group
			if recorder.RecordingGroup != nil {
				rg := recorder.RecordingGroup
				resource["all_supported"] = rg.AllSupported
				resource["include_global_resources"] = rg.IncludeGlobalResourceTypes

				resourceTypes := make([]string, 0, len(rg.ResourceTypes))
				for _, rt := range rg.ResourceTypes {
					resourceTypes = append(resourceTypes, string(rt))
				}
				resource["resource_types"] = resourceTypes
				resource["resource_type_count"] = len(resourceTypes)

				// Recording strategy
				if rg.RecordingStrategy != nil {
					resource["recording_strategy"] = string(rg.RecordingStrategy.UseOnly)
				}

				// Exclusion by resource types
				if rg.ExclusionByResourceTypes != nil {
					excludedTypes := make([]string, 0, len(rg.ExclusionByResourceTypes.ResourceTypes))
					for _, rt := range rg.ExclusionByResourceTypes.ResourceTypes {
						excludedTypes = append(excludedTypes, string(rt))
					}
					resource["excluded_resource_types"] = excludedTypes
				}
			}

			// Recording mode
			if recorder.RecordingMode != nil {
				resource["recording_frequency"] = string(recorder.RecordingMode.RecordingFrequency)
			}

			// Status
			if status, ok := statusMap[recorderName]; ok {
				resource["is_recording"] = status.Recording
				resource["last_status"] = string(status.LastStatus)
				if status.LastStartTime != nil {
					resource["last_start_time"] = status.LastStartTime.String()
				}
				if status.LastStopTime != nil {
					resource["last_stop_time"] = status.LastStopTime.String()
				}
				if status.LastStatusChangeTime != nil {
					resource["last_status_change_time"] = status.LastStatusChangeTime.String()
				}
				if status.LastErrorCode != nil {
					resource["last_error_code"] = aws.ToString(status.LastErrorCode)
					resource["has_error"] = true
				} else {
					resource["has_error"] = false
				}
				if status.LastErrorMessage != nil {
					resource["last_error_message"] = aws.ToString(status.LastErrorMessage)
				}
			} else {
				resource["is_recording"] = false
			}

			// Computed
			allSupported, _ := resource["all_supported"].(bool)             //nolint:errcheck // zero value ok
			includeGlobal, _ := resource["include_global_resources"].(bool) //nolint:errcheck // zero value ok
			isRecording, _ := resource["is_recording"].(bool)               //nolint:errcheck // zero value ok
			resource["is_enabled"] = isRecording
			resource["records_all_resources"] = allSupported && includeGlobal
			resource["is_compliant_config"] = isRecording && allSupported

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// ConfigRule fetches AWS Config rules.
type ConfigRule struct {
	clientFactory ConfigClientFactory
	accountID     string
	regions       []string
}

// NewConfigRule creates a new ConfigRule resource.
func NewConfigRule(clientFactory ConfigClientFactory, accountID string, regions []string) *ConfigRule {
	return &ConfigRule{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *ConfigRule) Name() string {
	return "aws_config_rule"
}

// Fetch retrieves all Config rules across configured regions.
func (r *ConfigRule) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := configservice.NewDescribeConfigRulesPaginator(client, &configservice.DescribeConfigRulesInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, rule := range page.ConfigRules {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(rule.ConfigRuleName)
				resource["arn"] = aws.ToString(rule.ConfigRuleArn)
				resource["name"] = aws.ToString(rule.ConfigRuleName)
				resource["region"] = region

				resource["description"] = aws.ToString(rule.Description)
				resource["state"] = string(rule.ConfigRuleState)
				resource["created_by"] = aws.ToString(rule.CreatedBy)

				// Source
				if rule.Source != nil {
					resource["source_owner"] = string(rule.Source.Owner)
					resource["source_identifier"] = aws.ToString(rule.Source.SourceIdentifier)

					resource["is_aws_managed"] = rule.Source.Owner == types.OwnerAws
					resource["is_custom_rule"] = rule.Source.Owner == types.OwnerCustomLambda || rule.Source.Owner == types.OwnerCustomPolicy

					// Source details
					triggerTypes := make([]string, 0, len(rule.Source.SourceDetails))
					for _, detail := range rule.Source.SourceDetails {
						triggerTypes = append(triggerTypes, string(detail.MessageType))
					}
					resource["trigger_types"] = triggerTypes

					// Custom policy details
					if rule.Source.CustomPolicyDetails != nil {
						resource["policy_runtime"] = aws.ToString(rule.Source.CustomPolicyDetails.PolicyRuntime)
						resource["enable_debug_log"] = rule.Source.CustomPolicyDetails.EnableDebugLogDelivery
					}
				}

				// Scope
				if rule.Scope != nil {
					resource["compliance_resource_types"] = rule.Scope.ComplianceResourceTypes
					resource["tag_key"] = aws.ToString(rule.Scope.TagKey)
					resource["tag_value"] = aws.ToString(rule.Scope.TagValue)
					resource["compliance_resource_id"] = aws.ToString(rule.Scope.ComplianceResourceId)
				}

				// Input parameters
				resource["input_parameters"] = aws.ToString(rule.InputParameters)
				resource["has_input_parameters"] = rule.InputParameters != nil && aws.ToString(rule.InputParameters) != ""

				// Maximum execution frequency
				resource["maximum_execution_frequency"] = string(rule.MaximumExecutionFrequency)

				// Evaluation mode
				evaluationModes := make([]string, 0, len(rule.EvaluationModes))
				for _, mode := range rule.EvaluationModes {
					evaluationModes = append(evaluationModes, string(mode.Mode))
				}
				resource["evaluation_modes"] = evaluationModes

				// Get compliance status
				complianceResp, err := client.DescribeComplianceByConfigRule(ctx, &configservice.DescribeComplianceByConfigRuleInput{
					ConfigRuleNames: []string{aws.ToString(rule.ConfigRuleName)},
				})
				if err == nil && len(complianceResp.ComplianceByConfigRules) > 0 {
					compliance := complianceResp.ComplianceByConfigRules[0]
					resource["compliance_type"] = string(compliance.Compliance.ComplianceType)
					resource["is_compliant"] = compliance.Compliance.ComplianceType == types.ComplianceTypeCompliant
					resource["is_non_compliant"] = compliance.Compliance.ComplianceType == types.ComplianceTypeNonCompliant
					resource["is_not_applicable"] = compliance.Compliance.ComplianceType == types.ComplianceTypeNotApplicable
					resource["is_insufficient_data"] = compliance.Compliance.ComplianceType == types.ComplianceTypeInsufficientData
				}

				// Computed
				resource["is_active"] = rule.ConfigRuleState == types.ConfigRuleStateActive
				resource["is_deleting"] = rule.ConfigRuleState == types.ConfigRuleStateDeleting
				resource["is_deleting_results"] = rule.ConfigRuleState == types.ConfigRuleStateDeletingResults
				resource["is_evaluating"] = rule.ConfigRuleState == types.ConfigRuleStateEvaluating

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// ConfigDeliveryChannel fetches AWS Config delivery channels.
type ConfigDeliveryChannel struct {
	clientFactory ConfigClientFactory
	accountID     string
	regions       []string
}

// NewConfigDeliveryChannel creates a new ConfigDeliveryChannel resource.
func NewConfigDeliveryChannel(clientFactory ConfigClientFactory, accountID string, regions []string) *ConfigDeliveryChannel {
	return &ConfigDeliveryChannel{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *ConfigDeliveryChannel) Name() string {
	return "aws_config_deliverychannel"
}

// Fetch retrieves all Config delivery channels across configured regions.
func (r *ConfigDeliveryChannel) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		resp, err := client.DescribeDeliveryChannels(ctx, &configservice.DescribeDeliveryChannelsInput{})
		if err != nil {
			continue
		}

		// Get delivery channel status
		statusResp, err := client.DescribeDeliveryChannelStatus(ctx, &configservice.DescribeDeliveryChannelStatusInput{})
		statusMap := make(map[string]types.DeliveryChannelStatus)
		if err == nil {
			for _, status := range statusResp.DeliveryChannelsStatus {
				statusMap[aws.ToString(status.Name)] = status
			}
		}

		for _, channel := range resp.DeliveryChannels {
			resource := make(core.Resource)
			channelName := aws.ToString(channel.Name)
			resource["id"] = channelName
			resource["name"] = channelName
			resource["region"] = region

			resource["s3_bucket_name"] = aws.ToString(channel.S3BucketName)
			resource["s3_key_prefix"] = aws.ToString(channel.S3KeyPrefix)
			resource["s3_kms_key_arn"] = aws.ToString(channel.S3KmsKeyArn)
			resource["sns_topic_arn"] = aws.ToString(channel.SnsTopicARN)

			resource["has_s3_bucket"] = channel.S3BucketName != nil && aws.ToString(channel.S3BucketName) != ""
			resource["has_sns_topic"] = channel.SnsTopicARN != nil && aws.ToString(channel.SnsTopicARN) != ""
			resource["has_s3_encryption"] = channel.S3KmsKeyArn != nil && aws.ToString(channel.S3KmsKeyArn) != ""

			// Config snapshot delivery properties
			if channel.ConfigSnapshotDeliveryProperties != nil {
				resource["delivery_frequency"] = string(channel.ConfigSnapshotDeliveryProperties.DeliveryFrequency)
			}

			// Status
			if status, ok := statusMap[channelName]; ok {
				// Config history delivery
				if status.ConfigHistoryDeliveryInfo != nil {
					info := status.ConfigHistoryDeliveryInfo
					resource["config_history_status"] = string(info.LastStatus)
					if info.LastSuccessfulTime != nil {
						resource["config_history_last_success"] = info.LastSuccessfulTime.String()
					}
					if info.LastErrorCode != nil {
						resource["config_history_error_code"] = aws.ToString(info.LastErrorCode)
					}
				}

				// Config snapshot delivery
				if status.ConfigSnapshotDeliveryInfo != nil {
					info := status.ConfigSnapshotDeliveryInfo
					resource["config_snapshot_status"] = string(info.LastStatus)
					if info.LastSuccessfulTime != nil {
						resource["config_snapshot_last_success"] = info.LastSuccessfulTime.String()
					}
					if info.LastErrorCode != nil {
						resource["config_snapshot_error_code"] = aws.ToString(info.LastErrorCode)
					}
				}

				// Config stream delivery
				if status.ConfigStreamDeliveryInfo != nil {
					info := status.ConfigStreamDeliveryInfo
					resource["config_stream_status"] = string(info.LastStatus)
					if info.LastStatusChangeTime != nil {
						resource["config_stream_last_change"] = info.LastStatusChangeTime.String()
					}
					if info.LastErrorCode != nil {
						resource["config_stream_error_code"] = aws.ToString(info.LastErrorCode)
					}
				}
			}

			// Computed
			hasS3, _ := resource["has_s3_bucket"].(bool)  //nolint:errcheck // zero value ok
			hasSNS, _ := resource["has_sns_topic"].(bool) //nolint:errcheck // zero value ok
			resource["is_configured"] = hasS3
			resource["has_notifications"] = hasSNS

			resources = append(resources, resource)
		}
	}

	return resources, nil
}

// ConfigConformancePack fetches AWS Config conformance packs.
type ConfigConformancePack struct {
	clientFactory ConfigClientFactory
	accountID     string
	regions       []string
}

// NewConfigConformancePack creates a new ConfigConformancePack resource.
func NewConfigConformancePack(clientFactory ConfigClientFactory, accountID string, regions []string) *ConfigConformancePack {
	return &ConfigConformancePack{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *ConfigConformancePack) Name() string {
	return "aws_config_conformancepack"
}

// Fetch retrieves all Config conformance packs across configured regions.
func (r *ConfigConformancePack) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := configservice.NewDescribeConformancePacksPaginator(client, &configservice.DescribeConformancePacksInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, pack := range page.ConformancePackDetails {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(pack.ConformancePackName)
				resource["arn"] = aws.ToString(pack.ConformancePackArn)
				resource["name"] = aws.ToString(pack.ConformancePackName)
				resource["region"] = region

				resource["delivery_s3_bucket"] = aws.ToString(pack.DeliveryS3Bucket)
				resource["delivery_s3_key_prefix"] = aws.ToString(pack.DeliveryS3KeyPrefix)
				resource["created_by"] = aws.ToString(pack.CreatedBy)

				resource["has_delivery_bucket"] = pack.DeliveryS3Bucket != nil && aws.ToString(pack.DeliveryS3Bucket) != ""

				// Template S3 URI
				resource["template_s3_uri"] = aws.ToString(pack.TemplateSSMDocumentDetails.DocumentName)

				// Input parameters
				paramNames := make([]string, 0, len(pack.ConformancePackInputParameters))
				for _, param := range pack.ConformancePackInputParameters {
					paramNames = append(paramNames, aws.ToString(param.ParameterName))
				}
				resource["input_parameter_names"] = paramNames
				resource["input_parameter_count"] = len(paramNames)

				// Last update
				if pack.LastUpdateRequestedTime != nil {
					resource["last_update_time"] = pack.LastUpdateRequestedTime.String()
				}

				// Get compliance status
				complianceResp, err := client.DescribeConformancePackCompliance(ctx, &configservice.DescribeConformancePackComplianceInput{
					ConformancePackName: pack.ConformancePackName,
				})
				if err == nil {
					compliantCount := 0
					nonCompliantCount := 0
					for _, rule := range complianceResp.ConformancePackRuleComplianceList {
						switch rule.ComplianceType { //nolint:exhaustive // only care about compliant/non-compliant
						case types.ConformancePackComplianceTypeCompliant:
							compliantCount++
						case types.ConformancePackComplianceTypeNonCompliant:
							nonCompliantCount++
						default:
							// Other compliance types (insufficient data, etc.)
						}
					}
					resource["compliant_rule_count"] = compliantCount
					resource["non_compliant_rule_count"] = nonCompliantCount
					resource["total_rule_count"] = len(complianceResp.ConformancePackRuleComplianceList)
					resource["is_fully_compliant"] = nonCompliantCount == 0 && compliantCount > 0
				}

				// Get status
				statusResp, err := client.DescribeConformancePackStatus(ctx, &configservice.DescribeConformancePackStatusInput{
					ConformancePackNames: []string{aws.ToString(pack.ConformancePackName)},
				})
				if err == nil && len(statusResp.ConformancePackStatusDetails) > 0 {
					status := statusResp.ConformancePackStatusDetails[0]
					resource["status"] = string(status.ConformancePackState)
					resource["status_reason"] = aws.ToString(status.ConformancePackStatusReason)
					resource["is_create_complete"] = status.ConformancePackState == types.ConformancePackStateCreateComplete
					resource["is_create_failed"] = status.ConformancePackState == types.ConformancePackStateCreateFailed
				}

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
