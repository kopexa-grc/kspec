package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/kopexa-grc/kspec/core"
)

// DynamoDBTableResource fetches DynamoDB tables.
type DynamoDBTableResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *DynamoDBTableResource) Name() string {
	return "aws_dynamodb_table"
}

// Fetch retrieves all DynamoDB tables across configured regions.
func (r *DynamoDBTableResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := dynamodb.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := dynamodb.NewListTablesPaginator(client, &dynamodb.ListTablesInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, tableName := range page.TableNames {
				// Describe table
				descResp, err := client.DescribeTable(ctx, &dynamodb.DescribeTableInput{
					TableName: aws.String(tableName),
				})
				if err != nil || descResp.Table == nil {
					continue
				}

				table := descResp.Table
				resource := make(core.Resource)
				resource["id"] = tableName
				resource["arn"] = aws.ToString(table.TableArn)
				resource["name"] = tableName
				resource["region"] = region

				resource["status"] = string(table.TableStatus)
				resource["item_count"] = table.ItemCount
				resource["table_size_bytes"] = table.TableSizeBytes

				// Billing mode
				if table.BillingModeSummary != nil {
					resource["billing_mode"] = string(table.BillingModeSummary.BillingMode)
					resource["is_on_demand"] = table.BillingModeSummary.BillingMode == "PAY_PER_REQUEST"
				} else {
					resource["billing_mode"] = "PROVISIONED"
					resource["is_on_demand"] = false
				}

				// Provisioned throughput
				if table.ProvisionedThroughput != nil {
					resource["read_capacity_units"] = table.ProvisionedThroughput.ReadCapacityUnits
					resource["write_capacity_units"] = table.ProvisionedThroughput.WriteCapacityUnits
				}

				// Encryption
				if table.SSEDescription != nil {
					resource["sse_status"] = string(table.SSEDescription.Status)
					resource["sse_type"] = string(table.SSEDescription.SSEType)
					resource["kms_key_arn"] = aws.ToString(table.SSEDescription.KMSMasterKeyArn)
					resource["has_encryption"] = table.SSEDescription.Status == statusEnabled
					resource["has_cmk_encryption"] = table.SSEDescription.SSEType == "KMS" && aws.ToString(table.SSEDescription.KMSMasterKeyArn) != ""
				} else {
					resource["has_encryption"] = false
					resource["has_cmk_encryption"] = false
				}

				// Stream
				if table.StreamSpecification != nil {
					resource["stream_enabled"] = table.StreamSpecification.StreamEnabled != nil && *table.StreamSpecification.StreamEnabled
					resource["stream_view_type"] = string(table.StreamSpecification.StreamViewType)
				} else {
					resource["stream_enabled"] = false
				}
				resource["latest_stream_arn"] = aws.ToString(table.LatestStreamArn)

				// Global secondary indexes
				resource["gsi_count"] = len(table.GlobalSecondaryIndexes)
				var gsiNames []string
				for _, gsi := range table.GlobalSecondaryIndexes {
					gsiNames = append(gsiNames, aws.ToString(gsi.IndexName))
				}
				resource["global_secondary_indexes"] = gsiNames

				// Local secondary indexes
				resource["lsi_count"] = len(table.LocalSecondaryIndexes)
				var lsiNames []string
				for _, lsi := range table.LocalSecondaryIndexes {
					lsiNames = append(lsiNames, aws.ToString(lsi.IndexName))
				}
				resource["local_secondary_indexes"] = lsiNames

				// Replicas (global tables)
				resource["replica_count"] = len(table.Replicas)
				var replicaRegions []string
				for _, replica := range table.Replicas {
					replicaRegions = append(replicaRegions, aws.ToString(replica.RegionName))
				}
				resource["replica_regions"] = replicaRegions
				resource["is_global_table"] = len(table.Replicas) > 0

				// Table class
				if table.TableClassSummary != nil {
					resource["table_class"] = string(table.TableClassSummary.TableClass)
				}

				// Deletion protection
				resource["deletion_protection_enabled"] = table.DeletionProtectionEnabled != nil && *table.DeletionProtectionEnabled
				resource["has_deletion_protection"] = table.DeletionProtectionEnabled != nil && *table.DeletionProtectionEnabled

				// Point-in-time recovery
				pitrResp, err := client.DescribeContinuousBackups(ctx, &dynamodb.DescribeContinuousBackupsInput{
					TableName: aws.String(tableName),
				})
				if err == nil && pitrResp.ContinuousBackupsDescription != nil {
					pitr := pitrResp.ContinuousBackupsDescription.PointInTimeRecoveryDescription
					if pitr != nil {
						resource["pitr_status"] = string(pitr.PointInTimeRecoveryStatus)
						resource["has_point_in_time_recovery"] = pitr.PointInTimeRecoveryStatus == "ENABLED"
						if pitr.EarliestRestorableDateTime != nil {
							resource["earliest_restorable_datetime"] = pitr.EarliestRestorableDateTime.String()
						}
						if pitr.LatestRestorableDateTime != nil {
							resource["latest_restorable_datetime"] = pitr.LatestRestorableDateTime.String()
						}
					} else {
						resource["has_point_in_time_recovery"] = false
					}
				} else {
					resource["has_point_in_time_recovery"] = false
				}

				// Tags
				tagsResp, err := client.ListTagsOfResource(ctx, &dynamodb.ListTagsOfResourceInput{
					ResourceArn: table.TableArn,
				})
				if err == nil {
					tags := make(map[string]string)
					for _, tag := range tagsResp.Tags {
						tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					}
					resource["tags"] = tags
				}

				// Computed
				resource["is_active"] = table.TableStatus == statusActive

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
