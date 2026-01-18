// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

//go:build fast

// Package suites provides policy integration test fixtures for AWS resources.
package suites

import "github.com/kopexa-grc/kspec/testing/suites"

// S3 Bucket Fixtures

// SecureS3Bucket represents a fully compliant S3 bucket configuration.
var SecureS3Bucket = suites.NewFixture().
	Set("id", "secure-bucket-001").
	Set("name", "secure-bucket").
	Set("arn", "arn:aws:s3:::secure-bucket").
	Set("region", "eu-central-1").
	Set("has_versioning", true).
	Set("has_encryption", true).
	Set("has_kms_encryption", true).
	Set("encryption_algorithm", "aws:kms").
	Set("has_logging", true).
	Set("public_access_block_enabled", true).
	Set("block_public_acls", true).
	Set("ignore_public_acls", true).
	Set("block_public_policy", true).
	Set("restrict_public_buckets", true).
	Set("has_public_acl", false).
	Set("has_bucket_policy", true).
	Set("has_public_policy", false).
	Set("has_secure_transport_policy", true).
	Set("has_lifecycle_rules", true).
	Set("object_lock_enabled", false).
	Set("is_public", false).
	Set("allows_http", false)

// InsecureS3Bucket represents a non-compliant S3 bucket with multiple issues.
var InsecureS3Bucket = suites.NewFixture().
	Set("id", "insecure-bucket-001").
	Set("name", "insecure-bucket").
	Set("arn", "arn:aws:s3:::insecure-bucket").
	Set("region", "eu-central-1").
	Set("has_versioning", false).
	Set("has_encryption", false).
	Set("has_kms_encryption", false).
	Set("has_logging", false).
	Set("public_access_block_enabled", false).
	Set("block_public_acls", false).
	Set("ignore_public_acls", false).
	Set("block_public_policy", false).
	Set("restrict_public_buckets", false).
	Set("has_public_acl", true).
	Set("has_bucket_policy", false).
	Set("has_public_policy", true).
	Set("has_secure_transport_policy", false).
	Set("has_lifecycle_rules", false).
	Set("object_lock_enabled", false).
	Set("is_public", true).
	Set("allows_http", true)

// IAM User Fixtures

// SecureIAMUser represents a compliant IAM user configuration.
var SecureIAMUser = suites.NewFixture().
	Set("id", "secure-user-001").
	Set("name", "secure-user").
	Set("arn", "arn:aws:iam::123456789012:user/secure-user").
	Set("has_console_access", true).
	Set("has_mfa", true).
	Set("inline_policy_count", 0).
	Set("has_access_keys", true).
	Set("access_key_age_days", 30).
	Set("last_activity_days", 5).
	Set("active_access_key_count", 1)

// InsecureIAMUser represents a non-compliant IAM user.
var InsecureIAMUser = suites.NewFixture().
	Set("id", "insecure-user-001").
	Set("name", "insecure-user").
	Set("arn", "arn:aws:iam::123456789012:user/insecure-user").
	Set("has_console_access", true).
	Set("has_mfa", false).
	Set("inline_policy_count", 3).
	Set("has_access_keys", true).
	Set("access_key_age_days", 120).
	Set("last_activity_days", 95).
	Set("active_access_key_count", 2)

// IAM Policy Fixtures

// SecureIAMPolicy represents a compliant IAM policy.
var SecureIAMPolicy = suites.NewFixture().
	Set("id", "secure-policy-001").
	Set("name", "secure-policy").
	Set("arn", "arn:aws:iam::123456789012:policy/secure-policy").
	Set("is_admin_policy", false).
	Set("has_wildcard_actions", false).
	Set("has_wildcard_resources", false)

// InsecureIAMPolicy represents an overly permissive IAM policy.
var InsecureIAMPolicy = suites.NewFixture().
	Set("id", "admin-policy-001").
	Set("name", "admin-policy").
	Set("arn", "arn:aws:iam::123456789012:policy/admin-policy").
	Set("is_admin_policy", true).
	Set("has_wildcard_actions", true).
	Set("has_wildcard_resources", true)

