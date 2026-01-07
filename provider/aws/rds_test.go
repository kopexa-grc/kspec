// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: BUSL-1.1

package aws

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"go.uber.org/mock/gomock"

	"github.com/kopexa-grc/kspec/core"
)

func TestRDSInstanceResource_Name(t *testing.T) {
	r := &RDSInstanceResource{}
	if got := r.Name(); got != "aws_rds_instance" {
		t.Errorf("Name() = %v, want aws_rds_instance", got)
	}
}

func TestRDSInstanceResource_Fetch_EmptyInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockRDSAPI(ctrl)

	mock.EXPECT().
		DescribeDBInstances(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&rds.DescribeDBInstancesOutput{DBInstances: []types.DBInstance{}}, nil)

	r := &RDSInstanceResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("Fetch() returned %d resources, want 0", len(resources))
	}
}

func TestRDSInstanceResource_Fetch_SecureInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockRDSAPI(ctrl)

	dbIdentifier := "secure-db"

	mock.EXPECT().
		DescribeDBInstances(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{
				{
					DBInstanceIdentifier:             aws.String(dbIdentifier),
					DBInstanceArn:                    aws.String("arn:aws:rds:us-east-1:123456789012:db:secure-db"),
					Engine:                           aws.String("postgres"),
					EngineVersion:                    aws.String("15.4"),
					DBInstanceClass:                  aws.String("db.t3.medium"),
					DBInstanceStatus:                 aws.String("available"),
					AllocatedStorage:                 aws.Int32(100),
					StorageType:                      aws.String("gp3"),
					AvailabilityZone:                 aws.String("us-east-1a"),
					MultiAZ:                          aws.Bool(true),
					StorageEncrypted:                 aws.Bool(true),
					KmsKeyId:                         aws.String("arn:aws:kms:us-east-1:123456789012:key/12345"),
					PubliclyAccessible:               aws.Bool(false),
					IAMDatabaseAuthenticationEnabled: aws.Bool(true),
					DeletionProtection:               aws.Bool(true),
					AutoMinorVersionUpgrade:          aws.Bool(true),
					BackupRetentionPeriod:            aws.Int32(7),
					PerformanceInsightsEnabled:       aws.Bool(true),
					MonitoringInterval:               aws.Int32(60),
					DBSubnetGroup: &types.DBSubnetGroup{
						VpcId:             aws.String("vpc-12345"),
						DBSubnetGroupName: aws.String("default"),
					},
					VpcSecurityGroups: []types.VpcSecurityGroupMembership{
						{VpcSecurityGroupId: aws.String("sg-12345")},
					},
					Endpoint: &types.Endpoint{
						Address: aws.String("secure-db.cluster.us-east-1.rds.amazonaws.com"),
						Port:    aws.Int32(5432),
					},
				},
			},
		}, nil)

	r := &RDSInstanceResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(resources))
	}

	res := resources[0]
	if res["id"] != dbIdentifier {
		t.Errorf("id = %v, want %v", res["id"], dbIdentifier)
	}
	if res["is_encrypted"] != true {
		t.Errorf("is_encrypted = %v, want true", res["is_encrypted"])
	}
	if res["has_multi_az"] != true {
		t.Errorf("has_multi_az = %v, want true", res["has_multi_az"])
	}
	if res["is_public"] != false {
		t.Errorf("is_public = %v, want false", res["is_public"])
	}
	if res["has_iam_auth"] != true {
		t.Errorf("has_iam_auth = %v, want true", res["has_iam_auth"])
	}
	if res["has_deletion_protection"] != true {
		t.Errorf("has_deletion_protection = %v, want true", res["has_deletion_protection"])
	}
	if res["has_backups"] != true {
		t.Errorf("has_backups = %v, want true", res["has_backups"])
	}
	if res["has_enhanced_monitoring"] != true {
		t.Errorf("has_enhanced_monitoring = %v, want true", res["has_enhanced_monitoring"])
	}
	if res["is_available"] != true {
		t.Errorf("is_available = %v, want true", res["is_available"])
	}
}

func TestRDSInstanceResource_Fetch_InsecureInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockRDSAPI(ctrl)

	mock.EXPECT().
		DescribeDBInstances(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&rds.DescribeDBInstancesOutput{
			DBInstances: []types.DBInstance{
				{
					DBInstanceIdentifier:  aws.String("insecure-db"),
					DBInstanceArn:         aws.String("arn:aws:rds:us-east-1:123456789012:db:insecure-db"),
					Engine:                aws.String("mysql"),
					EngineVersion:         aws.String("8.0"),
					DBInstanceClass:       aws.String("db.t3.small"),
					DBInstanceStatus:      aws.String("available"),
					MultiAZ:               aws.Bool(false),
					StorageEncrypted:      aws.Bool(false),
					PubliclyAccessible:    aws.Bool(true), // Bad!
					DeletionProtection:    aws.Bool(false),
					BackupRetentionPeriod: aws.Int32(0), // No backups!
				},
			},
		}, nil)

	r := &RDSInstanceResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["is_encrypted"] != false {
		t.Errorf("is_encrypted = %v, want false", res["is_encrypted"])
	}
	if res["is_public"] != true {
		t.Errorf("is_public = %v, want true", res["is_public"])
	}
	if res["has_deletion_protection"] != false {
		t.Errorf("has_deletion_protection = %v, want false", res["has_deletion_protection"])
	}
	if res["has_backups"] != false {
		t.Errorf("has_backups = %v, want false", res["has_backups"])
	}
}

