// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package resources

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ptr"
)

// EC2Client is the interface for EC2 operations.
type EC2Client interface {
	DescribeInstances(ctx context.Context, params *ec2.DescribeInstancesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error)
	DescribeVolumes(ctx context.Context, params *ec2.DescribeVolumesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVolumesOutput, error)
	DescribeSnapshots(ctx context.Context, params *ec2.DescribeSnapshotsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotsOutput, error)
	DescribeSnapshotAttribute(ctx context.Context, params *ec2.DescribeSnapshotAttributeInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSnapshotAttributeOutput, error)
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	DescribeNetworkInterfaces(ctx context.Context, params *ec2.DescribeNetworkInterfacesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNetworkInterfacesOutput, error)
	DescribeKeyPairs(ctx context.Context, params *ec2.DescribeKeyPairsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeKeyPairsOutput, error)
	DescribeVpcs(ctx context.Context, params *ec2.DescribeVpcsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error)
	DescribeVpcAttribute(ctx context.Context, params *ec2.DescribeVpcAttributeInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcAttributeOutput, error)
	DescribeSubnets(ctx context.Context, params *ec2.DescribeSubnetsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error)
	DescribeFlowLogs(ctx context.Context, params *ec2.DescribeFlowLogsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeFlowLogsOutput, error)
	DescribeInternetGateways(ctx context.Context, params *ec2.DescribeInternetGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error)
	DescribeNatGateways(ctx context.Context, params *ec2.DescribeNatGatewaysInput, optFns ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error)
	DescribeVpcEndpoints(ctx context.Context, params *ec2.DescribeVpcEndpointsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeVpcEndpointsOutput, error)
}

// EC2ClientFactory creates EC2 clients for specific regions.
type EC2ClientFactory func(region string) EC2Client

// EC2Instance fetches EC2 instances.
type EC2Instance struct {
	clientFactory EC2ClientFactory
	accountID     string
	regions       []string
}

