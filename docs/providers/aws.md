# AWS Provider

The AWS provider scans your Amazon Web Services accounts and resources for security compliance, covering compute, storage, networking, identity, and security services.

## Overview

Use the AWS provider to validate:
- IAM users, roles, groups, and policies
- S3 bucket security (encryption, public access, versioning)
- EC2 instances, security groups, and key pairs
- VPC configurations (flow logs, endpoints, subnets)
- RDS databases and clusters
- Lambda functions
- EKS and ECS clusters
- CloudTrail logging
- KMS encryption keys
- GuardDuty and Security Hub
- And many more AWS services

## Quick Start

```bash
# Set your AWS credentials
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"

# Scan with all AWS policies
kspec scan aws account -d policies

# Scan with specific policy
kspec scan aws account -f policies/aws-security.yml
```

## Prerequisites

- AWS account
- IAM credentials with **ReadOnly** permissions

## Authentication

### IAM Credentials

The AWS provider supports multiple authentication methods following the AWS SDK credential chain:

**Environment Variables (Recommended):**
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_REGION="us-east-1"
kspec scan aws account -f policy.yml
```

**AWS Profile:**
```bash
export AWS_PROFILE="your-profile"
kspec scan aws account -f policy.yml
```

**Temporary Credentials (STS):**
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_SESSION_TOKEN="your-session-token"
kspec scan aws account -f policy.yml
```

**Cross-Account Access (Assume Role):**
```bash
export AWS_ACCESS_KEY_ID="your-access-key"
export AWS_SECRET_ACCESS_KEY="your-secret-key"
export AWS_ROLE_ARN="arn:aws:iam::123456789012:role/SecurityAuditRole"
export AWS_EXTERNAL_ID="optional-external-id"
kspec scan aws account -f policy.yml
```

### Multi-Region Scanning

```bash
# Scan specific regions
export AWS_REGIONS="us-east-1,us-west-2,eu-west-1"
kspec scan aws account -f policy.yml
```

### Required IAM Permissions