// EC2 Instance Fixtures

// SecureEC2Instance represents a compliant EC2 instance.
var SecureEC2Instance = suites.NewFixture().
	Set("id", "i-secure001").
	Set("name", "secure-instance").
	Set("arn", "arn:aws:ec2:eu-central-1:123456789012:instance/i-secure001").
	Set("region", "eu-central-1").
	Set("instance_type", "t3.medium").
	Set("has_imdsv2", true).
	Set("is_public", false).
	Set("has_detailed_monitoring", true).
	Set("state", "running").
	Set("user_data_has_secrets", false)

// InsecureEC2Instance represents a non-compliant EC2 instance.
var InsecureEC2Instance = suites.NewFixture().
	Set("id", "i-insecure001").
	Set("name", "insecure-instance").
	Set("arn", "arn:aws:ec2:eu-central-1:123456789012:instance/i-insecure001").
	Set("region", "eu-central-1").
	Set("instance_type", "t3.medium").
	Set("has_imdsv2", false).
	Set("is_public", true).
	Set("has_detailed_monitoring", false).
	Set("state", "running").
	Set("user_data_has_secrets", true)

// Security Group Fixtures

// SecureSecurityGroup represents a compliant security group.
var SecureSecurityGroup = suites.NewFixture().
	Set("id", "sg-secure001").
	Set("name", "secure-sg").
	Set("description", "Secure security group").
	Set("vpc_id", "vpc-001").
	Set("allows_ssh_from_internet", false).
	Set("allows_rdp_from_internet", false).
	Set("allows_vnc_from_internet", false).
	Set("has_unrestricted_ingress", false).
	Set("is_default", false).
	Set("has_ingress_rules", true).
	Set("has_egress_rules", true)

// InsecureSecurityGroup represents a non-compliant security group.
var InsecureSecurityGroup = suites.NewFixture().
	Set("id", "sg-insecure001").
	Set("name", "insecure-sg").
	Set("description", "Insecure security group").
	Set("vpc_id", "vpc-001").
	Set("allows_ssh_from_internet", true).
	Set("allows_rdp_from_internet", true).
	Set("allows_vnc_from_internet", true).
	Set("has_unrestricted_ingress", true).
	Set("is_default", false).
	Set("has_ingress_rules", true).
	Set("has_egress_rules", true)

// DefaultSecurityGroupOpen represents a default SG with rules (non-compliant).
var DefaultSecurityGroupOpen = suites.NewFixture().
	Set("id", "sg-default001").
	Set("name", "default").
	Set("description", "Default security group").
	Set("vpc_id", "vpc-001").
	Set("allows_ssh_from_internet", false).
	Set("allows_rdp_from_internet", false).
	Set("allows_vnc_from_internet", false).
	Set("has_unrestricted_ingress", false).
	Set("is_default", true).
	Set("has_ingress_rules", true).
	Set("has_egress_rules", true)

// DefaultSecurityGroupClosed represents a properly closed default SG (compliant).
var DefaultSecurityGroupClosed = suites.NewFixture().
	Set("id", "sg-default002").
	Set("name", "default").
	Set("description", "Default security group").
	Set("vpc_id", "vpc-001").
	Set("allows_ssh_from_internet", false).
	Set("allows_rdp_from_internet", false).
	Set("allows_vnc_from_internet", false).
	Set("has_unrestricted_ingress", false).
	Set("is_default", true).
	Set("has_ingress_rules", false).
	Set("has_egress_rules", false)

// EBS Volume Fixtures

// SecureEBSVolume represents a compliant EBS volume.
var SecureEBSVolume = suites.NewFixture().
	Set("id", "vol-secure001").
	Set("arn", "arn:aws:ec2:eu-central-1:123456789012:volume/vol-secure001").
	Set("is_encrypted", true).
	Set("kms_key_id", "arn:aws:kms:eu-central-1:123456789012:key/abc123").
	Set("is_attached", true)