// NewEC2Instance creates a new EC2Instance resource.
func NewEC2Instance(clientFactory EC2ClientFactory, accountID string, regions []string) *EC2Instance {
	return &EC2Instance{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *EC2Instance) Name() string {
	return "aws_ec2_instance"
}

// Fetch retrieves all EC2 instances across configured regions.
func (r *EC2Instance) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := ec2.NewDescribeInstancesPaginator(client, &ec2.DescribeInstancesInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue // Skip region on error
			}

			for _, reservation := range page.Reservations {
				for _, instance := range reservation.Instances {
					resource := make(core.Resource)
					resource["id"] = aws.ToString(instance.InstanceId)
					resource["region"] = region

					// Build ARN
					resource["arn"] = "arn:aws:ec2:" + region + ":" + r.accountID + ":instance/" + aws.ToString(instance.InstanceId)

					// Get name from tags
					for _, tag := range instance.Tags {
						if aws.ToString(tag.Key) == tagKeyName {
							resource["name"] = aws.ToString(tag.Value)
							break
						}
					}

					resource["instance_type"] = string(instance.InstanceType)
					resource["state"] = string(instance.State.Name)
					resource["architecture"] = string(instance.Architecture)
					resource["platform"] = string(instance.Platform)
					resource["vpc_id"] = aws.ToString(instance.VpcId)
					resource["subnet_id"] = aws.ToString(instance.SubnetId)
					resource["availability_zone"] = aws.ToString(instance.Placement.AvailabilityZone)

					// Network info
					resource["private_ip"] = aws.ToString(instance.PrivateIpAddress)
					resource["private_dns"] = aws.ToString(instance.PrivateDnsName)
					resource["public_ip"] = aws.ToString(instance.PublicIpAddress)
					resource["public_dns"] = aws.ToString(instance.PublicDnsName)

					// Computed: Is public
					resource["is_public"] = instance.PublicIpAddress != nil && aws.ToString(instance.PublicIpAddress) != ""

					// Launch time
					if instance.LaunchTime != nil {
						resource["launch_time"] = instance.LaunchTime.String()
					}

					// Security groups
					sgIDs := make([]string, 0, len(instance.SecurityGroups))
					for _, sg := range instance.SecurityGroups {
						sgIDs = append(sgIDs, aws.ToString(sg.GroupId))
					}
					resource["security_groups"] = sgIDs
					resource["security_group_count"] = len(sgIDs)

					// IAM instance profile
					if instance.IamInstanceProfile != nil {
						resource["iam_instance_profile_arn"] = aws.ToString(instance.IamInstanceProfile.Arn)
						resource["has_iam_profile"] = true
					} else {
						resource["has_iam_profile"] = false
					}

					// Metadata options (IMDSv2)
					if instance.MetadataOptions != nil {
						resource["metadata_http_tokens"] = string(instance.MetadataOptions.HttpTokens)
						resource["metadata_http_endpoint"] = string(instance.MetadataOptions.HttpEndpoint)
						resource["has_imdsv2"] = instance.MetadataOptions.HttpTokens == types.HttpTokensStateRequired
					} else {
						resource["has_imdsv2"] = false
					}

					// Monitoring
					if instance.Monitoring != nil {
						resource["monitoring_state"] = string(instance.Monitoring.State)
						resource["has_detailed_monitoring"] = instance.Monitoring.State == types.MonitoringStateEnabled
					} else {
						resource["has_detailed_monitoring"] = false
					}

					// EBS optimization
					resource["ebs_optimized"] = ptr.Deref(instance.EbsOptimized, false)

					// Root device
					resource["root_device_type"] = string(instance.RootDeviceType)
					resource["root_device_name"] = aws.ToString(instance.RootDeviceName)

					// Block device mappings (volumes)
					volumeIDs := make([]string, 0, len(instance.BlockDeviceMappings))
					for _, bdm := range instance.BlockDeviceMappings {
						if bdm.Ebs != nil {
							volumeIDs = append(volumeIDs, aws.ToString(bdm.Ebs.VolumeId))
						}
					}
					resource["volume_ids"] = volumeIDs
					resource["volume_count"] = len(volumeIDs)

					// Check encryption of attached volumes
					allVolumesEncrypted := true
					if len(volumeIDs) > 0 {
						volumesResp, err := client.DescribeVolumes(ctx, &ec2.DescribeVolumesInput{
							VolumeIds: volumeIDs,
						})
						if err == nil {
							for _, vol := range volumesResp.Volumes {
								if !ptr.Deref(vol.Encrypted, false) {
									allVolumesEncrypted = false
									break
								}
							}
						}
					}
					resource["has_ebs_encryption"] = allVolumesEncrypted && len(volumeIDs) > 0

					// Source/dest check
					resource["source_dest_check"] = ptr.Deref(instance.SourceDestCheck, false)

					// Hibernation
					if instance.HibernationOptions != nil {
						resource["hibernation_configured"] = ptr.Deref(instance.HibernationOptions.Configured, false)
					} else {
						resource["hibernation_configured"] = false
					}

					// Tags
					tags := make(map[string]string)
					for _, tag := range instance.Tags {
						tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
					}
					resource["tags"] = tags

					// Computed: Is running
					resource["is_running"] = instance.State.Name == types.InstanceStateNameRunning
					resource["is_stopped"] = instance.State.Name == types.InstanceStateNameStopped
					resource["is_terminated"] = instance.State.Name == types.InstanceStateNameTerminated

					resources = append(resources, resource)
				}
			}
		}
	}

	return resources, nil
}

// EC2Volume fetches EBS volumes.
type EC2Volume struct {
	clientFactory EC2ClientFactory
	accountID     string
	regions       []string
}

