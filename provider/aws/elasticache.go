package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"

	"github.com/kopexa-grc/kspec/core"
)

// ElastiCacheClusterResource fetches ElastiCache clusters.
type ElastiCacheClusterResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *ElastiCacheClusterResource) Name() string {
	return "aws_elasticache_cluster"
}

// Fetch retrieves all ElastiCache clusters across configured regions.
func (r *ElastiCacheClusterResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := elasticache.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := elasticache.NewDescribeCacheClustersPaginator(client, &elasticache.DescribeCacheClustersInput{
			ShowCacheNodeInfo: aws.Bool(true),
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, cluster := range page.CacheClusters {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(cluster.CacheClusterId)
				resource["arn"] = aws.ToString(cluster.ARN)
				resource["name"] = aws.ToString(cluster.CacheClusterId)
				resource["region"] = region

				resource["engine"] = aws.ToString(cluster.Engine)
				resource["engine_version"] = aws.ToString(cluster.EngineVersion)
				resource["cache_node_type"] = aws.ToString(cluster.CacheNodeType)
				resource["status"] = aws.ToString(cluster.CacheClusterStatus)
				resource["num_cache_nodes"] = cluster.NumCacheNodes

				// Network
				resource["preferred_availability_zone"] = aws.ToString(cluster.PreferredAvailabilityZone)
				if cluster.CacheSubnetGroupName != nil {
					resource["subnet_group_name"] = aws.ToString(cluster.CacheSubnetGroupName)
				}

				// Security groups
				sgIDs := make([]string, 0, len(cluster.SecurityGroups))
				for _, sg := range cluster.SecurityGroups {
					sgIDs = append(sgIDs, aws.ToString(sg.SecurityGroupId))
				}
				resource["security_groups"] = sgIDs

				// Parameter group
				if cluster.CacheParameterGroup != nil {
					resource["parameter_group_name"] = aws.ToString(cluster.CacheParameterGroup.CacheParameterGroupName)
				}

				// Encryption
				resource["at_rest_encryption_enabled"] = cluster.AtRestEncryptionEnabled != nil && *cluster.AtRestEncryptionEnabled
				resource["transit_encryption_enabled"] = cluster.TransitEncryptionEnabled != nil && *cluster.TransitEncryptionEnabled
				resource["has_encryption_at_rest"] = cluster.AtRestEncryptionEnabled != nil && *cluster.AtRestEncryptionEnabled
				resource["has_encryption_in_transit"] = cluster.TransitEncryptionEnabled != nil && *cluster.TransitEncryptionEnabled

				// Auth token (Redis AUTH)
				resource["auth_token_enabled"] = cluster.AuthTokenEnabled != nil && *cluster.AuthTokenEnabled
				resource["has_auth_token"] = cluster.AuthTokenEnabled != nil && *cluster.AuthTokenEnabled

				// Snapshot
				resource["snapshot_retention_limit"] = cluster.SnapshotRetentionLimit
				resource["snapshot_window"] = aws.ToString(cluster.SnapshotWindow)
				resource["has_snapshots"] = cluster.SnapshotRetentionLimit != nil && *cluster.SnapshotRetentionLimit > 0

				// Maintenance
				resource["maintenance_window"] = aws.ToString(cluster.PreferredMaintenanceWindow)

				// Auto minor version upgrade
				resource["auto_minor_version_upgrade"] = cluster.AutoMinorVersionUpgrade

				// Notification
				resource["notification_topic_arn"] = aws.ToString(cluster.NotificationConfiguration.TopicArn)
				resource["has_notifications"] = cluster.NotificationConfiguration != nil && aws.ToString(cluster.NotificationConfiguration.TopicArn) != ""

				// Replication group
				resource["replication_group_id"] = aws.ToString(cluster.ReplicationGroupId)
				resource["is_replication_member"] = cluster.ReplicationGroupId != nil && aws.ToString(cluster.ReplicationGroupId) != ""

				// Network type
				resource["network_type"] = string(cluster.NetworkType)
				resource["ip_discovery"] = string(cluster.IpDiscovery)

				// Creation time
				if cluster.CacheClusterCreateTime != nil {
					resource["created_at"] = cluster.CacheClusterCreateTime.String()
				}

				// Nodes
				nodeEndpoints := make([]map[string]interface{}, 0, len(cluster.CacheNodes))
				for _, node := range cluster.CacheNodes {
					nodeInfo := map[string]interface{}{
						"id":                aws.ToString(node.CacheNodeId),
						"status":            aws.ToString(node.CacheNodeStatus),
						"availability_zone": aws.ToString(node.CustomerAvailabilityZone),
					}
					if node.Endpoint != nil {
						nodeInfo["address"] = aws.ToString(node.Endpoint.Address)
						nodeInfo["port"] = node.Endpoint.Port
					}
					nodeEndpoints = append(nodeEndpoints, nodeInfo)
				}
				resource["cache_nodes"] = nodeEndpoints
				resource["node_count"] = len(cluster.CacheNodes)

				// Computed
				resource["is_available"] = aws.ToString(cluster.CacheClusterStatus) == statusAvailable
				resource["is_redis"] = aws.ToString(cluster.Engine) == "redis"
				resource["is_memcached"] = aws.ToString(cluster.Engine) == "memcached"

				// Security baseline
				hasEncryptionAtRest, _ := resource["has_encryption_at_rest"].(bool)       //nolint:errcheck // zero value ok
				hasEncryptionInTransit, _ := resource["has_encryption_in_transit"].(bool) //nolint:errcheck // zero value ok
				hasAuth, _ := resource["has_auth_token"].(bool)                           //nolint:errcheck // zero value ok
				resource["meets_security_baseline"] = hasEncryptionAtRest && hasEncryptionInTransit && hasAuth

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// ElastiCacheReplicationGroupResource fetches ElastiCache replication groups.
type ElastiCacheReplicationGroupResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *ElastiCacheReplicationGroupResource) Name() string {
	return "aws_elasticache_replicationgroup"
}

// Fetch retrieves all ElastiCache replication groups across configured regions.
func (r *ElastiCacheReplicationGroupResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := elasticache.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := elasticache.NewDescribeReplicationGroupsPaginator(client, &elasticache.DescribeReplicationGroupsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, rg := range page.ReplicationGroups {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(rg.ReplicationGroupId)
				resource["arn"] = aws.ToString(rg.ARN)
				resource["name"] = aws.ToString(rg.ReplicationGroupId)
				resource["region"] = region

				resource["description"] = aws.ToString(rg.Description)
				resource["status"] = aws.ToString(rg.Status)

				// Global replication
				if rg.GlobalReplicationGroupInfo != nil {
					resource["global_replication_group_id"] = aws.ToString(rg.GlobalReplicationGroupInfo.GlobalReplicationGroupId)
					resource["global_replication_group_member_role"] = aws.ToString(rg.GlobalReplicationGroupInfo.GlobalReplicationGroupMemberRole)
					resource["is_global"] = true
				} else {
					resource["is_global"] = false
				}

				// Multi-AZ
				resource["automatic_failover"] = string(rg.AutomaticFailover)
				resource["multi_az"] = string(rg.MultiAZ)
				resource["has_automatic_failover"] = rg.AutomaticFailover == valueEnabled
				resource["has_multi_az"] = rg.MultiAZ == "enabled"

				// Cluster mode
				resource["cluster_enabled"] = rg.ClusterEnabled != nil && *rg.ClusterEnabled
				resource["is_cluster_mode"] = rg.ClusterEnabled != nil && *rg.ClusterEnabled

				// Node groups (shards)
				resource["node_group_count"] = len(rg.NodeGroups)
				shardInfo := make([]map[string]interface{}, 0, len(rg.NodeGroups))
				totalNodes := 0
				for _, ng := range rg.NodeGroups {
					shard := map[string]interface{}{
						"id":     aws.ToString(ng.NodeGroupId),
						"status": aws.ToString(ng.Status),
						"slots":  aws.ToString(ng.Slots),
						"nodes":  len(ng.NodeGroupMembers),
					}
					shardInfo = append(shardInfo, shard)
					totalNodes += len(ng.NodeGroupMembers)
				}
				resource["shards"] = shardInfo
				resource["total_node_count"] = totalNodes

				// Encryption
				resource["at_rest_encryption_enabled"] = rg.AtRestEncryptionEnabled != nil && *rg.AtRestEncryptionEnabled
				resource["transit_encryption_enabled"] = rg.TransitEncryptionEnabled != nil && *rg.TransitEncryptionEnabled
				resource["has_encryption_at_rest"] = rg.AtRestEncryptionEnabled != nil && *rg.AtRestEncryptionEnabled
				resource["has_encryption_in_transit"] = rg.TransitEncryptionEnabled != nil && *rg.TransitEncryptionEnabled
				resource["transit_encryption_mode"] = string(rg.TransitEncryptionMode)

				// KMS
				resource["kms_key_id"] = aws.ToString(rg.KmsKeyId)
				resource["has_cmk_encryption"] = rg.KmsKeyId != nil && aws.ToString(rg.KmsKeyId) != ""

				// Auth token
				resource["auth_token_enabled"] = rg.AuthTokenEnabled != nil && *rg.AuthTokenEnabled

				// Endpoints
				if rg.ConfigurationEndpoint != nil {
					resource["configuration_endpoint_address"] = aws.ToString(rg.ConfigurationEndpoint.Address)
					resource["configuration_endpoint_port"] = rg.ConfigurationEndpoint.Port
				}

				// Primary endpoint (non-cluster mode)
				var primaryEndpoint string
				var readerEndpoint string
				for _, ng := range rg.NodeGroups {
					if ng.PrimaryEndpoint != nil {
						primaryEndpoint = aws.ToString(ng.PrimaryEndpoint.Address)
					}
					if ng.ReaderEndpoint != nil {
						readerEndpoint = aws.ToString(ng.ReaderEndpoint.Address)
					}
				}
				resource["primary_endpoint"] = primaryEndpoint
				resource["reader_endpoint"] = readerEndpoint

				// Snapshot
				resource["snapshot_retention_limit"] = rg.SnapshotRetentionLimit
				resource["snapshot_window"] = aws.ToString(rg.SnapshotWindow)
				resource["snapshotting_cluster_id"] = aws.ToString(rg.SnapshottingClusterId)
				resource["has_snapshots"] = rg.SnapshotRetentionLimit != nil && *rg.SnapshotRetentionLimit > 0

				// Member clusters
				resource["member_clusters"] = rg.MemberClusters
				resource["member_cluster_count"] = len(rg.MemberClusters)

				// User groups
				resource["user_group_ids"] = rg.UserGroupIds
				resource["has_user_groups"] = len(rg.UserGroupIds) > 0

				// Log delivery
				logTypes := make([]string, 0, len(rg.LogDeliveryConfigurations))
				for _, ld := range rg.LogDeliveryConfigurations {
					logTypes = append(logTypes, string(ld.LogType))
				}
				resource["log_delivery_types"] = logTypes
				resource["has_log_delivery"] = len(rg.LogDeliveryConfigurations) > 0

				// Data tiering
				resource["data_tiering"] = string(rg.DataTiering)
				resource["has_data_tiering"] = rg.DataTiering == "enabled"

				// Computed
				resource["is_available"] = aws.ToString(rg.Status) == statusAvailable

				// Security baseline
				hasEncryptionAtRest, _ := resource["has_encryption_at_rest"].(bool)       //nolint:errcheck // zero value ok
				hasEncryptionInTransit, _ := resource["has_encryption_in_transit"].(bool) //nolint:errcheck // zero value ok
				hasFailover, _ := resource["has_automatic_failover"].(bool)               //nolint:errcheck // zero value ok
				hasSnapshots, _ := resource["has_snapshots"].(bool)                       //nolint:errcheck // zero value ok
				resource["meets_security_baseline"] = hasEncryptionAtRest && hasEncryptionInTransit && hasFailover && hasSnapshots

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