// InsecureEBSVolume represents a non-compliant EBS volume.
var InsecureEBSVolume = suites.NewFixture().
	Set("id", "vol-insecure001").
	Set("arn", "arn:aws:ec2:eu-central-1:123456789012:volume/vol-insecure001").
	Set("is_encrypted", false).
	Set("is_attached", true)

// OrphanedEBSVolume represents an unattached EBS volume.
var OrphanedEBSVolume = suites.NewFixture().
	Set("id", "vol-orphan001").
	Set("arn", "arn:aws:ec2:eu-central-1:123456789012:volume/vol-orphan001").
	Set("is_encrypted", true).
	Set("is_attached", false)

// VPC Fixtures

// SecureVPC represents a compliant VPC.
var SecureVPC = suites.NewFixture().
	Set("id", "vpc-secure001").
	Set("name", "secure-vpc").
	Set("arn", "arn:aws:ec2:eu-central-1:123456789012:vpc/vpc-secure001").
	Set("is_default", false).
	Set("has_flow_logs", true).
	Set("endpoint_count", 3).
	Set("block_public_access_enabled", true)

// InsecureVPC represents a non-compliant VPC (default VPC without flow logs).
var InsecureVPC = suites.NewFixture().
	Set("id", "vpc-insecure001").
	Set("name", "default-vpc").
	Set("arn", "arn:aws:ec2:eu-central-1:123456789012:vpc/vpc-insecure001").
	Set("is_default", true).
	Set("has_flow_logs", false).
	Set("endpoint_count", 0).
	Set("block_public_access_enabled", false)

// RDS Instance Fixtures

// SecureRDSInstance represents a compliant RDS instance.
var SecureRDSInstance = suites.NewFixture().
	Set("id", "rds-secure001").
	Set("name", "secure-db").
	Set("arn", "arn:aws:rds:eu-central-1:123456789012:db:secure-db").
	Set("is_encrypted", true).
	Set("is_public", false).
	Set("has_multi_az", true).
	Set("has_deletion_protection", true).
	Set("has_iam_auth", true).
	Set("has_backups", true).
	Set("backup_retention_period", 7).
	Set("has_auto_minor_upgrade", true).
	Set("has_pending_maintenance", false)

// InsecureRDSInstance represents a non-compliant RDS instance.
var InsecureRDSInstance = suites.NewFixture().
	Set("id", "rds-insecure001").
	Set("name", "insecure-db").
	Set("arn", "arn:aws:rds:eu-central-1:123456789012:db:insecure-db").
	Set("is_encrypted", false).
	Set("is_public", true).
	Set("has_multi_az", false).
	Set("has_deletion_protection", false).
	Set("has_iam_auth", false).
	Set("has_backups", false).
	Set("backup_retention_period", 0).
	Set("has_auto_minor_upgrade", false).
	Set("has_pending_maintenance", true)

// CloudTrail Fixtures

// SecureCloudTrail represents a compliant CloudTrail trail.
var SecureCloudTrail = suites.NewFixture().
	Set("id", "cloudtrail-secure001").
	Set("name", "secure-trail").
	Set("arn", "arn:aws:cloudtrail:eu-central-1:123456789012:trail/secure-trail").
	Set("is_logging", true).
	Set("is_multi_region_trail", true).
	Set("log_file_validation_enabled", true).
	Set("has_kms_encryption", true).
	Set("has_cloudwatch_logs", true)

// InsecureCloudTrail represents a non-compliant CloudTrail trail.
var InsecureCloudTrail = suites.NewFixture().
	Set("id", "cloudtrail-insecure001").
	Set("name", "insecure-trail").
	Set("arn", "arn:aws:cloudtrail:eu-central-1:123456789012:trail/insecure-trail").
	Set("is_logging", false).
	Set("is_multi_region_trail", false).
	Set("log_file_validation_enabled", false).
	Set("has_kms_encryption", false).
	Set("has_cloudwatch_logs", false)

