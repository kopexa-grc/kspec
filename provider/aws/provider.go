// Package aws provides a kspec provider for AWS cloud resources.
package aws

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/service/sts"

	"github.com/kopexa-grc/kspec/core"
)

// DefaultRegion is the default AWS region used if none is specified.
const DefaultRegion = "us-east-1"

// Provider implements the core.Provider interface for AWS.
type Provider struct{}

// NewProvider creates a new AWS provider.
func NewProvider() *Provider {
	return &Provider{}
}

// Name returns the provider name.
func (p *Provider) Name() string {
	return "aws"
}

// Connect establishes a connection to AWS using the SDK v2 credential chain.
// Supported config options:
//   - profile: AWS profile name from ~/.aws/credentials
//   - region: AWS region (default: us-east-1)
//   - regions: Comma-separated list of regions for multi-region scanning
//   - access_key: AWS access key ID (for static credentials)
//   - secret_key: AWS secret access key (for static credentials)
//   - session_token: AWS session token (for temporary credentials)
//   - role_arn: IAM role ARN to assume (for cross-account access)
//   - external_id: External ID for assume role
func (p *Provider) Connect(ctx context.Context, cfg map[string]string) (core.Connection, error) {
	// Build AWS config options
	opts := []func(*config.LoadOptions) error{}

	// Set region
	region := getRegion(cfg)
	opts = append(opts, config.WithRegion(region))

	// Profile support
	if profile := cfg["profile"]; profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	} else if profile := os.Getenv("AWS_PROFILE"); profile != "" {
		opts = append(opts, config.WithSharedConfigProfile(profile))
	}

	// Static credentials (explicit key/secret override credential chain)
	accessKey := cfg["access_key"]
	if accessKey == "" {
		accessKey = os.Getenv("AWS_ACCESS_KEY_ID")
	}
	secretKey := cfg["secret_key"]
	if secretKey == "" {
		secretKey = os.Getenv("AWS_SECRET_ACCESS_KEY")
	}
	sessionToken := cfg["session_token"]
	if sessionToken == "" {
		sessionToken = os.Getenv("AWS_SESSION_TOKEN")
	}

	if accessKey != "" && secretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, sessionToken),
		))
	}

	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("aws: failed to load config: %w", err)
	}

	// STS Assume Role support for cross-account access
	roleArn := cfg["role_arn"]
	if roleArn == "" {
		roleArn = os.Getenv("AWS_ROLE_ARN")
	}
	if roleArn != "" {
		stsClient := sts.NewFromConfig(awsCfg)
		assumeRoleOpts := func(o *stscreds.AssumeRoleOptions) {
			if externalID := cfg["external_id"]; externalID != "" {
				o.ExternalID = aws.String(externalID)
			} else if externalID := os.Getenv("AWS_EXTERNAL_ID"); externalID != "" {
				o.ExternalID = aws.String(externalID)
			}
		}
		creds := stscreds.NewAssumeRoleProvider(stsClient, roleArn, assumeRoleOpts)
		awsCfg.Credentials = aws.NewCredentialsCache(creds)
	}

	// Verify credentials by calling GetCallerIdentity
	stsClient := sts.NewFromConfig(awsCfg)
	identity, err := stsClient.GetCallerIdentity(ctx, &sts.GetCallerIdentityInput{})
	if err != nil {
		return nil, fmt.Errorf("aws: failed to verify credentials: %w", err)
	}

	// Parse regions for multi-region scanning
	regions := getRegions(cfg, region)

	return &Connection{
		cfg:       awsCfg,
		accountID: aws.ToString(identity.Account),
		userARN:   aws.ToString(identity.Arn),
		regions:   regions,
	}, nil
}

// Connection represents an active connection to AWS.
type Connection struct {
	cfg       aws.Config
	accountID string
	userARN   string
	regions   []string
}

// AccountID returns the AWS account ID.
func (c *Connection) AccountID() string {
	return c.accountID
}

// Regions returns the list of regions to scan.
func (c *Connection) Regions() []string {
	return c.regions
}

// Config returns the base AWS config.
func (c *Connection) Config() aws.Config {
	return c.cfg
}

// ConfigForRegion returns an AWS config for a specific region.
func (c *Connection) ConfigForRegion(region string) aws.Config {
	cfg := c.cfg.Copy()
	cfg.Region = region
	return cfg
}

