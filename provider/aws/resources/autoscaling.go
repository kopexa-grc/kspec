// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"

	"github.com/kopexa-grc/kspec/core"
)

// AutoScalingClient is the interface for Auto Scaling operations.
type AutoScalingClient interface {
	DescribeAutoScalingGroups(ctx context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error)
	DescribeLaunchConfigurations(ctx context.Context, params *autoscaling.DescribeLaunchConfigurationsInput, optFns ...func(*autoscaling.Options)) (*autoscaling.DescribeLaunchConfigurationsOutput, error)
}

// AutoScalingClientFactory creates Auto Scaling clients for specific regions.
type AutoScalingClientFactory func(region string) AutoScalingClient

// AutoScalingGroup fetches Auto Scaling groups.
type AutoScalingGroup struct {
	clientFactory AutoScalingClientFactory
	accountID     string
	regions       []string
}

// NewAutoScalingGroup creates a new AutoScalingGroup resource.
func NewAutoScalingGroup(clientFactory AutoScalingClientFactory, accountID string, regions []string) *AutoScalingGroup {
	return &AutoScalingGroup{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *AutoScalingGroup) Name() string {
	return "aws_autoscaling_group"
}

// Fetch retrieves all Auto Scaling groups across configured regions.
func (r *AutoScalingGroup) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := autoscaling.NewDescribeAutoScalingGroupsPaginator(client, &autoscaling.DescribeAutoScalingGroupsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, asg := range page.AutoScalingGroups {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(asg.AutoScalingGroupName)
				resource["arn"] = aws.ToString(asg.AutoScalingGroupARN)
				resource["name"] = aws.ToString(asg.AutoScalingGroupName)
				resource["region"] = region

				// Capacity
				resource["min_size"] = asg.MinSize
				resource["max_size"] = asg.MaxSize
				resource["desired_capacity"] = asg.DesiredCapacity

				// Current state
				resource["instance_count"] = len(asg.Instances)

				// Launch configuration/template
				resource["launch_configuration_name"] = aws.ToString(asg.LaunchConfigurationName)
				resource["has_launch_configuration"] = asg.LaunchConfigurationName != nil && aws.ToString(asg.LaunchConfigurationName) != ""

				if asg.LaunchTemplate != nil {
					resource["launch_template_id"] = aws.ToString(asg.LaunchTemplate.LaunchTemplateId)
					resource["launch_template_name"] = aws.ToString(asg.LaunchTemplate.LaunchTemplateName)
					resource["launch_template_version"] = aws.ToString(asg.LaunchTemplate.Version)
					resource["has_launch_template"] = true
				} else {
					resource["has_launch_template"] = false
				}

				// Mixed instances policy
				if asg.MixedInstancesPolicy != nil {
					resource["has_mixed_instances_policy"] = true
					if asg.MixedInstancesPolicy.LaunchTemplate != nil {
						var overrideTypes []string
						for _, override := range asg.MixedInstancesPolicy.LaunchTemplate.Overrides {
							if override.InstanceType != nil {
								overrideTypes = append(overrideTypes, aws.ToString(override.InstanceType))
							}
						}
						resource["instance_type_overrides"] = overrideTypes
					}
					if asg.MixedInstancesPolicy.InstancesDistribution != nil {
						dist := asg.MixedInstancesPolicy.InstancesDistribution
						resource["on_demand_percentage"] = dist.OnDemandPercentageAboveBaseCapacity
						resource["spot_allocation_strategy"] = aws.ToString(dist.SpotAllocationStrategy)
					}
				} else {
					resource["has_mixed_instances_policy"] = false
				}

				// Availability zones
				resource["availability_zones"] = asg.AvailabilityZones
				resource["az_count"] = len(asg.AvailabilityZones)
				resource["is_multi_az"] = len(asg.AvailabilityZones) > 1

				// Subnets
				resource["vpc_zone_identifier"] = aws.ToString(asg.VPCZoneIdentifier)
				resource["has_vpc"] = asg.VPCZoneIdentifier != nil && aws.ToString(asg.VPCZoneIdentifier) != ""

				// Load balancers
				resource["load_balancer_names"] = asg.LoadBalancerNames
				resource["target_group_arns"] = asg.TargetGroupARNs
				resource["has_load_balancer"] = len(asg.LoadBalancerNames) > 0 || len(asg.TargetGroupARNs) > 0

				// Health check
				resource["health_check_type"] = aws.ToString(asg.HealthCheckType)
				resource["health_check_grace_period"] = asg.HealthCheckGracePeriod
				resource["uses_elb_health_check"] = aws.ToString(asg.HealthCheckType) == "ELB"
				resource["uses_ec2_health_check"] = aws.ToString(asg.HealthCheckType) == "EC2"

				// Termination policies
				resource["termination_policies"] = asg.TerminationPolicies

				// Metrics
				enabledMetrics := make([]string, 0, len(asg.EnabledMetrics))
				for _, metric := range asg.EnabledMetrics {
					enabledMetrics = append(enabledMetrics, aws.ToString(metric.Metric))
				}
				resource["enabled_metrics"] = enabledMetrics
				resource["has_detailed_monitoring"] = len(enabledMetrics) > 0

				// Suspended processes
				suspendedProcesses := make([]string, 0, len(asg.SuspendedProcesses))
				for _, sp := range asg.SuspendedProcesses {
					suspendedProcesses = append(suspendedProcesses, aws.ToString(sp.ProcessName))
				}
				resource["suspended_processes"] = suspendedProcesses
				resource["has_suspended_processes"] = len(suspendedProcesses) > 0

				// Capacity rebalance
				resource["capacity_rebalance"] = asg.CapacityRebalance != nil && *asg.CapacityRebalance

				// Default cooldown
				resource["default_cooldown"] = asg.DefaultCooldown

				// Service-linked role
				resource["service_linked_role_arn"] = aws.ToString(asg.ServiceLinkedRoleARN)

				// Max instance lifetime
				if asg.MaxInstanceLifetime != nil {
					resource["max_instance_lifetime"] = *asg.MaxInstanceLifetime
					resource["has_max_instance_lifetime"] = true
				} else {
					resource["has_max_instance_lifetime"] = false
				}

				// New instances protected from scale in
				resource["new_instances_protected_from_scale_in"] = asg.NewInstancesProtectedFromScaleIn != nil && *asg.NewInstancesProtectedFromScaleIn

				// Warm pool
				if asg.WarmPoolConfiguration != nil {
					resource["has_warm_pool"] = true
					resource["warm_pool_status"] = string(asg.WarmPoolConfiguration.PoolState)
					if asg.WarmPoolConfiguration.MaxGroupPreparedCapacity != nil {
						resource["warm_pool_max_capacity"] = *asg.WarmPoolConfiguration.MaxGroupPreparedCapacity
					}
					if asg.WarmPoolConfiguration.MinSize != nil {
						resource["warm_pool_min_size"] = *asg.WarmPoolConfiguration.MinSize
					}
				} else {
					resource["has_warm_pool"] = false
				}

				// Instance maintenance policy
				if asg.InstanceMaintenancePolicy != nil {
					resource["has_instance_maintenance_policy"] = true
				} else {
					resource["has_instance_maintenance_policy"] = false
				}

				// Created time
				if asg.CreatedTime != nil {
					resource["created_time"] = asg.CreatedTime.String()
				}

				// Status
				resource["status"] = aws.ToString(asg.Status)

				// Tags
				tags := make(map[string]string)
				for _, tag := range asg.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				// Instances detail
				instanceIDs := make([]string, 0, len(asg.Instances))
				healthyCount := 0
				unhealthyCount := 0
				for _, instance := range asg.Instances {
					instanceIDs = append(instanceIDs, aws.ToString(instance.InstanceId))
					if aws.ToString(instance.HealthStatus) == "Healthy" {
						healthyCount++
					} else {
						unhealthyCount++
					}
				}
				resource["instance_ids"] = instanceIDs
				resource["healthy_instance_count"] = healthyCount
				resource["unhealthy_instance_count"] = unhealthyCount
				resource["all_instances_healthy"] = unhealthyCount == 0 && len(asg.Instances) > 0

				// Computed
				resource["is_at_desired_capacity"] = len(asg.Instances) == int(*asg.DesiredCapacity)
				resource["is_at_min_capacity"] = len(asg.Instances) == int(*asg.MinSize)
				resource["is_at_max_capacity"] = len(asg.Instances) == int(*asg.MaxSize)

				// Security baseline
				isMultiAZ, _ := resource["is_multi_az"].(bool)              //nolint:errcheck // zero value ok
				hasELBHealth, _ := resource["uses_elb_health_check"].(bool) //nolint:errcheck // zero value ok
				hasLB, _ := resource["has_load_balancer"].(bool)            //nolint:errcheck // zero value ok
				resource["meets_ha_baseline"] = isMultiAZ && ((hasLB && hasELBHealth) || !hasLB)

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// AutoScalingLaunchConfiguration fetches Auto Scaling launch configurations.
type AutoScalingLaunchConfiguration struct {
	clientFactory AutoScalingClientFactory
	accountID     string
	regions       []string
}

// NewAutoScalingLaunchConfiguration creates a new AutoScalingLaunchConfiguration resource.
func NewAutoScalingLaunchConfiguration(clientFactory AutoScalingClientFactory, accountID string, regions []string) *AutoScalingLaunchConfiguration {
	return &AutoScalingLaunchConfiguration{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *AutoScalingLaunchConfiguration) Name() string {
	return "aws_autoscaling_launchconfig"
}

// Fetch retrieves all Auto Scaling launch configurations across configured regions.
func (r *AutoScalingLaunchConfiguration) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := autoscaling.NewDescribeLaunchConfigurationsPaginator(client, &autoscaling.DescribeLaunchConfigurationsInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, lc := range page.LaunchConfigurations {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(lc.LaunchConfigurationName)
				resource["arn"] = aws.ToString(lc.LaunchConfigurationARN)
				resource["name"] = aws.ToString(lc.LaunchConfigurationName)
				resource["region"] = region

				resource["image_id"] = aws.ToString(lc.ImageId)
				resource["instance_type"] = aws.ToString(lc.InstanceType)
				resource["key_name"] = aws.ToString(lc.KeyName)
				resource["kernel_id"] = aws.ToString(lc.KernelId)
				resource["ramdisk_id"] = aws.ToString(lc.RamdiskId)

				// Security groups
				resource["security_groups"] = lc.SecurityGroups
				resource["security_group_count"] = len(lc.SecurityGroups)

				// IAM instance profile
				resource["iam_instance_profile"] = aws.ToString(lc.IamInstanceProfile)
				resource["has_iam_profile"] = lc.IamInstanceProfile != nil && aws.ToString(lc.IamInstanceProfile) != ""

				// Spot price
				resource["spot_price"] = aws.ToString(lc.SpotPrice)
				resource["is_spot"] = lc.SpotPrice != nil && aws.ToString(lc.SpotPrice) != ""

				// EBS optimized
				resource["ebs_optimized"] = lc.EbsOptimized != nil && *lc.EbsOptimized

				// Monitoring
				if lc.InstanceMonitoring != nil {
					resource["detailed_monitoring"] = lc.InstanceMonitoring.Enabled != nil && *lc.InstanceMonitoring.Enabled
				} else {
					resource["detailed_monitoring"] = false
				}

				// Placement tenancy
				resource["placement_tenancy"] = aws.ToString(lc.PlacementTenancy)

				// Associate public IP
				resource["associate_public_ip_address"] = lc.AssociatePublicIpAddress != nil && *lc.AssociatePublicIpAddress

				// Block device mappings
				var ebsVolumes []map[string]interface{}
				hasEncryptedVolumes := true
				for _, bdm := range lc.BlockDeviceMappings {
					if bdm.Ebs != nil {
						vol := map[string]interface{}{
							"device_name":           aws.ToString(bdm.DeviceName),
							"volume_size":           bdm.Ebs.VolumeSize,
							"volume_type":           aws.ToString(bdm.Ebs.VolumeType),
							"encrypted":             bdm.Ebs.Encrypted != nil && *bdm.Ebs.Encrypted,
							"delete_on_termination": bdm.Ebs.DeleteOnTermination != nil && *bdm.Ebs.DeleteOnTermination,
						}
						ebsVolumes = append(ebsVolumes, vol)
						if bdm.Ebs.Encrypted == nil || !*bdm.Ebs.Encrypted {
							hasEncryptedVolumes = false
						}
					}
				}
				resource["block_device_mappings"] = ebsVolumes
				resource["volume_count"] = len(ebsVolumes)
				resource["has_encrypted_volumes"] = hasEncryptedVolumes || len(ebsVolumes) == 0

				// User data
				resource["has_user_data"] = lc.UserData != nil && aws.ToString(lc.UserData) != ""

				// Classic link
				resource["classic_link_vpc_id"] = aws.ToString(lc.ClassicLinkVPCId)
				resource["classic_link_vpc_security_groups"] = lc.ClassicLinkVPCSecurityGroups

				// Metadata options
				if lc.MetadataOptions != nil {
					resource["http_tokens"] = string(lc.MetadataOptions.HttpTokens)
					resource["http_put_response_hop_limit"] = lc.MetadataOptions.HttpPutResponseHopLimit
					resource["http_endpoint"] = string(lc.MetadataOptions.HttpEndpoint)
					resource["requires_imdsv2"] = lc.MetadataOptions.HttpTokens == "required"
				} else {
					resource["requires_imdsv2"] = false
				}

				// Created time
				if lc.CreatedTime != nil {
					resource["created_time"] = lc.CreatedTime.String()
				}

				// Security baseline
				hasIAM, _ := resource["has_iam_profile"].(bool)              //nolint:errcheck // zero value ok
				hasEncryption, _ := resource["has_encrypted_volumes"].(bool) //nolint:errcheck // zero value ok
				requiresIMDSv2, _ := resource["requires_imdsv2"].(bool)      //nolint:errcheck // zero value ok
				resource["meets_security_baseline"] = hasIAM && hasEncryption && requiresIMDSv2

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
