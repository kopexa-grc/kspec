package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ptr"
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
				multiAZ := ptr.Deref(instance.MultiAZ, false)
				resource["multi_az"] = multiAZ
				resource["has_multi_az"] = multiAZ

				// Network
				resource["vpc_id"] = ""
				if instance.DBSubnetGroup != nil {
					resource["vpc_id"] = aws.ToString(instance.DBSubnetGroup.VpcId)
					resource["subnet_group"] = aws.ToString(instance.DBSubnetGroup.DBSubnetGroupName)
				}
				publiclyAccessible := ptr.Deref(instance.PubliclyAccessible, false)
				resource["publicly_accessible"] = publiclyAccessible
				resource["is_public"] = publiclyAccessible

				// Security
				storageEncrypted := ptr.Deref(instance.StorageEncrypted, false)
				resource["storage_encrypted"] = storageEncrypted
				resource["is_encrypted"] = storageEncrypted
				resource["kms_key_id"] = aws.ToString(instance.KmsKeyId)

				// IAM auth
				iamAuth := ptr.Deref(instance.IAMDatabaseAuthenticationEnabled, false)
				resource["iam_database_authentication_enabled"] = iamAuth
				resource["has_iam_auth"] = iamAuth

				// Deletion protection
				deletionProtection := ptr.Deref(instance.DeletionProtection, false)
				resource["deletion_protection"] = deletionProtection
				resource["has_deletion_protection"] = deletionProtection

				// Auto minor version upgrade
				autoUpgrade := ptr.Deref(instance.AutoMinorVersionUpgrade, false)
				resource["auto_minor_version_upgrade"] = autoUpgrade
				resource["has_auto_minor_upgrade"] = autoUpgrade

				// Backup
				backupRetention := ptr.Deref(instance.BackupRetentionPeriod, int32(0))
				resource["backup_retention_period"] = backupRetention
				resource["backup_retention_days"] = backupRetention
				resource["has_backups"] = backupRetention > 0

				// Performance Insights
				perfInsights := ptr.Deref(instance.PerformanceInsightsEnabled, false)
				resource["performance_insights_enabled"] = perfInsights
				resource["has_performance_insights"] = perfInsights

				// Enhanced monitoring
				monitoringInterval := ptr.Deref(instance.MonitoringInterval, int32(0))
				resource["monitoring_interval"] = monitoringInterval
				resource["has_enhanced_monitoring"] = monitoringInterval > 0

				// Endpoint
				if instance.Endpoint != nil {
					resource["endpoint_address"] = aws.ToString(instance.Endpoint.Address)
					resource["endpoint_port"] = instance.Endpoint.Port
				}

				// Security groups
				sgIDs := make([]string, 0, len(instance.VpcSecurityGroups))
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
				resource["multi_az"] = ptr.Deref(cluster.MultiAZ, false)

				// Storage encryption
				storageEncrypted := ptr.Deref(cluster.StorageEncrypted, false)
				resource["storage_encrypted"] = storageEncrypted
				resource["is_encrypted"] = storageEncrypted
				resource["kms_key_id"] = aws.ToString(cluster.KmsKeyId)

				// IAM auth
				iamAuth := ptr.Deref(cluster.IAMDatabaseAuthenticationEnabled, false)
				resource["iam_database_authentication_enabled"] = iamAuth
				resource["has_iam_auth"] = iamAuth

				// Deletion protection
				deletionProtection := ptr.Deref(cluster.DeletionProtection, false)
				resource["deletion_protection"] = deletionProtection
				resource["has_deletion_protection"] = deletionProtection

				// Backup
				backupRetention := ptr.Deref(cluster.BackupRetentionPeriod, int32(0))
				resource["backup_retention_period"] = backupRetention
				resource["has_backups"] = backupRetention > 0

				// Endpoints
				resource["endpoint"] = aws.ToString(cluster.Endpoint)
				resource["reader_endpoint"] = aws.ToString(cluster.ReaderEndpoint)
				resource["port"] = cluster.Port

				// Members
				resource["member_count"] = len(cluster.DBClusterMembers)

				// Security groups
				sgIDs := make([]string, 0, len(cluster.VpcSecurityGroups))
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
