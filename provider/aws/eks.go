package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"

	"github.com/kopexa-grc/kspec/core"
)

// EKSClusterResource fetches EKS clusters.
type EKSClusterResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *EKSClusterResource) Name() string {
	return "aws_eks_cluster"
}

// Fetch retrieves all EKS clusters across configured regions.
func (r *EKSClusterResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := eks.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := eks.NewListClustersPaginator(client, &eks.ListClustersInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, clusterName := range page.Clusters {
				// Get cluster details
				clusterResp, err := client.DescribeCluster(ctx, &eks.DescribeClusterInput{
					Name: aws.String(clusterName),
				})
				if err != nil || clusterResp.Cluster == nil {
					continue
				}

				cluster := clusterResp.Cluster
				resource := make(core.Resource)
				resource["id"] = clusterName
				resource["arn"] = aws.ToString(cluster.Arn)
				resource["name"] = clusterName
				resource["region"] = region

				resource["version"] = aws.ToString(cluster.Version)
				resource["status"] = string(cluster.Status)
				resource["platform_version"] = aws.ToString(cluster.PlatformVersion)
				resource["role_arn"] = aws.ToString(cluster.RoleArn)

				// Endpoint
				resource["endpoint"] = aws.ToString(cluster.Endpoint)

				// VPC config
				if cluster.ResourcesVpcConfig != nil {
					vpc := cluster.ResourcesVpcConfig
					resource["vpc_id"] = aws.ToString(vpc.VpcId)
					resource["subnet_ids"] = vpc.SubnetIds
					resource["security_group_ids"] = vpc.SecurityGroupIds
					resource["cluster_security_group_id"] = aws.ToString(vpc.ClusterSecurityGroupId)

					// Endpoint access
					resource["endpoint_public_access"] = vpc.EndpointPublicAccess
					resource["endpoint_private_access"] = vpc.EndpointPrivateAccess
					resource["public_access_cidrs"] = vpc.PublicAccessCidrs

					// Computed: Public endpoint with unrestricted access
					isPublicUnrestricted := vpc.EndpointPublicAccess
					if isPublicUnrestricted {
						for _, cidr := range vpc.PublicAccessCidrs {
							if cidr == "0.0.0.0/0" {
								resource["has_unrestricted_public_access"] = true
								break
							}
						}
					}
				}

				// Logging
				if cluster.Logging != nil {
					var enabledLogs []string
					for _, logSetup := range cluster.Logging.ClusterLogging {
						if logSetup.Enabled != nil && *logSetup.Enabled {
							for _, logType := range logSetup.Types {
								enabledLogs = append(enabledLogs, string(logType))
							}
						}
					}
					resource["enabled_log_types"] = enabledLogs
					resource["has_logging"] = len(enabledLogs) > 0
					resource["has_all_logging"] = len(enabledLogs) >= 5 // api, audit, authenticator, controllerManager, scheduler
				} else {
					resource["has_logging"] = false
				}

				// Encryption config
				hasEncryption := false
				for _, encConfig := range cluster.EncryptionConfig {
					if encConfig.Provider != nil && aws.ToString(encConfig.Provider.KeyArn) != "" {
						hasEncryption = true
						resource["encryption_key_arn"] = aws.ToString(encConfig.Provider.KeyArn)
						break
					}
				}
				resource["has_secrets_encryption"] = hasEncryption

				// Identity
				if cluster.Identity != nil && cluster.Identity.Oidc != nil {
					resource["oidc_issuer"] = aws.ToString(cluster.Identity.Oidc.Issuer)
				}

				// Created at
				if cluster.CreatedAt != nil {
					resource["created_at"] = cluster.CreatedAt.String()
				}

				// Tags
				resource["tags"] = cluster.Tags

				// Computed
				resource["is_active"] = cluster.Status == statusActive

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// EKSNodegroupResource fetches EKS node groups.
type EKSNodegroupResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *EKSNodegroupResource) Name() string {
	return "aws_eks_nodegroup"
}

// Fetch retrieves all EKS node groups across configured regions.
func (r *EKSNodegroupResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := eks.NewFromConfig(r.conn.ConfigForRegion(region))

		// First list clusters
		clusterPaginator := eks.NewListClustersPaginator(client, &eks.ListClustersInput{})

		for clusterPaginator.HasMorePages() {
			clusterPage, err := clusterPaginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, clusterName := range clusterPage.Clusters {
				// List node groups for each cluster
				ngPaginator := eks.NewListNodegroupsPaginator(client, &eks.ListNodegroupsInput{
					ClusterName: aws.String(clusterName),
				})

				for ngPaginator.HasMorePages() {
					ngPage, err := ngPaginator.NextPage(ctx)
					if err != nil {
						continue
					}

					for _, ngName := range ngPage.Nodegroups {
						ngResp, err := client.DescribeNodegroup(ctx, &eks.DescribeNodegroupInput{
							ClusterName:   aws.String(clusterName),
							NodegroupName: aws.String(ngName),
						})
						if err != nil || ngResp.Nodegroup == nil {
							continue
						}

						ng := ngResp.Nodegroup
						resource := make(core.Resource)
						resource["id"] = ngName
						resource["arn"] = aws.ToString(ng.NodegroupArn)
						resource["name"] = ngName
						resource["cluster_name"] = clusterName
						resource["region"] = region

						resource["status"] = string(ng.Status)
						resource["node_role"] = aws.ToString(ng.NodeRole)
						resource["ami_type"] = string(ng.AmiType)
						resource["capacity_type"] = string(ng.CapacityType)
						resource["disk_size"] = ng.DiskSize

						// Instance types
						resource["instance_types"] = ng.InstanceTypes

						// Scaling
						if ng.ScalingConfig != nil {
							resource["desired_size"] = ng.ScalingConfig.DesiredSize
							resource["min_size"] = ng.ScalingConfig.MinSize
							resource["max_size"] = ng.ScalingConfig.MaxSize
						}

						// Subnets
						resource["subnet_ids"] = ng.Subnets

						// Remote access
						if ng.RemoteAccess != nil {
							resource["remote_access_sg"] = aws.ToString(ng.RemoteAccess.Ec2SshKey)
							resource["has_remote_access"] = true
						} else {
							resource["has_remote_access"] = false
						}

						// Tags
						resource["tags"] = ng.Tags

						// Computed
						resource["is_active"] = ng.Status == statusActive
						resource["is_spot"] = ng.CapacityType == "SPOT"

						resources = append(resources, resource)
					}
				}
			}
		}
	}

	return resources, nil
}
