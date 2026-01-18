// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudtrail"
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

	"github.com/kopexa-grc/kspec/pkg/ratelimit"
)

// Client wraps AWS SDK v2 configuration with rate limiting support.
// AWS SDK v2 has built-in retry, but we add rate limiting for safety
// when making many API calls during scanning.
type Client struct {
	cfg       aws.Config
	limiter   *ratelimit.Limiter
	accountID string
	userARN   string
	regions   []string
}

// ClientConfig holds configuration for creating an AWS client.
type ClientConfig struct {
	// Config is the AWS SDK configuration.
	Config aws.Config
	// AccountID is the AWS account ID.
	AccountID string
	// UserARN is the ARN of the authenticated user/role.
	UserARN string
	// Regions is the list of regions to scan.
	Regions []string
	// Limiter is an optional rate limiter. If nil, no rate limiting is applied.
	Limiter *ratelimit.Limiter
}

// NewClient creates a new AWS client with the given configuration.
func NewClient(cfg ClientConfig) *Client {
	return &Client{
		cfg:       cfg.Config,
		limiter:   cfg.Limiter,
		accountID: cfg.AccountID,
		userARN:   cfg.UserARN,
		regions:   cfg.Regions,
	}
}

// Do executes a function with rate limiting.
// If no rate limiter is configured, the function is executed immediately.
func (c *Client) Do(ctx context.Context, fn func() error) error {
	if c.limiter != nil {
		if err := c.limiter.Wait(ctx); err != nil {
			return err
		}
	}
	return fn()
}

// Config returns the base AWS config.
func (c *Client) Config() aws.Config {
	return c.cfg
}

// ConfigForRegion returns an AWS config for a specific region.
func (c *Client) ConfigForRegion(region string) aws.Config {
	cfg := c.cfg.Copy()
	cfg.Region = region
	return cfg
}

// AccountID returns the AWS account ID.
func (c *Client) AccountID() string {
	return c.accountID
}

// UserARN returns the ARN of the authenticated user/role.
func (c *Client) UserARN() string {
	return c.userARN
}

// Regions returns the list of regions to scan.
func (c *Client) Regions() []string {
	return c.regions
}

// Service client getters - these create new clients for specific services.

// IAM returns an IAM client.
func (c *Client) IAM() *iam.Client {
	return iam.NewFromConfig(c.cfg)
}

// STS returns an STS client.
func (c *Client) STS() *sts.Client {
	return sts.NewFromConfig(c.cfg)
}

// Organizations returns an Organizations client.
func (c *Client) Organizations() *organizations.Client {
	return organizations.NewFromConfig(c.cfg)
}

// S3 returns an S3 client.
func (c *Client) S3() *s3.Client {
	return s3.NewFromConfig(c.cfg)
}

// EC2 returns an EC2 client for the specified region.
func (c *Client) EC2(region string) *ec2.Client {
	return ec2.NewFromConfig(c.ConfigForRegion(region))
}

// RDS returns an RDS client for the specified region.
func (c *Client) RDS(region string) *rds.Client {
	return rds.NewFromConfig(c.ConfigForRegion(region))
}

// Lambda returns a Lambda client for the specified region.
func (c *Client) Lambda(region string) *lambda.Client {
	return lambda.NewFromConfig(c.ConfigForRegion(region))
}

// KMS returns a KMS client for the specified region.
func (c *Client) KMS(region string) *kms.Client {
	return kms.NewFromConfig(c.ConfigForRegion(region))
}

// CloudTrail returns a CloudTrail client for the specified region.
func (c *Client) CloudTrail(region string) *cloudtrail.Client {
	return cloudtrail.NewFromConfig(c.ConfigForRegion(region))
}

// EKS returns an EKS client for the specified region.
func (c *Client) EKS(region string) *eks.Client {
	return eks.NewFromConfig(c.ConfigForRegion(region))
}

