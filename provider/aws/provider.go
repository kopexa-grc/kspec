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
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/configservice"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/guardduty"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/kms"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/organizations"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/wafv2"

	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/provider/aws/resources"
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
	// Global clients (IAM, S3, Organizations, STS are global services)
	iamClient := iam.NewFromConfig(c.cfg)
	s3Client := s3.NewFromConfig(c.cfg)
	orgClient := organizations.NewFromConfig(c.cfg)
	stsClient := sts.NewFromConfig(c.cfg)

	// Regional client factory functions
	ec2Factory := func(region string) resources.EC2Client {
		return ec2.NewFromConfig(c.ConfigForRegion(region))
	}
	kmsFactory := func(region string) resources.KMSClient {
		return kms.NewFromConfig(c.ConfigForRegion(region))
	}
	cloudtrailFactory := func(region string) resources.CloudTrailClient {
		return cloudtrail.NewFromConfig(c.ConfigForRegion(region))
	}
	rdsFactory := func(region string) resources.RDSClient {
		return rds.NewFromConfig(c.ConfigForRegion(region))
	}
	lambdaFactory := func(region string) resources.LambdaClient {
		return lambda.NewFromConfig(c.ConfigForRegion(region))
	}
	eksFactory := func(region string) resources.EKSClient {
		return eks.NewFromConfig(c.ConfigForRegion(region))
	}
	ecsFactory := func(region string) resources.ECSClient {
		return ecs.NewFromConfig(c.ConfigForRegion(region))
	}
	dynamodbFactory := func(region string) resources.DynamoDBClient {
		return dynamodb.NewFromConfig(c.ConfigForRegion(region))
	}
	elbFactory := func(region string) resources.ELBClient {
		return elasticloadbalancingv2.NewFromConfig(c.ConfigForRegion(region))
	}
	cloudwatchLogsFactory := func(region string) resources.CloudWatchLogsClient {
		return cloudwatchlogs.NewFromConfig(c.ConfigForRegion(region))
	}
	cloudwatchFactory := func(region string) resources.CloudWatchClient {
		return cloudwatch.NewFromConfig(c.ConfigForRegion(region))
	}
	configFactory := func(region string) resources.ConfigClient {
		return configservice.NewFromConfig(c.ConfigForRegion(region))
	}
	guarddutyFactory := func(region string) resources.GuardDutyClient {
		return guardduty.NewFromConfig(c.ConfigForRegion(region))
	}
	securityhubFactory := func(region string) resources.SecurityHubClient {
		return securityhub.NewFromConfig(c.ConfigForRegion(region))
	}
	secretsmanagerFactory := func(region string) resources.SecretsManagerClient {
		return secretsmanager.NewFromConfig(c.ConfigForRegion(region))
	}
	snsFactory := func(region string) resources.SNSClient {
		return sns.NewFromConfig(c.ConfigForRegion(region))
	}
	sqsFactory := func(region string) resources.SQSClient {
		return sqs.NewFromConfig(c.ConfigForRegion(region))
	}
	ecrFactory := func(region string) resources.ECRClient {
		return ecr.NewFromConfig(c.ConfigForRegion(region))
	}
	wafv2Factory := func(region string) resources.WAFv2Client {
		return wafv2.NewFromConfig(c.ConfigForRegion(region))
	}
	acmFactory := func(region string) resources.ACMClient {
		return acm.NewFromConfig(c.ConfigForRegion(region))
	}
	elasticacheFactory := func(region string) resources.ElastiCacheClient {
		return elasticache.NewFromConfig(c.ConfigForRegion(region))
	}
	apigatewayFactory := func(region string) resources.APIGatewayClient {
		return apigateway.NewFromConfig(c.ConfigForRegion(region))
	}
	apigatewayv2Factory := func(region string) resources.APIGatewayV2Client {
		return apigatewayv2.NewFromConfig(c.ConfigForRegion(region))
	}
	ssmFactory := func(region string) resources.SSMClient {
		return ssm.NewFromConfig(c.ConfigForRegion(region))
	}
	autoscalingFactory := func(region string) resources.AutoScalingClient {
		return autoscaling.NewFromConfig(c.ConfigForRegion(region))
	}

	return []core.ResourceSpec{
		// Account & Organization (global services)
		resources.NewAccount(stsClient, iamClient, orgClient),
		resources.NewOrganization(orgClient),

		// IAM (global service)
		resources.NewIAMUser(iamClient, c.accountID),
		resources.NewIAMRole(iamClient, c.accountID),
		resources.NewIAMGroup(iamClient),
		resources.NewIAMPolicy(iamClient),

		// S3 (global service)
		resources.NewS3Bucket(s3Client),

		// EC2 (regional)
		resources.NewEC2Instance(ec2Factory, c.accountID, c.regions),
		resources.NewEC2Volume(ec2Factory, c.accountID, c.regions),
		resources.NewEC2Snapshot(ec2Factory, c.accountID, c.regions),
		resources.NewEC2SecurityGroup(ec2Factory, c.accountID, c.regions),
		resources.NewEC2KeyPair(ec2Factory, c.accountID, c.regions),

		// VPC (regional, uses EC2 client)
		resources.NewVPC(ec2Factory, c.accountID, c.regions),
		resources.NewVPCSubnet(ec2Factory, c.accountID, c.regions),
		resources.NewVPCEndpoint(ec2Factory, c.accountID, c.regions),
		resources.NewVPCFlowLog(ec2Factory, c.accountID, c.regions),

		// KMS (regional)
		resources.NewKMSKey(kmsFactory, c.accountID, c.regions),

		// CloudTrail (regional)
		resources.NewCloudTrail(cloudtrailFactory, c.accountID, c.regions),

		// RDS (regional)
		resources.NewRDSInstance(rdsFactory, c.regions),
		resources.NewRDSCluster(rdsFactory, c.regions),

		// Lambda (regional)
		resources.NewLambdaFunction(lambdaFactory, c.regions),

		// EKS (regional)
		resources.NewEKSCluster(eksFactory, c.accountID, c.regions),
		resources.NewEKSNodegroup(eksFactory, c.accountID, c.regions),

		// ECS (regional)
		resources.NewECSCluster(ecsFactory, c.accountID, c.regions),
		resources.NewECSService(ecsFactory, c.accountID, c.regions),

		// DynamoDB (regional)
		resources.NewDynamoDBTable(dynamodbFactory, c.accountID, c.regions),

		// ELB (regional)
		resources.NewELBLoadBalancer(elbFactory, c.accountID, c.regions),
		resources.NewELBTargetGroup(elbFactory, c.accountID, c.regions),
		resources.NewELBListener(elbFactory, c.accountID, c.regions),

		// CloudWatch (regional)
		resources.NewCloudWatchLogGroup(cloudwatchLogsFactory, c.accountID, c.regions),
		resources.NewCloudWatchAlarm(cloudwatchFactory, c.accountID, c.regions),
		resources.NewCloudWatchMetricStream(cloudwatchFactory, c.accountID, c.regions),

		// Config (regional)
		resources.NewConfigRecorder(configFactory, c.accountID, c.regions),
		resources.NewConfigRule(configFactory, c.accountID, c.regions),
		resources.NewConfigDeliveryChannel(configFactory, c.accountID, c.regions),
		resources.NewConfigConformancePack(configFactory, c.accountID, c.regions),

		// GuardDuty (regional)
		resources.NewGuardDutyDetector(guarddutyFactory, c.accountID, c.regions),
		resources.NewGuardDutyFinding(guarddutyFactory, c.accountID, c.regions),

		// Security Hub (regional)
		resources.NewSecurityHub(securityhubFactory, c.accountID, c.regions),
		resources.NewSecurityHubFinding(securityhubFactory, c.accountID, c.regions),
		resources.NewSecurityHubStandard(securityhubFactory, c.accountID, c.regions),

		// Secrets Manager (regional)
		resources.NewSecretsManagerSecret(secretsmanagerFactory, c.accountID, c.regions),

		// SNS (regional)
		resources.NewSNSTopic(snsFactory, c.accountID, c.regions),
		resources.NewSNSSubscription(snsFactory, c.accountID, c.regions),

		// SQS (regional)
		resources.NewSQSQueue(sqsFactory, c.accountID, c.regions),

		// ECR (regional)
		resources.NewECRRepository(ecrFactory, c.accountID, c.regions),
		resources.NewECRImage(ecrFactory, c.accountID, c.regions),

		// WAF (regional)
		resources.NewWAFWebACL(wafv2Factory, c.accountID, c.regions),
		resources.NewWAFIPSet(wafv2Factory, c.accountID, c.regions),
		resources.NewWAFRuleGroup(wafv2Factory, c.accountID, c.regions),

		// ACM (regional)
		resources.NewACMCertificate(acmFactory, c.accountID, c.regions),

		// ElastiCache (regional)
		resources.NewElastiCacheCluster(elasticacheFactory, c.accountID, c.regions),
		resources.NewElastiCacheReplicationGroup(elasticacheFactory, c.accountID, c.regions),

		// CloudFront (global service, always us-east-1)
		resources.NewCloudFrontDistribution(cloudfront.NewFromConfig(c.ConfigForRegion("us-east-1")), c.accountID),

		// API Gateway (regional)
		resources.NewAPIGatewayRestAPI(apigatewayFactory, c.accountID, c.regions),
		resources.NewAPIGatewayStage(apigatewayFactory, c.accountID, c.regions),
		resources.NewAPIGatewayV2API(apigatewayv2Factory, c.accountID, c.regions),

		// SSM (regional)
		resources.NewSSMInstance(ssmFactory, c.accountID, c.regions),
		resources.NewSSMParameter(ssmFactory, c.accountID, c.regions),
		resources.NewSSMDocument(ssmFactory, c.accountID, c.regions),
		resources.NewSSMPatchBaseline(ssmFactory, c.accountID, c.regions),

		// Auto Scaling (regional)
		resources.NewAutoScalingGroup(autoscalingFactory, c.accountID, c.regions),
		resources.NewAutoScalingLaunchConfiguration(autoscalingFactory, c.accountID, c.regions),
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
