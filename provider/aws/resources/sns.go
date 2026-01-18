// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"

	"github.com/kopexa-grc/kspec/core"
)

// SNSClient is the interface for SNS operations.
type SNSClient interface {
	ListTopics(ctx context.Context, params *sns.ListTopicsInput, optFns ...func(*sns.Options)) (*sns.ListTopicsOutput, error)
	GetTopicAttributes(ctx context.Context, params *sns.GetTopicAttributesInput, optFns ...func(*sns.Options)) (*sns.GetTopicAttributesOutput, error)
	ListSubscriptionsByTopic(ctx context.Context, params *sns.ListSubscriptionsByTopicInput, optFns ...func(*sns.Options)) (*sns.ListSubscriptionsByTopicOutput, error)
	ListTagsForResource(ctx context.Context, params *sns.ListTagsForResourceInput, optFns ...func(*sns.Options)) (*sns.ListTagsForResourceOutput, error)
	ListSubscriptions(ctx context.Context, params *sns.ListSubscriptionsInput, optFns ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error)
	GetSubscriptionAttributes(ctx context.Context, params *sns.GetSubscriptionAttributesInput, optFns ...func(*sns.Options)) (*sns.GetSubscriptionAttributesOutput, error)
}

// SNSClientFactory creates SNS clients for specific regions.
type SNSClientFactory func(region string) SNSClient

// snsTopicsPaginator is an interface for paginating ListTopics results.
type snsTopicsPaginator interface {
	HasMorePages() bool
	NextPage(ctx context.Context, optFns ...func(*sns.Options)) (*sns.ListTopicsOutput, error)
}

// snsSubscriptionsPaginator is an interface for paginating ListSubscriptions results.
type snsSubscriptionsPaginator interface {
	HasMorePages() bool
	NextPage(ctx context.Context, optFns ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error)
}

// SNSTopic fetches SNS topics.
type SNSTopic struct {
	clientFactory    SNSClientFactory
	accountID        string
	regions          []string
	paginatorFactory func(client SNSClient) snsTopicsPaginator
}

// NewSNSTopic creates a new SNSTopic resource.
func NewSNSTopic(clientFactory SNSClientFactory, accountID string, regions []string) *SNSTopic {
	return &SNSTopic{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
		paginatorFactory: func(client SNSClient) snsTopicsPaginator {
			// Type assert to get the concrete client for the paginator
			if c, ok := client.(*sns.Client); ok {
				return sns.NewListTopicsPaginator(c, &sns.ListTopicsInput{})
			}
			return nil
		},
	}
}

// Name returns the resource type name.
func (r *SNSTopic) Name() string {
	return "aws_sns_topic"
}