For comprehensive security scanning, use the AWS managed policy `SecurityAudit` or create a custom policy with the following permissions:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "iam:Get*",
        "iam:List*",
        "s3:GetBucket*",
        "s3:GetEncryptionConfiguration",
        "s3:ListAllMyBuckets",
        "ec2:Describe*",
        "rds:Describe*",
        "lambda:List*",
        "lambda:GetFunction*",
        "eks:Describe*",
        "eks:List*",
        "ecs:Describe*",
        "ecs:List*",
        "cloudtrail:Describe*",
        "cloudtrail:GetTrailStatus",
        "kms:Describe*",
        "kms:List*",
        "guardduty:Get*",
        "guardduty:List*",
        "securityhub:Get*",
        "securityhub:Describe*",
        "sts:GetCallerIdentity",
        "organizations:Describe*",
        "organizations:List*"
      ],
      "Resource": "*"
    }
  ]
}
```

---

## Resources

The AWS provider discovers the following resources:

### Account & Organization

| Resource | Description |
|----------|-------------|
| `aws_account` | AWS account information and settings |
| `aws_organization` | AWS Organizations configuration |

### Identity & Access Management (IAM)

| Resource | Description |
|----------|-------------|
| `aws_iam_user` | IAM users with MFA status, access keys |
| `aws_iam_role` | IAM roles with trust policies |
| `aws_iam_group` | IAM groups with attached policies |
| `aws_iam_policy` | IAM policies (customer managed) |

### Storage

| Resource | Description |
|----------|-------------|
| `aws_s3_bucket` | S3 buckets with encryption, versioning, public access settings |

### Compute

| Resource | Description |
|----------|-------------|
| `aws_ec2_instance` | EC2 instances |
| `aws_ec2_volume` | EBS volumes |
| `aws_ec2_snapshot` | EBS snapshots |
| `aws_lambda_function` | Lambda functions |
| `aws_autoscaling_group` | Auto Scaling groups |
| `aws_autoscaling_launch_configuration` | Launch configurations |

### Security

| Resource | Description |
|----------|-------------|
| `aws_ec2_security_group` | Security groups with inbound/outbound rules |
| `aws_ec2_key_pair` | SSH key pairs |

### Networking

| Resource | Description |
|----------|-------------|
| `aws_vpc` | Virtual Private Clouds |
| `aws_vpc_subnet` | VPC subnets |
| `aws_vpc_endpoint` | VPC endpoints |
| `aws_vpc_flow_log` | VPC flow logs |

### Database

| Resource | Description |
|----------|-------------|
| `aws_rds_instance` | RDS database instances |
| `aws_rds_cluster` | RDS Aurora clusters |
| `aws_dynamodb_table` | DynamoDB tables |
| `aws_elasticache_cluster` | ElastiCache clusters |
| `aws_elasticache_replication_group` | ElastiCache replication groups |

### Containers

| Resource | Description |
|----------|-------------|
| `aws_eks_cluster` | EKS Kubernetes clusters |
| `aws_eks_nodegroup` | EKS node groups |
| `aws_ecs_cluster` | ECS clusters |
| `aws_ecs_service` | ECS services |
| `aws_ecr_repository` | ECR container repositories |
| `aws_ecr_image` | ECR container images |

### Load Balancing

| Resource | Description |
|----------|-------------|
| `aws_elb_load_balancer` | Application/Network Load Balancers |
| `aws_elb_target_group` | Load balancer target groups |
| `aws_elb_listener` | Load balancer listeners |

### Security Services

| Resource | Description |
|----------|-------------|
| `aws_cloudtrail` | CloudTrail trails |
| `aws_kms_key` | KMS encryption keys |
| `aws_guardduty_detector` | GuardDuty detectors |
| `aws_guardduty_finding` | GuardDuty findings |
| `aws_securityhub` | Security Hub configuration |
| `aws_securityhub_finding` | Security Hub findings |
| `aws_securityhub_standard` | Security Hub standards |
| `aws_waf_web_acl` | WAF Web ACLs |
| `aws_waf_ip_set` | WAF IP sets |
| `aws_waf_rule_group` | WAF rule groups |
| `aws_acm_certificate` | ACM SSL/TLS certificates |

### Configuration & Compliance

| Resource | Description |
|----------|-------------|
| `aws_config_recorder` | AWS Config recorders |
| `aws_config_rule` | AWS Config rules |
| `aws_config_delivery_channel` | AWS Config delivery channels |
| `aws_config_conformance_pack` | AWS Config conformance packs |

### Monitoring

| Resource | Description |
|----------|-------------|
| `aws_cloudwatch_log_group` | CloudWatch log groups |
| `aws_cloudwatch_alarm` | CloudWatch alarms |
| `aws_cloudwatch_metric_stream` | CloudWatch metric streams |

### Messaging

| Resource | Description |
|----------|-------------|
| `aws_sns_topic` | SNS topics |
| `aws_sns_subscription` | SNS subscriptions |
| `aws_sqs_queue` | SQS queues |

### Secrets & Parameters

| Resource | Description |
|----------|-------------|
| `aws_secretsmanager_secret` | Secrets Manager secrets |
| `aws_ssm_parameter` | SSM Parameter Store parameters |
| `aws_ssm_document` | SSM documents |
| `aws_ssm_patch_baseline` | SSM patch baselines |
| `aws_ssm_instance` | SSM managed instances |

### Content Delivery

| Resource | Description |
|----------|-------------|
| `aws_cloudfront_distribution` | CloudFront distributions |

### API Gateway

| Resource | Description |
|----------|-------------|
| `aws_apigateway_rest_api` | REST APIs |
| `aws_apigateway_stage` | API stages |
| `aws_apigateway_v2_api` | HTTP/WebSocket APIs |

---

## Example Policies

### S3 Bucket Encryption

```yaml
policies:
  - uid: aws-s3-security
    name: AWS S3 Security
    version: 1.0.0
    require:
      - provider: aws
    groups:
      - title: S3 Bucket Security
        checks:
          - uid: s3-encryption-enabled

queries:
  - uid: s3-encryption-enabled
    title: Ensure S3 buckets have encryption enabled
    resource: aws_s3_bucket
    impact: 90
    query: |
      resource.has_encryption == true
    docs:
      desc: S3 buckets should have server-side encryption enabled to protect data at rest.
      remediation: Enable default encryption on the S3 bucket using SSE-S3 or SSE-KMS.
```

### CloudTrail Logging

```yaml
queries:
  - uid: cloudtrail-enabled
    title: Ensure CloudTrail is enabled and logging
    resource: aws_cloudtrail
    impact: 95
    query: |
      resource.is_logging == true &&
      resource.is_multi_region_trail == true &&
      resource.log_file_validation_enabled == true
    docs:
      desc: CloudTrail should be enabled with multi-region logging and log file validation.
      remediation: Enable CloudTrail with multi-region support and log file validation.
```

### IAM User MFA

```yaml
queries:
  - uid: iam-user-mfa
    title: Ensure IAM users have MFA enabled
    resource: aws_iam_user
    impact: 90
    query: |
      resource.has_console_password == false || resource.mfa_enabled == true
    docs:
      desc: IAM users with console access should have MFA enabled.
      remediation: Enable MFA for all IAM users with console access.
```

---

## Troubleshooting

### Authentication Errors

```
aws: failed to verify credentials
```

**Solutions:**
- Verify AWS credentials are correctly set
- Check IAM user/role has required permissions
- Ensure credentials are not expired (for temporary credentials)

### Region Errors

```
aws: operation error: no such region
```

**Solutions:**
- Set a valid AWS region
- Check AWS_REGION environment variable
- Verify regions in AWS_REGIONS are valid

### Permission Denied

```
AccessDenied: User is not authorized to perform
```

**Solutions:**
- Attach the `SecurityAudit` managed policy
- Verify IAM policy allows the required actions
- Check for SCPs that might block access