// NewEC2Volume creates a new EC2Volume resource.
func NewEC2Volume(clientFactory EC2ClientFactory, accountID string, regions []string) *EC2Volume {
	return &EC2Volume{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *EC2Volume) Name() string {
	return "aws_ec2_volume"
}

// Fetch retrieves all EBS volumes across configured regions.
func (r *EC2Volume) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := ec2.NewDescribeVolumesPaginator(client, &ec2.DescribeVolumesInput{})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, volume := range page.Volumes {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(volume.VolumeId)
				resource["region"] = region
				resource["arn"] = "arn:aws:ec2:" + region + ":" + r.accountID + ":volume/" + aws.ToString(volume.VolumeId)

				// Get name from tags
				for _, tag := range volume.Tags {
					if aws.ToString(tag.Key) == tagKeyName {
						resource["name"] = aws.ToString(tag.Value)
						break
					}
				}

				resource["availability_zone"] = aws.ToString(volume.AvailabilityZone)
				resource["size_gb"] = volume.Size
				resource["volume_type"] = string(volume.VolumeType)
				resource["state"] = string(volume.State)
				resource["iops"] = volume.Iops
				resource["throughput"] = volume.Throughput

				// Encryption
				encrypted := ptr.Deref(volume.Encrypted, false)
				resource["encrypted"] = encrypted
				resource["kms_key_id"] = aws.ToString(volume.KmsKeyId)
				resource["has_encryption"] = encrypted

				// Multi-attach
				resource["multi_attach_enabled"] = ptr.Deref(volume.MultiAttachEnabled, false)

				// Snapshot
				resource["snapshot_id"] = aws.ToString(volume.SnapshotId)
				resource["created_from_snapshot"] = volume.SnapshotId != nil && aws.ToString(volume.SnapshotId) != ""

				// Create time
				if volume.CreateTime != nil {
					resource["created_at"] = volume.CreateTime.String()
				}

				// Attachments
				attachedInstances := make([]string, 0, len(volume.Attachments))
				for _, attachment := range volume.Attachments {
					attachedInstances = append(attachedInstances, aws.ToString(attachment.InstanceId))
				}
				resource["attached_instances"] = attachedInstances
				resource["is_attached"] = len(attachedInstances) > 0
				resource["attachment_count"] = len(attachedInstances)

				// Is root volume
				isRoot := false
				for _, attachment := range volume.Attachments {
					device := aws.ToString(attachment.Device)
					if device == "/dev/sda1" || device == "/dev/xvda" || device == "/dev/nvme0n1" {
						isRoot = true
						break
					}
				}
				resource["is_root_volume"] = isRoot

				// Tags
				tags := make(map[string]string)
				for _, tag := range volume.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				// Computed: Available (not attached)
				resource["is_available"] = volume.State == types.VolumeStateAvailable
				resource["is_in_use"] = volume.State == types.VolumeStateInUse

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}

// EC2Snapshot fetches EBS snapshots.
type EC2Snapshot struct {
	clientFactory EC2ClientFactory
	accountID     string
	regions       []string
}

// NewEC2Snapshot creates a new EC2Snapshot resource.
func NewEC2Snapshot(clientFactory EC2ClientFactory, accountID string, regions []string) *EC2Snapshot {
	return &EC2Snapshot{
		clientFactory: clientFactory,
		accountID:     accountID,
		regions:       regions,
	}
}

// Name returns the resource type name.
func (r *EC2Snapshot) Name() string {
	return "aws_ec2_snapshot"
}

// Fetch retrieves all EBS snapshots owned by the account.
func (r *EC2Snapshot) Fetch(ctx context.Context, asset core.Asset) ([]core.Resource, error) {
	var resources []core.Resource

	for _, region := range r.regions {
		client := r.clientFactory(region)

		paginator := ec2.NewDescribeSnapshotsPaginator(client, &ec2.DescribeSnapshotsInput{
			OwnerIds: []string{"self"}, // Only our snapshots
		})

		for paginator.HasMorePages() {
			page, err := paginator.NextPage(ctx)
			if err != nil {
				continue
			}

			for _, snapshot := range page.Snapshots {
				resource := make(core.Resource)
				resource["id"] = aws.ToString(snapshot.SnapshotId)
				resource["region"] = region
				resource["arn"] = "arn:aws:ec2:" + region + "::" + "snapshot/" + aws.ToString(snapshot.SnapshotId)

				// Get name from tags
				for _, tag := range snapshot.Tags {
					if aws.ToString(tag.Key) == tagKeyName {
						resource["name"] = aws.ToString(tag.Value)
						break
					}
				}

				resource["description"] = aws.ToString(snapshot.Description)
				resource["volume_id"] = aws.ToString(snapshot.VolumeId)
				resource["volume_size_gb"] = snapshot.VolumeSize
				resource["state"] = string(snapshot.State)
				resource["progress"] = aws.ToString(snapshot.Progress)
				resource["owner_id"] = aws.ToString(snapshot.OwnerId)

				// Encryption
				encrypted := ptr.Deref(snapshot.Encrypted, false)
				resource["encrypted"] = encrypted
				resource["kms_key_id"] = aws.ToString(snapshot.KmsKeyId)
				resource["has_encryption"] = encrypted

				// Start time
				if snapshot.StartTime != nil {
					resource["created_at"] = snapshot.StartTime.String()
				}

				// Check if snapshot is public
				attrResp, err := client.DescribeSnapshotAttribute(ctx, &ec2.DescribeSnapshotAttributeInput{
					SnapshotId: snapshot.SnapshotId,
					Attribute:  types.SnapshotAttributeNameCreateVolumePermission,
				})
				if err == nil {
					isPublic := false
					for _, perm := range attrResp.CreateVolumePermissions {
						if perm.Group == types.PermissionGroupAll {
							isPublic = true
							break
						}
					}
					resource["is_public"] = isPublic
				} else {
					resource["is_public"] = false
				}

				// Tags
				tags := make(map[string]string)
				for _, tag := range snapshot.Tags {
					tags[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
				}
				resource["tags"] = tags

				// Computed
				resource["is_completed"] = snapshot.State == types.SnapshotStateCompleted
				resource["is_pending"] = snapshot.State == types.SnapshotStatePending
				resource["is_error"] = snapshot.State == types.SnapshotStateError

				resources = append(resources, resource)
			}
		}
	}

	return resources, nil
}
