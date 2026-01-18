// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"

	"github.com/kopexa-grc/kspec/core"
)

// CloudWatchLogsClient is the interface for CloudWatch Logs operations.
type CloudWatchLogsClient interface {
	DescribeLogGroups(ctx context.Context, params *cloudwatchlogs.DescribeLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeLogGroupsOutput, error)
	DescribeSubscriptionFilters(ctx context.Context, params *cloudwatchlogs.DescribeSubscriptionFiltersInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.DescribeSubscriptionFiltersOutput, error)
	ListTagsForResource(ctx context.Context, params *cloudwatchlogs.ListTagsForResourceInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListTagsForResourceOutput, error)
}

// CloudWatchLogsClientFactory creates CloudWatch Logs clients for specific regions.
type CloudWatchLogsClientFactory func(region string) CloudWatchLogsClient

// CloudWatchClient is the interface for CloudWatch operations.
type CloudWatchClient interface {
	DescribeAlarms(ctx context.Context, params *cloudwatch.DescribeAlarmsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.DescribeAlarmsOutput, error)
	ListMetricStreams(ctx context.Context, params *cloudwatch.ListMetricStreamsInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.ListMetricStreamsOutput, error)
	GetMetricStream(ctx context.Context, params *cloudwatch.GetMetricStreamInput, optFns ...func(*cloudwatch.Options)) (*cloudwatch.GetMetricStreamOutput, error)
}

// CloudWatchClientFactory creates CloudWatch clients for specific regions.
type CloudWatchClientFactory func(region string) CloudWatchClient

// CloudWatchLogGroup fetches CloudWatch Log Groups.
type CloudWatchLogGroup struct {
	clientFactory CloudWatchLogsClientFactory
	accountID     string
	regions       []string
}

