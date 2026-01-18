// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"

	"github.com/kopexa-grc/kspec/core"
)

// CloudTrailClient is the interface for CloudTrail operations.
type CloudTrailClient interface {
	DescribeTrails(ctx context.Context, params *cloudtrail.DescribeTrailsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.DescribeTrailsOutput, error)
	GetTrailStatus(ctx context.Context, params *cloudtrail.GetTrailStatusInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetTrailStatusOutput, error)
	GetEventSelectors(ctx context.Context, params *cloudtrail.GetEventSelectorsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetEventSelectorsOutput, error)
	GetInsightSelectors(ctx context.Context, params *cloudtrail.GetInsightSelectorsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.GetInsightSelectorsOutput, error)
	ListTags(ctx context.Context, params *cloudtrail.ListTagsInput, optFns ...func(*cloudtrail.Options)) (*cloudtrail.ListTagsOutput, error)
}

// CloudTrailClientFactory creates CloudTrail clients for specific regions.
type CloudTrailClientFactory func(region string) CloudTrailClient

// CloudTrail fetches CloudTrail trails.
type CloudTrail struct {
	clientFactory CloudTrailClientFactory
	accountID     string
	regions       []string
}

// NewCloudTrail creates a new CloudTrail resource.
func NewCloudTrail(clientFactory CloudTrailClientFactory, accountID string, regions []string) *CloudTrail {
	return &CloudTrail{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *CloudTrail) Name() string {
	return "aws_cloudtrail"
}

// Fetch retrieves all CloudTrail trails.
func (r *CloudTrail) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	// CloudTrail trails are regional but can be multi-region
	// We'll query from each region to catch all trails
	seenTrails := make(map[string]bool)

	for _, region := range r.regions {
		client := r.clientFactory(region)

		resp, err := client.DescribeTrails(ctx, &cloudtrail.DescribeTrailsInput{
			IncludeShadowTrails: aws.Bool(false), // Don't include shadow trails
		})
		if err != nil {
			continue
		}

		for _, trail := range resp.TrailList {
			trailARN := aws.ToString(trail.TrailARN)

			// Skip if we've already processed this trail
			if seenTrails[trailARN] {
				continue
			}
			seenTrails[trailARN] = true

			resource := make(core.Resource)
			resource["id"] = aws.ToString(trail.Name)
			resource["name"] = aws.ToString(trail.Name)
			resource["arn"] = trailARN
			resource["home_region"] = aws.ToString(trail.HomeRegion)

			// S3 configuration
			resource["s3_bucket_name"] = aws.ToString(trail.S3BucketName)
			resource["s3_key_prefix"] = aws.ToString(trail.S3KeyPrefix)

			// CloudWatch Logs
			resource["cloudwatch_logs_log_group_arn"] = aws.ToString(trail.CloudWatchLogsLogGroupArn)
			resource["cloudwatch_logs_role_arn"] = aws.ToString(trail.CloudWatchLogsRoleArn)
			resource["has_cloudwatch_logs"] = trail.CloudWatchLogsLogGroupArn != nil && aws.ToString(trail.CloudWatchLogsLogGroupArn) != ""

			// SNS topic
			resource["sns_topic_arn"] = aws.ToString(trail.SnsTopicARN)
			resource["has_sns_topic"] = trail.SnsTopicARN != nil && aws.ToString(trail.SnsTopicARN) != ""

			// KMS encryption
			resource["kms_key_id"] = aws.ToString(trail.KmsKeyId)
			resource["has_kms_encryption"] = trail.KmsKeyId != nil && aws.ToString(trail.KmsKeyId) != ""

			// Multi-region
			resource["is_multi_region_trail"] = trail.IsMultiRegionTrail != nil && *trail.IsMultiRegionTrail

			// Organization trail
			resource["is_organization_trail"] = trail.IsOrganizationTrail != nil && *trail.IsOrganizationTrail

			// Include global service events (IAM, STS, etc.)
			resource["include_global_service_events"] = trail.IncludeGlobalServiceEvents != nil && *trail.IncludeGlobalServiceEvents

			// Log file validation
			resource["log_file_validation_enabled"] = trail.LogFileValidationEnabled != nil && *trail.LogFileValidationEnabled

			// Get trail status
			statusResp, err := client.GetTrailStatus(ctx, &cloudtrail.GetTrailStatusInput{
				Name: trail.Name,
			})
			if err == nil {
				resource["is_logging"] = statusResp.IsLogging != nil && *statusResp.IsLogging

				if statusResp.LatestDeliveryTime != nil {
					resource["latest_delivery_time"] = statusResp.LatestDeliveryTime.String()
				}
				if statusResp.LatestNotificationTime != nil {
					resource["latest_notification_time"] = statusResp.LatestNotificationTime.String()
				}
				if statusResp.StartLoggingTime != nil {
					resource["start_logging_time"] = statusResp.StartLoggingTime.String()
				}
				if statusResp.StopLoggingTime != nil {
					resource["stop_logging_time"] = statusResp.StopLoggingTime.String()
				}

				// Error info
				if statusResp.LatestDeliveryError != nil {
					resource["latest_delivery_error"] = aws.ToString(statusResp.LatestDeliveryError)
					resource["has_delivery_error"] = true
				} else {
					resource["has_delivery_error"] = false
				}

				if statusResp.LatestNotificationError != nil {
					resource["latest_notification_error"] = aws.ToString(statusResp.LatestNotificationError)
				}
			} else {
				resource["is_logging"] = false
			}

			// Get event selectors
			selectorsResp, err := client.GetEventSelectors(ctx, &cloudtrail.GetEventSelectorsInput{
				TrailName: trail.Name,
			})
			if err == nil {
				// Basic event selectors
				if len(selectorsResp.EventSelectors) > 0 {
					selectors := make([]map[string]any, 0, len(selectorsResp.EventSelectors))
					var logsManagementEvents, logsDataEvents bool
					var readWriteType string

					for _, es := range selectorsResp.EventSelectors {
						selector := map[string]any{
							"read_write_type":           string(es.ReadWriteType),
							"include_management_events": es.IncludeManagementEvents != nil && *es.IncludeManagementEvents,
						}

						if es.IncludeManagementEvents != nil && *es.IncludeManagementEvents {
							logsManagementEvents = true
						}

						readWriteType = string(es.ReadWriteType)

						// Data resources
						dataResources := make([]map[string]any, 0, len(es.DataResources))
						for _, dr := range es.DataResources {
							dataResource := map[string]any{
								"type":   aws.ToString(dr.Type),
								"values": dr.Values,
							}
							dataResources = append(dataResources, dataResource)
							if len(dr.Values) > 0 {
								logsDataEvents = true
							}
						}
						selector["data_resources"] = dataResources

						selectors = append(selectors, selector)
					}

					resource["event_selectors"] = selectors
					resource["logs_management_events"] = logsManagementEvents
					resource["logs_data_events"] = logsDataEvents
					resource["read_write_type"] = readWriteType
					resource["logs_all_events"] = readWriteType == readWriteTypeAll
					resource["logs_read_events"] = readWriteType == "ReadOnly" || readWriteType == readWriteTypeAll
					resource["logs_write_events"] = readWriteType == "WriteOnly" || readWriteType == readWriteTypeAll
				}

				// Advanced event selectors
				if len(selectorsResp.AdvancedEventSelectors) > 0 {
					resource["has_advanced_event_selectors"] = true
					resource["advanced_event_selector_count"] = len(selectorsResp.AdvancedEventSelectors)
				} else {
					resource["has_advanced_event_selectors"] = false
				}
			}

			// Get insight selectors
			insightsResp, err := client.GetInsightSelectors(ctx, &cloudtrail.GetInsightSelectorsInput{
				TrailName: trail.Name,
			})
			if err == nil && len(insightsResp.InsightSelectors) > 0 {
				resource["has_insights"] = true
				insightTypes := make([]string, 0, len(insightsResp.InsightSelectors))
				for _, is := range insightsResp.InsightSelectors {
					insightTypes = append(insightTypes, string(is.InsightType))
				}
				resource["insight_types"] = insightTypes
			} else {
				resource["has_insights"] = false
			}

			// Get tags
			tagsResp, err := client.ListTags(ctx, &cloudtrail.ListTagsInput{
				ResourceIdList: []string{trailARN},
			})
			if err == nil && len(tagsResp.ResourceTagList) > 0 {
				tags := make(map[string]string)
				for _, rt := range tagsResp.ResourceTagList {
					for _, tag := range rt.TagsList {
						tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					}
				}
				resource["tags"] = tags
			}

			// Computed: Security best practices checks
			isLogging, _ := resource["is_logging"].(bool)                         //nolint:errcheck // zero value ok
			isMultiRegion, _ := resource["is_multi_region_trail"].(bool)          //nolint:errcheck // zero value ok
			hasLogValidation, _ := resource["log_file_validation_enabled"].(bool) //nolint:errcheck // zero value ok
			hasKMS, _ := resource["has_kms_encryption"].(bool)                    //nolint:errcheck // zero value ok
			hasCloudWatch, _ := resource["has_cloudwatch_logs"].(bool)            //nolint:errcheck // zero value ok
			logsManagement, _ := resource["logs_management_events"].(bool)        //nolint:errcheck // zero value ok

			resource["meets_basic_security"] = isLogging && isMultiRegion && hasLogValidation
			resource["meets_enhanced_security"] = isLogging && isMultiRegion && hasLogValidation && hasKMS && hasCloudWatch && logsManagement

			resources = append(resources, resource)
		}
	}

	return resources, nil
}
