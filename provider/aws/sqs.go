package aws

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"

	"github.com/kopexa-grc/kspec/core"
)

// SQSQueueResource fetches SQS queues.
type SQSQueueResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *SQSQueueResource) Name() string {
	return "aws_sqs_queue"
}

// Fetch retrieves all SQS queues across configured regions.
func (r *SQSQueueResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := sqs.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := sqs.NewListQueuesPaginator(client, &sqs.ListQueuesInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, queueURL := range page.QueueUrls {
				// Get queue attributes
				attrsResp, err := client.GetQueueAttributes(ctx, &sqs.GetQueueAttributesInput{
					QueueUrl:       aws.String(queueURL),
					AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
				})
				if err != nil {
					continue
				}

				attrs := attrsResp.Attributes

				resource := make(core.Resource)
				resource["id"] = queueURL
				resource["url"] = queueURL
				resource["arn"] = attrs["QueueArn"]
				resource["region"] = region

				// Extract name from URL
				urlParts := strings.Split(queueURL, "/")
				if len(urlParts) > 0 {
					resource["name"] = urlParts[len(urlParts)-1]
				}

				// FIFO queue
				isFifo := strings.HasSuffix(queueURL, ".fifo")
				resource["fifo_queue"] = isFifo
				resource["is_fifo"] = isFifo

				// Message attributes
				if v, ok := attrs["DelaySeconds"]; ok {
					resource["delay_seconds"], _ = strconv.Atoi(v) //nolint:errcheck // zero value ok
				}
				if v, ok := attrs["MaximumMessageSize"]; ok {
					resource["maximum_message_size"], _ = strconv.Atoi(v) //nolint:errcheck // zero value ok
				}
				if v, ok := attrs["MessageRetentionPeriod"]; ok {
					resource["message_retention_period"], _ = strconv.Atoi(v) //nolint:errcheck // zero value ok
				}
				if v, ok := attrs["ReceiveMessageWaitTimeSeconds"]; ok {
					resource["receive_wait_time_seconds"], _ = strconv.Atoi(v) //nolint:errcheck // zero value ok
				}
				if v, ok := attrs["VisibilityTimeout"]; ok {
					resource["visibility_timeout"], _ = strconv.Atoi(v) //nolint:errcheck // zero value ok
				}

				// Queue stats
				if v, ok := attrs["ApproximateNumberOfMessages"]; ok {
					resource["approximate_message_count"], _ = strconv.Atoi(v) //nolint:errcheck // zero value ok
				}
				if v, ok := attrs["ApproximateNumberOfMessagesNotVisible"]; ok {
					resource["approximate_messages_not_visible"], _ = strconv.Atoi(v) //nolint:errcheck // zero value ok
				}
				if v, ok := attrs["ApproximateNumberOfMessagesDelayed"]; ok {
					resource["approximate_messages_delayed"], _ = strconv.Atoi(v) //nolint:errcheck // zero value ok
				}

				// KMS encryption
				resource["kms_master_key_id"] = attrs["KmsMasterKeyId"]
				resource["has_encryption"] = attrs["KmsMasterKeyId"] != ""
				resource["has_cmk_encryption"] = attrs["KmsMasterKeyId"] != "" && !strings.HasPrefix(attrs["KmsMasterKeyId"], "alias/aws/")

				// SSE type
				resource["sqs_managed_sse_enabled"] = attrs["SqsManagedSseEnabled"] == valueTrue
				resource["has_sse"] = attrs["SqsManagedSseEnabled"] == valueTrue || attrs["KmsMasterKeyId"] != ""

				if v, ok := attrs["KmsDataKeyReusePeriodSeconds"]; ok {
					resource["kms_data_key_reuse_period"], _ = strconv.Atoi(v) //nolint:errcheck // zero value ok
				}

				// FIFO specific
				if isFifo {
					resource["content_based_deduplication"] = attrs["ContentBasedDeduplication"] == valueTrue
					resource["deduplication_scope"] = attrs["DeduplicationScope"]
					resource["fifo_throughput_limit"] = attrs["FifoThroughputLimit"]
				}

				// Dead letter queue (Redrive policy)
				if attrs["RedrivePolicy"] != "" {
					resource["has_dlq"] = true
					resource["redrive_policy"] = attrs["RedrivePolicy"]

					var redrivePolicy struct {
						DeadLetterTargetArn string `json:"deadLetterTargetArn"`
						MaxReceiveCount     int    `json:"maxReceiveCount"`
					}
					if err := json.Unmarshal([]byte(attrs["RedrivePolicy"]), &redrivePolicy); err == nil {
						resource["dead_letter_target_arn"] = redrivePolicy.DeadLetterTargetArn
						resource["max_receive_count"] = redrivePolicy.MaxReceiveCount
					}
				} else {
					resource["has_dlq"] = false
				}

				// Redrive allow policy
				if attrs["RedriveAllowPolicy"] != "" {
					resource["has_redrive_allow_policy"] = true
					resource["redrive_allow_policy"] = attrs["RedriveAllowPolicy"]
				} else {
					resource["has_redrive_allow_policy"] = false
				}

				// Policy
				if attrs["Policy"] != "" {
					resource["has_policy"] = true
					resource["policy"] = attrs["Policy"]

					// Check for public access
					isPublic := checkPublicSQSPolicy(attrs["Policy"])
					resource["is_public"] = isPublic
					resource["allows_public_access"] = isPublic
				} else {
					resource["has_policy"] = false
					resource["is_public"] = false
				}

				// Created/modified timestamps
				resource["created_timestamp"] = attrs["CreatedTimestamp"]
				resource["last_modified_timestamp"] = attrs["LastModifiedTimestamp"]

				// Get tags
				tagsResp, err := client.ListQueueTags(ctx, &sqs.ListQueueTagsInput{
					QueueUrl: aws.String(queueURL),
				})
				if err == nil {
					resource["tags"] = tagsResp.Tags
				}

				// Computed
				hasEncryption, _ := resource["has_sse"].(bool) //nolint:errcheck // zero value ok
				hasDLQ, _ := resource["has_dlq"].(bool)        //nolint:errcheck // zero value ok
				resource["meets_security_baseline"] = hasEncryption && hasDLQ

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

func checkPublicSQSPolicy(policyJSON string) bool {
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