// ECS returns an ECS client for the specified region.
func (c *Client) ECS(region string) *ecs.Client {
	return ecs.NewFromConfig(c.ConfigForRegion(region))
}

// DynamoDB returns a DynamoDB client for the specified region.
func (c *Client) DynamoDB(region string) *dynamodb.Client {
	return dynamodb.NewFromConfig(c.ConfigForRegion(region))
}

// ELB returns an Elastic Load Balancing v2 client for the specified region.
func (c *Client) ELB(region string) *elasticloadbalancingv2.Client {
	return elasticloadbalancingv2.NewFromConfig(c.ConfigForRegion(region))
}

// CloudWatchLogs returns a CloudWatch Logs client for the specified region.
func (c *Client) CloudWatchLogs(region string) *cloudwatchlogs.Client {
	return cloudwatchlogs.NewFromConfig(c.ConfigForRegion(region))
}

// ConfigService returns a Config service client for the specified region.
func (c *Client) ConfigService(region string) *configservice.Client {
	return configservice.NewFromConfig(c.ConfigForRegion(region))
}

// GuardDuty returns a GuardDuty client for the specified region.
func (c *Client) GuardDuty(region string) *guardduty.Client {
	return guardduty.NewFromConfig(c.ConfigForRegion(region))
}

// SecurityHub returns a Security Hub client for the specified region.
func (c *Client) SecurityHub(region string) *securityhub.Client {
	return securityhub.NewFromConfig(c.ConfigForRegion(region))
}

// SecretsManager returns a Secrets Manager client for the specified region.
func (c *Client) SecretsManager(region string) *secretsmanager.Client {
	return secretsmanager.NewFromConfig(c.ConfigForRegion(region))
}

// SNS returns an SNS client for the specified region.
func (c *Client) SNS(region string) *sns.Client {
	return sns.NewFromConfig(c.ConfigForRegion(region))
}

// SQS returns an SQS client for the specified region.
func (c *Client) SQS(region string) *sqs.Client {
	return sqs.NewFromConfig(c.ConfigForRegion(region))
}

// ECR returns an ECR client for the specified region.
func (c *Client) ECR(region string) *ecr.Client {
	return ecr.NewFromConfig(c.ConfigForRegion(region))
}

// WAFv2 returns a WAFv2 client for the specified region.
func (c *Client) WAFv2(region string) *wafv2.Client {
	return wafv2.NewFromConfig(c.ConfigForRegion(region))
}

// ACM returns an ACM client for the specified region.
func (c *Client) ACM(region string) *acm.Client {
	return acm.NewFromConfig(c.ConfigForRegion(region))
}

// ElastiCache returns an ElastiCache client for the specified region.
func (c *Client) ElastiCache(region string) *elasticache.Client {
	return elasticache.NewFromConfig(c.ConfigForRegion(region))
}

// CloudFront returns a CloudFront client.
// CloudFront is a global service, so no region is needed.
func (c *Client) CloudFront() *cloudfront.Client {
	return cloudfront.NewFromConfig(c.ConfigForRegion("us-east-1"))
}

// APIGateway returns an API Gateway client for the specified region.
func (c *Client) APIGateway(region string) *apigateway.Client {
	return apigateway.NewFromConfig(c.ConfigForRegion(region))
}

// APIGatewayV2 returns an API Gateway V2 client for the specified region.
func (c *Client) APIGatewayV2(region string) *apigatewayv2.Client {
	return apigatewayv2.NewFromConfig(c.ConfigForRegion(region))
}

// SSM returns an SSM client for the specified region.
func (c *Client) SSM(region string) *ssm.Client {
	return ssm.NewFromConfig(c.ConfigForRegion(region))
}

// AutoScaling returns an Auto Scaling client for the specified region.
func (c *Client) AutoScaling(region string) *autoscaling.Client {
	return autoscaling.NewFromConfig(c.ConfigForRegion(region))
}