// Fetch retrieves all SNS topics across configured regions.
func (r *SNSTopic) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := r.paginatorFactory(client)
		if paginator == nil {
			continue
		}

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, topic := range page.Topics {
				topicArn := aws.ToString(topic.TopicArn)

				// Get topic attributes
				attrsResp, err := client.GetTopicAttributes(ctx, &sns.GetTopicAttributesInput{
					TopicArn: topic.TopicArn,
				})
				if err != nil {
					continue
				}

				resource := make(core.Resource)
				resource["id"] = topicArn
				resource["arn"] = topicArn
				resource["region"] = region

				// Extract name from ARN
				arnParts := strings.Split(topicArn, ":")
				if len(arnParts) > 0 {
					resource["name"] = arnParts[len(arnParts)-1]
				}

				attrs := attrsResp.Attributes

				// Owner
				resource["owner"] = attrs["Owner"]

				// Display name
				resource["display_name"] = attrs["DisplayName"]

				// KMS encryption
				resource["kms_master_key_id"] = attrs["KmsMasterKeyId"]
				resource["has_encryption"] = attrs["KmsMasterKeyId"] != ""

				// Subscription counts
				resource["subscriptions_confirmed"] = attrs["SubscriptionsConfirmed"]
				resource["subscriptions_pending"] = attrs["SubscriptionsPending"]
				resource["subscriptions_deleted"] = attrs["SubscriptionsDeleted"]

				// FIFO topic
				resource["fifo_topic"] = attrs["FifoTopic"] == valueTrue
				resource["is_fifo"] = attrs["FifoTopic"] == valueTrue

				// Content-based deduplication
				resource["content_based_deduplication"] = attrs["ContentBasedDeduplication"] == valueTrue

				// Effective delivery policy
				if attrs["EffectiveDeliveryPolicy"] != "" {
					resource["has_delivery_policy"] = true
					resource["delivery_policy"] = attrs["EffectiveDeliveryPolicy"]
				} else {
					resource["has_delivery_policy"] = false
				}

				// Policy
				if attrs["Policy"] != "" {
					resource["policy"] = attrs["Policy"]
					resource["has_policy"] = true

					// Check for public access
					isPublic := checkPublicSNSPolicy(attrs["Policy"])
					resource["is_public"] = isPublic
					resource["allows_public_access"] = isPublic
				} else {
					resource["has_policy"] = false
					resource["is_public"] = false
				}

				// Tracing config
				resource["tracing_config"] = attrs["TracingConfig"]
				resource["has_tracing"] = attrs["TracingConfig"] == "Active"

				// HTTP success feedback
				resource["http_success_feedback_role_arn"] = attrs["HTTPSuccessFeedbackRoleArn"]
				resource["http_failure_feedback_role_arn"] = attrs["HTTPFailureFeedbackRoleArn"]
				resource["has_http_feedback"] = attrs["HTTPSuccessFeedbackRoleArn"] != "" || attrs["HTTPFailureFeedbackRoleArn"] != ""

				// Lambda success feedback
				resource["lambda_success_feedback_role_arn"] = attrs["LambdaSuccessFeedbackRoleArn"]
				resource["lambda_failure_feedback_role_arn"] = attrs["LambdaFailureFeedbackRoleArn"]
				resource["has_lambda_feedback"] = attrs["LambdaSuccessFeedbackRoleArn"] != "" || attrs["LambdaFailureFeedbackRoleArn"] != ""

				// SQS success feedback
				resource["sqs_success_feedback_role_arn"] = attrs["SQSSuccessFeedbackRoleArn"]
				resource["sqs_failure_feedback_role_arn"] = attrs["SQSFailureFeedbackRoleArn"]
				resource["has_sqs_feedback"] = attrs["SQSSuccessFeedbackRoleArn"] != "" || attrs["SQSFailureFeedbackRoleArn"] != ""

				// Application success feedback
				resource["application_success_feedback_role_arn"] = attrs["ApplicationSuccessFeedbackRoleArn"]
				resource["application_failure_feedback_role_arn"] = attrs["ApplicationFailureFeedbackRoleArn"]

				// Firehose success feedback
				resource["firehose_success_feedback_role_arn"] = attrs["FirehoseSuccessFeedbackRoleArn"]
				resource["firehose_failure_feedback_role_arn"] = attrs["FirehoseFailureFeedbackRoleArn"]

				// Signature version
				resource["signature_version"] = attrs["SignatureVersion"]

				// Get subscriptions
				subsResp, err := client.ListSubscriptionsByTopic(ctx, &sns.ListSubscriptionsByTopicInput{
					TopicArn: topic.TopicArn,
				})
				if err == nil {
					protocols := make([]string, 0, len(subsResp.Subscriptions))
					endpoints := make([]string, 0, len(subsResp.Subscriptions))
					for _, sub := range subsResp.Subscriptions {
						protocols = append(protocols, aws.ToString(sub.Protocol))
						endpoints = append(endpoints, aws.ToString(sub.Endpoint))
					}
					resource["subscription_protocols"] = protocols
					resource["subscription_endpoints"] = endpoints
					resource["subscription_count"] = len(subsResp.Subscriptions)
					resource["has_subscriptions"] = len(subsResp.Subscriptions) > 0

					// Check for specific protocol types
					resource["has_email_subscriptions"] = containsProtocol(protocols, "email")
					resource["has_sms_subscriptions"] = containsProtocol(protocols, "sms")
					resource["has_lambda_subscriptions"] = containsProtocol(protocols, "lambda")
					resource["has_sqs_subscriptions"] = containsProtocol(protocols, "sqs")
					resource["has_https_subscriptions"] = containsProtocol(protocols, "https")
					resource["has_http_subscriptions"] = containsProtocol(protocols, "http")
				}

				// Get tags
				tagsResp, err := client.ListTagsForResource(ctx, &sns.ListTagsForResourceInput{
					ResourceArn: topic.TopicArn,
				})
				if err == nil {
					tags := make(map[string]string)
					for _, tag := range tagsResp.Tags {
						tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					}
					resource["tags"] = tags
				}

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// SNSSubscription fetches SNS subscriptions.
type SNSSubscription struct {
	clientFactory    SNSClientFactory
	accountID        string
	regions          []string
	paginatorFactory func(client SNSClient) snsSubscriptionsPaginator
}

// NewSNSSubscription creates a new SNSSubscription resource.
func NewSNSSubscription(clientFactory SNSClientFactory, accountID string, regions []string) *SNSSubscription {
	return &SNSSubscription{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
		paginatorFactory: func(client SNSClient) snsSubscriptionsPaginator {
			// Type assert to get the concrete client for the paginator
			if c, ok := client.(*sns.Client); ok {
				return sns.NewListSubscriptionsPaginator(c, &sns.ListSubscriptionsInput{})
			}
			return nil
		},
	}
}

// Name returns the resource type name.
func (r *SNSSubscription) Name() string {
	return "aws_sns_subscription"
}

// Fetch retrieves all SNS subscriptions across configured regions.
func (r *SNSSubscription) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := r.paginatorFactory(client)
		if paginator == nil {
			continue
		}

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, sub := range page.Subscriptions {
				subArn := aws.ToString(sub.SubscriptionArn)

				// Skip pending confirmations
				if subArn == "PendingConfirmation" {
					continue
				}

				resource := make(core.Resource)
				resource["id"] = subArn
				resource["arn"] = subArn
				resource["topic_arn"] = aws.ToString(sub.TopicArn)
				resource["region"] = region

				resource["protocol"] = aws.ToString(sub.Protocol)
				resource["endpoint"] = aws.ToString(sub.Endpoint)
				resource["owner"] = aws.ToString(sub.Owner)

				// Computed protocol checks
				protocol := aws.ToString(sub.Protocol)
				resource["is_email"] = protocol == "email" || protocol == "email-json"
				resource["is_sms"] = protocol == "sms"
				resource["is_lambda"] = protocol == "lambda"
				resource["is_sqs"] = protocol == "sqs"
				resource["is_http"] = protocol == "http"
				resource["is_https"] = protocol == "https"
				resource["is_firehose"] = protocol == "firehose"
				resource["is_application"] = protocol == "application"

				// Get subscription attributes
				attrsResp, err := client.GetSubscriptionAttributes(ctx, &sns.GetSubscriptionAttributesInput{
					SubscriptionArn: sub.SubscriptionArn,
				})
				if err == nil {
					attrs := attrsResp.Attributes

					resource["confirmation_was_authenticated"] = attrs["ConfirmationWasAuthenticated"] == valueTrue
					resource["pending_confirmation"] = attrs["PendingConfirmation"] == valueTrue
					resource["raw_message_delivery"] = attrs["RawMessageDelivery"] == valueTrue

					// Filter policy
					if attrs["FilterPolicy"] != "" {
						resource["has_filter_policy"] = true
						resource["filter_policy"] = attrs["FilterPolicy"]
					} else {
						resource["has_filter_policy"] = false
					}

					// Filter policy scope
					resource["filter_policy_scope"] = attrs["FilterPolicyScope"]

					// Redrive policy (DLQ)
					if attrs["RedrivePolicy"] != "" {
						resource["has_redrive_policy"] = true
						resource["redrive_policy"] = attrs["RedrivePolicy"]
					} else {
						resource["has_redrive_policy"] = false
					}

					// Delivery policy
					if attrs["DeliveryPolicy"] != "" {
						resource["has_delivery_policy"] = true
						resource["delivery_policy"] = attrs["DeliveryPolicy"]
					} else {
						resource["has_delivery_policy"] = false
					}

					// Subscription role ARN (for Firehose)
					resource["subscription_role_arn"] = attrs["SubscriptionRoleArn"]
				}

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

func containsProtocol(protocols []string, protocol string) bool {
	for _, p := range protocols {
		if p == protocol {
			return true
		}
	}
	return false
}

func checkPublicSNSPolicy(policyJSON string) bool {
	var policy struct {
		Statement []struct {
			Effect    string      `json:"Effect"`
			Principal interface{} `json:"Principal"`
		} `json:"Statement"`
	}

	if err := json.Unmarshal([]byte(policyJSON), &policy); err != nil {
		return false
	}

	for _, stmt := range policy.Statement {
		if stmt.Effect != "Allow" {
			continue
		}

		switch p := stmt.Principal.(type) {
		case string:
			if p == "*" {
				return true
			}
		case map[string]interface{}:
			if awsPrincipal, ok := p["AWS"]; ok {
				switch v := awsPrincipal.(type) {
				case string:
					if v == "*" {
						return true
					}
				case []interface{}:
					for _, item := range v {
						if s, ok := item.(string); ok && s == "*" {
							return true
						}
					}
				}
			}
		}
	}

	return false
}