// KMS Key Fixtures

// SecureKMSKey represents a compliant KMS key.
var SecureKMSKey = suites.NewFixture().
	Set("id", "kms-secure001").
	Set("key_id", "abc123-def456").
	Set("arn", "arn:aws:kms:eu-central-1:123456789012:key/abc123-def456").
	Set("is_aws_managed", false).
	Set("key_rotation_enabled", true).
	Set("has_public_access", false)

// InsecureKMSKey represents a non-compliant KMS key.
var InsecureKMSKey = suites.NewFixture().
	Set("id", "kms-insecure001").
	Set("key_id", "xyz789").
	Set("arn", "arn:aws:kms:eu-central-1:123456789012:key/xyz789").
	Set("is_aws_managed", false).
	Set("key_rotation_enabled", false).
	Set("has_public_access", true)

// Lambda Function Fixtures

// SecureLambdaFunction represents a compliant Lambda function.
var SecureLambdaFunction = suites.NewFixture().
	Set("id", "lambda-secure001").
	Set("name", "secure-function").
	Set("arn", "arn:aws:lambda:eu-central-1:123456789012:function:secure-function").
	Set("runtime", "python3.12").
	Set("in_vpc", true).
	Set("function_url_is_public", false).
	Set("has_deprecated_runtime", false).
	Set("has_dead_letter_config", true).
	Set("reserved_concurrent_executions", 100)

// InsecureLambdaFunction represents a non-compliant Lambda function.
var InsecureLambdaFunction = suites.NewFixture().
	Set("id", "lambda-insecure001").
	Set("name", "insecure-function").
	Set("arn", "arn:aws:lambda:eu-central-1:123456789012:function:insecure-function").
	Set("runtime", "python2.7").
	Set("in_vpc", false).
	Set("function_url_is_public", true).
	Set("has_deprecated_runtime", true).
	Set("has_dead_letter_config", false).
	Set("reserved_concurrent_executions", 0)

// EKS Cluster Fixtures

// SecureEKSCluster represents a compliant EKS cluster.
var SecureEKSCluster = suites.NewFixture().
	Set("id", "eks-secure001").
	Set("name", "secure-cluster").
	Set("arn", "arn:aws:eks:eu-central-1:123456789012:cluster/secure-cluster").
	Set("has_secrets_encryption", true).
	Set("has_logging", true).
	Set("has_unrestricted_public_access", false)

// InsecureEKSCluster represents a non-compliant EKS cluster.
var InsecureEKSCluster = suites.NewFixture().
	Set("id", "eks-insecure001").
	Set("name", "insecure-cluster").
	Set("arn", "arn:aws:eks:eu-central-1:123456789012:cluster/insecure-cluster").
	Set("has_secrets_encryption", false).
	Set("has_logging", false).
	Set("has_unrestricted_public_access", true)

// IAM Password Policy Fixtures

// SecurePasswordPolicy represents a compliant password policy.
var SecurePasswordPolicy = suites.NewFixture().
	Set("id", "password-policy").
	Set("minimum_password_length", 14).
	Set("require_uppercase", true).
	Set("require_lowercase", true).
	Set("require_numbers", true).
	Set("require_symbols", true).
	Set("max_password_age", 90).
	Set("password_reuse_prevention", 24)

// InsecurePasswordPolicy represents a non-compliant password policy.
var InsecurePasswordPolicy = suites.NewFixture().
	Set("id", "password-policy").
	Set("minimum_password_length", 8).
	Set("require_uppercase", false).
	Set("require_lowercase", true).
	Set("require_numbers", false).
	Set("require_symbols", false).
	Set("max_password_age", 365).
	Set("password_reuse_prevention", 0)

// IAM Group Fixtures

// IAMGroupWithUsers represents a compliant IAM group with users.
var IAMGroupWithUsers = suites.NewFixture().
	Set("id", "group-001").
	Set("name", "developers").
	Set("arn", "arn:aws:iam::123456789012:group/developers").
	Set("user_count", 5)

