// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package aws

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"go.uber.org/mock/gomock"

	"github.com/kopexa-grc/kspec/core"
)

func TestEC2InstanceResource_Name(t *testing.T) {
	r := &EC2InstanceResource{}
	if got := r.Name(); got != "aws_ec2_instance" {
		t.Errorf("Name() = %v, want aws_ec2_instance", got)
	}
}

func TestEC2InstanceResource_Fetch_EmptyInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEC2API(ctrl)

	// Paginator passes optional functions, match any args
	mock.EXPECT().
		DescribeInstances(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeInstancesOutput{Reservations: []types.Reservation{}}, nil)

	r := &EC2InstanceResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}, accountID: "123456789012"},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("Fetch() returned %d resources, want 0", len(resources))
	}
}

func TestEC2InstanceResource_Fetch_SingleInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEC2API(ctrl)

	instanceID := "i-1234567890abcdef0"
	launchTime := time.Now().Add(-24 * time.Hour)

	mock.EXPECT().
		DescribeInstances(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:       aws.String(instanceID),
							InstanceType:     types.InstanceTypeT3Micro,
							State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
							Architecture:     types.ArchitectureValuesX8664,
							VpcId:            aws.String("vpc-12345"),
							SubnetId:         aws.String("subnet-12345"),
							PrivateIpAddress: aws.String("10.0.1.100"),
							LaunchTime:       &launchTime,
							Placement:        &types.Placement{AvailabilityZone: aws.String("us-east-1a")},
							Tags: []types.Tag{
								{Key: aws.String("Name"), Value: aws.String("test-instance")},
							},
							SecurityGroups: []types.GroupIdentifier{
								{GroupId: aws.String("sg-12345")},
							},
							MetadataOptions: &types.InstanceMetadataOptionsResponse{
								HttpTokens:   types.HttpTokensStateRequired,
								HttpEndpoint: types.InstanceMetadataEndpointStateEnabled,
							},
							Monitoring: &types.Monitoring{State: types.MonitoringStateDisabled},
						},
					},
				},
			},
		}, nil)

	r := &EC2InstanceResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}, accountID: "123456789012"},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(resources))
	}

	res := resources[0]
	if res["id"] != instanceID {
		t.Errorf("id = %v, want %v", res["id"], instanceID)
	}
	if res["name"] != "test-instance" {
		t.Errorf("name = %v, want test-instance", res["name"])
	}
	if res["is_running"] != true {
		t.Errorf("is_running = %v, want true", res["is_running"])
	}
	if res["has_imdsv2"] != true {
		t.Errorf("has_imdsv2 = %v, want true", res["has_imdsv2"])
	}
	if res["is_public"] != false {
		t.Errorf("is_public = %v, want false (no public IP)", res["is_public"])
	}
}

func TestEC2InstanceResource_Fetch_PublicInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEC2API(ctrl)

	mock.EXPECT().
		DescribeInstances(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeInstancesOutput{
			Reservations: []types.Reservation{
				{
					Instances: []types.Instance{
						{
							InstanceId:       aws.String("i-public"),
							InstanceType:     types.InstanceTypeT3Small,
							State:            &types.InstanceState{Name: types.InstanceStateNameRunning},
							PublicIpAddress:  aws.String("203.0.113.50"),
							PrivateIpAddress: aws.String("10.0.1.101"),
							Placement:        &types.Placement{AvailabilityZone: aws.String("us-east-1a")},
						},
					},
				},
			},
		}, nil)

	r := &EC2InstanceResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}, accountID: "123456789012"},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["is_public"] != true {
		t.Errorf("is_public = %v, want true", res["is_public"])
	}
	if res["public_ip"] != "203.0.113.50" {
		t.Errorf("public_ip = %v, want 203.0.113.50", res["public_ip"])
	}
}

func TestEC2SecurityGroupResource_Name(t *testing.T) {
	r := &EC2SecurityGroupResource{}
	if got := r.Name(); got != "aws_ec2_securitygroup" {
		t.Errorf("Name() = %v, want aws_ec2_securitygroup", got)
	}
}

func TestEC2SecurityGroupResource_Fetch_EmptySecurityGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEC2API(ctrl)

	mock.EXPECT().
		DescribeSecurityGroups(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeSecurityGroupsOutput{SecurityGroups: []types.SecurityGroup{}}, nil)

	r := &EC2SecurityGroupResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}, accountID: "123456789012"},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("Fetch() returned %d resources, want 0", len(resources))
	}
}

func TestEC2SecurityGroupResource_Fetch_SecureSecurityGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEC2API(ctrl)

	sgID := "sg-secure123"

	mock.EXPECT().
		DescribeSecurityGroups(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{
				{
					GroupId:     aws.String(sgID),
					GroupName:   aws.String("secure-sg"),
					Description: aws.String("Secure security group"),
					VpcId:       aws.String("vpc-12345"),
					OwnerId:     aws.String("123456789012"),
					IpPermissions: []types.IpPermission{
						{
							IpProtocol: aws.String("tcp"),
							FromPort:   aws.Int32(443),
							ToPort:     aws.Int32(443),
							IpRanges: []types.IpRange{
								{CidrIp: aws.String("10.0.0.0/8")},
							},
						},
					},
					IpPermissionsEgress: []types.IpPermission{
						{
							IpProtocol: aws.String("tcp"),
							FromPort:   aws.Int32(443),
							ToPort:     aws.Int32(443),
							IpRanges: []types.IpRange{
								{CidrIp: aws.String("0.0.0.0/0")},
							},
						},
					},
				},
			},
		}, nil)

	mock.EXPECT().
		DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeNetworkInterfacesOutput{NetworkInterfaces: []types.NetworkInterface{}}, nil)

	r := &EC2SecurityGroupResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}, accountID: "123456789012"},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(resources))
	}

	res := resources[0]
	if res["id"] != sgID {
		t.Errorf("id = %v, want %v", res["id"], sgID)
	}
	if res["has_unrestricted_ingress"] != false {
		t.Errorf("has_unrestricted_ingress = %v, want false", res["has_unrestricted_ingress"])
	}
	if res["allows_ssh_from_internet"] != false {
		t.Errorf("allows_ssh_from_internet = %v, want false", res["allows_ssh_from_internet"])
	}
}

