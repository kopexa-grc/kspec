// Copyright (c) Kopexa GmbH
// SPDX-License-Identifier: Elastic-2.0

package aws

import (
	"github.com/kopexa-grc/kspec/core"
	"github.com/kopexa-grc/kspec/pkg/ratelimit"
	"github.com/kopexa-grc/kspec/provider"
)

func init() {
	provider.Register(&provider.ProviderDefinition{
		Name:        "aws",
		Aliases:     []string{"amazon", "ec2", "s3"},
		Description: "Scan AWS accounts for security compliance (IAM, S3, EC2, RDS, Lambda, EKS, and 50+ services)",
		Flags: []provider.FlagDefinition{
			{
				Name:        "profile",
				Description: "AWS profile name from ~/.aws/credentials",
				EnvVar:      "AWS_PROFILE",
			},
			{
				Name:        "region",
				Description: "AWS region",
				Default:     "us-east-1",
				EnvVar:      "AWS_REGION",
			},
			{
				Name:        "regions",
				Description: "Multiple AWS regions for multi-region scanning (comma-separated)",
				Type:        provider.FlagTypeStringSlice,
			},
			{
				Name:        "access-key-id",
				Description: "AWS access key ID",
				EnvVar:      "AWS_ACCESS_KEY_ID",
			},
			{
				Name:        "secret-access-key",
				Description: "AWS secret access key",
				EnvVar:      "AWS_SECRET_ACCESS_KEY",
			},
			{
				Name:        "session-token",
				Description: "AWS session token",
				EnvVar:      "AWS_SESSION_TOKEN",
			},
			{
				Name:        "role-arn",
				Description: "IAM role ARN to assume for cross-account access",
				EnvVar:      "AWS_ROLE_ARN",
			},
			{
				Name:        "external-id",
				Description: "External ID for assume role",
				EnvVar:      "AWS_EXTERNAL_ID",
			},
			{
				Name:        "endpoint-url",
				Description: "Custom endpoint URL (for LocalStack, etc.)",
				EnvVar:      "AWS_ENDPOINT_URL",
			},
		},
		AssetTypes: []provider.AssetDefinition{
			{
				Name:        "account",
				Description: "Scan entire AWS account",
				DefaultName: "default",
				ScannerKey:  "aws-account",
			},
		},
		ConfigMapping: map[string]string{
			"access-key-id":     "access_key",
			"secret-access-key": "secret_key",
			"session-token":     "session_token",
			"role-arn":          "role_arn",
			"external-id":       "external_id",
			"endpoint-url":      "endpoint_url",
		},
		// AWS SDK v2 has built-in retry, but we add rate limiting for safety
		// when making many API calls during scanning. Conservative limit for IAM.
		RateLimitConfig: &ratelimit.Config{
			RequestsPerSecond: 10,
			Burst:             20,
		},
		Factory: func() core.Provider {
			return NewProvider()
		},
	})
}