// IAMGroupEmpty represents an empty IAM group (non-compliant).
var IAMGroupEmpty = suites.NewFixture().
	Set("id", "group-002").
	Set("name", "empty-group").
	Set("arn", "arn:aws:iam::123456789012:group/empty-group").
	Set("user_count", 0)

// S3 Account Fixtures

// SecureS3Account represents a compliant S3 account-level configuration.
var SecureS3Account = suites.NewFixture().
	Set("id", "123456789012").
	Set("block_public_acls", true).
	Set("ignore_public_acls", true).
	Set("block_public_policy", true).
	Set("restrict_public_buckets", true)

// InsecureS3Account represents a non-compliant S3 account-level configuration.
var InsecureS3Account = suites.NewFixture().
	Set("id", "123456789012").
	Set("block_public_acls", false).
	Set("ignore_public_acls", false).
	Set("block_public_policy", false).
	Set("restrict_public_buckets", false)

// EBS Snapshot Fixtures

// SecureEBSSnapshot represents a compliant EBS snapshot.
var SecureEBSSnapshot = suites.NewFixture().
	Set("id", "snap-secure001").
	Set("arn", "arn:aws:ec2:eu-central-1:123456789012:snapshot/snap-secure001").
	Set("is_public", false).
	Set("is_encrypted", true)

// InsecureEBSSnapshot represents a non-compliant EBS snapshot.
var InsecureEBSSnapshot = suites.NewFixture().
	Set("id", "snap-insecure001").
	Set("arn", "arn:aws:ec2:eu-central-1:123456789012:snapshot/snap-insecure001").
	Set("is_public", true).
	Set("is_encrypted", false)

// EC2 EBS Settings Fixtures

// SecureEBSSettings represents compliant EC2 EBS default settings.
var SecureEBSSettings = suites.NewFixture().
	Set("id", "ebs-settings").
	Set("encryption_by_default", true)

// InsecureEBSSettings represents non-compliant EC2 EBS default settings.
var InsecureEBSSettings = suites.NewFixture().
	Set("id", "ebs-settings").
	Set("encryption_by_default", false)

// RDS Cluster Fixtures

// SecureRDSCluster represents a compliant RDS Aurora cluster.
var SecureRDSCluster = suites.NewFixture().
	Set("id", "rds-cluster-secure001").
	Set("name", "secure-cluster").
	Set("arn", "arn:aws:rds:eu-central-1:123456789012:cluster:secure-cluster").
	Set("ssl_enforced", true).
	Set("is_encrypted", true)

// InsecureRDSCluster represents a non-compliant RDS Aurora cluster.
var InsecureRDSCluster = suites.NewFixture().
	Set("id", "rds-cluster-insecure001").
	Set("name", "insecure-cluster").
	Set("arn", "arn:aws:rds:eu-central-1:123456789012:cluster:insecure-cluster").
	Set("ssl_enforced", false).
	Set("is_encrypted", false)

// API Gateway Stage Fixtures

// SecureAPIGatewayStage represents a compliant API Gateway stage.
var SecureAPIGatewayStage = suites.NewFixture().
	Set("id", "stage-secure001").
	Set("name", "prod").
	Set("has_access_logging", true).
	Set("has_waf", true).
	Set("cache_enabled", true).
	Set("cache_encrypted", true).
	Set("minimum_tls_version", "TLS_1_2").
	Set("xray_tracing_enabled", true)

// InsecureAPIGatewayStage represents a non-compliant API Gateway stage.
var InsecureAPIGatewayStage = suites.NewFixture().
	Set("id", "stage-insecure001").
	Set("name", "dev").
	Set("has_access_logging", false).
	Set("has_waf", false).
	Set("cache_enabled", true).
	Set("cache_encrypted", false).
	Set("minimum_tls_version", "TLS_1_0").
	Set("xray_tracing_enabled", false)

// API Gateway Method Fixtures