func TestEC2SecurityGroupResource_Fetch_OpenSSH(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEC2API(ctrl)

	mock.EXPECT().
		DescribeSecurityGroups(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeSecurityGroupsOutput{
			SecurityGroups: []types.SecurityGroup{
				{
					GroupId:     aws.String("sg-open-ssh"),
					GroupName:   aws.String("open-ssh-sg"),
					Description: aws.String("Insecure - SSH open to world"),
					VpcId:       aws.String("vpc-12345"),
					OwnerId:     aws.String("123456789012"),
					IpPermissions: []types.IpPermission{
						{
							IpProtocol: aws.String("tcp"),
							FromPort:   aws.Int32(22),
							ToPort:     aws.Int32(22),
							IpRanges: []types.IpRange{
								{CidrIp: aws.String("0.0.0.0/0")}, // Open to world!
							},
						},
					},
				},
			},
		}, nil)

	mock.EXPECT().
		DescribeNetworkInterfaces(gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeNetworkInterfacesOutput{NetworkInterfaces: []types.NetworkInterface{}}, nil)

	r := &EC2SecurityGroupResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}, accountID: "123456789012"},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["has_unrestricted_ingress"] != true {
		t.Errorf("has_unrestricted_ingress = %v, want true", res["has_unrestricted_ingress"])
	}
	if res["allows_ssh_from_internet"] != true {
		t.Errorf("allows_ssh_from_internet = %v, want true", res["allows_ssh_from_internet"])
	}
}

func TestEC2VolumeResource_Name(t *testing.T) {
	r := &EC2VolumeResource{}
	if got := r.Name(); got != "aws_ec2_volume" {
		t.Errorf("Name() = %v, want aws_ec2_volume", got)
	}
}

func TestEC2VolumeResource_Fetch_EncryptedVolume(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEC2API(ctrl)

	volumeID := "vol-encrypted"

	mock.EXPECT().
		DescribeVolumes(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeVolumesOutput{
			Volumes: []types.Volume{
				{
					VolumeId:         aws.String(volumeID),
					AvailabilityZone: aws.String("us-east-1a"),
					Size:             aws.Int32(100),
					VolumeType:       types.VolumeTypeGp3,
					State:            types.VolumeStateInUse,
					Encrypted:        aws.Bool(true),
					KmsKeyId:         aws.String("arn:aws:kms:us-east-1:123456789012:key/12345"),
					Attachments: []types.VolumeAttachment{
						{
							InstanceId: aws.String("i-12345"),
							Device:     aws.String("/dev/xvda"),
						},
					},
				},
			},
		}, nil)

	r := &EC2VolumeResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}, accountID: "123456789012"},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(resources))
	}

	res := resources[0]
	if res["id"] != volumeID {
		t.Errorf("id = %v, want %v", res["id"], volumeID)
	}
	if res["has_encryption"] != true {
		t.Errorf("has_encryption = %v, want true", res["has_encryption"])
	}
	if res["is_attached"] != true {
		t.Errorf("is_attached = %v, want true", res["is_attached"])
	}
	if res["is_root_volume"] != true {
		t.Errorf("is_root_volume = %v, want true", res["is_root_volume"])
	}
}

func TestEC2SnapshotResource_Name(t *testing.T) {
	r := &EC2SnapshotResource{}
	if got := r.Name(); got != "aws_ec2_snapshot" {
		t.Errorf("Name() = %v, want aws_ec2_snapshot", got)
	}
}

func TestEC2KeyPairResource_Name(t *testing.T) {
	r := &EC2KeyPairResource{}
	if got := r.Name(); got != "aws_ec2_keypair" {
		t.Errorf("Name() = %v, want aws_ec2_keypair", got)
	}
}

func TestEC2KeyPairResource_Fetch_KeyPairs(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockEC2API(ctrl)

	mock.EXPECT().
		DescribeKeyPairs(gomock.Any(), gomock.Any()).
		Return(&ec2.DescribeKeyPairsOutput{
			KeyPairs: []types.KeyPairInfo{
				{
					KeyPairId:      aws.String("key-12345"),
					KeyName:        aws.String("my-key"),
					KeyFingerprint: aws.String("ab:cd:ef:12:34"),
					KeyType:        types.KeyTypeEd25519,
				},
				{
					KeyPairId:      aws.String("key-67890"),
					KeyName:        aws.String("legacy-key"),
					KeyFingerprint: aws.String("12:34:56:78:90"),
					KeyType:        types.KeyTypeRsa,
				},
			},
		}, nil)

	r := &EC2KeyPairResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}, accountID: "123456789012"},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 2 {
		t.Fatalf("Fetch() returned %d resources, want 2", len(resources))
	}

	// ED25519 key
	res0 := resources[0]
	if res0["is_modern_key_type"] != true {
		t.Errorf("is_modern_key_type = %v, want true for ed25519", res0["is_modern_key_type"])
	}

	// RSA key
	res1 := resources[1]
	if res1["is_modern_key_type"] != false {
		t.Errorf("is_modern_key_type = %v, want false for RSA", res1["is_modern_key_type"])
	}
}
