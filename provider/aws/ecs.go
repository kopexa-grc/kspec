package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/ecs/types"

	"github.com/kopexa-grc/kspec/core"
)

// ECSClusterResource fetches ECS clusters.
type ECSClusterResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *ECSClusterResource) Name() string {
	return "aws_ecs_cluster"
}

// Fetch retrieves all ECS clusters across configured regions.
func (r *ECSClusterResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := ecs.NewFromConfig(r.conn.ConfigForRegion(region))

		paginator := ecs.NewListClustersPaginator(client, &ecs.ListClustersInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			if len(page.ClusterArns) == 0 {
				continue
			}

			// Describe clusters
			descResp, err := client.DescribeClusters(ctx, &ecs.DescribeClustersInput{
				Clusters: page.ClusterArns,
				Include:  []types.ClusterField{"SETTINGS", "CONFIGURATIONS", "STATISTICS", "TAGS"},
			})
			if err != nil {
				continue
			}

			for _, cluster := range descResp.Clusters {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(cluster.ClusterName)
				resource["arn"] = aws.ToString(cluster.ClusterArn)
				resource["name"] = aws.ToString(cluster.ClusterName)
				resource["region"] = region

				resource["status"] = aws.ToString(cluster.Status)

				// Statistics
				resource["running_tasks_count"] = cluster.RunningTasksCount
				resource["pending_tasks_count"] = cluster.PendingTasksCount
				resource["active_services_count"] = cluster.ActiveServicesCount
				resource["registered_container_instances_count"] = cluster.RegisteredContainerInstancesCount

				// Settings
				for _, setting := range cluster.Settings {
					if string(setting.Name) == "containerInsights" {
						resource["container_insights"] = aws.ToString(setting.Value)
						resource["has_container_insights"] = aws.ToString(setting.Value) == valueEnabled
					}
				}

				// Configuration
				if cluster.Configuration != nil {
					if cluster.Configuration.ExecuteCommandConfiguration != nil {
						execConfig := cluster.Configuration.ExecuteCommandConfiguration
						resource["execute_command_kms_key_id"] = aws.ToString(execConfig.KmsKeyId)
						resource["execute_command_logging"] = string(execConfig.Logging)
						resource["has_execute_command_logging"] = execConfig.Logging != "" && execConfig.Logging != valueNone
					}
				}

				// Capacity providers
				resource["capacity_providers"] = cluster.CapacityProviders
				resource["has_fargate"] = false
				for _, cp := range cluster.CapacityProviders {
					if cp == "FARGATE" || cp == "FARGATE_SPOT" {
						resource["has_fargate"] = true
						break
					}
				}

				// Tags
				tags := make(map[string]string)
				for _, tag := range cluster.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				// Computed
				resource["is_active"] = aws.ToString(cluster.Status) == statusActive
				resource["has_services"] = cluster.ActiveServicesCount > 0
				resource["has_tasks"] = cluster.RunningTasksCount > 0

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// ECSServiceResource fetches ECS services.
type ECSServiceResource struct {
	conn *Connection
}

// Name returns the resource type name.
func (r *ECSServiceResource) Name() string {
	return "aws_ecs_service"
}

// Fetch retrieves all ECS services across configured regions.
func (r *ECSServiceResource) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.conn.regions {
		client := ecs.NewFromConfig(r.conn.ConfigForRegion(region))

		// First list clusters
		clusterPaginator := ecs.NewListClustersPaginator(client, &ecs.ListClustersInput{})

		for clusterPaginator.HasMorePages() {
			clusterPage, err := clusterPaginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, clusterArn := range clusterPage.ClusterArns {
				// List services for each cluster
				svcPaginator := ecs.NewListServicesPaginator(client, &ecs.ListServicesInput{
					Cluster: aws.String(clusterArn),
				})

				for svcPaginator.HasMorePages() {
					svcPage, err := svcPaginator.NextPage(ctx)
					if err != nil {
						continue
					}

					if len(svcPage.ServiceArns) == 0 {
						continue
					}

					// Describe services
					descResp, err := client.DescribeServices(ctx, &ecs.DescribeServicesInput{
						Cluster:  aws.String(clusterArn),
						Services: svcPage.ServiceArns,
						Include:  []types.ServiceField{"TAGS"},
					})
					if err != nil {
						continue
					}

					for _, svc := range descResp.Services {
						resource := make(core.Resource)
						resource["id"] = aws.ToString(svc.ServiceName)
						resource["arn"] = aws.ToString(svc.ServiceArn)
						resource["name"] = aws.ToString(svc.ServiceName)
						resource["cluster_arn"] = clusterArn
						resource["region"] = region

						resource["status"] = aws.ToString(svc.Status)
						resource["launch_type"] = string(svc.LaunchType)
						resource["task_definition"] = aws.ToString(svc.TaskDefinition)
						resource["desired_count"] = svc.DesiredCount
						resource["running_count"] = svc.RunningCount
						resource["pending_count"] = svc.PendingCount

						// Network configuration
						if svc.NetworkConfiguration != nil && svc.NetworkConfiguration.AwsvpcConfiguration != nil {
							vpc := svc.NetworkConfiguration.AwsvpcConfiguration
							resource["subnets"] = vpc.Subnets
							resource["security_groups"] = vpc.SecurityGroups
							resource["assign_public_ip"] = string(vpc.AssignPublicIp)
							resource["has_public_ip"] = vpc.AssignPublicIp == statusEnabled
						}

						// Load balancers
						var lbTargets []string
						for _, lb := range svc.LoadBalancers {
							lbTargets = append(lbTargets, aws.ToString(lb.TargetGroupArn))
						}
						resource["load_balancers"] = lbTargets
						resource["has_load_balancer"] = len(lbTargets) > 0

						// Deployment configuration
						if svc.DeploymentConfiguration != nil {
							resource["deployment_maximum_percent"] = svc.DeploymentConfiguration.MaximumPercent
							resource["deployment_minimum_healthy_percent"] = svc.DeploymentConfiguration.MinimumHealthyPercent
						}

						// Tags
						tags := make(map[string]string)
						for _, tag := range svc.Tags {
							tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
						}
						resource["tags"] = tags

						// Computed
						resource["is_active"] = aws.ToString(svc.Status) == statusActive
						resource["is_fargate"] = svc.LaunchType == "FARGATE"
						resource["is_healthy"] = svc.RunningCount == svc.DesiredCount

						resources = append(resources, resource)
					}
				}
			}
		}
	}

	return resources, nil
}