// SecureAPIGatewayMethod represents a compliant API Gateway method.
var SecureAPIGatewayMethod = suites.NewFixture().
	Set("id", "method-secure001").
	Set("http_method", "POST").
	Set("authorization_type", "AWS_IAM")

// InsecureAPIGatewayMethod represents a non-compliant API Gateway method.
var InsecureAPIGatewayMethod = suites.NewFixture().
	Set("id", "method-insecure001").
	Set("http_method", "GET").
	Set("authorization_type", "NONE")

// Redshift Cluster Fixtures

// SecureRedshiftCluster represents a compliant Redshift cluster.
var SecureRedshiftCluster = suites.NewFixture().
	Set("id", "redshift-secure001").
	Set("name", "secure-cluster").
	Set("arn", "arn:aws:redshift:eu-central-1:123456789012:cluster:secure-cluster").
	Set("is_public", false).
	Set("is_encrypted", true).
	Set("audit_logging_enabled", true)

// InsecureRedshiftCluster represents a non-compliant Redshift cluster.
var InsecureRedshiftCluster = suites.NewFixture().
	Set("id", "redshift-insecure001").
	Set("name", "insecure-cluster").
	Set("arn", "arn:aws:redshift:eu-central-1:123456789012:cluster:insecure-cluster").
	Set("is_public", true).
	Set("is_encrypted", false).
	Set("audit_logging_enabled", false)

// EFS Filesystem Fixtures

// SecureEFSFilesystem represents a compliant EFS filesystem.
var SecureEFSFilesystem = suites.NewFixture().
	Set("id", "fs-secure001").
	Set("name", "secure-fs").
	Set("arn", "arn:aws:elasticfilesystem:eu-central-1:123456789012:file-system/fs-secure001").
	Set("is_encrypted", true).
	Set("encryption_in_transit_enforced", true)

// InsecureEFSFilesystem represents a non-compliant EFS filesystem.
var InsecureEFSFilesystem = suites.NewFixture().
	Set("id", "fs-insecure001").
	Set("name", "insecure-fs").
	Set("arn", "arn:aws:elasticfilesystem:eu-central-1:123456789012:file-system/fs-insecure001").
	Set("is_encrypted", false).
	Set("encryption_in_transit_enforced", false)

// OpenSearch Domain Fixtures

// SecureOpenSearchDomain represents a compliant OpenSearch domain.
var SecureOpenSearchDomain = suites.NewFixture().
	Set("id", "opensearch-secure001").
	Set("name", "secure-domain").
	Set("arn", "arn:aws:es:eu-central-1:123456789012:domain/secure-domain").
	Set("encryption_at_rest_enabled", true).
	Set("node_to_node_encryption_enabled", true).
	Set("is_public", false)

// InsecureOpenSearchDomain represents a non-compliant OpenSearch domain.
var InsecureOpenSearchDomain = suites.NewFixture().
	Set("id", "opensearch-insecure001").
	Set("name", "insecure-domain").
	Set("arn", "arn:aws:es:eu-central-1:123456789012:domain/insecure-domain").
	Set("encryption_at_rest_enabled", false).
	Set("node_to_node_encryption_enabled", false).
	Set("is_public", true)

// SageMaker Notebook Fixtures

// SecureSageMakerNotebook represents a compliant SageMaker notebook.
var SecureSageMakerNotebook = suites.NewFixture().
	Set("id", "notebook-secure001").
	Set("name", "secure-notebook").
	Set("arn", "arn:aws:sagemaker:eu-central-1:123456789012:notebook-instance/secure-notebook").
	Set("has_kms_encryption", true).
	Set("direct_internet_access", false)

// InsecureSageMakerNotebook represents a non-compliant SageMaker notebook.
var InsecureSageMakerNotebook = suites.NewFixture().
	Set("id", "notebook-insecure001").
	Set("name", "insecure-notebook").
	Set("arn", "arn:aws:sagemaker:eu-central-1:123456789012:notebook-instance/insecure-notebook").
	Set("has_kms_encryption", false).
	Set("direct_internet_access", true)