// NewCloudWatchLogGroup creates a new CloudWatchLogGroup resource.
func NewCloudWatchLogGroup(clientFactory CloudWatchLogsClientFactory, accountID string, regions []string) *CloudWatchLogGroup {
	return &CloudWatchLogGroup{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *CloudWatchLogGroup) Name() string {
	return "aws_cloudwatch_loggroup"
}

// Fetch retrieves all CloudWatch Log Groups across configured regions.
func (r *CloudWatchLogGroup) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := cloudwatchlogs.NewDescribeLogGroupsPaginator(client, &cloudwatchlogs.DescribeLogGroupsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, lg := range page.LogGroups {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(lg.LogGroupName)
				resource["arn"] = aws.ToString(lg.Arn)
				resource["name"] = aws.ToString(lg.LogGroupName)
				resource["region"] = region

				// Retention
				if lg.RetentionInDays != nil {
					resource["retention_in_days"] = *lg.RetentionInDays
					resource["has_retention"] = true
				} else {
					resource["retention_in_days"] = 0
					resource["has_retention"] = false
				}

				// Storage
				resource["stored_bytes"] = lg.StoredBytes

				// KMS encryption
				resource["kms_key_id"] = aws.ToString(lg.KmsKeyId)
				resource["has_encryption"] = lg.KmsKeyId != nil && aws.ToString(lg.KmsKeyId) != ""

				// Data protection
				resource["data_protection_status"] = string(lg.DataProtectionStatus)
				resource["has_data_protection"] = lg.DataProtectionStatus != ""

				// Log group class
				resource["log_group_class"] = string(lg.LogGroupClass)

				// Metric filter count
				resource["metric_filter_count"] = lg.MetricFilterCount

				// Creation time
				if lg.CreationTime != nil {
					resource["creation_time"] = *lg.CreationTime
				}

				// Get subscription filters
				subsResp, err := client.DescribeSubscriptionFilters(ctx, &cloudwatchlogs.DescribeSubscriptionFiltersInput{
					LogGroupName: lg.LogGroupName,
				})
				if err == nil {
					resource["subscription_filter_count"] = len(subsResp.SubscriptionFilters)
					resource["has_subscription_filters"] = len(subsResp.SubscriptionFilters) > 0

					filterDestinations := make([]string, 0, len(subsResp.SubscriptionFilters))
					for _, sf := range subsResp.SubscriptionFilters {
						filterDestinations = append(filterDestinations, aws.ToString(sf.DestinationArn))
					}
					resource["subscription_destinations"] = filterDestinations
				} else {
					resource["subscription_filter_count"] = 0
					resource["has_subscription_filters"] = false
				}

				// Get tags
				tagsResp, err := client.ListTagsForResource(ctx, &cloudwatchlogs.ListTagsForResourceInput{
					ResourceArn: lg.Arn,
				})
				if err == nil {
					resource["tags"] = tagsResp.Tags
				}

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// CloudWatchAlarm fetches CloudWatch Alarms.
type CloudWatchAlarm struct {
	clientFactory CloudWatchClientFactory
	accountID     string
	regions       []string
}

// NewCloudWatchAlarm creates a new CloudWatchAlarm resource.
func NewCloudWatchAlarm(clientFactory CloudWatchClientFactory, accountID string, regions []string) *CloudWatchAlarm {
	return &CloudWatchAlarm{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *CloudWatchAlarm) Name() string {
	return "aws_cloudwatch_alarm"
}

// Fetch retrieves all CloudWatch Alarms across configured regions.
func (r *CloudWatchAlarm) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := cloudwatch.NewDescribeAlarmsPaginator(client, &cloudwatch.DescribeAlarmsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			// Metric alarms
			for _, alarm := range page.MetricAlarms {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(alarm.AlarmName)
				resource["arn"] = aws.ToString(alarm.AlarmArn)
				resource["name"] = aws.ToString(alarm.AlarmName)
				resource["region"] = region
				resource["alarm_type"] = "MetricAlarm"

				resource["state"] = string(alarm.StateValue)
				resource["state_reason"] = aws.ToString(alarm.StateReason)
				resource["description"] = aws.ToString(alarm.AlarmDescription)

				// Metric info
				resource["metric_name"] = aws.ToString(alarm.MetricName)
				resource["namespace"] = aws.ToString(alarm.Namespace)
				resource["statistic"] = string(alarm.Statistic)
				resource["period"] = alarm.Period
				resource["evaluation_periods"] = alarm.EvaluationPeriods
				resource["threshold"] = alarm.Threshold
				resource["comparison_operator"] = string(alarm.ComparisonOperator)
				resource["treat_missing_data"] = aws.ToString(alarm.TreatMissingData)

				// Actions
				resource["alarm_actions"] = alarm.AlarmActions
				resource["ok_actions"] = alarm.OKActions
				resource["insufficient_data_actions"] = alarm.InsufficientDataActions
				resource["actions_enabled"] = alarm.ActionsEnabled != nil && *alarm.ActionsEnabled

				resource["has_alarm_actions"] = len(alarm.AlarmActions) > 0
				resource["has_ok_actions"] = len(alarm.OKActions) > 0

				// Dimensions
				dimensions := make([]map[string]string, 0, len(alarm.Dimensions))
				for _, dim := range alarm.Dimensions {
					dimensions = append(dimensions, map[string]string{
						"name":  aws.ToString(dim.Name),
						"value": aws.ToString(dim.Value),
					})
				}
				resource["dimensions"] = dimensions
				resource["dimension_count"] = len(dimensions)

				// Computed
				resource["is_in_alarm"] = alarm.StateValue == cwtypes.StateValueAlarm
				resource["is_ok"] = alarm.StateValue == cwtypes.StateValueOk
				resource["is_insufficient_data"] = alarm.StateValue == cwtypes.StateValueInsufficientData

				resources = append(resources, resource)
			}

			// Composite alarms
			for _, alarm := range page.CompositeAlarms {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(alarm.AlarmName)
				resource["arn"] = aws.ToString(alarm.AlarmArn)
				resource["name"] = aws.ToString(alarm.AlarmName)
				resource["region"] = region
				resource["alarm_type"] = "CompositeAlarm"

				resource["state"] = string(alarm.StateValue)
				resource["state_reason"] = aws.ToString(alarm.StateReason)
				resource["description"] = aws.ToString(alarm.AlarmDescription)
				resource["alarm_rule"] = aws.ToString(alarm.AlarmRule)

				// Actions
				resource["alarm_actions"] = alarm.AlarmActions
				resource["ok_actions"] = alarm.OKActions
				resource["insufficient_data_actions"] = alarm.InsufficientDataActions
				resource["actions_enabled"] = alarm.ActionsEnabled != nil && *alarm.ActionsEnabled

				resource["has_alarm_actions"] = len(alarm.AlarmActions) > 0
				resource["has_ok_actions"] = len(alarm.OKActions) > 0

				// Computed
				resource["is_in_alarm"] = alarm.StateValue == cwtypes.StateValueAlarm
				resource["is_ok"] = alarm.StateValue == cwtypes.StateValueOk
				resource["is_insufficient_data"] = alarm.StateValue == cwtypes.StateValueInsufficientData
				resource["is_composite"] = true

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// CloudWatchMetricStream fetches CloudWatch Metric Streams.
type CloudWatchMetricStream struct {
	clientFactory CloudWatchClientFactory
	accountID     string
	regions       []string
}

// NewCloudWatchMetricStream creates a new CloudWatchMetricStream resource.
func NewCloudWatchMetricStream(clientFactory CloudWatchClientFactory, accountID string, regions []string) *CloudWatchMetricStream {
	return &CloudWatchMetricStream{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *CloudWatchMetricStream) Name() string {
	return "aws_cloudwatch_metricstream"
}

// Fetch retrieves all CloudWatch Metric Streams across configured regions.
func (r *CloudWatchMetricStream) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		resp, err := client.ListMetricStreams(ctx, &cloudwatch.ListMetricStreamsInput{})
		if err != nil {
			continue
		}

		for _, stream := range resp.Entries {
			// Get full stream details
			detailResp, err := client.GetMetricStream(ctx, &cloudwatch.GetMetricStreamInput{
				Name: stream.Name,
			})
			if err != nil {
				continue
			}

			resource := make(core.Resource)
			resource["id"] = aws.ToString(detailResp.Name)
			resource["arn"] = aws.ToString(detailResp.Arn)
			resource["name"] = aws.ToString(detailResp.Name)
			resource["region"] = region

			resource["state"] = aws.ToString(detailResp.State)
			resource["firehose_arn"] = aws.ToString(detailResp.FirehoseArn)
			resource["role_arn"] = aws.ToString(detailResp.RoleArn)
			resource["output_format"] = string(detailResp.OutputFormat)

			// Include/Exclude filters
			includeFilters := make([]string, 0, len(detailResp.IncludeFilters))
			for _, f := range detailResp.IncludeFilters {
				includeFilters = append(includeFilters, aws.ToString(f.Namespace))
			}
			resource["include_filters"] = includeFilters
			resource["include_filter_count"] = len(includeFilters)

			excludeFilters := make([]string, 0, len(detailResp.ExcludeFilters))
			for _, f := range detailResp.ExcludeFilters {
				excludeFilters = append(excludeFilters, aws.ToString(f.Namespace))
			}
			resource["exclude_filters"] = excludeFilters
			resource["exclude_filter_count"] = len(excludeFilters)

			// Statistics configurations
			resource["statistics_configuration_count"] = len(detailResp.StatisticsConfigurations)
			resource["include_linked_accounts"] = detailResp.IncludeLinkedAccountsMetrics != nil && *detailResp.IncludeLinkedAccountsMetrics

			// Creation/update time
			if detailResp.CreationDate != nil {
				resource["creation_date"] = detailResp.CreationDate.String()
			}
			if detailResp.LastUpdateDate != nil {
				resource["last_update_date"] = detailResp.LastUpdateDate.String()
			}

			// Computed
			resource["is_running"] = aws.ToString(detailResp.State) == "running"
			resource["is_stopped"] = aws.ToString(detailResp.State) == "stopped"

			resources = append(resources, resource)
		}
	}

	return resources, nil
}