func TestRDSClusterResource_Name(t *testing.T) {
	r := &RDSClusterResource{}
	if got := r.Name(); got != "aws_rds_cluster" {
		t.Errorf("Name() = %v, want aws_rds_cluster", got)
	}
}

func TestRDSClusterResource_Fetch_EmptyClusters(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockRDSAPI(ctrl)

	mock.EXPECT().
		DescribeDBClusters(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&rds.DescribeDBClustersOutput{DBClusters: []types.DBCluster{}}, nil)

	r := &RDSClusterResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 0 {
		t.Errorf("Fetch() returned %d resources, want 0", len(resources))
	}
}

func TestRDSClusterResource_Fetch_AuroraCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockRDSAPI(ctrl)

	clusterID := "aurora-cluster"

	mock.EXPECT().
		DescribeDBClusters(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&rds.DescribeDBClustersOutput{
			DBClusters: []types.DBCluster{
				{
					DBClusterIdentifier:              aws.String(clusterID),
					DBClusterArn:                     aws.String("arn:aws:rds:us-east-1:123456789012:cluster:aurora-cluster"),
					Engine:                           aws.String("aurora-postgresql"),
					EngineVersion:                    aws.String("15.4"),
					EngineMode:                       aws.String("provisioned"),
					Status:                           aws.String("available"),
					DatabaseName:                     aws.String("mydb"),
					MultiAZ:                          aws.Bool(true),
					StorageEncrypted:                 aws.Bool(true),
					KmsKeyId:                         aws.String("arn:aws:kms:us-east-1:123456789012:key/12345"),
					IAMDatabaseAuthenticationEnabled: aws.Bool(true),
					DeletionProtection:               aws.Bool(true),
					BackupRetentionPeriod:            aws.Int32(14),
					Endpoint:                         aws.String("aurora-cluster.cluster.us-east-1.rds.amazonaws.com"),
					ReaderEndpoint:                   aws.String("aurora-cluster.cluster-ro.us-east-1.rds.amazonaws.com"),
					Port:                             aws.Int32(5432),
					DBClusterMembers: []types.DBClusterMember{
						{DBInstanceIdentifier: aws.String("aurora-1")},
						{DBInstanceIdentifier: aws.String("aurora-2")},
					},
					VpcSecurityGroups: []types.VpcSecurityGroupMembership{
						{VpcSecurityGroupId: aws.String("sg-aurora")},
					},
				},
			},
		}, nil)

	r := &RDSClusterResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}
	if len(resources) != 1 {
		t.Fatalf("Fetch() returned %d resources, want 1", len(resources))
	}

	res := resources[0]
	if res["id"] != clusterID {
		t.Errorf("id = %v, want %v", res["id"], clusterID)
	}
	if res["is_encrypted"] != true {
		t.Errorf("is_encrypted = %v, want true", res["is_encrypted"])
	}
	if res["has_iam_auth"] != true {
		t.Errorf("has_iam_auth = %v, want true", res["has_iam_auth"])
	}
	if res["has_deletion_protection"] != true {
		t.Errorf("has_deletion_protection = %v, want true", res["has_deletion_protection"])
	}
	if res["member_count"] != 2 {
		t.Errorf("member_count = %v, want 2", res["member_count"])
	}
	if res["is_serverless"] != false {
		t.Errorf("is_serverless = %v, want false", res["is_serverless"])
	}
}

func TestRDSClusterResource_Fetch_ServerlessCluster(t *testing.T) {
	ctrl := gomock.NewController(t)
	mock := NewMockRDSAPI(ctrl)

	mock.EXPECT().
		DescribeDBClusters(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&rds.DescribeDBClustersOutput{
			DBClusters: []types.DBCluster{
				{
					DBClusterIdentifier: aws.String("serverless-cluster"),
					DBClusterArn:        aws.String("arn:aws:rds:us-east-1:123456789012:cluster:serverless-cluster"),
					Engine:              aws.String("aurora-postgresql"),
					EngineMode:          aws.String("serverless"),
					Status:              aws.String("available"),
					StorageEncrypted:    aws.Bool(true),
				},
			},
		}, nil)

	r := &RDSClusterResource{
		client: mock,
		conn:   &Connection{regions: []string{"us-east-1"}},
	}
	resources, err := r.Fetch(context.Background(), core.Asset{})

	if err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	res := resources[0]
	if res["is_serverless"] != true {
		t.Errorf("is_serverless = %v, want true", res["is_serverless"])
	}
}