// Resources returns all available AWS resources.
func (c *Connection) Resources() []core.ResourceSpec {
	return []core.ResourceSpec{
		// Account & Organization
		&AccountResource{conn: c},
		&OrganizationResource{conn: c},
		// IAM
		&IAMUserResource{conn: c},
		&IAMRoleResource{conn: c},
		&IAMGroupResource{conn: c},
		&IAMPolicyResource{conn: c},
		// S3
		&S3BucketResource{conn: c},
		// EC2
		&EC2InstanceResource{conn: c},
		&EC2VolumeResource{conn: c},
		&EC2SnapshotResource{conn: c},
		// EC2 Security
		&EC2SecurityGroupResource{conn: c},
		&EC2KeyPairResource{conn: c},
		// VPC
		&VPCResource{conn: c},
		&VPCSubnetResource{conn: c},
		&VPCEndpointResource{conn: c},
		&VPCFlowLogResource{conn: c},
		// KMS
		&KMSKeyResource{conn: c},
		// CloudTrail
		&CloudTrailResource{conn: c},
		// RDS
		&RDSInstanceResource{conn: c},
		&RDSClusterResource{conn: c},
		// Lambda
		&LambdaFunctionResource{conn: c},
		// EKS
		&EKSClusterResource{conn: c},
		&EKSNodegroupResource{conn: c},
		// ECS
		&ECSClusterResource{conn: c},
		&ECSServiceResource{conn: c},
		// DynamoDB
		&DynamoDBTableResource{conn: c},
		// ELB
		&ELBLoadBalancerResource{conn: c},
		&ELBTargetGroupResource{conn: c},
		&ELBListenerResource{conn: c},
		// CloudWatch
		&CloudWatchLogGroupResource{conn: c},
		&CloudWatchAlarmResource{conn: c},
		&CloudWatchMetricStreamResource{conn: c},
		// Config
		&ConfigRecorderResource{conn: c},
		&ConfigRuleResource{conn: c},
		&ConfigDeliveryChannelResource{conn: c},
		&ConfigConformancePackResource{conn: c},
		// GuardDuty
		&GuardDutyDetectorResource{conn: c},
		&GuardDutyFindingResource{conn: c},
		// Security Hub
		&SecurityHubResource{conn: c},
		&SecurityHubFindingResource{conn: c},
		&SecurityHubStandardResource{conn: c},
		// Secrets Manager
		&SecretsManagerSecretResource{conn: c},
		// SNS
		&SNSTopicResource{conn: c},
		&SNSSubscriptionResource{conn: c},
		// SQS
		&SQSQueueResource{conn: c},
		// ECR
		&ECRRepositoryResource{conn: c},
		&ECRImageResource{conn: c},
		// WAF
		&WAFWebACLResource{conn: c},
		&WAFIPSetResource{conn: c},
		&WAFRuleGroupResource{conn: c},
		// ACM
		&ACMCertificateResource{conn: c},
		// ElastiCache
		&ElastiCacheClusterResource{conn: c},
		&ElastiCacheReplicationGroupResource{conn: c},
		// CloudFront
		&CloudFrontDistributionResource{conn: c},
		// API Gateway
		&APIGatewayRestAPIResource{conn: c},
		&APIGatewayStageResource{conn: c},
		&APIGatewayV2APIResource{conn: c},
		// SSM
		&SSMInstanceResource{conn: c},
		&SSMParameterResource{conn: c},
		&SSMDocumentResource{conn: c},
		&SSMPatchBaselineResource{conn: c},
		// Auto Scaling
		&AutoScalingGroupResource{conn: c},
		&AutoScalingLaunchConfigurationResource{conn: c},
	}
}

// getRegion returns the AWS region from config or environment.
func getRegion(cfg map[string]string) string {
	if region := cfg["region"]; region != "" {
		return region
	}
	if region := os.Getenv("AWS_REGION"); region != "" {
		return region
	}
	if region := os.Getenv("AWS_DEFAULT_REGION"); region != "" {
		return region
	}
	return DefaultRegion
}

// getRegions returns the list of regions for multi-region scanning.
func getRegions(cfg map[string]string, defaultRegion string) []string {
	if regions := cfg["regions"]; regions != "" {
		return strings.Split(regions, ",")
	}
	if regions := os.Getenv("AWS_REGIONS"); regions != "" {
		return strings.Split(regions, ",")
	}
	return []string{defaultRegion}
}
