package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/kopexa-grc/kspec/core"
)

// RDSInstanceResource fetches RDS DB instances.
type RDSInstanceResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *RDSInstanceResource) Name() string {
	return "aws_rds_instance"
}

// Fetch retrieves all RDS DB instances across configured regions.
func (r *RDSInstanceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := rds.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := rds.NewDescribeDBInstancesPaginator(client, &rds.DescribeDBInstancesInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, instance := range page.DBInstances {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(instance.DBInstanceIdentifier)
				resource["arn"] = aws.ToString(instance.DBInstanceArn)
				resource["region"] = region

				resource["engine"] = aws.ToString(instance.Engine)
				resource["engine_version"] = aws.ToString(instance.EngineVersion)
				resource["instance_class"] = aws.ToString(instance.DBInstanceClass)
				resource["status"] = aws.ToString(instance.DBInstanceStatus)
				resource["allocated_storage"] = instance.AllocatedStorage
				resource["storage_type"] = aws.ToString(instance.StorageType)

				// Availability
				resource["availability_zone"] = aws.ToString(instance.AvailabilityZone)
				resource["multi_az"] = instance.MultiAZ != nil && *instance.MultiAZ
				resource["has_multi_az"] = instance.MultiAZ != nil && *instance.MultiAZ

				// Network
				resource["vpc_id"] = ""
				if instance.DBSubnetGroup != nil {
					resource["vpc_id"] = aws.ToString(instance.DBSubnetGroup.VpcId)
					resource["subnet_group"] = aws.ToString(instance.DBSubnetGroup.DBSubnetGroupName)
				}
				resource["publicly_accessible"] = instance.PubliclyAccessible != nil && *instance.PubliclyAccessible
				resource["is_public"] = instance.PubliclyAccessible != nil && *instance.PubliclyAccessible

				// Security
				resource["storage_encrypted"] = instance.StorageEncrypted != nil && *instance.StorageEncrypted
				resource["is_encrypted"] = instance.StorageEncrypted != nil && *instance.StorageEncrypted
				resource["kms_key_id"] = aws.ToString(instance.KmsKeyId)

				// IAM auth
				resource["iam_database_authentication_enabled"] = instance.IAMDatabaseAuthenticationEnabled != nil && *instance.IAMDatabaseAuthenticationEnabled
				resource["has_iam_auth"] = instance.IAMDatabaseAuthenticationEnabled != nil && *instance.IAMDatabaseAuthenticationEnabled

				// Deletion protection
				resource["deletion_protection"] = instance.DeletionProtection != nil && *instance.DeletionProtection
				resource["has_deletion_protection"] = instance.DeletionProtection != nil && *instance.DeletionProtection

				// Auto minor version upgrade
				resource["auto_minor_version_upgrade"] = instance.AutoMinorVersionUpgrade != nil && *instance.AutoMinorVersionUpgrade
				resource["has_auto_minor_upgrade"] = instance.AutoMinorVersionUpgrade != nil && *instance.AutoMinorVersionUpgrade

				// Backup
				resource["backup_retention_period"] = instance.BackupRetentionPeriod
				resource["backup_retention_days"] = instance.BackupRetentionPeriod
				resource["has_backups"] = instance.BackupRetentionPeriod != nil && *instance.BackupRetentionPeriod > 0

				// Performance Insights
				resource["performance_insights_enabled"] = instance.PerformanceInsightsEnabled != nil && *instance.PerformanceInsightsEnabled
				resource["has_performance_insights"] = instance.PerformanceInsightsEnabled != nil && *instance.PerformanceInsightsEnabled

				// Enhanced monitoring
				resource["monitoring_interval"] = instance.MonitoringInterval
				resource["has_enhanced_monitoring"] = instance.MonitoringInterval != nil && *instance.MonitoringInterval > 0

				// Endpoint
				if instance.Endpoint != nil {
					resource["endpoint_address"] = aws.ToString(instance.Endpoint.Address)
					resource["endpoint_port"] = instance.Endpoint.Port
				}

				// Security groups
				var sgIDs []string
				for _, sg := range instance.VpcSecurityGroups {
					sgIDs = append(sgIDs, aws.ToString(sg.VpcSecurityGroupId))
				}
				resource["security_groups"] = sgIDs

				// Tags
				tags := make(map[string]string)
				for _, tag := range instance.TagList {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				// Computed
				resource["is_available"] = aws.ToString(instance.DBInstanceStatus) == "available"

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// RDSClusterResource fetches RDS Aurora clusters.
type RDSClusterResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *RDSClusterResource) Name() string {
	return "aws_rds_cluster"
}

// Fetch retrieves all RDS Aurora clusters across configured regions.
func (r *RDSClusterResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := rds.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := rds.NewDescribeDBClustersPaginator(client, &rds.DescribeDBClustersInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, cluster := range page.DBClusters {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(cluster.DBClusterIdentifier)
				resource["arn"] = aws.ToString(cluster.DBClusterArn)
				resource["region"] = region

				resource["engine"] = aws.ToString(cluster.Engine)
				resource["engine_version"] = aws.ToString(cluster.EngineVersion)
				resource["engine_mode"] = aws.ToString(cluster.EngineMode)
				resource["status"] = aws.ToString(cluster.Status)
				resource["database_name"] = aws.ToString(cluster.DatabaseName)

				// Multi-AZ
				resource["multi_az"] = cluster.MultiAZ != nil && *cluster.MultiAZ

				// Storage encryption
				resource["storage_encrypted"] = cluster.StorageEncrypted != nil && *cluster.StorageEncrypted
				resource["is_encrypted"] = cluster.StorageEncrypted != nil && *cluster.StorageEncrypted
				resource["kms_key_id"] = aws.ToString(cluster.KmsKeyId)

				// IAM auth
				resource["iam_database_authentication_enabled"] = cluster.IAMDatabaseAuthenticationEnabled != nil && *cluster.IAMDatabaseAuthenticationEnabled
				resource["has_iam_auth"] = cluster.IAMDatabaseAuthenticationEnabled != nil && *cluster.IAMDatabaseAuthenticationEnabled

				// Deletion protection
				resource["deletion_protection"] = cluster.DeletionProtection != nil && *cluster.DeletionProtection
				resource["has_deletion_protection"] = cluster.DeletionProtection != nil && *cluster.DeletionProtection

				// Backup
				resource["backup_retention_period"] = cluster.BackupRetentionPeriod
				resource["has_backups"] = cluster.BackupRetentionPeriod != nil && *cluster.BackupRetentionPeriod > 0

				// Endpoints
				resource["endpoint"] = aws.ToString(cluster.Endpoint)
				resource["reader_endpoint"] = aws.ToString(cluster.ReaderEndpoint)
				resource["port"] = cluster.Port

				// Members
				resource["member_count"] = len(cluster.DBClusterMembers)

				// Security groups
				var sgIDs []string
				for _, sg := range cluster.VpcSecurityGroups {
					sgIDs = append(sgIDs, aws.ToString(sg.VpcSecurityGroupId))
				}
				resource["security_groups"] = sgIDs

				// Tags
				tags := make(map[string]string)
				for _, tag := range cluster.TagList {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				// Computed
				resource["is_available"] = aws.ToString(cluster.Status) == "available"
				resource["is_serverless"] = aws.ToString(cluster.EngineMode) == "serverless"

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
